package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

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
