package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/service"
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

	const query = `
SELECT id, name, app_id, jwt_secret
FROM tenants
WHERE app_id = ?
LIMIT 1
`

	var tenantApp domain.TenantApp
	err := r.db.QueryRowContext(ctx, query, appID).Scan(
		&tenantApp.TenantID,
		&tenantApp.TenantName,
		&tenantApp.AppID,
		&tenantApp.JWTSecret,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TenantApp{}, service.ErrTenantAppNotFound
		}

		return domain.TenantApp{}, fmt.Errorf("query tenant app by app id: %w", err)
	}

	return tenantApp, nil
}
