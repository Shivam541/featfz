package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shivam/featfz/feat-manager/internal/domain"
)

func TestHS256JWTVerifierVerify(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	secret := "phase2-secret"

	tests := []struct {
		name       string
		token      string
		wantClaims JWTClaims
		wantErr    error
	}{
		{
			name: "valid token",
			token: testJWT(t, secret, map[string]any{
				"app_id": "app-acme",
				"sub":    "user-123",
				"iat":    now.Add(-time.Minute).Unix(),
				"exp":    now.Add(time.Hour).Unix(),
			}),
			wantClaims: JWTClaims{
				AppID:   "app-acme",
				Subject: "user-123",
			},
		},
		{
			name:    "rejects malformed token",
			token:   "abc.def",
			wantErr: ErrInvalidToken,
		},
		{
			name: "rejects invalid signature",
			token: testJWT(t, "other-secret", map[string]any{
				"app_id": "app-acme",
				"exp":    now.Add(time.Hour).Unix(),
			}),
			wantErr: ErrInvalidToken,
		},
		{
			name: "rejects expired token",
			token: testJWT(t, secret, map[string]any{
				"app_id": "app-acme",
				"exp":    now.Add(-time.Minute).Unix(),
			}),
			wantErr: ErrTokenExpired,
		},
		{
			name: "rejects token without app id claim",
			token: testJWT(t, secret, map[string]any{
				"sub": "user-123",
				"exp": now.Add(time.Hour).Unix(),
			}),
			wantErr: ErrInvalidToken,
		},
	}

	verifier := HS256JWTVerifier{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := verifier.Verify(tt.token, secret, now)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if claims != tt.wantClaims {
				t.Fatalf("expected claims %+v, got %+v", tt.wantClaims, claims)
			}
		})
	}
}

func TestAuthenticationServiceAuthenticate(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	repo := stubTenantAppRepository{
		records: map[string]domain.TenantApp{
			"app-acme": {
				TenantID:  11,
				AppID:     "app-acme",
				JWTSecret: "phase2-secret",
			},
		},
	}

	authenticator := AuthenticationService{
		TenantApps:    repo,
		TokenVerifier: HS256JWTVerifier{},
		Now:           func() time.Time { return now },
	}

	tests := []struct {
		name       string
		appID      string
		authHeader string
		wantTenant AuthenticatedTenant
		wantCode   string
	}{
		{
			name:  "authenticates valid tenant request",
			appID: "app-acme",
			authHeader: "Bearer " + testJWT(t, "phase2-secret", map[string]any{
				"app_id": "app-acme",
				"sub":    "user-123",
				"exp":    now.Add(time.Hour).Unix(),
			}),
			wantTenant: AuthenticatedTenant{
				TenantID: 11,
				AppID:    "app-acme",
				Subject:  "user-123",
			},
		},
		{
			name:       "rejects missing app id header",
			appID:      "",
			authHeader: "Bearer token",
			wantCode:   "missing_app_id",
		},
		{
			name:  "rejects unknown app id",
			appID: "unknown-app",
			authHeader: "Bearer " + testJWT(t, "phase2-secret", map[string]any{
				"app_id": "unknown-app",
				"exp":    now.Add(time.Hour).Unix(),
			}),
			wantCode: "unknown_app_id",
		},
		{
			name:  "rejects app id mismatch",
			appID: "app-acme",
			authHeader: "Bearer " + testJWT(t, "phase2-secret", map[string]any{
				"app_id": "app-globex",
				"exp":    now.Add(time.Hour).Unix(),
			}),
			wantCode: "app_id_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := authenticator.Authenticate(context.Background(), tt.appID, tt.authHeader)
			if tt.wantCode != "" {
				var authErr *AuthError
				if !errors.As(err, &authErr) {
					t.Fatalf("expected auth error, got %v", err)
				}
				if authErr.Code != tt.wantCode {
					t.Fatalf("expected auth code %q, got %q", tt.wantCode, authErr.Code)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if tenant != tt.wantTenant {
				t.Fatalf("expected tenant %+v, got %+v", tt.wantTenant, tenant)
			}
		})
	}
}

type stubTenantAppRepository struct {
	records map[string]domain.TenantApp
}

func (s stubTenantAppRepository) FindByAppID(_ context.Context, appID string) (domain.TenantApp, error) {
	record, ok := s.records[appID]
	if !ok {
		return domain.TenantApp{}, ErrTenantAppNotFound
	}

	return record, nil
}

func testJWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]any{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := strings.Join([]string{headerPart, payloadPart}, ".")

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return strings.Join([]string{headerPart, payloadPart, signaturePart}, ".")
}
