package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
)

func TestSettingsStoreReadsBaselineAndSeededUnits(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "settings-read.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	settings, err := store.GetSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if settings.BusinessName().String() != "Sweeters" {
		t.Fatalf("business name = %q, want Sweeters", settings.BusinessName().String())
	}
	if settings.Locale().String() != "pt-BR" {
		t.Fatalf("locale = %q, want pt-BR", settings.Locale().String())
	}
	if settings.Timezone().Name() != "America/Sao_Paulo" {
		t.Fatalf("timezone = %q, want America/Sao_Paulo", settings.Timezone().Name())
	}
	if settings.Currency().Code().String() != "BRL" || settings.Currency().MinorDigits().Int() != 2 {
		t.Fatalf("currency = %s/%d, want BRL/2", settings.Currency().Code().String(), settings.Currency().MinorDigits().Int())
	}
	if settings.HourlyLaborCost().IsSome() || settings.DefaultGrossMargin().IsSome() {
		t.Fatal("optional baseline settings must be absent")
	}
	if settings.CreatedAt().UnixMilli() != 0 || settings.UpdatedAt().UnixMilli() != 0 {
		t.Fatalf("settings timestamps = %d/%d, want 0/0", settings.CreatedAt().UnixMilli(), settings.UpdatedAt().UnixMilli())
	}

	units, err := store.ListMeasurementUnits(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if units == nil {
		t.Fatal("ListMeasurementUnits returned a nil slice")
	}
	if len(units) != 9 {
		t.Fatalf("unit count = %d, want 9", len(units))
	}
	wantCodes := []string{"mg", "g", "kg", "ul", "ml", "l", "milli_each", "each", "dozen"}
	for index, want := range wantCodes {
		if got := units[index].Code().String(); got != want {
			t.Fatalf("unit[%d] = %q, want %q", index, got, want)
		}
	}

	gramCode := mustSettingsUnitCode(t, "g")
	gram, err := store.GetMeasurementUnit(ctx, gramCode)
	if err != nil {
		t.Fatal(err)
	}
	if gram.Dimension() != domain.DimensionMass || !gram.IsItemBase() || !gram.IsSeeded() {
		t.Fatalf("unexpected gram metadata: dimension=%s item_base=%t seeded=%t", gram.Dimension(), gram.IsItemBase(), gram.IsSeeded())
	}
	if gram.Conversion().NumeratorAtomic() != 1000 || gram.Conversion().Denominator() != 1 {
		t.Fatalf("gram conversion = %d/%d, want 1000/1", gram.Conversion().NumeratorAtomic(), gram.Conversion().Denominator())
	}
}

func TestSettingsStoreUpdatesAtomicallyAndRejectsStaleOrNonAdvancingVersions(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "settings-update.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	expected := mustSettingsInstant(t, 0)
	updatedAt := mustSettingsInstant(t, 10)
	hourlyCost := mustSettingsMinorAmount(t, 12_345)
	margin := mustSettingsBasisPoints(t, 2_500)

	input := UpdateSettingsInput{
		BusinessName:       mustSettingsDisplayName(t, "  Joao's Bakery  "),
		Locale:             mustSettingsLocale(t, "en-US"),
		Timezone:           mustSettingsTimezone(t, "UTC"),
		Currency:           mustSettingsCurrency(t, "USD"),
		HourlyLaborCost:    domain.Some(hourlyCost),
		DefaultGrossMargin: domain.Some(margin),
		ExpectedUpdatedAt:  expected,
		UpdatedAt:          updatedAt,
	}
	updated, err := store.UpdateSettings(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if updated.BusinessName().String() != "Joao's Bakery" || updated.Locale().String() != "en-US" || updated.Timezone().Name() != "UTC" {
		t.Fatalf("unexpected updated settings: %q/%s/%s", updated.BusinessName().String(), updated.Locale().String(), updated.Timezone().Name())
	}
	if updated.Currency().Code().String() != "USD" || updated.Currency().MinorDigits().Int() != 2 {
		t.Fatalf("currency = %s/%d, want USD/2", updated.Currency().Code().String(), updated.Currency().MinorDigits().Int())
	}
	if value, ok := updated.HourlyLaborCost().Get(); !ok || value.Int64() != 12_345 {
		t.Fatalf("hourly cost = %#v/%t, want 12345/present", value, ok)
	}
	if value, ok := updated.DefaultGrossMargin().Get(); !ok || value.Int64() != 2_500 {
		t.Fatalf("margin = %#v/%t, want 2500/present", value, ok)
	}
	if !updated.UpdatedAt().Equal(updatedAt) || !updated.CreatedAt().Equal(expected) {
		t.Fatalf("timestamps = %d/%d, want 0/10", updated.CreatedAt().UnixMilli(), updated.UpdatedAt().UnixMilli())
	}

	stale := input
	stale.UpdatedAt = mustSettingsInstant(t, 20)
	if _, err := store.UpdateSettings(ctx, stale); !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale update error = %v, want domain.ErrStale", err)
	}

	nonAdvancing := input
	nonAdvancing.ExpectedUpdatedAt = updatedAt
	nonAdvancing.UpdatedAt = updatedAt
	if _, err := store.UpdateSettings(ctx, nonAdvancing); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("non-advancing update error = %v, want domain.ErrValidation", err)
	}

	reloaded, err := store.GetSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded.UpdatedAt().Equal(updatedAt) || reloaded.BusinessName().String() != "Joao's Bakery" {
		t.Fatalf("failed writes changed persisted settings: name=%q updated=%d", reloaded.BusinessName().String(), reloaded.UpdatedAt().UnixMilli())
	}
}

