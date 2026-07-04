package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/entity"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

type FlagRepository struct {
	db *sql.DB
}

func NewFlagRepository(db *sql.DB) *FlagRepository {
	return &FlagRepository{db: db}
}

func (r *FlagRepository) Create(ctx context.Context, flag domain.Flag) (domain.Flag, error) {
	flag.Key = strings.TrimSpace(flag.Key)
	if flag.Key == "" {
		return domain.Flag{}, fmt.Errorf("create flag: key is required")
	}

	description := strings.TrimSpace(flag.Description)
	gormDB, err := openGormDB(r.db)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("open gorm db: %w", err)
	}

	record := entity.Flag{
		TenantID:       flag.TenantID,
		Key:            flag.Key,
		Description:    description,
		DefaultEnabled: flag.DefaultEnabled,
	}
	if err := gormDB.WithContext(ctx).Create(&record).Error; err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return domain.Flag{}, service.ErrFlagAlreadyExists
		}

		return domain.Flag{}, fmt.Errorf("insert flag: %w", err)
	}

	return toDomainFlag(record), nil
}

func (r *FlagRepository) FindByKey(ctx context.Context, tenantID int64, key string) (domain.Flag, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return domain.Flag{}, service.ErrFlagNotFound
	}

	gormDB, err := openGormDB(r.db)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("open gorm db: %w", err)
	}

	var record entity.Flag
	err = gormDB.WithContext(ctx).
		Where("tenant_id = ? AND `key` = ? AND archived_at IS NULL", tenantID, key).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Flag{}, service.ErrFlagNotFound
		}

		return domain.Flag{}, fmt.Errorf("find flag by key: %w", err)
	}

	return toDomainFlag(record), nil
}

func (r *FlagRepository) ListActive(ctx context.Context, tenantID int64) ([]domain.Flag, error) {
	gormDB, err := openGormDB(r.db)
	if err != nil {
		return nil, fmt.Errorf("open gorm db: %w", err)
	}

	var records []entity.Flag
	if err := gormDB.WithContext(ctx).
		Where("tenant_id = ? AND archived_at IS NULL", tenantID).
		Order("id ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list active flags: %w", err)
	}

	flags := make([]domain.Flag, 0, len(records))
	for _, record := range records {
		flags = append(flags, toDomainFlag(record))
	}

	return flags, nil
}

func (r *FlagRepository) Update(ctx context.Context, flag domain.Flag) (domain.Flag, error) {
	flag.Key = strings.TrimSpace(flag.Key)
	if flag.Key == "" {
		return domain.Flag{}, service.ErrFlagNotFound
	}

	description := strings.TrimSpace(flag.Description)
	gormDB, err := openGormDB(r.db)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("open gorm db: %w", err)
	}

	result := gormDB.WithContext(ctx).
		Model(&entity.Flag{}).
		Where("tenant_id = ? AND `key` = ? AND archived_at IS NULL", flag.TenantID, flag.Key).
		Updates(map[string]any{
			"description":     description,
			"default_enabled": flag.DefaultEnabled,
			"updated_at":      time.Now().UTC(),
		})
	if result.Error != nil {
		return domain.Flag{}, fmt.Errorf("update flag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.Flag{}, service.ErrFlagNotFound
	}

	return r.FindByKey(ctx, flag.TenantID, flag.Key)
}

func (r *FlagRepository) Archive(ctx context.Context, tenantID int64, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return service.ErrFlagNotFound
	}

	gormDB, err := openGormDB(r.db)
	if err != nil {
		return fmt.Errorf("open gorm db: %w", err)
	}

	now := time.Now().UTC()
	result := gormDB.WithContext(ctx).
		Model(&entity.Flag{}).
		Where("tenant_id = ? AND `key` = ? AND archived_at IS NULL", tenantID, key).
		Updates(map[string]any{
			"archived_at": now,
			"updated_at":  now,
		})
	if result.Error != nil {
		return fmt.Errorf("archive flag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return service.ErrFlagNotFound
	}

	return nil
}

func toDomainFlag(record entity.Flag) domain.Flag {
	flag := domain.Flag{
		ID:             record.ID,
		TenantID:       record.TenantID,
		Key:            record.Key,
		Description:    record.Description,
		DefaultEnabled: record.DefaultEnabled,
		CreatedAt:      record.CreatedAt.UTC(),
		UpdatedAt:      record.UpdatedAt.UTC(),
	}

	if record.ArchivedAt != nil {
		archivedAt := record.ArchivedAt.UTC()
		flag.ArchivedAt = &archivedAt
	}

	return flag
}
