package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrTenantAppNotFound    = errors.New("tenant app not found")
	ErrFlagNotFound         = errors.New("flag not found")
	ErrFlagOverrideNotFound = errors.New("flag override not found")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
)

type AuthenticatedTenant struct {
	TenantID int64
	AppID    string
	Subject  string
}

type Authenticator interface {
	Authenticate(context.Context, string, string) (AuthenticatedTenant, error)
}

type JWTClaims struct {
	AppID   string
	Subject string
}

type JWTVerifier interface {
	Verify(string, string, time.Time) (JWTClaims, error)
}

type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	if e == nil {
		return ""
	}

	return e.Code
}

type AuthenticationService struct {
	TenantApps    TenantAppRepository
	TokenVerifier JWTVerifier
	Now           func() time.Time
}

func (s AuthenticationService) Authenticate(ctx context.Context, appID string, authHeader string) (AuthenticatedTenant, error) {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return AuthenticatedTenant{}, &AuthError{
			Code:    "missing_app_id",
			Message: "Authentication failed.",
		}
	}

	token, err := bearerToken(authHeader)
	if err != nil {
		return AuthenticatedTenant{}, err
	}

	tenantApp, err := s.TenantApps.FindByAppID(ctx, appID)
	if err != nil {
		if errors.Is(err, ErrTenantAppNotFound) {
			return AuthenticatedTenant{}, &AuthError{
				Code:    "unknown_app_id",
				Message: "Authentication failed.",
			}
		}

		return AuthenticatedTenant{}, fmt.Errorf("find tenant app: %w", err)
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}

	claims, err := s.TokenVerifier.Verify(token, tenantApp.JWTSecret, now)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenExpired):
			return AuthenticatedTenant{}, &AuthError{
				Code:    "token_expired",
				Message: "Authentication failed.",
			}
		case errors.Is(err, ErrInvalidToken):
			return AuthenticatedTenant{}, &AuthError{
				Code:    "invalid_token",
				Message: "Authentication failed.",
			}
		default:
			return AuthenticatedTenant{}, fmt.Errorf("verify token: %w", err)
		}
	}

	if claims.AppID != tenantApp.AppID {
		return AuthenticatedTenant{}, &AuthError{
			Code:    "app_id_mismatch",
			Message: "Authentication failed.",
		}
	}

	return AuthenticatedTenant{
		TenantID: tenantApp.TenantID,
		AppID:    tenantApp.AppID,
		Subject:  claims.Subject,
	}, nil
}

func bearerToken(authHeader string) (string, error) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", &AuthError{
			Code:    "missing_authorization_header",
			Message: "Authentication failed.",
		}
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", &AuthError{
			Code:    "invalid_token",
			Message: "Authentication failed.",
		}
	}

	return strings.TrimSpace(parts[1]), nil
}

type HS256JWTVerifier struct{}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type jwtPayload struct {
	AppID   string `json:"app_id"`
	Subject string `json:"sub"`
	Exp     *int64 `json:"exp"`
	Iat     *int64 `json:"iat"`
}

func (HS256JWTVerifier) Verify(token string, secret string, now time.Time) (JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JWTClaims{}, ErrInvalidToken
	}

	var header jwtHeader
	if err := decodeJWTPart(parts[0], &header); err != nil {
		return JWTClaims{}, ErrInvalidToken
	}

	if header.Algorithm != "HS256" {
		return JWTClaims{}, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := signHS256(signingInput, secret)

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return JWTClaims{}, ErrInvalidToken
	}

	if subtle.ConstantTimeCompare(signature, expectedSignature) != 1 {
		return JWTClaims{}, ErrInvalidToken
	}

	var payload jwtPayload
	if err := decodeJWTPart(parts[1], &payload); err != nil {
		return JWTClaims{}, ErrInvalidToken
	}

	if strings.TrimSpace(payload.AppID) == "" {
		return JWTClaims{}, ErrInvalidToken
	}

	if payload.Exp != nil && now.Unix() >= *payload.Exp {
		return JWTClaims{}, ErrTokenExpired
	}

	return JWTClaims{
		AppID:   payload.AppID,
		Subject: payload.Subject,
	}, nil
}

func decodeJWTPart(part string, target any) error {
	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}

	return json.Unmarshal(decoded, target)
}

func signHS256(input string, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(input))
	return mac.Sum(nil)
}

func NewAuthenticationService(repo TenantAppRepository) AuthenticationService {
	return AuthenticationService{
		TenantApps:    repo,
		TokenVerifier: HS256JWTVerifier{},
		Now:           time.Now,
	}
}
