package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
)

func TestAuthCheck(t *testing.T) {
	t.Run("returns tenant context when present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/auth/check", nil)
		req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{
			TenantID: 7,
			AppID:    "app-acme",
			Subject:  "user-123",
		}))
		rec := httptest.NewRecorder()

		NewAuthCheck().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var body struct {
			Success bool `json:"success"`
			Data    struct {
				TenantID int64  `json:"tenant_id"`
				AppID    string `json:"app_id"`
				Subject  string `json:"subject"`
			} `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("expected valid json body, got %v", err)
		}

		if !body.Success {
			t.Fatal("expected success=true")
		}
		if body.Data.TenantID != 7 || body.Data.AppID != "app-acme" || body.Data.Subject != "user-123" {
			t.Fatalf("unexpected tenant payload: %+v", body.Data)
		}
	})

	t.Run("returns internal error when tenant context missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/auth/check", nil)
		rec := httptest.NewRecorder()

		NewAuthCheck().ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}

		var body struct {
			Success bool `json:"success"`
			Error   struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("expected valid json body, got %v", err)
		}
		if body.Error.Code != "tenant_context_missing" {
			t.Fatalf("expected tenant_context_missing, got %q", body.Error.Code)
		}
		if body.Error.Message != "Something went wrong." {
			t.Fatalf("expected generic error message, got %q", body.Error.Message)
		}
	})
}
