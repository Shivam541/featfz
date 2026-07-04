package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shivam/featfz/feat-manager/internal/domain"
)

var ErrFlagAlreadyExists = errors.New("flag already exists")

type CreateFlagInput struct {
	Key            string
	Description    string
	DefaultEnabled bool
}

type UpdateFlagInput struct {
	Description    *string
	DefaultEnabled *bool
}

type FlagService struct {
	Flags     FlagRepository
	Overrides FlagOverrideRepository
}

func NewFlagService(repo FlagRepository, overrideRepo FlagOverrideRepository) FlagService {
	return FlagService{Flags: repo, Overrides: overrideRepo}
}

func (s FlagService) Create(ctx context.Context, tenantID int64, input CreateFlagInput) (domain.Flag, error) {
	input.Key = strings.TrimSpace(input.Key)
	input.Description = strings.TrimSpace(input.Description)

	flag, err := s.Flags.Create(ctx, domain.Flag{
		TenantID:       tenantID,
		Key:            input.Key,
		Description:    input.Description,
		DefaultEnabled: input.DefaultEnabled,
	})
	if err != nil {
		return domain.Flag{}, fmt.Errorf("create flag: %w", err)
	}

	return flag, nil
}

func (s FlagService) List(ctx context.Context, tenantID int64) ([]domain.Flag, error) {
	flags, err := s.Flags.ListActive(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list flags: %w", err)
	}

	return flags, nil
}

func (s FlagService) Get(ctx context.Context, tenantID int64, key string) (domain.Flag, error) {
	flag, err := s.Flags.FindByKey(ctx, tenantID, key)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("get flag: %w", err)
	}

	return flag, nil
}

func (s FlagService) Update(ctx context.Context, tenantID int64, key string, input UpdateFlagInput) (domain.Flag, error) {
	flag, err := s.Flags.FindByKey(ctx, tenantID, key)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("get flag before update: %w", err)
	}

	if input.Description != nil {
		flag.Description = strings.TrimSpace(*input.Description)
	}
	if input.DefaultEnabled != nil {
		flag.DefaultEnabled = *input.DefaultEnabled
	}

	updated, err := s.Flags.Update(ctx, flag)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("update flag: %w", err)
	}

	return updated, nil
}

func (s FlagService) Archive(ctx context.Context, tenantID int64, key string) error {
	if err := s.Flags.Archive(ctx, tenantID, key); err != nil {
		return fmt.Errorf("archive flag: %w", err)
	}

	return nil
}

func (s FlagService) BulkSetOverrides(ctx context.Context, tenantID int64, key string, overrides []FlagUserOverrideInput) (int, error) {
	flag, err := s.Flags.FindByKey(ctx, tenantID, key)
	if err != nil {
		return 0, fmt.Errorf("get flag before bulk set overrides: %w", err)
	}

	if len(overrides) == 0 {
		return 0, nil
	}

	deduped := make(map[string]domain.FlagUserOverride, len(overrides))
	for _, override := range overrides {
		userID := strings.TrimSpace(override.UserID)
		if userID == "" {
			return 0, fmt.Errorf("bulk set flag overrides: user id is required")
		}

		deduped[userID] = domain.FlagUserOverride{
			UserID:  userID,
			Enabled: override.Enabled,
		}
	}

	records := make([]domain.FlagUserOverride, 0, len(deduped))
	for _, override := range deduped {
		records = append(records, override)
	}

	if err := s.Overrides.BulkUpsert(ctx, tenantID, flag.ID, records); err != nil {
		return 0, fmt.Errorf("bulk set flag overrides: %w", err)
	}

	return len(records), nil
}
