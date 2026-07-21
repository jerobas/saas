package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain/settings"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteSettingsStore struct {
	store *sqlite.Store
}

func NewSQLiteSettingsStore(store *sqlite.Store) SettingsStore {
	if store == nil {
		panic("sqlite settings store requires a store")
	}
	return &sqliteSettingsStore{store: store}
}

func (s *sqliteSettingsStore) GetSettings(ctx context.Context) (settings.Settings, error) {
	return s.store.GetSettings(ctx)
}

func (s *sqliteSettingsStore) UpdateSettings(ctx context.Context, input settingsUpdateStoreInput) (settings.Settings, error) {
	return s.store.UpdateSettings(ctx, sqlite.UpdateSettingsInput{
		BusinessName:       input.BusinessName,
		Locale:             input.Locale,
		Timezone:           input.Timezone,
		Currency:           input.Currency,
		HourlyLaborCost:    input.HourlyLaborCost,
		DefaultGrossMargin: input.DefaultGrossMargin,
		ExpectedUpdatedAt:  input.ExpectedUpdatedAt,
		UpdatedAt:          input.UpdatedAt,
	})
}
