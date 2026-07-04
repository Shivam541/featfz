package validation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequiredValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/eval?flag=test-flag", nil)
	req.Header.Set("X-App-ID", " app-123 ")
	req.SetPathValue("flagKey", " feature_a ")

	headerValue, headerErr := RequiredHeader(req, "X-App-ID")
	if headerErr != nil {
		t.Fatalf("expected nil header error, got %+v", headerErr)
	}
	if headerValue != "app-123" {
		t.Fatalf("expected trimmed header value, got %q", headerValue)
	}

	queryValue, queryErr := RequiredQuery(req, "flag")
	if queryErr != nil {
		t.Fatalf("expected nil query error, got %+v", queryErr)
	}
	if queryValue != "test-flag" {
		t.Fatalf("expected test-flag, got %q", queryValue)
	}

	pathValue, pathErr := RequiredPath(req, "flagKey")
	if pathErr != nil {
		t.Fatalf("expected nil path error, got %+v", pathErr)
	}
	if pathValue != "feature_a" {
		t.Fatalf("expected feature_a, got %q", pathValue)
	}
}

func TestRequiredHeaderMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/eval", nil)

	_, err := RequiredHeader(req, "X-App-ID")
	if err == nil {
		t.Fatal("expected error for missing header")
	}

	if err.Field != "X-App-ID" || err.Message != "is required" {
		t.Fatalf("unexpected error detail: %+v", err)
	}
}

func TestDecodeJSONBody(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{"key":"new_dashboard","default_enabled":true}`))
		req.Header.Set("Content-Type", "application/json")

		var body struct {
			Key            string `json:"key"`
			DefaultEnabled bool   `json:"default_enabled"`
		}

		errs := DecodeJSONBody(req, &body)
		if len(errs) != 0 {
			t.Fatalf("expected no errors, got %+v", errs)
		}

		if body.Key != "new_dashboard" || !body.DefaultEnabled {
			t.Fatalf("unexpected decoded body: %+v", body)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{"key":"new_dashboard","extra":true}`))
		req.Header.Set("Content-Type", "application/json")

		var body struct {
			Key string `json:"key"`
		}

		errs := DecodeJSONBody(req, &body)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %+v", errs)
		}

		if errs[0].Field != "extra" || errs[0].Message != "is not allowed" {
			t.Fatalf("unexpected error detail: %+v", errs[0])
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{"key":`))
		req.Header.Set("Content-Type", "application/json")

		var body struct {
			Key string `json:"key"`
		}

		errs := DecodeJSONBody(req, &body)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %+v", errs)
		}

		if errs[0].Field != "body" || errs[0].Message != "contains invalid JSON" {
			t.Fatalf("unexpected error detail: %+v", errs[0])
		}
	})

	t.Run("wrong content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/flags", strings.NewReader(`{"key":"new_dashboard"}`))
		req.Header.Set("Content-Type", "text/plain")

		var body struct {
			Key string `json:"key"`
		}

		errs := DecodeJSONBody(req, &body)
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %+v", errs)
		}

		if errs[0].Field != "body" || errs[0].Message != "must be application/json" {
			t.Fatalf("unexpected error detail: %+v", errs[0])
		}
	})
}
