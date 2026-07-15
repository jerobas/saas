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

func TestPhase5BackendSurfaceForSettingsUnitsCatalogAndCounterparties(t *testing.T) {
	db := newSurfaceDatabase(t)
	store := sqlite.NewStore(db)
	clock := &surfaceClock{now: must(domain.UTCInstantFromUnixMilli(2_000))}
	ctx := context.Background()

	settingsHandler := NewSettingsHandler(application.NewSettingsService(application.NewSQLiteSettingsStore(store), clock))
	referenceDataHandler := NewReferenceDataHandler(application.NewReferenceDataService(application.NewSQLiteReferenceDataStore(store)))
	catalogHandler := NewCatalogHandler(application.NewCatalogService(application.NewSQLiteCatalogStore(store), clock))
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
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if updatedSettings.HourlyLaborCost == nil || *updatedSettings.HourlyLaborCost != hourlyLaborCost {
		t.Fatalf("hourly labor cost = %#v", updatedSettings.HourlyLaborCost)
	}
	if updatedSettings.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("settings updated at = %d, want %d", updatedSettings.UpdatedAtMs, clock.now.UnixMilli())
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

	gram, err := referenceDataHandler.GetMeasurementUnit(ctx, "g")
	if err != nil {
		t.Fatalf("get gram unit: %v", err)
	}
	if !gram.IsItemBase {
		t.Fatalf("gram unit should be an item base unit: %#v", gram)
	}
	kilogram, err := referenceDataHandler.GetMeasurementUnit(ctx, "kg")
	if err != nil {
		t.Fatalf("get kilogram unit: %v", err)
	}
	if kilogram.Dimension != gram.Dimension {
		t.Fatalf("kilogram dimension = %q, want %q", kilogram.Dimension, gram.Dimension)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(3_000))
	defaultSalePrice := int64(1_250)
	reorderQuantity := int64(5_000)
	item, err := catalogHandler.CreateItem(ctx, dto.ItemWriteRequest{
		Name:         "Flour",
		SKU:          stringPointer("FLOUR-001"),
		Description:  stringPointer("Wheat flour"),
		BaseUnitCode: "g",
		Capabilities: dto.CapabilitiesRequest{
			Purchasable: true,
			Sellable:    true,
		},
		DefaultSalePrice: &defaultSalePrice,
		ReorderQuantity:  &reorderQuantity,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if item.ID == 0 || item.BaseUnit.Code != "g" || item.CreatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("created item = %#v", item)
	}

	itemPage, err := catalogHandler.ListItems(ctx, dto.ItemListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(itemPage.Items) != 1 || itemPage.Items[0].ID != item.ID {
		t.Fatalf("item page = %#v", itemPage)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(4_000))
	updatedItem, err := catalogHandler.UpdateItem(ctx, item.ID, dto.ItemUpdateRequest{
		ItemWriteRequest: dto.ItemWriteRequest{
			Name:         "Premium Flour",
			SKU:          stringPointer("FLOUR-001"),
			Description:  stringPointer("Premium wheat flour"),
			BaseUnitCode: "g",
			Capabilities: dto.CapabilitiesRequest{
				Purchasable: true,
				Sellable:    true,
			},
			DefaultSalePrice: &defaultSalePrice,
			ReorderQuantity:  &reorderQuantity,
		},
		ExpectedUpdatedAtMs: item.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("update item: %v", err)
	}
	if updatedItem.Name != "Premium Flour" || updatedItem.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("updated item = %#v", updatedItem)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(5_000))
	packaging, err := catalogHandler.CreateItemPackaging(ctx, dto.PackagingCreateRequest{
		ItemID: updatedItem.ID,
		PackagingWriteRequest: dto.PackagingWriteRequest{
			Name:                  "Kilogram bag",
			EnteredUnitCode:       "kg",
			ConversionNumerator:   2_000_000,
			ConversionDenominator: 2,
		},
	})
	if err != nil {
		t.Fatalf("create packaging: %v", err)
	}
	if packaging.ItemID != updatedItem.ID || packaging.EnteredUnit.Code != "kg" || packaging.ConversionNumerator != 1_000_000 {
		t.Fatalf("created packaging = %#v", packaging)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(6_000))
	updatedPackaging, err := catalogHandler.UpdateItemPackaging(ctx, packaging.ID, dto.PackagingUpdateRequest{
		PackagingWriteRequest: dto.PackagingWriteRequest{
			Name:                  "Kilogram sack",
			EnteredUnitCode:       "kg",
			ConversionNumerator:   1_000_000,
			ConversionDenominator: 1,
		},
		ExpectedUpdatedAtMs: packaging.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("update packaging: %v", err)
	}
	if updatedPackaging.Name != "Kilogram sack" || updatedPackaging.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("updated packaging = %#v", updatedPackaging)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(7_000))
	archivedPackaging, err := catalogHandler.ArchiveItemPackaging(ctx, updatedPackaging.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: updatedPackaging.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive packaging: %v", err)
	}
	if archivedPackaging.ArchivedAtMs == nil || *archivedPackaging.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived packaging = %#v", archivedPackaging)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(8_000))
	reconfiguredPackaging, err := catalogHandler.ReconfigureArchivedItemPackaging(ctx, archivedPackaging.ID, dto.PackagingUpdateRequest{
		PackagingWriteRequest: dto.PackagingWriteRequest{
			Name:                  "Kilogram package",
			EnteredUnitCode:       "kg",
			ConversionNumerator:   1_000_000,
			ConversionDenominator: 1,
		},
		ExpectedUpdatedAtMs: archivedPackaging.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("reconfigure archived packaging: %v", err)
	}
	if reconfiguredPackaging.ArchivedAtMs == nil || *reconfiguredPackaging.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("reconfigured packaging = %#v", reconfiguredPackaging)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(9_000))
	restoredPackaging, err := catalogHandler.RestoreItemPackaging(ctx, reconfiguredPackaging.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: reconfiguredPackaging.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("restore packaging: %v", err)
	}
	if restoredPackaging.ArchivedAtMs != nil || restoredPackaging.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("restored packaging = %#v", restoredPackaging)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(10_000))
	archivedItem, err := catalogHandler.ArchiveItem(ctx, updatedItem.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: updatedItem.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive item: %v", err)
	}
	if archivedItem.ArchivedAtMs == nil || *archivedItem.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived item = %#v", archivedItem)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(11_000))
	restoredItem, err := catalogHandler.RestoreItem(ctx, archivedItem.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: archivedItem.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("restore item: %v", err)
	}
	if restoredItem.ArchivedAtMs != nil || restoredItem.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("restored item = %#v", restoredItem)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(12_000))
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

	clock.now = must(domain.UTCInstantFromUnixMilli(13_000))
	archived, err := counterpartyHandler.ArchiveCounterparty(ctx, created.ID, dto.VersionedCounterpartyRequest{
		ExpectedUpdatedAtMs: created.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive counterparty: %v", err)
	}
	if archived.ArchivedAtMs == nil || *archived.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived counterparty = %#v", archived)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(14_000))
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

func stringPointer(value string) *string {
	return &value
}
