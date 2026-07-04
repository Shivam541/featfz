package dao

import (
	"context"
	"errors"
	"testing"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

func TestFlagOverrideRepository(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()

	tenantOneID := insertIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	tenantTwoID := insertIntegrationTenant(t, db, "globex", "app-globex", "globex-secret")

	flagRepo := NewFlagRepository(db)
	flagOne, err := flagRepo.Create(ctx, domain.Flag{
		TenantID:       tenantOneID,
		Key:            "beta_access",
		Description:    "beta users",
		DefaultEnabled: false,
	})
	if err != nil {
		t.Fatalf("create flag: %v", err)
	}

	_, err = flagRepo.Create(ctx, domain.Flag{
		TenantID:       tenantTwoID,
		Key:            "beta_access",
		Description:    "beta users in tenant two",
		DefaultEnabled: true,
	})
	if err != nil {
		t.Fatalf("create second tenant flag: %v", err)
	}

	repo := NewFlagOverrideRepository(db)

	if err := repo.BulkUpsert(ctx, tenantOneID, flagOne.ID, []domain.FlagUserOverride{
		{UserID: "user_123", Enabled: true},
		{UserID: " user_456 ", Enabled: false},
	}); err != nil {
		t.Fatalf("bulk upsert overrides: %v", err)
	}

	got, err := repo.FindByUser(ctx, tenantOneID, flagOne.ID, "user_123")
	if err != nil {
		t.Fatalf("find override: %v", err)
	}
	if !got.Enabled {
		t.Fatal("expected enabled override")
	}
	if got.UserID != "user_123" {
		t.Fatalf("expected trimmed user id, got %q", got.UserID)
	}
	if got.TenantID != tenantOneID || got.FlagID != flagOne.ID {
		t.Fatalf("unexpected override ownership: %+v", got)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatal("expected override timestamps")
	}

	if err := repo.BulkUpsert(ctx, tenantOneID, flagOne.ID, []domain.FlagUserOverride{
		{UserID: "user_123", Enabled: false},
	}); err != nil {
		t.Fatalf("update override: %v", err)
	}

	got, err = repo.FindByUser(ctx, tenantOneID, flagOne.ID, "user_123")
	if err != nil {
		t.Fatalf("find updated override: %v", err)
	}
	if got.Enabled {
		t.Fatal("expected updated override to be disabled")
	}

	got, err = repo.FindByUser(ctx, tenantOneID, flagOne.ID, "user_456")
	if err != nil {
		t.Fatalf("find second override: %v", err)
	}
	if got.Enabled {
		t.Fatal("expected user_456 override to be disabled")
	}

	if _, err := repo.FindByUser(ctx, tenantTwoID, flagOne.ID, "user_123"); !errors.Is(err, service.ErrFlagOverrideNotFound) {
		t.Fatalf("expected other tenant to not see override, got %v", err)
	}

	if err := repo.BulkUpsert(ctx, tenantOneID, flagOne.ID, nil); err != nil {
		t.Fatalf("empty bulk upsert should be a no-op, got %v", err)
	}
}

func TestFlagOverrideRepositoryRejectsMissingUserID(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()

	tenantID := insertIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	flagRepo := NewFlagRepository(db)
	flag, err := flagRepo.Create(ctx, domain.Flag{
		TenantID:       tenantID,
		Key:            "beta_access",
		DefaultEnabled: false,
	})
	if err != nil {
		t.Fatalf("create flag: %v", err)
	}

	repo := NewFlagOverrideRepository(db)
	if err := repo.BulkUpsert(ctx, tenantID, flag.ID, []domain.FlagUserOverride{{UserID: "   ", Enabled: true}}); err == nil {
		t.Fatal("expected missing user id error")
	}
}
