package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/settings"
)

type SettingsStore interface {
	GetSettings(ctx context.Context) (settings.Settings, error)
	UpdateSettings(ctx context.Context, input SettingsUpdateInput) (settings.Settings, error)
}

type SettingsService struct {
	store SettingsStore
}

func NewSettingsService(store SettingsStore) *SettingsService {
	if store == nil {
		panic("settings service requires a store")
	}
	return &SettingsService{store: store}
}

func (s *SettingsService) GetSettings(ctx context.Context) (settings.Settings, error) {
	return s.store.GetSettings(ctx)
}

func (s *SettingsService) UpdateSettings(ctx context.Context, input SettingsUpdateInput) (settings.Settings, error) {
	if input.UpdatedAt.IsZero() {
		return settings.Settings{}, domain.Invalid("updated_at", domain.ViolationRequired, "SET-004")
	}
	updated, err := s.store.UpdateSettings(ctx, SettingsUpdateInput{
		BusinessName:       input.BusinessName,
		Locale:             input.Locale,
		Timezone:           input.Timezone,
		Currency:           input.Currency,
		HourlyLaborCost:    input.HourlyLaborCost,
		DefaultGrossMargin: input.DefaultGrossMargin,
		ExpectedUpdatedAt:  input.ExpectedUpdatedAt,
		UpdatedAt:          input.UpdatedAt,
	})
	if err != nil {
		return settings.Settings{}, fmt.Errorf("update settings: %w", err)
	}
	return updated, nil
}
