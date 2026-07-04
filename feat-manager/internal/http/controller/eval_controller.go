package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/http/response"
	"github.com/shivam/featfz/feat-manager/internal/http/validation"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type EvalController struct {
	Evaluator service.Evaluator
}

func NewEvalController(evaluator service.Evaluator) *EvalController {
	return &EvalController{Evaluator: evaluator}
}

func (c *EvalController) EvaluateFlag(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	flagKey, flagErr := validation.RequiredQuery(r, "flag")
	userID, userErr := validation.RequiredQuery(r, "user")
	if flagErr != nil || userErr != nil {
		details := make([]response.ErrorDetail, 0, 2)
		if flagErr != nil {
			details = append(details, *flagErr)
		}
		if userErr != nil {
			details = append(details, *userErr)
		}
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", details)
		return
	}

	if len(flagKey) > 255 || len(userID) > 255 {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", []response.ErrorDetail{{
			Field:   "query",
			Message: "flag and user must be at most 255 characters",
		}})
		return
	}

	result, err := c.Evaluator.Evaluate(r.Context(), tenant.TenantID, strings.TrimSpace(flagKey), strings.TrimSpace(userID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFlagNotFound):
			response.WriteError(w, http.StatusNotFound, "flag_not_found", "The requested flag was not found.", nil)
		case errors.Is(err, service.ErrInvalidEvalInput):
			response.WriteError(w, http.StatusBadRequest, "invalid_request", "The request is invalid.", nil)
		default:
			response.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "The service is temporarily unavailable.", nil)
		}
		return
	}

	status := "off"
	if result.Enabled {
		status = "on"
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"result":  status,
	})
}
