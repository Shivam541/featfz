package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/entity"
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

	gormDB, err := openGormDB(r.db)
	if err != nil {
		return fmt.Errorf("open gorm db: %w", err)
	}

	for _, override := range overrides {
		userID := strings.TrimSpace(override.UserID)
		if userID == "" {
			return fmt.Errorf("bulk upsert flag overrides: user id is required")
		}

		record := entity.FlagUserOverride{
			TenantID:  tenantID,
			FlagID:    flagID,
			UserID:    userID,
			Enabled:   override.Enabled,
			UpdatedAt: time.Now().UTC(),
		}

		result := gormDB.WithContext(ctx).
			Model(&entity.FlagUserOverride{}).
			Where("tenant_id = ? AND flag_id = ? AND user_id = ?", tenantID, flagID, userID).
			Updates(map[string]any{
				"enabled":    record.Enabled,
				"updated_at": record.UpdatedAt,
			})
		if result.Error != nil {
			return fmt.Errorf("upsert flag override for user %q: %w", userID, result.Error)
		}

		if result.RowsAffected > 0 {
			continue
		}

		record.CreatedAt = record.UpdatedAt
		if err := gormDB.WithContext(ctx).Create(&record).Error; err != nil {
			return fmt.Errorf("insert flag override for user %q: %w", userID, err)
		}
	}

	return nil
}

func (r *FlagOverrideRepository) FindByUser(ctx context.Context, tenantID int64, flagID int64, userID string) (domain.FlagUserOverride, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.FlagUserOverride{}, service.ErrFlagOverrideNotFound
	}

	gormDB, err := openGormDB(r.db)
	if err != nil {
		return domain.FlagUserOverride{}, fmt.Errorf("open gorm db: %w", err)
	}

	var record entity.FlagUserOverride
	err = gormDB.WithContext(ctx).
		Where("tenant_id = ? AND flag_id = ? AND user_id = ?", tenantID, flagID, userID).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.FlagUserOverride{}, service.ErrFlagOverrideNotFound
		}

		return domain.FlagUserOverride{}, fmt.Errorf("find flag override by user: %w", err)
	}

	return domain.FlagUserOverride{
		ID:        record.ID,
		TenantID:  record.TenantID,
		FlagID:    record.FlagID,
		UserID:    record.UserID,
		Enabled:   record.Enabled,
		CreatedAt: record.CreatedAt.UTC(),
		UpdatedAt: record.UpdatedAt.UTC(),
	}, nil
}
