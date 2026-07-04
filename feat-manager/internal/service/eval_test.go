package service

import (
	"context"
	"errors"
	"testing"

	"github.com/shivam/featfz/feat-manager/internal/domain"
)

func TestEvalServiceEvaluate(t *testing.T) {
	nowFlag := domain.Flag{
		ID:             11,
		TenantID:       7,
		Key:            "new_dashboard",
		DefaultEnabled: true,
	}

	tests := []struct {
		name        string
		flag        domain.Flag
		override    domain.FlagUserOverride
		overrideErr error
		want        bool
	}{
		{
			name:        "default on",
			flag:        nowFlag,
			overrideErr: ErrFlagOverrideNotFound,
			want:        true,
		},
		{
			name: "default off",
			flag: domain.Flag{
				ID:             11,
				TenantID:       7,
				Key:            "new_dashboard",
				DefaultEnabled: false,
			},
			overrideErr: ErrFlagOverrideNotFound,
			want:        false,
		},
		{
			name: "override true wins",
			flag: nowFlag,
			override: domain.FlagUserOverride{
				Enabled: true,
			},
			want: true,
		},
		{
			name: "override false wins",
			flag: nowFlag,
			override: domain.FlagUserOverride{
				Enabled: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := stubFlagRepository{
				findFn: func(_ context.Context, tenantID int64, key string) (domain.Flag, error) {
					if tenantID != 7 || key != "new_dashboard" {
						t.Fatalf("unexpected flag lookup: tenant=%d key=%q", tenantID, key)
					}

					return tt.flag, nil
				},
			}
			overrideRepo := stubFlagOverrideRepository{
				findFn: func(_ context.Context, tenantID int64, flagID int64, userID string) (domain.FlagUserOverride, error) {
					if tenantID != 7 || flagID != 11 || userID != "user_123" {
						t.Fatalf("unexpected override lookup: tenant=%d flag=%d user=%q", tenantID, flagID, userID)
					}

					if tt.overrideErr != nil {
						return domain.FlagUserOverride{}, tt.overrideErr
					}

					result := tt.override
					result.TenantID = tenantID
					result.FlagID = flagID
					result.UserID = userID
					return result, nil
				},
			}

			result, err := NewEvalService(repo, overrideRepo).Evaluate(context.Background(), 7, "new_dashboard", "user_123")
			if err != nil {
				t.Fatalf("evaluate flag: %v", err)
			}
			if result.Enabled != tt.want {
				t.Fatalf("expected enabled=%v, got %v", tt.want, result.Enabled)
			}
		})
	}
}

func TestEvalServiceEvaluateSurfacesErrors(t *testing.T) {
	repoErr := errors.New("repo down")

	service := NewEvalService(stubFlagRepository{
		findFn: func(context.Context, int64, string) (domain.Flag, error) {
			return domain.Flag{}, repoErr
		},
	}, stubFlagOverrideRepository{})

	if _, err := service.Evaluate(context.Background(), 7, "new_dashboard", "user_123"); err == nil || !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestEvalServiceRejectsBlankInput(t *testing.T) {
	service := NewEvalService(stubFlagRepository{}, stubFlagOverrideRepository{})

	if _, err := service.Evaluate(context.Background(), 7, "   ", "user_123"); !errors.Is(err, ErrInvalidEvalInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
	if _, err := service.Evaluate(context.Background(), 7, "new_dashboard", "   "); !errors.Is(err, ErrInvalidEvalInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}
