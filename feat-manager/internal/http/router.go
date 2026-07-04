package httpapi

import (
	"net/http"

	"github.com/shivam/featfz/feat-manager/internal/http/handlers"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handlers.Health)
	return mux
}
