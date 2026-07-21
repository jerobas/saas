package settings

import "github.com/jerobas/saas/internal/domain"

type Params struct {
	BusinessName       domain.DisplayName
	Locale             domain.Locale
	Timezone           domain.BusinessTimezone
	Currency           domain.Currency
	HourlyLaborCost    domain.Option[domain.MinorAmount]
	DefaultGrossMargin domain.Option[domain.BasisPoints]
	CreatedAt          domain.UTCInstant
	UpdatedAt          domain.UTCInstant
}

// Settings is the validated singleton read model. Currency includes its
// persisted minor-digit snapshot and is not inferred from locale.
type Settings struct {
	businessName       domain.DisplayName
	locale             domain.Locale
	timezone           domain.BusinessTimezone
	currency           domain.Currency
	hourlyLaborCost    domain.Option[domain.MinorAmount]
	defaultGrossMargin domain.Option[domain.BasisPoints]
	createdAt          domain.UTCInstant
	updatedAt          domain.UTCInstant
}

func New(params Params) (Settings, error) {
	violations := make([]domain.Violation, 0, 6)
	if params.BusinessName.String() == "" {
		violations = append(violations, required("business_name"))
	}
	if params.Locale.IsZero() {
		violations = append(violations, required("locale_code"))
	}
	if params.Timezone.IsZero() {
		violations = append(violations, required("timezone_name"))
	}
	if params.Currency.IsZero() {
		violations = append(violations, required("currency"))
	}
	if err := domain.ValidateTimestampOrder(params.CreatedAt, params.UpdatedAt, domain.None[domain.UTCInstant]()); err != nil {
		if validation, ok := err.(*domain.ValidationError); ok {
			violations = append(violations, validation.Violations()...)
		}
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Settings{}, err
	}
	return Settings{
		businessName: params.BusinessName, locale: params.Locale,
		timezone: params.Timezone, currency: params.Currency,
		hourlyLaborCost:    params.HourlyLaborCost,
		defaultGrossMargin: params.DefaultGrossMargin,
		createdAt:          params.CreatedAt, updatedAt: params.UpdatedAt,
	}, nil
}

func (s Settings) BusinessName() domain.DisplayName                      { return s.businessName }
func (s Settings) Locale() domain.Locale                                 { return s.locale }
func (s Settings) Timezone() domain.BusinessTimezone                     { return s.timezone }
func (s Settings) Currency() domain.Currency                             { return s.currency }
func (s Settings) HourlyLaborCost() domain.Option[domain.MinorAmount]    { return s.hourlyLaborCost }
func (s Settings) DefaultGrossMargin() domain.Option[domain.BasisPoints] { return s.defaultGrossMargin }
func (s Settings) CreatedAt() domain.UTCInstant                          { return s.createdAt }
func (s Settings) UpdatedAt() domain.UTCInstant                          { return s.updatedAt }

func required(field string) domain.Violation {
	return domain.Violation{Field: field, Code: domain.ViolationRequired}
}
