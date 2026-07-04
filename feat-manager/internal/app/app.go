package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/shivam/featfz/feat-manager/internal/config"
	"github.com/shivam/featfz/feat-manager/internal/dao"
	httpapi "github.com/shivam/featfz/feat-manager/internal/http"
	"github.com/shivam/featfz/feat-manager/internal/http/controller"
	"github.com/shivam/featfz/feat-manager/internal/http/validation"
	"github.com/shivam/featfz/feat-manager/internal/mysql"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type Dependencies struct {
	OpenDB                    func(context.Context, config.Config) (*sql.DB, error)
	Logger                    *slog.Logger
	HealthChecker             service.HealthChecker
	NewTenantAppRepository    func(*sql.DB) service.TenantAppRepository
	NewFlagRepository         func(*sql.DB) service.FlagRepository
	NewFlagOverrideRepository func(*sql.DB) service.FlagOverrideRepository
	NewAuthenticator          func(service.TenantAppRepository) service.Authenticator
	NewFlagService            func(service.FlagRepository, service.FlagOverrideRepository) service.FlagManager
	NewFlagController         func(service.FlagManager) *controller.FlagController
}

type Runtime struct {
	DB      *sql.DB
	Handler http.Handler
}

func New(ctx context.Context, cfg config.Config, dependencies Dependencies) (*Runtime, error) {
	openDB := dependencies.OpenDB
	if openDB == nil {
		openDB = mysql.Open
	}

	logger := dependencies.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	healthChecker := dependencies.HealthChecker
	if healthChecker == nil {
		healthChecker = service.StaticHealthChecker{}
	}

	newTenantAppRepository := dependencies.NewTenantAppRepository
	if newTenantAppRepository == nil {
		newTenantAppRepository = func(db *sql.DB) service.TenantAppRepository {
			return dao.NewTenantAppRepository(db)
		}
	}

	newFlagRepository := dependencies.NewFlagRepository
	if newFlagRepository == nil {
		newFlagRepository = func(db *sql.DB) service.FlagRepository {
			return dao.NewFlagRepository(db)
		}
	}

	newFlagOverrideRepository := dependencies.NewFlagOverrideRepository
	if newFlagOverrideRepository == nil {
		newFlagOverrideRepository = func(db *sql.DB) service.FlagOverrideRepository {
			return dao.NewFlagOverrideRepository(db)
		}
	}

	newAuthenticator := dependencies.NewAuthenticator
	if newAuthenticator == nil {
		newAuthenticator = func(repo service.TenantAppRepository) service.Authenticator {
			authenticator := service.NewAuthenticationService(repo)
			return authenticator
		}
	}

	newFlagService := dependencies.NewFlagService
	if newFlagService == nil {
		newFlagService = func(repo service.FlagRepository, overrideRepo service.FlagOverrideRepository) service.FlagManager {
			return service.NewFlagService(repo, overrideRepo)
		}
	}

	newFlagController := dependencies.NewFlagController
	if newFlagController == nil {
		newFlagController = func(flagService service.FlagManager) *controller.FlagController {
			return controller.NewFlagController(flagService, validation.NewValidator())
		}
	}

	db, err := openDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	tenantAppRepository := newTenantAppRepository(db)
	flagRepository := newFlagRepository(db)
	flagOverrideRepository := newFlagOverrideRepository(db)
	authenticator := newAuthenticator(tenantAppRepository)
	flagService := newFlagService(flagRepository, flagOverrideRepository)
	flagController := newFlagController(flagService)

	return &Runtime{
		DB: db,
		Handler: httpapi.NewRouter(httpapi.RouterDependencies{
			Logger:         logger,
			HealthChecker:  healthChecker,
			Authenticator:  authenticator,
			FlagController: flagController,
		}),
	}, nil
}

func (r *Runtime) Close() error {
	if r == nil || r.DB == nil {
		return nil
	}

	return r.DB.Close()
}