func TestSettingsStoreRejectsChangedCurrencyWithWrongRegistryScale(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "settings-currency-scale.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	spoofedCurrency, err := domain.RestoreCurrency("BRL", 6)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.UpdateSettings(ctx, UpdateSettingsInput{
		BusinessName:      mustSettingsDisplayName(t, "Sweeters"),
		Locale:            mustSettingsLocale(t, "pt-BR"),
		Timezone:          mustSettingsTimezone(t, "America/Sao_Paulo"),
		Currency:          spoofedCurrency,
		ExpectedUpdatedAt: mustSettingsInstant(t, 0),
		UpdatedAt:         mustSettingsInstant(t, 10),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("wrong currency scale error = %v, want domain.ErrValidation", err)
	}

	reloaded, err := store.GetSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Currency().Code().String() != "BRL" || reloaded.Currency().MinorDigits().Int() != 2 || reloaded.UpdatedAt().UnixMilli() != 0 {
		t.Fatalf("rejected currency changed settings: currency=%s/%d updated=%d",
			reloaded.Currency().Code().String(), reloaded.Currency().MinorDigits().Int(), reloaded.UpdatedAt().UnixMilli())
	}
}

func TestSettingsStoreRejectsNoncanonicalPersistedText(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "settings-corrupt.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	if _, err := store.database.ExecContext(ctx, "UPDATE app_settings SET business_name = '  Sweeters  '"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetSettings(ctx); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("noncanonical settings error = %v, want domain.ErrCorruptData", err)
	}
}

func mustSettingsDisplayName(t *testing.T, raw string) domain.DisplayName {
	t.Helper()
	value, err := domain.NewDisplayName(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSettingsLocale(t *testing.T, raw string) domain.Locale {
	t.Helper()
	value, err := domain.NewLocale(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSettingsTimezone(t *testing.T, raw string) domain.BusinessTimezone {
	t.Helper()
	value, err := domain.NewBusinessTimezone(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSettingsCurrency(t *testing.T, raw string) domain.Currency {
	t.Helper()
	value, err := domain.NewCurrency(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSettingsMinorAmount(t *testing.T, raw int64) domain.MinorAmount {
	t.Helper()
	value, err := domain.NewMinorAmount(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSettingsBasisPoints(t *testing.T, raw int64) domain.BasisPoints {
	t.Helper()
	value, err := domain.NewBasisPoints(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSettingsInstant(t *testing.T, milliseconds int64) domain.UTCInstant {
	t.Helper()
	value, err := domain.UTCInstantFromUnixMilli(milliseconds)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustSettingsUnitCode(t *testing.T, raw string) domain.UnitCode {
	t.Helper()
	value, err := domain.NewUnitCode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
