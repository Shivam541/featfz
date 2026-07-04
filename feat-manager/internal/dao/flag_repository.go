package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/shivam/featfz/feat-manager/internal/domain"
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

	var description any
	if trimmed := strings.TrimSpace(flag.Description); trimmed != "" {
		description = trimmed
	}

	const query = `
INSERT INTO flags (tenant_id, ` + "`key`" + `, description, default_enabled)
VALUES (?, ?, ?, ?)
`

	result, err := r.db.ExecContext(ctx, query, flag.TenantID, flag.Key, description, flag.DefaultEnabled)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return domain.Flag{}, service.ErrFlagAlreadyExists
		}

		return domain.Flag{}, fmt.Errorf("insert flag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Flag{}, fmt.Errorf("last insert flag id: %w", err)
	}

	return r.findByID(ctx, id)
}

func (r *FlagRepository) FindByKey(ctx context.Context, tenantID int64, key string) (domain.Flag, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return domain.Flag{}, service.ErrFlagNotFound
	}

	const query = `
SELECT id, tenant_id, ` + "`key`" + `, description, default_enabled, archived_at, created_at, updated_at
FROM flags
WHERE tenant_id = ? AND ` + "`key`" + ` = ? AND archived_at IS NULL
LIMIT 1
`

	return r.scanFlag(r.db.QueryRowContext(ctx, query, tenantID, key))
}

func (r *FlagRepository) ListActive(ctx context.Context, tenantID int64) ([]domain.Flag, error) {
	const query = `
SELECT id, tenant_id, ` + "`key`" + `, description, default_enabled, archived_at, created_at, updated_at
FROM flags
WHERE tenant_id = ? AND archived_at IS NULL
ORDER BY id ASC
`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list active flags: %w", err)
	}
	defer rows.Close()

	flags := make([]domain.Flag, 0)
	for rows.Next() {
		flag, err := scanFlagRow(rows)
		if err != nil {
			return nil, err
		}
		flags = append(flags, flag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list active flags: %w", err)
	}

	return flags, nil
}

func (r *FlagRepository) Update(ctx context.Context, flag domain.Flag) (domain.Flag, error) {
	flag.Key = strings.TrimSpace(flag.Key)
	if flag.Key == "" {
		return domain.Flag{}, service.ErrFlagNotFound
	}

	var description any
	if trimmed := strings.TrimSpace(flag.Description); trimmed != "" {
		description = trimmed
	}

	const query = `
UPDATE flags
SET description = ?, default_enabled = ?
WHERE tenant_id = ? AND ` + "`key`" + ` = ? AND archived_at IS NULL
`

	result, err := r.db.ExecContext(ctx, query, description, flag.DefaultEnabled, flag.TenantID, flag.Key)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("update flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Flag{}, fmt.Errorf("rows affected updating flag: %w", err)
	}
	if rowsAffected == 0 {
		return domain.Flag{}, service.ErrFlagNotFound
	}

	return r.FindByKey(ctx, flag.TenantID, flag.Key)
}

func (r *FlagRepository) Archive(ctx context.Context, tenantID int64, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return service.ErrFlagNotFound
	}

	const query = `
UPDATE flags
SET archived_at = CURRENT_TIMESTAMP
WHERE tenant_id = ? AND ` + "`key`" + ` = ? AND archived_at IS NULL
`

	result, err := r.db.ExecContext(ctx, query, tenantID, key)
	if err != nil {
		return fmt.Errorf("archive flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected archiving flag: %w", err)
	}
	if rowsAffected == 0 {
		return service.ErrFlagNotFound
	}

	return nil
}

func (r *FlagRepository) findByID(ctx context.Context, id int64) (domain.Flag, error) {
	const query = `
SELECT id, tenant_id, ` + "`key`" + `, description, default_enabled, archived_at, created_at, updated_at
FROM flags
WHERE id = ?
LIMIT 1
`

	return r.scanFlag(r.db.QueryRowContext(ctx, query, id))
}

func (r *FlagRepository) scanFlag(row rowScanner) (domain.Flag, error) {
	flag, err := scanFlagRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Flag{}, service.ErrFlagNotFound
		}

		return domain.Flag{}, err
	}

	return flag, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFlagRow(row rowScanner) (domain.Flag, error) {
	var flag domain.Flag
	var description sql.NullString
	var archivedAt sql.NullTime

	err := row.Scan(
		&flag.ID,
		&flag.TenantID,
		&flag.Key,
		&description,
		&flag.DefaultEnabled,
		&archivedAt,
		&flag.CreatedAt,
		&flag.UpdatedAt,
	)
	if err != nil {
		return domain.Flag{}, err
	}

	if description.Valid {
		flag.Description = description.String
	}
	if archivedAt.Valid {
		archived := archivedAt.Time.UTC()
		flag.ArchivedAt = &archived
	}

	flag.CreatedAt = flag.CreatedAt.UTC()
	flag.UpdatedAt = flag.UpdatedAt.UTC()

	return flag, nil
}
