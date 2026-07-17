package wails

import (
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

	settingsHandler := NewSettingsHandler(application.NewSettingsService(application.NewSQLiteSettingsStore(store), clock))
	referenceDataHandler := NewReferenceDataHandler(application.NewReferenceDataService(application.NewSQLiteReferenceDataStore(store)))
	catalogHandler := NewCatalogHandler(application.NewCatalogService(application.NewSQLiteCatalogStore(store), clock))
	counterpartyHandler := NewCounterpartyHandler(application.NewCounterpartyService(
		application.NewSQLiteCounterpartyStore(store),
		clock,
	))
	purchaseHandler := NewPurchaseHandler(application.NewPurchaseService(
		application.NewSQLitePurchaseStore(store),
		clock,
	))
	adjustmentHandler := NewAdjustmentHandler(application.NewAdjustmentService(
		application.NewSQLiteAdjustmentStore(store),
		clock,
	))
	reversalHandler := NewReversalHandler(application.NewReversalService(
		application.NewSQLiteReversalStore(store),
		clock,
	))
	productionHandler := NewProductionHandler(application.NewProductionService(
		application.NewSQLiteProductionStore(store),
		clock,
	))
	saleHandler := NewSaleHandler(application.NewSaleService(
		application.NewSQLiteSaleStore(store),
		clock,
	))
	recipeHandler := NewRecipeHandler(application.NewRecipeService(
		application.NewSQLiteRecipeStore(store),
		clock,
	))
	inventoryHandler := NewInventoryHandler(application.NewInventoryService(
		application.NewSQLiteInventoryStore(store),
	))
	reportingHandler := NewReportingHandler(application.NewReportingService())

	settingsValue, err := settingsHandler.GetSettings()
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	hourlyLaborCost := int64(12_500)
	defaultGrossMargin := int64(2_500)
	updatedSettings, err := settingsHandler.UpdateSettings(dto.SettingsUpdateRequest{
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

	units, err := referenceDataHandler.ListMeasurementUnits()
	if err != nil {
		t.Fatalf("list measurement units: %v", err)
	}
	if len(units) == 0 {
		t.Fatal("expected seeded measurement units")
	}
	unit, err := referenceDataHandler.GetMeasurementUnit(units[0].Code)
	if err != nil {
		t.Fatalf("get measurement unit: %v", err)
	}
	if unit.Code != units[0].Code || unit.NumeratorAtomic <= 0 || unit.Denominator <= 0 {
		t.Fatalf("measurement unit = %#v", unit)
	}

	gram, err := referenceDataHandler.GetMeasurementUnit("g")
	if err != nil {
		t.Fatalf("get gram unit: %v", err)
	}
	if !gram.IsItemBase {
		t.Fatalf("gram unit should be an item base unit: %#v", gram)
	}
	kilogram, err := referenceDataHandler.GetMeasurementUnit("kg")
	if err != nil {
		t.Fatalf("get kilogram unit: %v", err)
	}
	if kilogram.Dimension != gram.Dimension {
		t.Fatalf("kilogram dimension = %q, want %q", kilogram.Dimension, gram.Dimension)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(3_000))
	defaultSalePrice := int64(1_250)
	reorderQuantity := int64(5_000)
	item, err := catalogHandler.CreateItem(dto.ItemWriteRequest{
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

	itemPage, err := catalogHandler.ListItems(dto.ItemListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(itemPage.Items) != 1 || itemPage.Items[0].ID != item.ID {
		t.Fatalf("item page = %#v", itemPage)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(4_000))
	updatedItem, err := catalogHandler.UpdateItem(item.ID, dto.ItemUpdateRequest{
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
	packaging, err := catalogHandler.CreateItemPackaging(dto.PackagingCreateRequest{
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
	updatedPackaging, err := catalogHandler.UpdateItemPackaging(packaging.ID, dto.PackagingUpdateRequest{
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
	archivedPackaging, err := catalogHandler.ArchiveItemPackaging(updatedPackaging.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: updatedPackaging.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive packaging: %v", err)
	}
	if archivedPackaging.ArchivedAtMs == nil || *archivedPackaging.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived packaging = %#v", archivedPackaging)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(8_000))
	reconfiguredPackaging, err := catalogHandler.ReconfigureArchivedItemPackaging(archivedPackaging.ID, dto.PackagingUpdateRequest{
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
	restoredPackaging, err := catalogHandler.RestoreItemPackaging(reconfiguredPackaging.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: reconfiguredPackaging.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("restore packaging: %v", err)
	}
	if restoredPackaging.ArchivedAtMs != nil || restoredPackaging.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("restored packaging = %#v", restoredPackaging)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(10_000))
	archivedItem, err := catalogHandler.ArchiveItem(updatedItem.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: updatedItem.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive item: %v", err)
	}
	if archivedItem.ArchivedAtMs == nil || *archivedItem.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived item = %#v", archivedItem)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(11_000))
	restoredItem, err := catalogHandler.RestoreItem(archivedItem.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: archivedItem.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("restore item: %v", err)
	}
	if restoredItem.ArchivedAtMs != nil || restoredItem.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("restored item = %#v", restoredItem)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(11_500))
	outputItem, err := catalogHandler.CreateItem(dto.ItemWriteRequest{
		Name:         "Cake",
		BaseUnitCode: "g",
		Capabilities: dto.CapabilitiesRequest{
			Producible: true,
			Sellable:   true,
		},
		DefaultSalePrice: &defaultSalePrice,
	})
	if err != nil {
		t.Fatalf("create recipe output item: %v", err)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(11_600))
	recipeUnitCode := "g"
	recipeValue, err := recipeHandler.CreateRecipe(dto.RecipeCreateRequest{
		Name:         "Cake recipe",
		OutputItemID: outputItem.ID,
		Revision: dto.RecipeRevisionWriteRequest{
			StandardYieldQuantity:  1_000,
			Instructions:           "Mix and bake.",
			PreparationTimeMinutes: 45,
			Components: []dto.RecipeComponentRequest{
				{
					Order:          1,
					ItemID:         restoredItem.ID,
					QuantityAtomic: 500,
					SourceType:     "UNIT",
					UnitCode:       &recipeUnitCode,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create recipe: %v", err)
	}
	if recipeValue.ID == 0 || recipeValue.CurrentRevision.Number != 1 ||
		recipeValue.CurrentRevision.CreatedAtMs != clock.now.UnixMilli() ||
		len(recipeValue.CurrentRevision.Components) != 1 {
		t.Fatalf("recipe = %#v", recipeValue)
	}

	recipePage, err := recipeHandler.ListRecipes(dto.RecipeListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list recipes: %v", err)
	}
	if len(recipePage.Items) != 1 || recipePage.Items[0].ID != recipeValue.ID ||
		recipePage.Items[0].OutputItemName != outputItem.Name {
		t.Fatalf("recipe page = %#v", recipePage)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(11_700))
	publishedRevision, err := recipeHandler.PublishRecipeRevision(recipeValue.ID, dto.RecipePublishRevisionRequest{
		ExpectedLatestRevision: recipeValue.CurrentRevision.Number,
		ExpectedUpdatedAtMs:    recipeValue.UpdatedAtMs,
		Revision: dto.RecipeRevisionWriteRequest{
			StandardYieldQuantity:  1_000,
			Instructions:           "Mix, rest, and bake.",
			PreparationTimeMinutes: 50,
			Components: []dto.RecipeComponentRequest{
				{
					Order:          1,
					ItemID:         restoredItem.ID,
					QuantityAtomic: 550,
					SourceType:     "UNIT",
					UnitCode:       &recipeUnitCode,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("publish recipe revision: %v", err)
	}
	if publishedRevision.Number != 2 || publishedRevision.CreatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("published recipe revision = %#v", publishedRevision)
	}
	revisions, err := recipeHandler.ListRecipeRevisions(recipeValue.ID)
	if err != nil {
		t.Fatalf("list recipe revisions: %v", err)
	}
	if len(revisions) != 2 || revisions[0].Number != 2 || revisions[1].Number != 1 {
		t.Fatalf("recipe revisions = %#v", revisions)
	}
	recipeValue, err = recipeHandler.GetRecipe(recipeValue.ID)
	if err != nil {
		t.Fatalf("get recipe: %v", err)
	}
	if recipeValue.CurrentRevision.Number != 2 || recipeValue.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("current recipe = %#v", recipeValue)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(11_800))
	archivedRecipe, err := recipeHandler.ArchiveRecipe(recipeValue.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: recipeValue.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive recipe: %v", err)
	}
	if archivedRecipe.ArchivedAtMs == nil || *archivedRecipe.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived recipe = %#v", archivedRecipe)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(11_900))
	restoredRecipe, err := recipeHandler.RestoreRecipe(archivedRecipe.ID, dto.VersionedRequest{
		ExpectedUpdatedAtMs: archivedRecipe.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("restore recipe: %v", err)
	}
	if restoredRecipe.ArchivedAtMs != nil || restoredRecipe.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("restored recipe = %#v", restoredRecipe)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(12_000))
	phone := "+55 11 99999-0000"
	created, err := counterpartyHandler.CreateCounterparty(dto.CounterpartyWriteRequest{
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

	page, err := counterpartyHandler.ListCounterparties(dto.CounterpartyListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list counterparties: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != created.ID {
		t.Fatalf("counterparty page = %#v", page)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(13_000))
	archived, err := counterpartyHandler.ArchiveCounterparty(created.ID, dto.VersionedCounterpartyRequest{
		ExpectedUpdatedAtMs: created.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("archive counterparty: %v", err)
	}
	if archived.ArchivedAtMs == nil || *archived.ArchivedAtMs != clock.now.UnixMilli() {
		t.Fatalf("archived counterparty = %#v", archived)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(14_000))
	restored, err := counterpartyHandler.RestoreCounterparty(archived.ID, dto.VersionedCounterpartyRequest{
		ExpectedUpdatedAtMs: archived.UpdatedAtMs,
	})
	if err != nil {
		t.Fatalf("restore counterparty: %v", err)
	}
	if restored.ArchivedAtMs != nil || restored.UpdatedAtMs != clock.now.UnixMilli() {
		t.Fatalf("restored counterparty = %#v", restored)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(15_000))
	expiresOn := "2026-12-31"
	purchase, err := purchaseHandler.PostPurchase(dto.PurchasePostRequest{
		IdempotencyKey: "purchase-flour-1",
		CounterpartyID: &restored.ID,
		OccurredOn:     "2026-07-15",
		Lines: []dto.PurchaseLineRequest{
			{
				ItemID:                    restoredItem.ID,
				QuantityAtomic:            1_000,
				EnteredUnitCode:           "g",
				ConversionNumeratorAtomic: 1_000,
				ConversionDenominator:     1,
				CommercialTotalMinor:      500,
				LotCode:                   stringPointer("LOT-1"),
				ExpiresOn:                 &expiresOn,
			},
		},
	})
	if err != nil {
		t.Fatalf("post purchase: %v", err)
	}
	if purchase.ID == 0 || purchase.PostingSequence != 1 || purchase.PostedAtMs != clock.now.UnixMilli() {
		t.Fatalf("purchase = %#v", purchase)
	}
	if len(purchase.Lines) != 1 || purchase.Lines[0].LotID == 0 || purchase.Lines[0].InventoryValueMicro != 5_000_000 {
		t.Fatalf("purchase lines = %#v", purchase.Lines)
	}

	loadedPurchase, err := purchaseHandler.GetPurchase(purchase.ID)
	if err != nil {
		t.Fatalf("get purchase: %v", err)
	}
	if loadedPurchase.ID != purchase.ID || len(loadedPurchase.Lines) != 1 {
		t.Fatalf("loaded purchase = %#v", loadedPurchase)
	}

	purchasePage, err := purchaseHandler.ListPurchases(dto.PurchaseListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list purchases: %v", err)
	}
	if len(purchasePage.Items) != 1 || purchasePage.Items[0].ID != purchase.ID {
		t.Fatalf("purchase page = %#v", purchasePage)
	}

	balance, err := inventoryHandler.GetInventoryBalance(restoredItem.ID)
	if err != nil {
		t.Fatalf("get inventory balance: %v", err)
	}
	if balance.QuantityAtomic != 1_000 || balance.InventoryValueMicro != 5_000_000 {
		t.Fatalf("balance = %#v", balance)
	}

	balancePage, err := inventoryHandler.ListInventoryBalances(dto.InventoryBalanceListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list inventory balances: %v", err)
	}
	foundPurchasedBalance := false
	for _, item := range balancePage.Items {
		if item.ItemID == restoredItem.ID {
			foundPurchasedBalance = true
			break
		}
	}
	if !foundPurchasedBalance {
		t.Fatalf("balance page = %#v", balancePage)
	}

	lots, err := inventoryHandler.ListItemLotFacts(restoredItem.ID)
	if err != nil {
		t.Fatalf("list item lot facts: %v", err)
	}
	if len(lots) != 1 || lots[0].AvailableQuantity != 1_000 || lots[0].SourceDocumentID != purchase.ID {
		t.Fatalf("lots = %#v", lots)
	}

	ledger, err := inventoryHandler.ListItemLedgerPage(dto.ItemLedgerPageRequest{
		ItemID: restoredItem.ID, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list item ledger: %v", err)
	}
	if len(ledger.Items) != 1 || ledger.Items[0].LineID != purchase.Lines[0].ID {
		t.Fatalf("ledger = %#v", ledger)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(16_000))
	adjustment, err := adjustmentHandler.PostAdjustment(dto.AdjustmentPostRequest{
		IdempotencyKey: "waste-flour-1",
		OccurredOn:     "2026-07-16",
		ReasonCode:     "WASTE",
		Lines: []dto.AdjustmentLineRequest{
			{
				ItemID:                    restoredItem.ID,
				Direction:                 "OUT",
				QuantityAtomic:            250,
				EnteredUnitCode:           "g",
				ConversionNumeratorAtomic: 1_000,
				ConversionDenominator:     1,
			},
		},
	})
	if err != nil {
		t.Fatalf("post adjustment: %v", err)
	}
	if adjustment.ID == 0 || adjustment.PostingSequence != 2 || adjustment.Lines[0].InventoryValueMicro != 1_250_000 {
		t.Fatalf("adjustment = %#v", adjustment)
	}
	if len(adjustment.Lines[0].Allocations) != 1 || adjustment.Lines[0].Allocations[0].QuantityAtomic != 250 {
		t.Fatalf("adjustment allocations = %#v", adjustment.Lines[0].Allocations)
	}

	adjustedBalance, err := inventoryHandler.GetInventoryBalance(restoredItem.ID)
	if err != nil {
		t.Fatalf("get adjusted inventory balance: %v", err)
	}
	if adjustedBalance.QuantityAtomic != 750 || adjustedBalance.InventoryValueMicro != 3_750_000 {
		t.Fatalf("adjusted balance = %#v", adjustedBalance)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(17_000))
	reversal, err := reversalHandler.PostReversal(dto.ReversalPostRequest{
		IdempotencyKey:   "reverse-waste-flour-1",
		TargetDocumentID: adjustment.ID,
		OccurredOn:       "2026-07-16",
	})
	if err != nil {
		t.Fatalf("post reversal: %v", err)
	}
	if reversal.ID == 0 || reversal.PostingSequence != 3 || reversal.TargetDocumentID != adjustment.ID ||
		reversal.ReasonCode != "EXACT_REVERSAL" || reversal.PostedAtMs != clock.now.UnixMilli() {
		t.Fatalf("reversal = %#v", reversal)
	}
	if len(reversal.Lines) != 1 || reversal.Lines[0].Direction != "IN" ||
		reversal.Lines[0].QuantityAtomic != 250 || reversal.Lines[0].InventoryValueMicro != 1_250_000 ||
		reversal.Lines[0].ReversesLineID != adjustment.Lines[0].ID {
		t.Fatalf("reversal lines = %#v", reversal.Lines)
	}
	if len(reversal.Lines[0].Allocations) != 1 ||
		reversal.Lines[0].Allocations[0].RestoresAllocationID == nil ||
		*reversal.Lines[0].Allocations[0].RestoresAllocationID != adjustment.Lines[0].Allocations[0].ID {
		t.Fatalf("reversal allocations = %#v", reversal.Lines[0].Allocations)
	}

	restoredBalance, err := inventoryHandler.GetInventoryBalance(restoredItem.ID)
	if err != nil {
		t.Fatalf("get restored inventory balance: %v", err)
	}
	if restoredBalance.QuantityAtomic != 1_000 || restoredBalance.InventoryValueMicro != 5_000_000 {
		t.Fatalf("restored balance = %#v", restoredBalance)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(18_000))
	outputExpiresOn := "2026-07-20"
	production, err := productionHandler.PostProduction(dto.ProductionPostRequest{
		IdempotencyKey:   "produce-cake-1",
		RecipeRevisionID: recipeValue.CurrentRevision.ID,
		OccurredOn:       "2026-07-17",
		DirectCostMicro:  500_000,
		Output: dto.ProductionOutputRequest{
			QuantityAtomic:            100,
			EnteredUnitCode:           "g",
			ConversionNumeratorAtomic: 1_000,
			ConversionDenominator:     1,
			LotCode:                   stringPointer("CAKE-1"),
			ExpiresOn:                 &outputExpiresOn,
		},
		Inputs: []dto.ProductionComponentRequest{
			{
				ItemID:                    restoredItem.ID,
				QuantityAtomic:            500,
				EnteredUnitCode:           "g",
				ConversionNumeratorAtomic: 1_000,
				ConversionDenominator:     1,
			},
		},
	})
	if err != nil {
		t.Fatalf("post production: %v", err)
	}
	if production.ID == 0 || production.PostingSequence != 4 || production.RecipeRevisionID != recipeValue.CurrentRevision.ID ||
		production.OutputItemID != outputItem.ID || production.PostedAtMs != clock.now.UnixMilli() {
		t.Fatalf("production = %#v", production)
	}
	if len(production.InputLines) != 1 || production.InputLines[0].Direction != "OUT" ||
		production.InputLines[0].InventoryValueMicro != 2_500_000 ||
		len(production.InputLines[0].Allocations) != 1 ||
		production.InputLines[0].Allocations[0].QuantityAtomic != 500 {
		t.Fatalf("production input lines = %#v", production.InputLines)
	}
	if production.OutputLine.Direction != "IN" || production.OutputLine.ItemID != outputItem.ID ||
		production.OutputLine.InventoryValueMicro != 3_000_000 || production.OutputLine.LotID == nil {
		t.Fatalf("production output line = %#v", production.OutputLine)
	}

	componentBalance, err := inventoryHandler.GetInventoryBalance(restoredItem.ID)
	if err != nil {
		t.Fatalf("get component balance after production: %v", err)
	}
	outputBalance, err := inventoryHandler.GetInventoryBalance(outputItem.ID)
	if err != nil {
		t.Fatalf("get output balance after production: %v", err)
	}
	if componentBalance.QuantityAtomic != 500 || componentBalance.InventoryValueMicro != 2_500_000 ||
		outputBalance.QuantityAtomic != 100 || outputBalance.InventoryValueMicro != 3_000_000 {
		t.Fatalf("production balances component=%#v output=%#v", componentBalance, outputBalance)
	}

	clock.now = must(domain.UTCInstantFromUnixMilli(19_000))
	sale, err := saleHandler.PostSale(dto.SalePostRequest{
		IdempotencyKey: "sell-cake-1",
		OccurredOn:     "2026-07-18",
		Lines: []dto.SaleLineRequest{
			{
				ItemID:                    outputItem.ID,
				QuantityAtomic:            20,
				EnteredUnitCode:           "g",
				ConversionNumeratorAtomic: 1_000,
				ConversionDenominator:     1,
				CommercialTotalMinor:      1_000,
			},
		},
	})
	if err != nil {
		t.Fatalf("post sale: %v", err)
	}
	if sale.ID == 0 || sale.PostingSequence != 5 || sale.PostedAtMs != clock.now.UnixMilli() ||
		sale.ReasonCode != nil || sale.CounterpartyID != nil {
		t.Fatalf("sale = %#v", sale)
	}
	if len(sale.Lines) != 1 || sale.Lines[0].Direction != "OUT" ||
		sale.Lines[0].ItemID != outputItem.ID ||
		sale.Lines[0].QuantityAtomic != 20 ||
		sale.Lines[0].InventoryValueMicro != 600_000 ||
		sale.Lines[0].CommercialTotalMinor != 1_000 {
		t.Fatalf("sale lines = %#v", sale.Lines)
	}
	if len(sale.Lines[0].Allocations) != 1 || sale.Lines[0].Allocations[0].QuantityAtomic != 20 {
		t.Fatalf("sale allocations = %#v", sale.Lines[0].Allocations)
	}

	loadedSale, err := saleHandler.GetSale(sale.ID)
	if err != nil {
		t.Fatalf("get sale: %v", err)
	}
	if loadedSale.ID != sale.ID || len(loadedSale.Lines) != 1 {
		t.Fatalf("loaded sale = %#v", loadedSale)
	}

	salePage, err := saleHandler.ListSales(dto.SaleListRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("list sales: %v", err)
	}
	if len(salePage.Items) != 1 || salePage.Items[0].ID != sale.ID {
		t.Fatalf("sale page = %#v", salePage)
	}

	soldOutputBalance, err := inventoryHandler.GetInventoryBalance(outputItem.ID)
	if err != nil {
		t.Fatalf("get output balance after sale: %v", err)
	}
	if soldOutputBalance.QuantityAtomic != 80 || soldOutputBalance.InventoryValueMicro != 2_400_000 {
		t.Fatalf("sold output balance = %#v", soldOutputBalance)
	}

	categoryMix, err := reportingHandler.GetCategoryMixReport(dto.ReportingPeriodRequest{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-31",
		Granularity:    "MONTH",
	})
	if err != nil {
		t.Fatalf("get category mix report: %v", err)
	}
	if categoryMix.Available || categoryMix.UnavailableReason == nil || len(categoryMix.Rows) != 0 {
		t.Fatalf("category mix = %#v", categoryMix)
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
