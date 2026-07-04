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

type FlagOverrideRepository interface {
	BulkUpsert(context.Context, int64, int64, []domain.FlagUserOverride) error
	FindByUser(context.Context, int64, int64, string) (domain.FlagUserOverride, error)
}
