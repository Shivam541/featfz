package controller

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
	"github.com/shivam/featfz/feat-manager/internal/http/validation"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type stubFlagManager struct {
	createCalled  bool
	listCalled    bool
	getCalled     bool
	updateCalled  bool
	archiveCalled bool
	bulkCalled    bool
	tenantID      int64
	key           string
	createInput   service.CreateFlagInput
	updateInput   service.UpdateFlagInput
	bulkInput     []service.FlagUserOverrideInput
	createdResult domain.Flag
	listResult    []domain.Flag
	getResult     domain.Flag
	updateResult  domain.Flag
	createErr     error
	listErr       error
	getErr        error
	updateErr     error
	archiveErr    error
	bulkErr       error
}

func (s *stubFlagManager) Create(_ context.Context, tenantID int64, input service.CreateFlagInput) (domain.Flag, error) {
	s.createCalled = true
	s.tenantID = tenantID
	s.createInput = input
	return s.createdResult, s.createErr
}

func (s *stubFlagManager) List(_ context.Context, tenantID int64) ([]domain.Flag, error) {
	s.listCalled = true
	s.tenantID = tenantID
	return s.listResult, s.listErr
}

func (s *stubFlagManager) Get(_ context.Context, tenantID int64, key string) (domain.Flag, error) {
	s.getCalled = true
	s.tenantID = tenantID
	s.key = key
	return s.getResult, s.getErr
}

func (s *stubFlagManager) Update(_ context.Context, tenantID int64, key string, input service.UpdateFlagInput) (domain.Flag, error) {
	s.updateCalled = true
	s.tenantID = tenantID
	s.key = key
	s.updateInput = input
	return s.updateResult, s.updateErr
}

func (s *stubFlagManager) Archive(_ context.Context, tenantID int64, key string) error {
	s.archiveCalled = true
	s.tenantID = tenantID
	s.key = key
	return s.archiveErr
}

func (s *stubFlagManager) BulkSetOverrides(_ context.Context, tenantID int64, key string, input []service.FlagUserOverrideInput) (int, error) {
	s.bulkCalled = true
	s.tenantID = tenantID
	s.key = key
	s.bulkInput = input
	return len(input), s.bulkErr
}

