package wails

import (
	"testing"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/settings"
)

func TestMapSettingsIncludesOptionalValues(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1_000))
	updated := must(domain.UTCInstantFromUnixMilli(2_000))
	value := mustSettings(settings.New(settings.Params{
		BusinessName:       must(domain.NewDisplayName("Sweeters")),
		Locale:             must(domain.NewLocale("pt-BR")),
		Timezone:           must(domain.NewBusinessTimezone("America/Sao_Paulo")),
		Currency:           must(domain.RestoreCurrency("BRL", 2)),
		HourlyLaborCost:    domain.Some(must(domain.NewMinorAmount(2_500))),
		DefaultGrossMargin: domain.Some(must(domain.NewBasisPoints(3_000))),
		CreatedAt:          created,
		UpdatedAt:          updated,
	}))

	response := mapSettings(value)
	if response.BusinessName != "Sweeters" {
		t.Fatalf("business name = %q", response.BusinessName)
	}
	if response.HourlyLaborCost == nil || *response.HourlyLaborCost != 2_500 {
		t.Fatalf("hourly labor cost = %#v", response.HourlyLaborCost)
	}
	if response.DefaultGrossMargin == nil || *response.DefaultGrossMargin != 3_000 {
		t.Fatalf("default gross margin = %#v", response.DefaultGrossMargin)
	}
	if response.CreatedAtMs != created.UnixMilli() || response.UpdatedAtMs != updated.UnixMilli() {
		t.Fatalf("timestamps = %d/%d", response.CreatedAtMs, response.UpdatedAtMs)
	}
}

func TestMapSettingsKeepsOptionalValuesEmpty(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1_000))
	updated := must(domain.UTCInstantFromUnixMilli(2_000))
	value := mustSettings(settings.New(settings.Params{
		BusinessName:       must(domain.NewDisplayName("Sweeters")),
		Locale:             must(domain.NewLocale("pt-BR")),
		Timezone:           must(domain.NewBusinessTimezone("America/Sao_Paulo")),
		Currency:           must(domain.RestoreCurrency("BRL", 2)),
		HourlyLaborCost:    domain.None[domain.MinorAmount](),
		DefaultGrossMargin: domain.None[domain.BasisPoints](),
		CreatedAt:          created,
		UpdatedAt:          updated,
	}))

	response := mapSettings(value)
	if response.HourlyLaborCost != nil {
		t.Fatalf("hourly labor cost = %#v", response.HourlyLaborCost)
	}
	if response.DefaultGrossMargin != nil {
		t.Fatalf("default gross margin = %#v", response.DefaultGrossMargin)
	}
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func mustSettings(value settings.Settings, err error) settings.Settings {
	if err != nil {
		panic(err)
	}
	return value
}
