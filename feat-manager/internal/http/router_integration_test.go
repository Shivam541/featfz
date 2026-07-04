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
	"github.com/shivam/featfz/feat-manager/internal/http/controller"
	"github.com/shivam/featfz/feat-manager/internal/http/validation"
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
		FlagController: controller.NewFlagController(service.NewFlagService(dao.NewFlagRepository(db), dao.NewFlagOverrideRepository(db)), validation.NewValidator()),
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

func TestFlagCrudRoutesIntegration(t *testing.T) {
	db := openRouterIntegrationDB(t)
	resetRouterIntegrationTables(t, db)
	ctx := context.Background()

	tenantID := insertRouterIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	now := time.Unix(1_720_000_000, 0).UTC()

	router := NewRouter(RouterDependencies{
		HealthChecker: service.StaticHealthChecker{},
		Authenticator: service.AuthenticationService{
			TenantApps:    dao.NewTenantAppRepository(db),
			TokenVerifier: service.HS256JWTVerifier{},
			Now:           func() time.Time { return now },
		},
		FlagController: controller.NewFlagController(service.NewFlagService(dao.NewFlagRepository(db), dao.NewFlagOverrideRepository(db)), validation.NewValidator()),
	})

	createReq := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{
		"key": "new_dashboard",
		"description": "Enable the new dashboard experience",
		"default_enabled": false
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-App-ID", "app-acme")
	createReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/flags", nil)
	listReq.Header.Set("X-App-ID", "app-acme")
	listReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d: %s", listRec.Code, listRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/flags/new_dashboard", nil)
	getReq.SetPathValue("flagKey", "new_dashboard")
	getReq.Header.Set("X-App-ID", "app-acme")
	getReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/v1/flags/new_dashboard", strings.NewReader(`{
		"description": "Updated rollout",
		"default_enabled": true
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.SetPathValue("flagKey", "new_dashboard")
	updateReq.Header.Set("X-App-ID", "app-acme")
	updateReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update 200, got %d: %s", updateRec.Code, updateRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/flags/new_dashboard", nil)
	deleteReq.SetPathValue("flagKey", "new_dashboard")
	deleteReq.Header.Set("X-App-ID", "app-acme")
	deleteReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("expected delete 200, got %d: %s", deleteRec.Code, deleteRec.Body.String())
	}

	afterDeleteGetReq := httptest.NewRequest(http.MethodGet, "/v1/flags/new_dashboard", nil)
	afterDeleteGetReq.SetPathValue("flagKey", "new_dashboard")
	afterDeleteGetReq.Header.Set("X-App-ID", "app-acme")
	afterDeleteGetReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	afterDeleteRec := httptest.NewRecorder()
	router.ServeHTTP(afterDeleteRec, afterDeleteGetReq)
	if afterDeleteRec.Code != http.StatusNotFound {
		t.Fatalf("expected archived get 404, got %d: %s", afterDeleteRec.Code, afterDeleteRec.Body.String())
	}

	var remaining int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM flags WHERE tenant_id = ? AND archived_at IS NULL`, tenantID).Scan(&remaining); err != nil {
		t.Fatalf("count remaining flags: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected no active flags remaining, got %d", remaining)
	}
}

func TestBulkSetOverridesRouteIntegration(t *testing.T) {
	db := openRouterIntegrationDB(t)
	resetRouterIntegrationTables(t, db)
	ctx := context.Background()

	tenantID := insertRouterIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	now := time.Unix(1_720_000_000, 0).UTC()

	router := NewRouter(RouterDependencies{
		HealthChecker: service.StaticHealthChecker{},
		Authenticator: service.AuthenticationService{
			TenantApps:    dao.NewTenantAppRepository(db),
			TokenVerifier: service.HS256JWTVerifier{},
			Now:           func() time.Time { return now },
		},
		FlagController: controller.NewFlagController(service.NewFlagService(dao.NewFlagRepository(db), dao.NewFlagOverrideRepository(db)), validation.NewValidator()),
	})

	createReq := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{
		"key": "new_dashboard",
		"description": "Enable the new dashboard experience",
		"default_enabled": false
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-App-ID", "app-acme")
	createReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	bulkReq := httptest.NewRequest(http.MethodPost, "/v1/flags/new_dashboard/users/bulk-set", strings.NewReader(`{
		"overrides": [
			{"user_id":" user_123 ","enabled":true},
			{"user_id":"user_456","enabled":false},
			{"user_id":"user_123","enabled":false}
		]
	}`))
	bulkReq.Header.Set("Content-Type", "application/json")
	bulkReq.SetPathValue("flagKey", "new_dashboard")
	bulkReq.Header.Set("X-App-ID", "app-acme")
	bulkReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	bulkRec := httptest.NewRecorder()
	router.ServeHTTP(bulkRec, bulkReq)
	if bulkRec.Code != http.StatusOK {
		t.Fatalf("expected bulk set 200, got %d: %s", bulkRec.Code, bulkRec.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Applied int `json:"applied"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bulkRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json body, got %v", err)
	}
	if !body.Success || body.Data.Applied != 2 {
		t.Fatalf("unexpected bulk set response: %+v", body)
	}

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM flag_user_overrides WHERE tenant_id = ? AND flag_id = (SELECT id FROM flags WHERE tenant_id = ? AND `+"`key`"+` = ?)`, tenantID, tenantID, "new_dashboard").Scan(&count); err != nil {
		t.Fatalf("count overrides: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 override rows, got %d", count)
	}

	var enabled bool
	if err := db.QueryRowContext(ctx, `SELECT enabled FROM flag_user_overrides WHERE tenant_id = ? AND user_id = ?`, tenantID, "user_123").Scan(&enabled); err != nil {
		t.Fatalf("fetch override: %v", err)
	}
	if enabled {
		t.Fatal("expected last user_123 override to be disabled")
	}
}

func TestEvalRouteIntegration(t *testing.T) {
	db := openRouterIntegrationDB(t)
	resetRouterIntegrationTables(t, db)

	tenantOneID := insertRouterIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	tenantTwoID := insertRouterIntegrationTenant(t, db, "globex", "app-globex", "globex-secret")
	now := time.Unix(1_720_000_000, 0).UTC()

	router := NewRouter(RouterDependencies{
		HealthChecker: service.StaticHealthChecker{},
		Authenticator: service.AuthenticationService{
			TenantApps:    dao.NewTenantAppRepository(db),
			TokenVerifier: service.HS256JWTVerifier{},
			Now:           func() time.Time { return now },
		},
		FlagController: controller.NewFlagController(service.NewFlagService(dao.NewFlagRepository(db), dao.NewFlagOverrideRepository(db)), validation.NewValidator()),
		EvalController: controller.NewEvalController(service.NewEvalService(dao.NewFlagRepository(db), dao.NewFlagOverrideRepository(db))),
	})

	createTenantOneReq := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{
		"key": "new_dashboard",
		"description": "Tenant one dashboard",
		"default_enabled": false
	}`))
	createTenantOneReq.Header.Set("Content-Type", "application/json")
	createTenantOneReq.Header.Set("X-App-ID", "app-acme")
	createTenantOneReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	createTenantOneRec := httptest.NewRecorder()
	router.ServeHTTP(createTenantOneRec, createTenantOneReq)
	if createTenantOneRec.Code != http.StatusCreated {
		t.Fatalf("expected tenant one flag create 201, got %d: %s", createTenantOneRec.Code, createTenantOneRec.Body.String())
	}

	createTenantTwoReq := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{
		"key": "new_dashboard",
		"description": "Tenant two dashboard",
		"default_enabled": false
	}`))
	createTenantTwoReq.Header.Set("Content-Type", "application/json")
	createTenantTwoReq.Header.Set("X-App-ID", "app-globex")
	createTenantTwoReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "globex-secret", map[string]any{
		"app_id": "app-globex",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	createTenantTwoRec := httptest.NewRecorder()
	router.ServeHTTP(createTenantTwoRec, createTenantTwoReq)
	if createTenantTwoRec.Code != http.StatusCreated {
		t.Fatalf("expected tenant two flag create 201, got %d: %s", createTenantTwoRec.Code, createTenantTwoRec.Body.String())
	}

	bulkReq := httptest.NewRequest(http.MethodPost, "/v1/flags/new_dashboard/users/bulk-set", strings.NewReader(`{
		"overrides": [
			{"user_id":"user_123","enabled":true}
		]
	}`))
	bulkReq.Header.Set("Content-Type", "application/json")
	bulkReq.SetPathValue("flagKey", "new_dashboard")
	bulkReq.Header.Set("X-App-ID", "app-acme")
	bulkReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	bulkRec := httptest.NewRecorder()
	router.ServeHTTP(bulkRec, bulkReq)
	if bulkRec.Code != http.StatusOK {
		t.Fatalf("expected tenant one bulk set 200, got %d: %s", bulkRec.Code, bulkRec.Body.String())
	}

	evalTenantOneReq := httptest.NewRequest(http.MethodGet, "/eval?flag=new_dashboard&user=user_123", nil)
	evalTenantOneReq.Header.Set("X-App-ID", "app-acme")
	evalTenantOneReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "acme-secret", map[string]any{
		"app_id": "app-acme",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	evalTenantOneRec := httptest.NewRecorder()
	router.ServeHTTP(evalTenantOneRec, evalTenantOneReq)
	if evalTenantOneRec.Code != http.StatusOK {
		t.Fatalf("expected tenant one eval 200, got %d: %s", evalTenantOneRec.Code, evalTenantOneRec.Body.String())
	}

	var tenantOneBody struct {
		Success bool   `json:"success"`
		Result  string `json:"result"`
	}
	if err := json.Unmarshal(evalTenantOneRec.Body.Bytes(), &tenantOneBody); err != nil {
		t.Fatalf("expected valid tenant one eval body, got %v", err)
	}
	if !tenantOneBody.Success || tenantOneBody.Result != "on" {
		t.Fatalf("unexpected tenant one eval response: %+v", tenantOneBody)
	}

	evalTenantTwoReq := httptest.NewRequest(http.MethodGet, "/eval?flag=new_dashboard&user=user_123", nil)
	evalTenantTwoReq.Header.Set("X-App-ID", "app-globex")
	evalTenantTwoReq.Header.Set("Authorization", "Bearer "+testRouterJWT(t, "globex-secret", map[string]any{
		"app_id": "app-globex",
		"sub":    "user-123",
		"exp":    now.Add(time.Hour).Unix(),
	}))
	evalTenantTwoRec := httptest.NewRecorder()
	router.ServeHTTP(evalTenantTwoRec, evalTenantTwoReq)
	if evalTenantTwoRec.Code != http.StatusOK {
		t.Fatalf("expected tenant two eval 200, got %d: %s", evalTenantTwoRec.Code, evalTenantTwoRec.Body.String())
	}

	var tenantTwoBody struct {
		Success bool   `json:"success"`
		Result  string `json:"result"`
	}
	if err := json.Unmarshal(evalTenantTwoRec.Body.Bytes(), &tenantTwoBody); err != nil {
		t.Fatalf("expected valid tenant two eval body, got %v", err)
	}
	if !tenantTwoBody.Success || tenantTwoBody.Result != "off" {
		t.Fatalf("unexpected tenant two eval response: %+v", tenantTwoBody)
	}

	var count int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM flag_user_overrides WHERE tenant_id = ?`, tenantOneID).Scan(&count); err != nil {
		t.Fatalf("count tenant one overrides: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 tenant one override row, got %d", count)
	}
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM flags WHERE tenant_id = ?`, tenantTwoID).Scan(&count); err != nil {
		t.Fatalf("count tenant two flags: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 tenant two flag row, got %d", count)
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
