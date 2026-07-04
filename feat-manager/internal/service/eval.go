package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidEvalInput = errors.New("invalid evaluation input")

type EvalService struct {
	Flags     FlagRepository
	Overrides FlagOverrideRepository
}

func NewEvalService(repo FlagRepository, overrideRepo FlagOverrideRepository) EvalService {
	return EvalService{Flags: repo, Overrides: overrideRepo}
}

func (s EvalService) Evaluate(ctx context.Context, tenantID int64, flagKey, userID string) (EvalResult, error) {
	flagKey = strings.TrimSpace(flagKey)
	userID = strings.TrimSpace(userID)
	if flagKey == "" || userID == "" {
		return EvalResult{}, ErrInvalidEvalInput
	}

	flag, err := s.Flags.FindByKey(ctx, tenantID, flagKey)
	if err != nil {
		return EvalResult{}, fmt.Errorf("find flag for evaluation: %w", err)
	}

	override, err := s.Overrides.FindByUser(ctx, tenantID, flag.ID, userID)
	switch {
	case err == nil:
		return EvalResult{Enabled: override.Enabled}, nil
	case errors.Is(err, ErrFlagOverrideNotFound):
		return EvalResult{Enabled: flag.DefaultEnabled}, nil
	default:
		return EvalResult{}, fmt.Errorf("find override for evaluation: %w", err)
	}
}

var _ Evaluator = EvalService{}
