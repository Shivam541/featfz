package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/http/response"
	"github.com/shivam/featfz/feat-manager/internal/http/validation"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type CreateFlagHandler struct {
	flags service.FlagCreator
}

type createFlagRequest struct {
	Key            string `json:"key"`
	Description    string `json:"description"`
	DefaultEnabled *bool  `json:"default_enabled"`
}

func NewCreateFlag(flags service.FlagCreator) http.Handler {
	return CreateFlagHandler{flags: flags}
}

func (h CreateFlagHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	var req createFlagRequest
	details := validation.DecodeJSONBody(r, &req)
	if len(details) > 0 {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", details)
		return
	}

	req.Key = strings.TrimSpace(req.Key)
	if req.Key == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", []response.ErrorDetail{
			{Field: "key", Message: "is required"},
		})
		return
	}

	if len(req.Key) > 255 {
		response.WriteError(w, http.StatusUnprocessableEntity, "validation_failed", "The request could not be processed.", []response.ErrorDetail{
			{Field: "key", Message: "must be at most 255 characters"},
		})
		return
	}

	if req.DefaultEnabled == nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", []response.ErrorDetail{
			{Field: "default_enabled", Message: "is required"},
		})
		return
	}

	created, err := h.flags.Create(r.Context(), tenant.TenantID, service.CreateFlagInput{
		Key:            req.Key,
		Description:    req.Description,
		DefaultEnabled: *req.DefaultEnabled,
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
