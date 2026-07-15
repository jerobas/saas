package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	domainsettings "github.com/jerobas/saas/internal/domain/settings"
)

type SettingsRepository interface {
	GetSettings(ctx context.Context) (domainsettings.Settings, error)
	UpdateSettings(ctx context.Context, input interface{ /* no-op interface for composition */ }) (domainsettings.Settings, error)
}

type CatalogRepository interface {
	ListMeasurementUnits(ctx context.Context) ([]catalog.MeasurementUnit, error)
	GetSettings(ctx context.Context) (domainsettings.Settings, error)
}

type SettingsUpdateInput struct {
	BusinessName       domain.DisplayName
	Locale             domain.Locale
	Timezone           domain.BusinessTimezone
	Currency           domain.Currency
	HourlyLaborCost    domain.Option[domain.MinorAmount]
	DefaultGrossMargin domain.Option[domain.BasisPoints]
	ExpectedUpdatedAt  domain.UTCInstant
	UpdatedAt          domain.UTCInstant
}
