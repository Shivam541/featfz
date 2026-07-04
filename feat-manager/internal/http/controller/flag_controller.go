package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/http/response"
	"github.com/shivam/featfz/feat-manager/internal/http/validation"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type FlagController struct {
	FlagService service.FlagCreator
	Validator   *validator.Validate
}

type CreateFlagRequest struct {
	Key            string `json:"key" validate:"required,max=255"`
	Description    string `json:"description" validate:"omitempty,max=500"`
	DefaultEnabled *bool  `json:"default_enabled" validate:"required"`
}

func NewFlagController(flagService service.FlagCreator, validator *validator.Validate) *FlagController {
	if validator == nil {
		validator = validation.NewValidator()
	}

	return &FlagController{
		FlagService: flagService,
		Validator:   validator,
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

	created, err := c.FlagService.Create(r.Context(), tenant.TenantID, service.CreateFlagInput{
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

	response.WriteJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":              created.ID,
			"tenant_id":       created.TenantID,
			"key":             created.Key,
			"description":     created.Description,
			"default_enabled": created.DefaultEnabled,
			"archived_at":     created.ArchivedAt,
			"created_at":      created.CreatedAt,
			"updated_at":      created.UpdatedAt,
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
