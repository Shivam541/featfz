package handlers

import (
	"net/http"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/http/response"
)

type AuthCheckHandler struct{}

func NewAuthCheck() http.Handler {
	return AuthCheckHandler{}
}

func (AuthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, ok := requestctx.TenantFrom(r.Context())
	if !ok {
		response.WriteError(w, http.StatusInternalServerError, "tenant_context_missing", "Something went wrong.", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"tenant_id": tenant.TenantID,
			"app_id":    tenant.AppID,
			"subject":   tenant.Subject,
		},
	})
}
