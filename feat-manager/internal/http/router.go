package httpapi

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/shivam/featfz/feat-manager/internal/http/controller"
	"github.com/shivam/featfz/feat-manager/internal/http/handlers"
	"github.com/shivam/featfz/feat-manager/internal/http/middleware"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type RouterDependencies struct {
	Logger         *slog.Logger
	HealthChecker  service.HealthChecker
	Authenticator  service.Authenticator
	FlagController *controller.FlagController
	EvalController *controller.EvalController
}

func NewRouter(dependencies RouterDependencies) http.Handler {
	logger := dependencies.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	healthChecker := dependencies.HealthChecker
	if healthChecker == nil {
		healthChecker = service.StaticHealthChecker{}
	}

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", handlers.NewHealth(healthChecker))
	mux.Handle("GET /v1/auth/check", middleware.RequireAuth(dependencies.Authenticator)(handlers.NewAuthCheck()))
	if dependencies.EvalController != nil {
		mux.Handle("GET /eval", middleware.RequireAuth(dependencies.Authenticator)(http.HandlerFunc(dependencies.EvalController.EvaluateFlag)))
	}
	if dependencies.FlagController != nil {
		mux.Handle("POST /v1/flags", middleware.RequireAuth(dependencies.Authenticator)(http.HandlerFunc(dependencies.FlagController.CreateFlag)))
		mux.Handle("GET /v1/flags", middleware.RequireAuth(dependencies.Authenticator)(http.HandlerFunc(dependencies.FlagController.ListFlags)))
		mux.Handle("GET /v1/flags/{flagKey}", middleware.RequireAuth(dependencies.Authenticator)(http.HandlerFunc(dependencies.FlagController.GetFlag)))
		mux.Handle("PATCH /v1/flags/{flagKey}", middleware.RequireAuth(dependencies.Authenticator)(http.HandlerFunc(dependencies.FlagController.UpdateFlag)))
		mux.Handle("DELETE /v1/flags/{flagKey}", middleware.RequireAuth(dependencies.Authenticator)(http.HandlerFunc(dependencies.FlagController.ArchiveFlag)))
		mux.Handle("POST /v1/flags/{flagKey}/users/bulk-set", middleware.RequireAuth(dependencies.Authenticator)(http.HandlerFunc(dependencies.FlagController.BulkSetOverrides)))
	}

	return middleware.Chain(mux,
		middleware.Recover(logger),
		middleware.RequestContext(),
		middleware.RequestLogging(logger),
	)
}
