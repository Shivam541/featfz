package handlers

import (
	"net/http"

	"github.com/shivam/featfz/feat-manager/internal/http/response"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type HealthHandler struct {
	health service.HealthChecker
}

func NewHealth(health service.HealthChecker) http.Handler {
	return HealthHandler{health: health}
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := h.health.Check(r.Context())

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"status":  status.Status,
	})
}
