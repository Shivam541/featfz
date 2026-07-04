package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/shivam/featfz/feat-manager/internal/dao"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

func TestCreateFlagRouteIntegration(t *testing.T) {
	db := openRouterIntegrationDB(t)
	ctx := context.Background()
	resetRouterIntegrationTables(t, db)

	tenantID := insertRouterIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	now := time.Unix(1_720_000_000, 0).UTC()

	router := NewRouter(RouterDependencies{
		HealthChecker: service.StaticHealthChecker{},
		Authenticator: service.AuthenticationService{
			TenantApps:    dao.NewTenantAppRepository(db),
			TokenVerifier: service.HS256JWTVerifier{},
			Now:           func() time.Time { return now },
		},
		FlagCreator: service.NewFlagCreationService(dao.NewFlagRepository(db)),
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{
		"key": "new_dashboard",
		"description": "Enable the new dashboard experience",
		"default_enabled": false
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-ID", "app-acme")
	req.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Key        string `json:"key"`
			TenantID   int64  `json:"tenant_id"`
			ID         int64  `json:"id"`
			ArchivedAt any    `json:"archived_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json body, got %v", err)
	}
	if !body.Success {
		t.Fatal("expected success=true")
	}
	if body.Data.Key != "new_dashboard" || body.Data.TenantID != tenantID || body.Data.ID == 0 {
		t.Fatalf("unexpected created flag response: %+v", body.Data)
	}

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM flags WHERE tenant_id = ? AND `+"`key`"+` = ?`, tenantID, "new_dashboard").Scan(&count); err != nil {
		t.Fatalf("count created flags: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 created flag row, got %d", count)
	}
}

func openRouterIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("TEST_DB_DSN"))
	if dsn == "" {
		t.Skip("set TEST_DB_DSN to run router integration tests")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open integration db: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Fatalf("ping integration db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func resetRouterIntegrationTables(t *testing.T, db *sql.DB) {
	t.Helper()

	for _, query := range []string{
		"DELETE FROM flag_user_overrides",
		"DELETE FROM flags",
		"DELETE FROM tenants",
	} {
		if _, err := db.ExecContext(context.Background(), query); err != nil {
			t.Fatalf("reset tables with %q: %v", query, err)
		}
	}
}

func insertRouterIntegrationTenant(t *testing.T, db *sql.DB, name, appID, secret string) int64 {
	t.Helper()

	result, err := db.ExecContext(context.Background(), `
INSERT INTO tenants (name, app_id, jwt_secret)
VALUES (?, ?, ?)
`, name, appID, secret)
	if err != nil {
		t.Fatalf("insert tenant: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("tenant last insert id: %v", err)
	}

	return id
}

func testRouterJWT(t *testing.T, secret string, claims map[string]any) string {
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
