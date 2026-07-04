package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/shivam/featfz/feat-manager/internal/config"
	httpapi "github.com/shivam/featfz/feat-manager/internal/http"
	"github.com/shivam/featfz/feat-manager/internal/mysql"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type Dependencies struct {
	OpenDB        func(context.Context, config.Config) (*sql.DB, error)
	Logger        *slog.Logger
	HealthChecker service.HealthChecker
}

type Runtime struct {
	DB      *sql.DB
	Handler http.Handler
}

func New(ctx context.Context, cfg config.Config, deps Dependencies) (*Runtime, error) {
	openDB := deps.OpenDB
	if openDB == nil {
		openDB = mysql.Open
	}

	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	healthChecker := deps.HealthChecker
	if healthChecker == nil {
		healthChecker = service.StaticHealthChecker{}
	}

	db, err := openDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	return &Runtime{
		DB: db,
		Handler: httpapi.NewRouter(httpapi.RouterDependencies{
			Logger:        logger,
			HealthChecker: healthChecker,
		}),
	}, nil
}

func (r *Runtime) Close() error {
	if r == nil || r.DB == nil {
		return nil
	}

	return r.DB.Close()
}
