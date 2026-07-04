package httpapi

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/shivam/featfz/feat-manager/internal/http/handlers"
	"github.com/shivam/featfz/feat-manager/internal/http/middleware"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type RouterDependencies struct {
	Logger        *slog.Logger
	HealthChecker service.HealthChecker
	Authenticator service.Authenticator
	FlagCreator   service.FlagCreator
}

func NewRouter(deps RouterDependencies) http.Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	healthChecker := deps.HealthChecker
	if healthChecker == nil {
		healthChecker = service.StaticHealthChecker{}
	}

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", handlers.NewHealth(healthChecker))
	mux.Handle("GET /v1/auth/check", middleware.RequireAuth(deps.Authenticator)(handlers.NewAuthCheck()))
	mux.Handle("POST /v1/flags", middleware.RequireAuth(deps.Authenticator)(handlers.NewCreateFlag(deps.FlagCreator)))

	return middleware.Chain(mux,
		middleware.Recover(logger),
		middleware.RequestContext(),
		middleware.RequestLogging(logger),
	)
}
