package service

import (
	"context"

	"github.com/shivam/featfz/feat-manager/internal/domain"
)

type HealthStatus struct {
	Status string
}

type HealthChecker interface {
	Check(context.Context) HealthStatus
}

type StaticHealthChecker struct{}

func (StaticHealthChecker) Check(context.Context) HealthStatus {
	return HealthStatus{Status: "ok"}
}

type TenantAppRepository interface {
	FindByAppID(context.Context, string) (domain.TenantApp, error)
}

type FlagRepository interface {
	Create(context.Context, domain.Flag) (domain.Flag, error)
	FindByKey(context.Context, int64, string) (domain.Flag, error)
	ListActive(context.Context, int64) ([]domain.Flag, error)
	Update(context.Context, domain.Flag) (domain.Flag, error)
	Archive(context.Context, int64, string) error
}

type FlagCreator interface {
	Create(context.Context, int64, CreateFlagInput) (domain.Flag, error)
}

type FlagLister interface {
	List(context.Context, int64) ([]domain.Flag, error)
}

type FlagGetter interface {
	Get(context.Context, int64, string) (domain.Flag, error)
}

type FlagUpdater interface {
	Update(context.Context, int64, string, UpdateFlagInput) (domain.Flag, error)
}

type FlagArchiver interface {
	Archive(context.Context, int64, string) error
}

type FlagManager interface {
	FlagCreator
	FlagLister
	FlagGetter
	FlagUpdater
	FlagArchiver
	FlagOverrideBulkSetter
}

type FlagOverrideRepository interface {
	BulkUpsert(context.Context, int64, int64, []domain.FlagUserOverride) error
	FindByUser(context.Context, int64, int64, string) (domain.FlagUserOverride, error)
}

type FlagOverrideBulkSetter interface {
	BulkSetOverrides(context.Context, int64, string, []FlagUserOverrideInput) (int, error)
}

type FlagUserOverrideInput struct {
	UserID  string
	Enabled bool
}
