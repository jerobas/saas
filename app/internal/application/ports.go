package application

import (
	"github.com/jerobas/saas/internal/domain"
)

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
