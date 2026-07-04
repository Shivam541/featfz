package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/entity"
	"github.com/shivam/featfz/feat-manager/internal/service"
	"gorm.io/gorm"
)

type TenantAppRepository struct {
	db *sql.DB
}

func NewTenantAppRepository(db *sql.DB) *TenantAppRepository {
	return &TenantAppRepository{db: db}
}

func (r *TenantAppRepository) FindByAppID(ctx context.Context, appID string) (domain.TenantApp, error) {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return domain.TenantApp{}, service.ErrTenantAppNotFound
	}

	gormDB, err := openGormDB(r.db)
	if err != nil {
		return domain.TenantApp{}, fmt.Errorf("open gorm db: %w", err)
	}

	var tenant entity.Tenant
	err = gormDB.WithContext(ctx).Where("app_id = ?", appID).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.TenantApp{}, service.ErrTenantAppNotFound
		}

		return domain.TenantApp{}, fmt.Errorf("query tenant app by app id: %w", err)
	}

	return domain.TenantApp{
		TenantID:   tenant.ID,
		TenantName: tenant.Name,
		AppID:      tenant.AppID,
		JWTSecret:  tenant.JWTSecret,
	}, nil
}
