package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type createFlagCreatorStub struct {
	called   bool
	tenantID int64
	input    service.CreateFlagInput
	result   domain.Flag
	err      error
}

func (s *createFlagCreatorStub) Create(_ context.Context, tenantID int64, input service.CreateFlagInput) (domain.Flag, error) {
	s.called = true
	s.tenantID = tenantID
	s.input = input
	return s.result, s.err
}

func TestCreateFlagHandler(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	creator := &createFlagCreatorStub{
		result: domain.Flag{
			ID:             9,
			TenantID:       7,
			Key:            "new_dashboard",
			Description:    "Enable the new dashboard experience",
			DefaultEnabled: false,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/flags", bytes.NewBufferString(`{
		"key": " new_dashboard ",
		"description": "Enable the new dashboard experience",
		"default_enabled": false
	}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{
		TenantID: 7,
		AppID:    "app-acme",
		Subject:  "user-123",
	}))
	rec := httptest.NewRecorder()

	NewCreateFlag(creator).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if !creator.called {
		t.Fatal("expected flag creator to be called")
	}
	if creator.tenantID != 7 {
		t.Fatalf("expected tenant id 7, got %d", creator.tenantID)
	}
	if creator.input.Key != "new_dashboard" {
		t.Fatalf("expected trimmed key, got %q", creator.input.Key)
	}
	if creator.input.Description != "Enable the new dashboard experience" {
		t.Fatalf("expected description to pass through, got %q", creator.input.Description)
	}
	if creator.input.DefaultEnabled {
		t.Fatal("expected default_enabled=false")
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			ID             int64  `json:"id"`
			TenantID       int64  `json:"tenant_id"`
			Key            string `json:"key"`
			Description    string `json:"description"`
			DefaultEnabled bool   `json:"default_enabled"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json body, got %v", err)
	}
	if !body.Success {
		t.Fatal("expected success=true")
	}
	if body.Data.ID != 9 || body.Data.TenantID != 7 || body.Data.Key != "new_dashboard" || body.Data.Description != "Enable the new dashboard experience" || body.Data.DefaultEnabled {
		t.Fatalf("unexpected create payload: %+v", body.Data)
	}
}

func TestCreateFlagHandlerValidatesInput(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing key",
			body:       `{"default_enabled":true}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
		},
		{
			name:       "missing default enabled",
			body:       `{"key":"new_dashboard"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
		},
		{
			name:       "key too long",
			body:       `{"key":"` + strings.Repeat("a", 256) + `","default_enabled":true}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &createFlagCreatorStub{}
			req := httptest.NewRequest(http.MethodPost, "/v1/flags", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
			rec := httptest.NewRecorder()

			NewCreateFlag(creator).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if creator.called {
				t.Fatal("expected creator not to be called on validation error")
			}

			var body struct {
				Success bool `json:"success"`
				Error   struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("expected valid error body, got %v", err)
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

func TestCreateFlagHandlerRequiresTenantContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/flags", bytes.NewBufferString(`{"key":"new_dashboard","default_enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	NewCreateFlag(&createFlagCreatorStub{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
