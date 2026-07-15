package wails

import (
	"context"
	"testing"
	"time"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type surfaceClock struct {
	now domain.UTCInstant
}

func (c *surfaceClock) Now() (domain.UTCInstant, error) {
	return c.now, nil
}

func TestPhase5BackendSurfaceForSettingsUnitsAndCounterparties(t *testing.T) {
	db := newSurfaceDatabase(t)
	store := sqlite.NewStore(db)
	clock := &surfaceClock{now: must(domain.UTCInstantFromUnixMilli(2_000))}
	ctx := context.Background()

	settingsHandler := NewSettingsHandler(application.NewSettingsService(application.NewSQLiteSettingsStore(store)))
	referenceDataHandler := NewReferenceDataHandler(application.NewReferenceDataService(application.NewSQLiteReferenceDataStore(store)))
	counterpartyHandler := NewCounterpartyHandler(application.NewCounterpartyService(
		application.NewSQLiteCounterpartyStore(store),
		clock,
	))

	settingsValue, err := settingsHandler.GetSettings(ctx)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	hourlyLaborCost := int64(12_500)
	defaultGrossMargin := int64(2_500)
	updatedSettings, err := settingsHandler.UpdateSettings(ctx, dto.SettingsUpdateRequest{
		BusinessName:        "Sweeters Test",
		Locale:              "pt-BR",
		Timezone:            "America/Sao_Paulo",
		CurrencyCode:        settingsValue.CurrencyCode,
		CurrencyMinorDigits: settingsValue.CurrencyMinorDigits,
		HourlyLaborCost:     &hourlyLaborCost,
		DefaultGrossMargin:  &defaultGrossMargin,
		ExpectedUpdatedAtMs: settingsValue.UpdatedAtMs,
		UpdatedAtMs:         settingsValue.UpdatedAtMs + 1_000,
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if updatedSettings.HourlyLaborCost == nil || *updatedSettings.HourlyLaborCost != hourlyLaborCost {
		t.Fatalf("hourly labor cost = %#v", updatedSettings.HourlyLaborCost)
	}

	units, err := referenceDataHandler.ListMeasurementUnits(ctx)
	if err != nil {
		t.Fatalf("list measurement units: %v", err)
	}
	if len(units) == 0 {
		t.Fatal("expected seeded measurement units")
	}
	unit, err := referenceDataHandler.GetMeasurementUnit(ctx, units[0].Code)
	if err != nil {
		t.Fatalf("get measurement unit: %v", err)
	}
	if unit.Code != units[0].Code || unit.NumeratorAtomic <= 0 || unit.Denominator <= 0 {
		t.Fatalf("measurement unit = %#v", unit)
	}

	phone := "+55 11 99999-0000"
	created, err := counterpartyHandler.CreateCounterparty(ctx, dto.CounterpartyWriteRequest{
		Name:  "Supplier One",
		Phone: &phone,
		Roles: []string{"SUPPLIER"},
	})
	if err != nil {
		t.Fatalf("create counterparty: %v", err)
	}
	if created.ID == 0 || created.CreatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("created counterparty = %#v", created)
	}

	page, err := counterpartyHandler.ListCounterparties(ctx, dto.CounterpartyListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list counterparties: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != created.ID {
		t.Fatalf("counterparty page = %#v", page)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(3_000))
	archived, err := counterpartyHandler.ArchiveCounterparty(ctx, created.ID, dto.VersionedCounterpartyRequest{
		ExpectedUpdatedAtMs: created.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive counterparty: %v", err)
	}
	if archived.ArchivedAtMs == nil || *archived.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived counterparty = %#v", archived)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(4_000))
	restored, err := counterpartyHandler.RestoreCounterparty(ctx, archived.ID, dto.VersionedCounterpartyRequest{
		ExpectedUpdatedAtMs: archived.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("restore counterparty: %v", err)
	}
	if restored.ArchivedAtMs != nil || restored.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("restored counterparty = %#v", restored)
	}
}

func newSurfaceDatabase(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.NewDatabaseWithOptions(":memory:", database.OpenOptions{BusyTimeout: time.Second})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})
	return db
}
