package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shivam/featfz/feat-manager/internal/domain"
)

type stubFlagRepository struct {
	createFn  func(context.Context, domain.Flag) (domain.Flag, error)
	findFn    func(context.Context, int64, string) (domain.Flag, error)
	listFn    func(context.Context, int64) ([]domain.Flag, error)
	updateFn  func(context.Context, domain.Flag) (domain.Flag, error)
	archiveFn func(context.Context, int64, string) error
}

type stubFlagOverrideRepository struct {
	bulkUpsertFn func(context.Context, int64, int64, []domain.FlagUserOverride) error
}

func (s stubFlagRepository) Create(ctx context.Context, flag domain.Flag) (domain.Flag, error) {
	if s.createFn != nil {
		return s.createFn(ctx, flag)
	}
	return flag, nil
}

func (s stubFlagRepository) FindByKey(ctx context.Context, tenantID int64, key string) (domain.Flag, error) {
	if s.findFn != nil {
		return s.findFn(ctx, tenantID, key)
	}
	return domain.Flag{TenantID: tenantID, Key: key}, nil
}

func (s stubFlagRepository) ListActive(ctx context.Context, tenantID int64) ([]domain.Flag, error) {
	if s.listFn != nil {
		return s.listFn(ctx, tenantID)
	}
	return []domain.Flag{}, nil
}

func (s stubFlagRepository) Update(ctx context.Context, flag domain.Flag) (domain.Flag, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, flag)
	}
	return flag, nil
}

func (s stubFlagRepository) Archive(ctx context.Context, tenantID int64, key string) error {
	if s.archiveFn != nil {
		return s.archiveFn(ctx, tenantID, key)
	}
	return nil
}

func (s stubFlagOverrideRepository) BulkUpsert(ctx context.Context, tenantID int64, flagID int64, overrides []domain.FlagUserOverride) error {
	if s.bulkUpsertFn != nil {
		return s.bulkUpsertFn(ctx, tenantID, flagID, overrides)
	}
	return nil
}

func (s stubFlagOverrideRepository) FindByUser(context.Context, int64, int64, string) (domain.FlagUserOverride, error) {
	return domain.FlagUserOverride{}, ErrFlagOverrideNotFound
}

func TestFlagServiceCRUD(t *testing.T) {
	now := time.Unix(1_720_000_000, 0).UTC()
	repo := stubFlagRepository{
		createFn: func(_ context.Context, flag domain.Flag) (domain.Flag, error) {
			flag.ID = 5
			flag.CreatedAt = now
			flag.UpdatedAt = now
			return flag, nil
		},
		findFn: func(_ context.Context, tenantID int64, key string) (domain.Flag, error) {
			return domain.Flag{
				ID:             5,
				TenantID:       tenantID,
				Key:            key,
				Description:    "Existing",
				DefaultEnabled: false,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil
		},
		listFn: func(_ context.Context, tenantID int64) ([]domain.Flag, error) {
			return []domain.Flag{{
				ID:             5,
				TenantID:       tenantID,
				Key:            "new_dashboard",
				DefaultEnabled: false,
				CreatedAt:      now,
				UpdatedAt:      now,
			}}, nil
		},
		updateFn: func(_ context.Context, flag domain.Flag) (domain.Flag, error) {
			flag.UpdatedAt = now.Add(time.Minute)
			return flag, nil
		},
	}

	service := NewFlagService(repo, stubFlagOverrideRepository{})

	created, err := service.Create(context.Background(), 7, CreateFlagInput{
		Key:            " new_dashboard ",
		Description:    " rollout ",
		DefaultEnabled: true,
	})
	if err != nil {
		t.Fatalf("create flag: %v", err)
	}
	if created.Key != "new_dashboard" || created.Description != "rollout" {
		t.Fatalf("unexpected create result: %+v", created)
	}

	listed, err := service.List(context.Background(), 7)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	if len(listed) != 1 || listed[0].Key != "new_dashboard" {
		t.Fatalf("unexpected list result: %+v", listed)
	}

	got, err := service.Get(context.Background(), 7, "new_dashboard")
	if err != nil {
		t.Fatalf("get flag: %v", err)
	}
	if got.Key != "new_dashboard" {
		t.Fatalf("unexpected get result: %+v", got)
	}

	description := "Updated rollout"
	enabled := true
	updated, err := service.Update(context.Background(), 7, "new_dashboard", UpdateFlagInput{
		Description:    &description,
		DefaultEnabled: &enabled,
	})
	if err != nil {
		t.Fatalf("update flag: %v", err)
	}
	if updated.Description != "Updated rollout" || !updated.DefaultEnabled {
		t.Fatalf("unexpected update result: %+v", updated)
	}

	if err := service.Archive(context.Background(), 7, "new_dashboard"); err != nil {
		t.Fatalf("archive flag: %v", err)
	}
}

func TestFlagServiceSurfacesRepositoryErrors(t *testing.T) {
	repoErr := errors.New("repo down")
	service := NewFlagService(stubFlagRepository{
		listFn: func(context.Context, int64) ([]domain.Flag, error) {
			return nil, repoErr
		},
		findFn: func(context.Context, int64, string) (domain.Flag, error) {
			return domain.Flag{}, repoErr
		},
		updateFn: func(context.Context, domain.Flag) (domain.Flag, error) {
			return domain.Flag{}, repoErr
		},
		archiveFn: func(context.Context, int64, string) error {
			return repoErr
		},
	}, stubFlagOverrideRepository{})

	if _, err := service.List(context.Background(), 7); err == nil || !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped list error, got %v", err)
	}
	if _, err := service.Get(context.Background(), 7, "new_dashboard"); err == nil || !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped get error, got %v", err)
	}
	if _, err := service.Update(context.Background(), 7, "new_dashboard", UpdateFlagInput{}); err == nil || !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped update error, got %v", err)
	}
	if err := service.Archive(context.Background(), 7, "new_dashboard"); err == nil || !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped archive error, got %v", err)
	}
}

