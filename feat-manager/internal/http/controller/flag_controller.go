package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/http/response"
	"github.com/shivam/featfz/feat-manager/internal/http/validation"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type FlagController struct {
	Flags     service.FlagManager
	Validator *validator.Validate
}

type CreateFlagRequest struct {
	Key            string `json:"key" validate:"required,max=255"`
	Description    string `json:"description" validate:"omitempty,max=500"`
	DefaultEnabled *bool  `json:"default_enabled" validate:"required"`
}

type UpdateFlagRequest struct {
	Description    *string `json:"description" validate:"omitempty,max=500"`
	DefaultEnabled *bool   `json:"default_enabled"`
}

func NewFlagController(flagService service.FlagManager, validator *validator.Validate) *FlagController {
	if validator == nil {
		validator = validation.NewValidator()
	}

	return &FlagController{
		Flags:     flagService,
		Validator: validator,
	}
}

func (c *FlagController) CreateFlag(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	var request CreateFlagRequest
	details := validation.DecodeJSONBody(r, &request)
	if len(details) > 0 {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", details)
		return
	}

	if err := c.Validator.Struct(request); err != nil {
		details := validation.ValidationDetails(err)
		if hasValidationRequiredDetails(details) {
			response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", details)
			return
		}

		response.WriteError(w, http.StatusUnprocessableEntity, "validation_failed", "The request could not be processed.", details)
		return
	}

	created, err := c.Flags.Create(r.Context(), tenant.TenantID, service.CreateFlagInput{
		Key:            strings.TrimSpace(request.Key),
		Description:    request.Description,
		DefaultEnabled: *request.DefaultEnabled,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFlagAlreadyExists):
			response.WriteError(w, http.StatusConflict, "flag_already_exists", "A flag with this key already exists.", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
		}
		return
	}

	writeFlagData(w, http.StatusCreated, created)
}

func (c *FlagController) ListFlags(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	flags, err := c.Flags.List(r.Context(), tenant.TenantID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
		return
	}

	writeFlagsList(w, http.StatusOK, flags)
}

func (c *FlagController) GetFlag(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	key, detail := validation.RequiredPath(r, "flagKey")
	if detail != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", []response.ErrorDetail{*detail})
		return
	}

	flag, err := c.Flags.Get(r.Context(), tenant.TenantID, key)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFlagNotFound):
			response.WriteError(w, http.StatusNotFound, "flag_not_found", "The requested flag was not found.", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
		}
		return
	}

	writeFlagData(w, http.StatusOK, flag)
}

func (c *FlagController) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	key, detail := validation.RequiredPath(r, "flagKey")
	if detail != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", []response.ErrorDetail{*detail})
		return
	}

	var request UpdateFlagRequest
	details := validation.DecodeJSONBody(r, &request)
	if len(details) > 0 {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", details)
		return
	}

	if request.Description == nil && request.DefaultEnabled == nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", []response.ErrorDetail{{
			Field:   "body",
			Message: "must include description or default_enabled",
		}})
		return
	}

	if err := c.Validator.Struct(request); err != nil {
		details := validation.ValidationDetails(err)
		response.WriteError(w, http.StatusUnprocessableEntity, "validation_failed", "The request could not be processed.", details)
		return
	}

	updated, err := c.Flags.Update(r.Context(), tenant.TenantID, key, service.UpdateFlagInput{
		Description:    request.Description,
		DefaultEnabled: request.DefaultEnabled,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFlagNotFound):
			response.WriteError(w, http.StatusNotFound, "flag_not_found", "The requested flag was not found.", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
		}
		return
	}

	writeFlagData(w, http.StatusOK, updated)
}

func (c *FlagController) ArchiveFlag(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	key, detail := validation.RequiredPath(r, "flagKey")
	if detail != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", []response.ErrorDetail{*detail})
		return
	}

	if err := c.Flags.Archive(r.Context(), tenant.TenantID, key); err != nil {
		switch {
		case errors.Is(err, service.ErrFlagNotFound):
			response.WriteError(w, http.StatusNotFound, "flag_not_found", "The requested flag was not found.", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

func writeFlagData(w http.ResponseWriter, status int, flag domain.Flag) {
	response.WriteJSON(w, status, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":              flag.ID,
			"tenant_id":       flag.TenantID,
			"key":             flag.Key,
			"description":     flag.Description,
			"default_enabled": flag.DefaultEnabled,
			"archived_at":     flag.ArchivedAt,
			"created_at":      flag.CreatedAt,
			"updated_at":      flag.UpdatedAt,
		},
	})
}

func writeFlagsList(w http.ResponseWriter, status int, flags []domain.Flag) {
	items := make([]map[string]any, 0, len(flags))
	for _, flag := range flags {
		items = append(items, map[string]any{
			"id":              flag.ID,
			"tenant_id":       flag.TenantID,
			"key":             flag.Key,
			"description":     flag.Description,
			"default_enabled": flag.DefaultEnabled,
			"archived_at":     flag.ArchivedAt,
			"created_at":      flag.CreatedAt,
			"updated_at":      flag.UpdatedAt,
		})
	}

	response.WriteJSON(w, status, map[string]any{
		"success": true,
		"data": map[string]any{
			"flags": items,
		},
	})
}

func hasValidationRequiredDetails(details []response.ErrorDetail) bool {
	for _, detail := range details {
		if strings.Contains(strings.ToLower(detail.Message), "required") {
			return true
		}
	}

	return false
}
