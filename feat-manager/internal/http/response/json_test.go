package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, http.StatusBadRequest, "validation_error", "request validation failed", []ErrorDetail{
		{Field: "X-App-ID", Message: "is required"},
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json content type, got %q", got)
	}

	var body struct {
		Success bool      `json:"success"`
		Error   ErrorBody `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json body, got %v", err)
	}

	if body.Success {
		t.Fatal("expected success=false")
	}

	if body.Error.Code != "validation_error" {
		t.Fatalf("expected validation_error code, got %q", body.Error.Code)
	}

	if len(body.Error.Details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(body.Error.Details))
	}
}