func TestCreateFlag(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	flagService := &stubFlagManager{
		createdResult: domain.Flag{
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

	NewFlagController(flagService, validation.NewValidator()).CreateFlag(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if !flagService.createCalled {
		t.Fatal("expected flag service to be called")
	}
	if flagService.tenantID != 7 {
		t.Fatalf("expected tenant id 7, got %d", flagService.tenantID)
	}
	if flagService.createInput.Key != "new_dashboard" {
		t.Fatalf("expected trimmed key, got %q", flagService.createInput.Key)
	}
	if flagService.createInput.Description != "Enable the new dashboard experience" {
		t.Fatalf("expected description to pass through, got %q", flagService.createInput.Description)
	}
	if flagService.createInput.DefaultEnabled {
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

func TestListGetUpdateArchiveFlags(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	tests := []struct {
		name string
		run  func(t *testing.T, controller *FlagController, recorder *httptest.ResponseRecorder, request *http.Request, stub *stubFlagManager)
	}{
		{
			name: "list flags",
			run: func(t *testing.T, controller *FlagController, recorder *httptest.ResponseRecorder, request *http.Request, stub *stubFlagManager) {
				controller.ListFlags(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Fatalf("expected 200, got %d", recorder.Code)
				}
				if !stub.listCalled {
					t.Fatal("expected list to be called")
				}
				var body struct {
					Success bool `json:"success"`
					Data    struct {
						Flags []struct {
							ID             int64  `json:"id"`
							TenantID       int64  `json:"tenant_id"`
							Key            string `json:"key"`
							Description    string `json:"description"`
							DefaultEnabled bool   `json:"default_enabled"`
						} `json:"flags"`
					} `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatalf("expected valid json body, got %v", err)
				}
				if !body.Success || len(body.Data.Flags) != 1 || body.Data.Flags[0].Key != "new_dashboard" {
					t.Fatalf("unexpected list payload: %+v", body.Data.Flags)
				}
			},
		},
		{
			name: "get flag",
			run: func(t *testing.T, controller *FlagController, recorder *httptest.ResponseRecorder, request *http.Request, stub *stubFlagManager) {
				controller.GetFlag(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Fatalf("expected 200, got %d", recorder.Code)
				}
				if !stub.getCalled {
					t.Fatal("expected get to be called")
				}
			},
		},
		{
			name: "update flag",
			run: func(t *testing.T, controller *FlagController, recorder *httptest.ResponseRecorder, request *http.Request, stub *stubFlagManager) {
				controller.UpdateFlag(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Fatalf("expected 200, got %d", recorder.Code)
				}
				if !stub.updateCalled {
					t.Fatal("expected update to be called")
				}
				if stub.updateInput.Description == nil || *stub.updateInput.Description != "Updated rollout" {
					t.Fatalf("expected description update, got %#v", stub.updateInput.Description)
				}
				if stub.updateInput.DefaultEnabled == nil || !*stub.updateInput.DefaultEnabled {
					t.Fatalf("expected default_enabled update, got %#v", stub.updateInput.DefaultEnabled)
				}
			},
		},
		{
			name: "archive flag",
			run: func(t *testing.T, controller *FlagController, recorder *httptest.ResponseRecorder, request *http.Request, stub *stubFlagManager) {
				controller.ArchiveFlag(recorder, request)
				if recorder.Code != http.StatusOK {
					t.Fatalf("expected 200, got %d", recorder.Code)
				}
				if !stub.archiveCalled {
					t.Fatal("expected archive to be called")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubFlagManager{
				listResult: []domain.Flag{{
					ID:             9,
					TenantID:       7,
					Key:            "new_dashboard",
					Description:    "Enable the new dashboard experience",
					DefaultEnabled: false,
					CreatedAt:      now,
					UpdatedAt:      now,
				}},
				getResult: domain.Flag{
					ID:             9,
					TenantID:       7,
					Key:            "new_dashboard",
					Description:    "Enable the new dashboard experience",
					DefaultEnabled: false,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				updateResult: domain.Flag{
					ID:             9,
					TenantID:       7,
					Key:            "new_dashboard",
					Description:    "Updated rollout",
					DefaultEnabled: true,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			}

			controller := NewFlagController(stub, validation.NewValidator())
			request := httptest.NewRequest(http.MethodGet, "/v1/flags/new_dashboard", nil)
			request = request.WithContext(requestctx.WithTenant(request.Context(), requestctx.Tenant{TenantID: 7}))
			request.SetPathValue("flagKey", "new_dashboard")
			if tt.name == "update flag" {
				request = httptest.NewRequest(http.MethodPatch, "/v1/flags/new_dashboard", bytes.NewBufferString(`{
					"description":"Updated rollout",
					"default_enabled":true
				}`))
				request.Header.Set("Content-Type", "application/json")
				request = request.WithContext(requestctx.WithTenant(request.Context(), requestctx.Tenant{TenantID: 7}))
				request.SetPathValue("flagKey", "new_dashboard")
			}
			if tt.name == "archive flag" {
				request = httptest.NewRequest(http.MethodDelete, "/v1/flags/new_dashboard", nil)
				request = request.WithContext(requestctx.WithTenant(request.Context(), requestctx.Tenant{TenantID: 7}))
				request.SetPathValue("flagKey", "new_dashboard")
			}
			if tt.name == "list flags" {
				request = httptest.NewRequest(http.MethodGet, "/v1/flags", nil)
				request = request.WithContext(requestctx.WithTenant(request.Context(), requestctx.Tenant{TenantID: 7}))
			}

			recorder := httptest.NewRecorder()
			tt.run(t, controller, recorder, request, stub)
		})
	}
}

func TestCreateFlagValidatesInput(t *testing.T) {
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
			flagService := &stubFlagManager{}
			req := httptest.NewRequest(http.MethodPost, "/v1/flags", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
			rec := httptest.NewRecorder()

			NewFlagController(flagService, validation.NewValidator()).CreateFlag(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if flagService.createCalled {
				t.Fatal("expected service not to be called on validation error")
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

func TestUpdateFlagRequiresChanges(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/v1/flags/new_dashboard", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
	req.SetPathValue("flagKey", "new_dashboard")
	rec := httptest.NewRecorder()

	NewFlagController(&stubFlagManager{}, validation.NewValidator()).UpdateFlag(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBulkSetOverrides(t *testing.T) {
	flagService := &stubFlagManager{}
	req := httptest.NewRequest(http.MethodPost, "/v1/flags/new_dashboard/users/bulk-set", bytes.NewBufferString(`{
		"overrides": [
			{"user_id":" user_123 ","enabled":true},
			{"user_id":"user_456","enabled":false},
			{"user_id":"user_123","enabled":false}
		]
	}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
	req.SetPathValue("flagKey", "new_dashboard")
	rec := httptest.NewRecorder()

	NewFlagController(flagService, validation.NewValidator()).BulkSetOverrides(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !flagService.bulkCalled {
		t.Fatal("expected bulk set to be called")
	}
	if len(flagService.bulkInput) != 2 {
		t.Fatalf("expected deduped overrides, got %d", len(flagService.bulkInput))
	}

	got := map[string]bool{}
	for _, override := range flagService.bulkInput {
		got[override.UserID] = override.Enabled
	}
	if got["user_123"] {
		t.Fatal("expected last user_123 value to win and be false")
	}
	if got["user_456"] {
		t.Fatal("expected user_456 to remain false")
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Applied int `json:"applied"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected valid json body, got %v", err)
	}
	if !body.Success || body.Data.Applied != 2 {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestBulkSetOverridesValidatesInput(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing overrides",
			body:       `{}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation_failed",
		},
		{
			name:       "missing enabled",
			body:       `{"overrides":[{"user_id":"user_123"}]}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation_failed",
		},
		{
			name:       "blank user id",
			body:       `{"overrides":[{"user_id":"   ","enabled":true}]}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "validation_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubFlagManager{}
			req := httptest.NewRequest(http.MethodPost, "/v1/flags/new_dashboard/users/bulk-set", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(requestctx.WithTenant(req.Context(), requestctx.Tenant{TenantID: 7}))
			req.SetPathValue("flagKey", "new_dashboard")
			rec := httptest.NewRecorder()

			NewFlagController(stub, validation.NewValidator()).BulkSetOverrides(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if stub.bulkCalled {
				t.Fatal("expected service not to be called on validation error")
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

func TestCreateFlagRequiresTenantContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/flags", bytes.NewBufferString(`{"key":"new_dashboard","default_enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	NewFlagController(&stubFlagManager{}, validation.NewValidator()).CreateFlag(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
