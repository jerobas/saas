package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/settings"
)

type SettingsStore interface {
	GetSettings(ctx context.Context) (settings.Settings, error)
	UpdateSettings(ctx context.Context, input settingsUpdateStoreInput) (settings.Settings, error)
}

type SettingsService struct {
	store SettingsStore
	clock Clock
}

type settingsUpdateStoreInput struct {
	SettingsUpdateInput
	UpdatedAt domain.UTCInstant
}

func NewSettingsService(store SettingsStore, clock Clock) *SettingsService {
	if store == nil {
		panic("settings service requires a store")
	}
	if clock == nil {
		panic("settings service requires a clock")
	}
	return &SettingsService{store: store, clock: clock}
}

func (s *SettingsService) GetSettings(ctx context.Context) (settings.Settings, error) {
	return s.store.GetSettings(ctx)
}

func (s *SettingsService) UpdateSettings(ctx context.Context, input SettingsUpdateInput) (settings.Settings, error) {
	updatedAt, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return settings.Settings{}, fmt.Errorf("read clock: %w", err)
	}
	updated, err := s.store.UpdateSettings(ctx, settingsUpdateStoreInput{
		SettingsUpdateInput: input,
		UpdatedAt:           updatedAt,
	})
	if err != nil {
		return settings.Settings{}, fmt.Errorf("update settings: %w", err)
	}
	if !updated.UpdatedAt().Equal(updatedAt) {
		return settings.Settings{}, domain.ErrInvariant
	}
	return updated, nil
}
