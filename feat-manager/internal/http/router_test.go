package httpapi

import (
	"bytes"
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
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type stubHealthChecker struct {
	status string
}

func (s stubHealthChecker) Check(context.Context) service.HealthStatus {
	return service.HealthStatus{Status: s.status}
}

func TestNewRouter(t *testing.T) {
	router := NewRouter(RouterDependencies{
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		HealthChecker: stubHealthChecker{status: "ok"},
		Authenticator: service.AuthenticationService{
			TenantApps: routerTenantApps{
				records: map[string]domain.TenantApp{
					"app-acme": {
						TenantID:  21,
						AppID:     "app-acme",
						JWTSecret: "phase2-secret",
					},
				},
			},
			TokenVerifier: service.HS256JWTVerifier{},
			Now:           func() time.Time { return time.Unix(1_720_000_000, 0).UTC() },
		},
		FlagCreator: routerFlagCreator{},
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if got := rec.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("expected request id header")
	}
}

func TestNewRouterProtectedRoute(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	router := NewRouter(RouterDependencies{
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		HealthChecker: stubHealthChecker{status: "ok"},
		Authenticator: service.AuthenticationService{
			TenantApps: routerTenantApps{
				records: map[string]domain.TenantApp{
					"app-acme": {
						TenantID:  21,
						AppID:     "app-acme",
						JWTSecret: "phase2-secret",
					},
				},
			},
			TokenVerifier: service.HS256JWTVerifier{},
			Now:           func() time.Time { return now },
		},
		FlagCreator: routerFlagCreator{},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/check", nil)
	req.Header.Set("X-App-ID", "app-acme")
	req.Header.Set("Authorization", "Bearer "+routerJWT(t, "phase2-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNewRouterCreateFlagRoute(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	router := NewRouter(RouterDependencies{
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		HealthChecker: stubHealthChecker{status: "ok"},
		Authenticator: service.AuthenticationService{
			TenantApps: routerTenantApps{
				records: map[string]domain.TenantApp{
					"app-acme": {
						TenantID:  21,
						AppID:     "app-acme",
						JWTSecret: "phase2-secret",
					},
				},
			},
			TokenVerifier: service.HS256JWTVerifier{},
			Now:           func() time.Time { return now },
		},
		FlagCreator: routerFlagCreator{},
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/flags", bytes.NewBufferString(`{"key":"new_dashboard","default_enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-ID", "app-acme")
	req.Header.Set("Authorization", "Bearer "+routerJWT(t, "phase2-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

type routerTenantApps struct {
	records map[string]domain.TenantApp
}

func (r routerTenantApps) FindByAppID(_ context.Context, appID string) (domain.TenantApp, error) {
	record, ok := r.records[appID]
	if !ok {
		return domain.TenantApp{}, service.ErrTenantAppNotFound
	}

	return record, nil
}

func routerJWT(t *testing.T, secret string, claims map[string]any) string {
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

func middlewareJWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	return routerJWT(t, secret, claims)
}

type routerFlagCreator struct{}

func (routerFlagCreator) Create(context.Context, int64, service.CreateFlagInput) (domain.Flag, error) {
	return domain.Flag{
		ID:             42,
		TenantID:       21,
		Key:            "new_dashboard",
		DefaultEnabled: true,
	}, nil
}
