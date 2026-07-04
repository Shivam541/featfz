package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

func TestRequestContextUsesIncomingRequestID(t *testing.T) {
	handler := RequestContext()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := requestctx.RequestID(r.Context()); got != "req-123" {
			t.Fatalf("expected request id req-123, got %q", got)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(RequestIDHeader, "req-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != "req-123" {
		t.Fatalf("expected response request id req-123, got %q", got)
	}
}

func TestRequestContextGeneratesRequestID(t *testing.T) {
	handler := RequestContext()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := requestctx.RequestID(r.Context()); got == "" {
			t.Fatal("expected generated request id")
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got == "" {
		t.Fatal("expected request id response header")
	}
}

func TestRecoverWritesSafeInternalError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), RequestContext(), Recover(logger))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json body, got %v", err)
	}

	if body.Success {
		t.Fatal("expected success=false")
	}

	if body.Error.Code != "internal_error" {
		t.Fatalf("expected internal_error, got %q", body.Error.Code)
	}
}

func TestRequestLoggingPassesThrough(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := requestctx.RequestID(r.Context()); got != "req-456" {
			t.Fatalf("expected request id req-456, got %q", got)
		}

		w.WriteHeader(http.StatusAccepted)
	}), RequestContext(), RequestLogging(logger))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(RequestIDHeader, "req-456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req.WithContext(context.Background()))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestRequireAuth(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	authenticator := service.AuthenticationService{
		TenantApps: middlewareTenantAppRepository{
			records: map[string]domain.TenantApp{
				"app-acme": {
					TenantID:  15,
					AppID:     "app-acme",
					JWTSecret: "phase2-secret",
				},
			},
		},
		TokenVerifier: service.HS256JWTVerifier{},
		Now:           func() time.Time { return now },
	}

	tests := []struct {
		name           string
		headers        map[string]string
		wantStatus     int
		wantCode       string
		wantTenantID   int64
		wantHandlerRun bool
	}{
		{
			name:       "rejects missing authorization header",
			headers:    map[string]string{"X-App-ID": "app-acme"},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "missing_authorization_header",
		},
		{
			name:       "rejects missing app id header",
			headers:    map[string]string{"Authorization": "Bearer abc"},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "missing_app_id",
		},
		{
			name:       "rejects invalid token",
			headers:    map[string]string{"X-App-ID": "app-acme", "Authorization": "Bearer invalid.token.value"},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "invalid_token",
		},
		{
			name: "rejects expired token",
			headers: map[string]string{
				"X-App-ID":      "app-acme",
				"Authorization": "Bearer " + middlewareJWT(t, "phase2-secret", map[string]any{"app_id": "app-acme", "exp": now.Add(-time.Minute).Unix()}),
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "token_expired",
		},
		{
			name: "rejects app id mismatch before handler runs",
			headers: map[string]string{
				"X-App-ID":      "app-acme",
				"Authorization": "Bearer " + middlewareJWT(t, "phase2-secret", map[string]any{"app_id": "app-globex", "exp": now.Add(time.Hour).Unix()}),
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "app_id_mismatch",
		},
		{
			name: "allows valid tenant request",
			headers: map[string]string{
				"X-App-ID":      "app-acme",
				"Authorization": "Bearer " + middlewareJWT(t, "phase2-secret", map[string]any{"app_id": "app-acme", "sub": "user-123", "exp": now.Add(time.Hour).Unix()}),
			},
			wantStatus:     http.StatusNoContent,
			wantTenantID:   15,
			wantHandlerRun: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerRan := false
			handler := RequireAuth(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerRan = true

				tenant, ok := requestctx.TenantFrom(r.Context())
				if !ok {
					t.Fatal("expected tenant in request context")
				}
				if tenant.TenantID != tt.wantTenantID {
					t.Fatalf("expected tenant id %d, got %d", tt.wantTenantID, tenant.TenantID)
				}

				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, "/v1/auth/check", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if handlerRan != tt.wantHandlerRun {
				t.Fatalf("expected handler run %t, got %t", tt.wantHandlerRun, handlerRan)
			}

			if tt.wantCode == "" {
				return
			}

			var body struct {
				Success bool `json:"success"`
				Error   struct {
					Code string `json:"code"`
				} `json:"error"`
			}

			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("expected valid json body, got %v", err)
			}

			if body.Error.Code != tt.wantCode {
				t.Fatalf("expected error code %q, got %q", tt.wantCode, body.Error.Code)
			}
		})
	}
}

type middlewareTenantAppRepository struct {
	records map[string]domain.TenantApp
}

func (m middlewareTenantAppRepository) FindByAppID(_ context.Context, appID string) (domain.TenantApp, error) {
	record, ok := m.records[appID]
	if !ok {
		return domain.TenantApp{}, service.ErrTenantAppNotFound
	}

	return record, nil
}

func middlewareJWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]any{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerPart + "." + payloadPart

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return strings.Join([]string{headerPart, payloadPart, signaturePart}, ".")
}
