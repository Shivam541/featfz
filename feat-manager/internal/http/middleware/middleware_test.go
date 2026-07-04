package middleware

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
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