func TestFlagServiceBulkSetOverrides(t *testing.T) {
	seen := make([]domain.FlagUserOverride, 0)
	repo := stubFlagRepository{
		findFn: func(_ context.Context, tenantID int64, key string) (domain.Flag, error) {
			if tenantID != 7 || key != "new_dashboard" {
				t.Fatalf("unexpected flag lookup: tenant=%d key=%q", tenantID, key)
			}

			return domain.Flag{
				ID:       42,
				TenantID: tenantID,
				Key:      key,
			}, nil
		},
	}
	overrideRepo := stubFlagOverrideRepository{
		bulkUpsertFn: func(_ context.Context, tenantID int64, flagID int64, overrides []domain.FlagUserOverride) error {
			if tenantID != 7 || flagID != 42 {
				t.Fatalf("unexpected ownership: tenant=%d flag=%d", tenantID, flagID)
			}

			seen = append(seen, overrides...)
			return nil
		},
	}

	service := NewFlagService(repo, overrideRepo)
	enabled := true
	disabled := false

	applied, err := service.BulkSetOverrides(context.Background(), 7, "new_dashboard", []FlagUserOverrideInput{
		{UserID: " user_123 ", Enabled: enabled},
		{UserID: "user_456", Enabled: disabled},
		{UserID: "user_123", Enabled: disabled},
	})
	if err != nil {
		t.Fatalf("bulk set overrides: %v", err)
	}
	if applied != 2 {
		t.Fatalf("expected 2 applied overrides, got %d", applied)
	}

	if len(seen) != 2 {
		t.Fatalf("expected 2 deduped overrides, got %d", len(seen))
	}

	got := map[string]bool{}
	for _, override := range seen {
		got[override.UserID] = override.Enabled
	}

	if got["user_123"] {
		t.Fatal("expected last user_123 value to win and be false")
	}
	if got["user_456"] {
		t.Fatal("expected user_456 to remain false")
	}
}

func TestFlagServiceBulkSetOverridesRejectsInvalidUser(t *testing.T) {
	service := NewFlagService(stubFlagRepository{
		findFn: func(context.Context, int64, string) (domain.Flag, error) {
			return domain.Flag{ID: 42, TenantID: 7, Key: "new_dashboard"}, nil
		},
	}, stubFlagOverrideRepository{
		bulkUpsertFn: func(context.Context, int64, int64, []domain.FlagUserOverride) error {
			t.Fatal("expected bulk upsert not to be called")
			return nil
		},
	})

	applied, err := service.BulkSetOverrides(context.Background(), 7, "new_dashboard", []FlagUserOverrideInput{{UserID: "   ", Enabled: true}})
	if err == nil {
		t.Fatal("expected invalid user error")
	}
	if applied != 0 {
		t.Fatalf("expected no applied overrides on error, got %d", applied)
	}
}
