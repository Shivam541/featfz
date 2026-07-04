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

type FlagCreator interface {
	Create(context.Context, int64, CreateFlagInput) (domain.Flag, error)
}

type FlagCreationService struct {
	Flags FlagRepository
}

func NewFlagCreationService(repo FlagRepository) FlagCreationService {
	return FlagCreationService{Flags: repo}
}

func (s FlagCreationService) Create(ctx context.Context, tenantID int64, input CreateFlagInput) (domain.Flag, error) {
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
