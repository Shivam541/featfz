package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type stubEvaluator struct {
	called   bool
	tenantID int64
	flagKey  string
	userID   string
	result   service.EvalResult
	err      error
}

func (s *stubEvaluator) Evaluate(_ context.Context, tenantID int64, flagKey, userID string) (service.EvalResult, error) {
	s.called = true
	s.tenantID = tenantID
	s.flagKey = flagKey
	s.userID = userID
	return s.result, s.err
}

func TestEvaluateFlag(t *testing.T) {
	tests := []struct {
		name   string
		result service.EvalResult
		want   string
	}{
		{name: "on", result: service.EvalResult{Enabled: true}, want: "on"},
		{name: "off", result: service.EvalResult{Enabled: false}, want: "off"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubEvaluator{result: tt.result}
			req := httptest.NewRequest(http.MethodGet, "/eval?flag=new_dashboard&user=user_123", nil)
			req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
			rec := httptest.NewRecorder()

			NewEvalController(stub).EvaluateFlag(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
			if !stub.called {
				t.Fatal("expected evaluator to be called")
			}
			if stub.tenantID != 7 || stub.flagKey != "new_dashboard" || stub.userID != "user_123" {
				t.Fatalf("unexpected evaluator call: %+v", stub)
			}

			var body struct {
				Success bool   `json:"success"`
				Result  string `json:"result"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("expected valid json body, got %v", err)
			}
			if !body.Success || body.Result != tt.want {
				t.Fatalf("unexpected response body: %+v", body)
			}
		})
	}
}

func TestEvaluateFlagValidatesInput(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
	}{
		{name: "missing flag", target: "/eval?user=user_123", wantStatus: http.StatusBadRequest},
		{name: "missing user", target: "/eval?flag=new_dashboard", wantStatus: http.StatusBadRequest},
		{name: "blank user", target: "/eval?flag=new_dashboard&user=%20%20%20", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubEvaluator{}
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
			rec := httptest.NewRecorder()

			NewEvalController(stub).EvaluateFlag(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
			if stub.called {
				t.Fatal("expected evaluator not to be called")
			}
		})
	}
}

func TestEvaluateFlagMapsErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "not found",
			err:        service.ErrFlagNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "flag_not_found",
		},
		{
			name:       "dependency failure",
			err:        errors.New("db down"),
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "service_unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubEvaluator{err: tt.err}
			req := httptest.NewRequest(http.MethodGet, "/eval?flag=new_dashboard&user=user_123", nil)
			req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
			rec := httptest.NewRecorder()

			NewEvalController(stub).EvaluateFlag(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rec.Code)
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
			if body.Error.Code != tt.wantCode {
				t.Fatalf("expected error code %q, got %q", tt.wantCode, body.Error.Code)
			}
		})
	}
}

func TestEvaluateFlagRequiresTenantContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/eval?flag=new_dashboard&user=user_123", nil)
	rec := httptest.NewRecorder()

	NewEvalController(&stubEvaluator{}).EvaluateFlag(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
