package dao

import (
	"context"
	"errors"
	"testing"

	"github.com/shivam/featfz/feat-manager/internal/domain"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

func TestFlagRepository(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()

	tenantOneID := insertIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	tenantTwoID := insertIntegrationTenant(t, db, "globex", "app-globex", "globex-secret")

	repo := NewFlagRepository(db)

	t.Run("create list update archive with tenant scoping", func(t *testing.T) {
		created, err := repo.Create(ctx, domain.Flag{
			TenantID:       tenantOneID,
			Key:            "new_dashboard",
			Description:    "  New dashboard rollout  ",
			DefaultEnabled: false,
		})
		if err != nil {
			t.Fatalf("create flag: %v", err)
		}

		if created.ID == 0 {
			t.Fatal("expected generated flag id")
		}
		if created.TenantID != tenantOneID {
			t.Fatalf("expected tenant id %d, got %d", tenantOneID, created.TenantID)
		}
		if created.Key != "new_dashboard" {
			t.Fatalf("expected key new_dashboard, got %q", created.Key)
		}
		if created.Description != "New dashboard rollout" {
			t.Fatalf("expected trimmed description, got %q", created.Description)
		}
		if created.DefaultEnabled {
			t.Fatal("expected default_enabled=false")
		}
		if created.ArchivedAt != nil {
			t.Fatal("expected active flag to have nil archived_at")
		}
		if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
			t.Fatal("expected timestamps to be populated")
		}

		got, err := repo.FindByKey(ctx, tenantOneID, "new_dashboard")
		if err != nil {
			t.Fatalf("find flag: %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("expected flag id %d, got %d", created.ID, got.ID)
		}

		listed, err := repo.ListActive(ctx, tenantOneID)
		if err != nil {
			t.Fatalf("list active flags: %v", err)
		}
		if len(listed) != 1 {
			t.Fatalf("expected 1 active flag, got %d", len(listed))
		}

		updated, err := repo.Update(ctx, domain.Flag{
			TenantID:       tenantOneID,
			Key:            "new_dashboard",
			Description:    "Updated rollout",
			DefaultEnabled: true,
		})
		if err != nil {
			t.Fatalf("update flag: %v", err)
		}
		if !updated.DefaultEnabled {
			t.Fatal("expected default_enabled=true after update")
		}
		if updated.Description != "Updated rollout" {
			t.Fatalf("expected updated description, got %q", updated.Description)
		}
		if updated.UpdatedAt.Before(updated.CreatedAt) {
			t.Fatal("expected updated_at to be on or after created_at")
		}

		if err := repo.Archive(ctx, tenantOneID, "new_dashboard"); err != nil {
			t.Fatalf("archive flag: %v", err)
		}

		if _, err := repo.FindByKey(ctx, tenantOneID, "new_dashboard"); !errors.Is(err, service.ErrFlagNotFound) {
			t.Fatalf("expected archived flag to be hidden, got %v", err)
		}

		listed, err = repo.ListActive(ctx, tenantOneID)
		if err != nil {
			t.Fatalf("list active after archive: %v", err)
		}
		if len(listed) != 0 {
			t.Fatalf("expected no active flags after archive, got %d", len(listed))
		}

		if _, err := repo.Update(ctx, domain.Flag{
			TenantID:       tenantOneID,
			Key:            "new_dashboard",
			Description:    "Should not work",
			DefaultEnabled: false,
		}); !errors.Is(err, service.ErrFlagNotFound) {
			t.Fatalf("expected archived flag update to fail with not found, got %v", err)
		}

		if _, err := repo.FindByKey(ctx, tenantTwoID, "new_dashboard"); !errors.Is(err, service.ErrFlagNotFound) {
			t.Fatalf("expected other tenant to not see flag, got %v", err)
		}
	})

	t.Run("enforces unique tenant key constraint", func(t *testing.T) {
		_, err := repo.Create(ctx, domain.Flag{
			TenantID:       tenantOneID,
			Key:            "duplicate_key",
			Description:    "first",
			DefaultEnabled: true,
		})
		if err != nil {
			t.Fatalf("create first flag: %v", err)
		}

		_, err = repo.Create(ctx, domain.Flag{
			TenantID:       tenantOneID,
			Key:            "duplicate_key",
			Description:    "second",
			DefaultEnabled: false,
		})
		if !errors.Is(err, service.ErrFlagAlreadyExists) {
			t.Fatalf("expected duplicate key error, got %v", err)
		}

		otherTenantFlag, err := repo.Create(ctx, domain.Flag{
			TenantID:       tenantTwoID,
			Key:            "duplicate_key",
			Description:    "allowed in other tenant",
			DefaultEnabled: true,
		})
		if err != nil {
			t.Fatalf("create same key in other tenant: %v", err)
		}
		if otherTenantFlag.TenantID != tenantTwoID {
			t.Fatalf("expected tenant id %d, got %d", tenantTwoID, otherTenantFlag.TenantID)
		}
	})

	t.Run("returns not found for missing active flag", func(t *testing.T) {
		if _, err := repo.FindByKey(ctx, tenantOneID, "missing"); !errors.Is(err, service.ErrFlagNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
	})
}

func TestFlagRepositoryRejectsMissingKey(t *testing.T) {
	db := openIntegrationDB(t)
	repo := NewFlagRepository(db)

	if _, err := repo.Create(context.Background(), domain.Flag{TenantID: insertIntegrationTenant(t, db, "acme", "app-acme", "acme-secret"), DefaultEnabled: true}); err == nil {
		t.Fatal("expected missing key error")
	}
}

func TestFlagRepositoryDuplicateErrorType(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()

	tenantID := insertIntegrationTenant(t, db, "acme", "app-acme", "acme-secret")
	repo := NewFlagRepository(db)

	if _, err := repo.Create(ctx, domain.Flag{TenantID: tenantID, Key: "dup", DefaultEnabled: true}); err != nil {
		t.Fatalf("create flag: %v", err)
	}

	_, err := repo.Create(ctx, domain.Flag{TenantID: tenantID, Key: "dup", DefaultEnabled: false})
	if !errors.Is(err, service.ErrFlagAlreadyExists) {
		t.Fatalf("expected duplicate key error, got %v", err)
	}
}
