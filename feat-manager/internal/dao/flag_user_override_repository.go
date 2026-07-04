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

type FlagOverrideRepository struct {
	db *sql.DB
}

func NewFlagOverrideRepository(db *sql.DB) *FlagOverrideRepository {
	return &FlagOverrideRepository{db: db}
}

func (r *FlagOverrideRepository) BulkUpsert(ctx context.Context, tenantID int64, flagID int64, overrides []domain.FlagUserOverride) error {
	if len(overrides) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin bulk override upsert: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const query = `
INSERT INTO flag_user_overrides (tenant_id, flag_id, user_id, enabled)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  enabled = VALUES(enabled),
  updated_at = CURRENT_TIMESTAMP
`

	for _, override := range overrides {
		userID := strings.TrimSpace(override.UserID)
		if userID == "" {
			return fmt.Errorf("bulk upsert flag overrides: user id is required")
		}

		if _, err := tx.ExecContext(ctx, query, tenantID, flagID, userID, override.Enabled); err != nil {
			return fmt.Errorf("upsert flag override for user %q: %w", userID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit bulk override upsert: %w", err)
	}

	return nil
}

func (r *FlagOverrideRepository) FindByUser(ctx context.Context, tenantID int64, flagID int64, userID string) (domain.FlagUserOverride, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.FlagUserOverride{}, service.ErrFlagOverrideNotFound
	}

	const query = `
SELECT id, tenant_id, flag_id, user_id, enabled, created_at, updated_at
FROM flag_user_overrides
WHERE tenant_id = ? AND flag_id = ? AND user_id = ?
LIMIT 1
`

	var override domain.FlagUserOverride
	err := r.db.QueryRowContext(ctx, query, tenantID, flagID, userID).Scan(
		&override.ID,
		&override.TenantID,
		&override.FlagID,
		&override.UserID,
		&override.Enabled,
		&override.CreatedAt,
		&override.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FlagUserOverride{}, service.ErrFlagOverrideNotFound
		}

		return domain.FlagUserOverride{}, fmt.Errorf("find flag override by user: %w", err)
	}

	override.CreatedAt = override.CreatedAt.UTC()
	override.UpdatedAt = override.UpdatedAt.UTC()

	return override, nil
}
