package settings_test

import (
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/settings"
)

func TestSettingsSnapshot(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1000))
	updated := must(domain.UTCInstantFromUnixMilli(2000))
	value, err := settings.New(settings.Params{
		BusinessName: must(domain.NewDisplayName("Sweeters")), Locale: must(domain.NewLocale("pt-BR")),
		Timezone:           must(domain.NewBusinessTimezone("America/Sao_Paulo")),
		Currency:           must(domain.RestoreCurrency("BRL", 2)),
		HourlyLaborCost:    domain.Some(must(domain.NewMinorAmount(2500))),
		DefaultGrossMargin: domain.Some(must(domain.NewBasisPoints(3000))),
		CreatedAt:          created, UpdatedAt: updated,
	})
	if err != nil || value.Currency().Code().String() != "BRL" || value.DefaultGrossMargin().IsNone() {
		t.Fatalf("settings = %#v, %v", value, err)
	}
}

func TestSettingsRejectsIncompleteOrBackwardsSnapshot(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(2000))
	updated := must(domain.UTCInstantFromUnixMilli(1000))
	_, err := settings.New(settings.Params{CreatedAt: created, UpdatedAt: updated})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid settings error = %v", err)
	}
	var validation *domain.ValidationError
	if !errors.As(err, &validation) || len(validation.Violations()) < 5 {
		t.Fatalf("expected deterministic aggregate violations: %v", err)
	}
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
