package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

func TestReportingStoreSalesReportAggregatesSalesAndExcludesReversedDocuments(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "reporting-sales.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	itemID := createSaleTestItem(t, store, "Reported cake", true)
	postAdjustmentTestPurchase(t, store, itemID, "reporting-stock", "REPORT-LOT", "2026-12-31", 100, 1_000)
	customer, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name:      counterpartyName(t, "Dashboard customer"),
		Roles:     counterpartyRoles(t, domain.RoleCustomer),
		CreatedAt: counterpartyInstant(t, 2_000),
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	if _, err := store.PostSale(ctx, reportSaleInput(t, itemID, "report-previous-sale", "2026-07-03", 5, 500, domain.None[domain.CounterpartyID](), domain.None[domain.DocumentReason]())); err != nil {
		t.Fatalf("post previous sale: %v", err)
	}
	if _, err := store.PostSale(ctx, reportSaleInput(t, itemID, "report-current-sale", "2026-07-10", 10, 2_000, domain.Some(customer.ID()), domain.None[domain.DocumentReason]())); err != nil {
		t.Fatalf("post current sale: %v", err)
	}
	reversedSale, err := store.PostSale(ctx, reportSaleInput(t, itemID, "report-reversed-sale", "2026-07-11", 4, 800, domain.None[domain.CounterpartyID](), domain.None[domain.DocumentReason]()))
	if err != nil {
		t.Fatalf("post sale to reverse: %v", err)
	}
	if _, err := store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-report-sale"),
		TargetDocumentID: reversedSale.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-11"),
		PostedAt:         mustCatalogInstant(t, 8_000),
	}); err != nil {
		t.Fatalf("reverse sale: %v", err)
	}
	if _, err := store.PostSale(ctx, reportSaleInput(t, itemID, "report-sample-sale", "2026-07-12", 5, 0, domain.None[domain.CounterpartyID](), domain.Some(domain.ReasonSample))); err != nil {
		t.Fatalf("post sample sale: %v", err)
	}

	report, err := store.GetSalesReportData(ctx, ReportingPeriodFilter{
		FromOccurredOn: "2026-07-10",
		ToOccurredOn:   "2026-07-31",
		Granularity:    "DAY",
	}, ReportingPeriodFilter{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-09",
		Granularity:    "DAY",
	}, 5)
	if err != nil {
		t.Fatalf("get sales report data: %v", err)
	}

	if report.CurrentTotals.SalesCount != 2 ||
		report.CurrentTotals.QuantityAtomic != 15 ||
		report.CurrentTotals.RevenueMinor != 2_000 ||
		report.CurrentTotals.COGSMicro != 1_500_000 {
		t.Fatalf("current totals = %#v", report.CurrentTotals)
	}
	if report.PreviousTotals.SalesCount != 1 || report.PreviousTotals.RevenueMinor != 500 {
		t.Fatalf("previous totals = %#v", report.PreviousTotals)
	}
	if len(report.SalesRevenueSeries) != 2 ||
		report.SalesRevenueSeries[0].Bucket != "2026-07-10" ||
		report.SalesRevenueSeries[0].RevenueMinor != 2_000 ||
		report.SalesRevenueSeries[1].Bucket != "2026-07-12" ||
		report.SalesRevenueSeries[1].RevenueMinor != 0 {
		t.Fatalf("daily series = %#v", report.SalesRevenueSeries)
	}
	if len(report.MonthlySeries) != 1 ||
		report.MonthlySeries[0].Bucket != "2026-07" ||
		report.MonthlySeries[0].SalesCount != 2 ||
		report.MonthlySeries[0].RevenueMinor != 2_000 {
		t.Fatalf("monthly series = %#v", report.MonthlySeries)
	}
	if len(report.TopProductsByQuantity) != 1 ||
		report.TopProductsByQuantity[0].QuantityAtomic != 15 ||
		report.TopProductsByQuantity[0].RevenueMinor != 2_000 ||
		report.TopProductsByQuantity[0].COGSMicro != 1_500_000 ||
		report.TopProductsByQuantity[0].InventoryValueMicro != 0 {
		t.Fatalf("top products = %#v", report.TopProductsByQuantity)
	}
	if report.FreeSales.DocumentCount != 1 ||
		report.FreeSales.QuantityAtomic != 5 ||
		report.FreeSales.COGSMicro != 500_000 {
		t.Fatalf("free sales = %#v", report.FreeSales)
	}
	if len(report.SalesByCustomer) != 1 ||
		report.SalesByCustomer[0].DocumentCount != 1 ||
		report.SalesByCustomer[0].RevenueMinor != 2_000 {
		t.Fatalf("sales by customer = %#v", report.SalesByCustomer)
	}
	if report.AnonymousSales.DocumentCount != 1 || report.AnonymousSales.RevenueMinor != 0 {
		t.Fatalf("anonymous sales = %#v", report.AnonymousSales)
	}
}

func TestReportingStoreInventoryReportSummarizesBalancesLowStockAndLotRisk(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "reporting-inventory.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	lowStockItemID := createReportingItem(t, store, "Low stock chocolate", true, domain.Some(mustPurchaseQuantity(t, 100)))
	valueItemID := createReportingItem(t, store, "Valuable sugar", true, domain.None[domain.AtomicQuantity]())
	expiredItemID := createReportingItem(t, store, "Expired filling", true, domain.None[domain.AtomicQuantity]())
	createReportingItem(t, store, "Zero sellable cookie", true, domain.None[domain.AtomicQuantity]())

	postAdjustmentTestPurchase(t, store, lowStockItemID, "reporting-low-stock", "LOW-LOT", "2026-07-20", 50, 500)
	postAdjustmentTestPurchase(t, store, valueItemID, "reporting-value-stock", "VALUE-LOT", "2026-08-10", 200, 1_000)
	postAdjustmentTestPurchase(t, store, expiredItemID, "reporting-expired-stock", "EXPIRED-LOT", "2026-07-01", 10, 100)

	report, err := store.GetInventoryReportData(ctx, ReportingPeriodFilter{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-15",
		Granularity:    "DAY",
	}, 10)
	if err != nil {
		t.Fatalf("get inventory report data: %v", err)
	}

	if report.TotalInventoryValueMicro != 16_000_000 ||
		report.LowStockItemCount != 1 ||
		report.ZeroStockSellableCount != 1 {
		t.Fatalf("inventory totals = value %d low %d zero %d", report.TotalInventoryValueMicro, report.LowStockItemCount, report.ZeroStockSellableCount)
	}
	if len(report.LowStockItems) != 1 ||
		report.LowStockItems[0].ItemName != "Low stock chocolate" ||
		report.LowStockItems[0].QuantityAtomic != 50 ||
		report.LowStockItems[0].ReorderQuantityAtomic.IsNone() {
		t.Fatalf("low stock items = %#v", report.LowStockItems)
	}
	if len(report.ExpiringLots7Days) != 1 ||
		report.ExpiringLots7Days[0].ItemName != "Low stock chocolate" ||
		report.ExpiringLots7Days[0].AvailableQuantity != 50 {
		t.Fatalf("expiring 7 = %#v", report.ExpiringLots7Days)
	}
	if len(report.ExpiringLots30Days) != 2 {
		t.Fatalf("expiring 30 = %#v", report.ExpiringLots30Days)
	}
	if len(report.ExpiredLotsWithStock) != 1 ||
		report.ExpiredLotsWithStock[0].ItemName != "Expired filling" ||
		report.ExpiredLotsWithStock[0].InventoryValueMicro != 1_000_000 {
		t.Fatalf("expired lots = %#v", report.ExpiredLotsWithStock)
	}
	if len(report.InventoryValueByItem) != 3 ||
		report.InventoryValueByItem[0].ItemName != "Valuable sugar" ||
		report.InventoryValueByItem[0].InventoryValueMicro != 10_000_000 {
		t.Fatalf("value by item = %#v", report.InventoryValueByItem)
	}
}

func TestReportingStorePurchaseReportAggregatesSpendAndExcludesReversedDocuments(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "reporting-purchases.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	itemID := createReportingItem(t, store, "Reported cocoa", false, domain.None[domain.AtomicQuantity]())
	supplierA := createReportingSupplier(t, store, "Supplier A")
	supplierB := createReportingSupplier(t, store, "Supplier B")

	postReportingPurchase(t, store, itemID, "report-purchase-a", "2026-07-05", 2_000, domain.Some(supplierA), domain.None[domain.DocumentReason](), 100, 1_000)
	postReportingPurchase(t, store, itemID, "report-purchase-b", "2026-07-20", 3_000, domain.Some(supplierB), domain.None[domain.DocumentReason](), 50, 250)
	postReportingPurchase(t, store, itemID, "report-free-stock", "2026-07-20", 4_000, domain.None[domain.CounterpartyID](), domain.Some(domain.ReasonFreeStock), 10, 0)
	reversed := postReportingPurchase(t, store, itemID, "report-reversed-purchase", "2026-07-21", 5_000, domain.Some(supplierA), domain.None[domain.DocumentReason](), 20, 999)
	if _, err := store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-report-purchase"),
		TargetDocumentID: reversed.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-21"),
		PostedAt:         mustCatalogInstant(t, 6_000),
	}); err != nil {
		t.Fatalf("reverse purchase: %v", err)
	}

	report, err := store.GetPurchaseReportData(ctx, ReportingPeriodFilter{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-31",
		Granularity:    "DAY",
	}, 10)
	if err != nil {
		t.Fatalf("get purchase report data: %v", err)
	}

	if len(report.PurchaseSpendSeries) != 2 ||
		report.PurchaseSpendSeries[0].Bucket != "2026-07-05" ||
		report.PurchaseSpendSeries[0].DocumentCount != 1 ||
		report.PurchaseSpendSeries[0].SpendMinor != 1_000 ||
		report.PurchaseSpendSeries[0].InventoryValueMicro != 10_000_000 ||
		report.PurchaseSpendSeries[1].Bucket != "2026-07-20" ||
		report.PurchaseSpendSeries[1].DocumentCount != 2 ||
		report.PurchaseSpendSeries[1].QuantityAtomic != 60 ||
		report.PurchaseSpendSeries[1].SpendMinor != 250 {
		t.Fatalf("purchase spend series = %#v", report.PurchaseSpendSeries)
	}
	if len(report.TopSuppliersBySpend) != 2 ||
		report.TopSuppliersBySpend[0].CounterpartyName.IsNone() ||
		report.TopSuppliersBySpend[0].SpendMinor != 1_000 ||
		report.TopSuppliersBySpend[1].SpendMinor != 250 {
		t.Fatalf("top suppliers = %#v", report.TopSuppliersBySpend)
	}
	if len(report.FreeStockEntrySeries) != 1 ||
		report.FreeStockEntrySeries[0].Bucket != "2026-07-20" ||
		report.FreeStockEntrySeries[0].DocumentCount != 1 ||
		report.FreeStockEntrySeries[0].QuantityAtomic != 10 ||
		report.FreeStockEntrySeries[0].SpendMinor != 0 {
		t.Fatalf("free stock entries = %#v", report.FreeStockEntrySeries)
	}
}

func TestReportingStoreProductionReportAggregatesRunsYieldAndExcludesReversedDocuments(t *testing.T) {
	store := recipeTestStore(t, "reporting-production.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Reported cake", false, true)
	componentID := recipeTestItem(t, store, "Reported flour", true, false)
	recipeValue, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name:         recipeName(t, "Reported cake recipe"),
		OutputItemID: outputID,
		CreatedAt:    recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "mix", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 500, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatalf("create reporting recipe: %v", err)
	}
	postAdjustmentTestPurchase(t, store, componentID, "reporting-production-stock", "RPT-FLOUR", "2026-12-31", 1_000, 1_000)

	posted, err := store.PostProduction(ctx, productionReportInput(
		t,
		"reporting-production-run",
		recipeValue.CurrentRevision().ID(),
		componentID,
		500,
		250,
		2_500_000,
	))
	if err != nil {
		t.Fatalf("post production: %v", err)
	}
	if posted.OutputLine().InventoryValue().Int64() != 7_500_000 {
		t.Fatalf("production output = %#v", posted.OutputLine())
	}

	reversed, err := store.PostProduction(ctx, productionReportInput(
		t,
		"reporting-production-reversed",
		recipeValue.CurrentRevision().ID(),
		componentID,
		100,
		50,
		123_000,
	))
	if err != nil {
		t.Fatalf("post reversed production fixture: %v", err)
	}
	if _, err := store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-reporting-production"),
		TargetDocumentID: reversed.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-18"),
		PostedAt:         mustCatalogInstant(t, 9_000),
	}); err != nil {
		t.Fatalf("reverse production: %v", err)
	}

	report, err := store.GetProductionReportData(ctx, ReportingPeriodFilter{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-31",
		Granularity:    "MONTH",
	}, 10)
	if err != nil {
		t.Fatalf("get production report data: %v", err)
	}

	if len(report.ProductionByRecipeProduct) != 1 {
		t.Fatalf("production by recipe/product = %#v", report.ProductionByRecipeProduct)
	}
	product := report.ProductionByRecipeProduct[0]
	standardYield, standardOK := product.StandardYieldAtomic.Get()
	actualYield, actualOK := product.ActualYieldAtomic.Get()
	variance, varianceOK := product.VarianceAtomic.Get()
	if product.ItemName != "Reported cake" ||
		product.RecipeName.IsNone() ||
		product.DocumentCount != 1 ||
		product.QuantityAtomic != 250 ||
		product.InventoryValueMicro != 7_500_000 ||
		product.DirectCostMicro != 2_500_000 ||
		!standardOK || standardYield != 1_000 ||
		!actualOK || actualYield != 250 ||
		!varianceOK || variance != -750 {
		t.Fatalf("production product metric = %#v", product)
	}
	if len(report.DirectCostSeries) != 1 ||
		report.DirectCostSeries[0].Bucket != "2026-07" ||
		report.DirectCostSeries[0].DocumentCount != 1 ||
		report.DirectCostSeries[0].QuantityAtomic != 250 ||
		report.DirectCostSeries[0].InventoryValueMicro != 7_500_000 ||
		report.DirectCostSeries[0].DirectCostMicro != 2_500_000 {
		t.Fatalf("direct cost series = %#v", report.DirectCostSeries)
	}
	if len(report.YieldVariance) != 1 ||
		report.YieldVariance[0].ItemName != "Reported cake" ||
		report.YieldVariance[0].VarianceAtomic.IsNone() {
		t.Fatalf("yield variance = %#v", report.YieldVariance)
	}
}

func TestReportingStoreAdjustmentReportAggregatesReasonsAndExactReversals(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "reporting-adjustments.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	itemID := createReportingItem(t, store, "Reported butter", false, domain.None[domain.AtomicQuantity]())
	postAdjustmentTestPurchase(t, store, itemID, "reporting-adjustment-stock", "RPT-BUTTER", "2026-12-31", 100, 1_000)

	postReportingAdjustment(t, store, "report-positive-adjustment", "2026-07-05", 2_000, domain.ReasonOpeningBalance, itemID, domain.DirectionIn, 20, domain.Some(mustInventoryValue(t, 2_000_000)))
	postReportingAdjustment(t, store, "report-negative-adjustment", "2026-07-06", 3_000, domain.ReasonWaste, itemID, domain.DirectionOut, 10, domain.None[domain.InventoryValue]())
	reversed := postReportingAdjustment(t, store, "report-reversed-adjustment", "2026-07-07", 4_000, domain.ReasonDamage, itemID, domain.DirectionOut, 5, domain.None[domain.InventoryValue]())
	if _, err := store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-report-adjustment"),
		TargetDocumentID: reversed.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-08"),
		PostedAt:         mustCatalogInstant(t, 5_000),
	}); err != nil {
		t.Fatalf("reverse adjustment: %v", err)
	}

	report, err := store.GetAdjustmentReportData(ctx, ReportingPeriodFilter{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-31",
		Granularity:    "MONTH",
	})
	if err != nil {
		t.Fatalf("get adjustment report data: %v", err)
	}

	if len(report.PositiveByReason) != 1 ||
		report.PositiveByReason[0].ReasonCode != domain.ReasonOpeningBalance.String() ||
		report.PositiveByReason[0].DocumentCount != 1 ||
		report.PositiveByReason[0].QuantityAtomic != 20 ||
		report.PositiveByReason[0].InventoryValueMicro != 2_000_000 {
		t.Fatalf("positive adjustments = %#v", report.PositiveByReason)
	}
	if len(report.NegativeByReason) != 1 ||
		report.NegativeByReason[0].ReasonCode != domain.ReasonWaste.String() ||
		report.NegativeByReason[0].DocumentCount != 1 ||
		report.NegativeByReason[0].QuantityAtomic != 10 ||
		report.NegativeByReason[0].InventoryValueMicro != 1_000_000 {
		t.Fatalf("negative adjustments = %#v", report.NegativeByReason)
	}
	if len(report.ExactReversals) != 1 ||
		report.ExactReversals[0].Bucket != "2026-07" ||
		report.ExactReversals[0].DocumentCount != 1 ||
		report.ExactReversals[0].QuantityAtomic != 5 ||
		report.ExactReversals[0].InventoryValueMicro != 500_000 {
		t.Fatalf("exact reversals = %#v", report.ExactReversals)
	}
}

func reportSaleInput(
	t *testing.T,
	itemID domain.ItemID,
	idempotencyKey string,
	occurredOn string,
	quantityAtomic int64,
	commercialTotalMinor int64,
	counterpartyID domain.Option[domain.CounterpartyID],
	reason domain.Option[domain.DocumentReason],
) PostSaleInput {
	t.Helper()
	return PostSaleInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, idempotencyKey),
		CounterpartyID: counterpartyID,
		OccurredOn:     mustPurchaseDate(t, occurredOn),
		PostedAt:       mustCatalogInstant(t, quantityAtomic*1_000),
		Reason:         reason,
		Lines: []PostSaleLineInput{
			{
				ItemID:          itemID,
				Quantity:        mustPurchaseQuantity(t, quantityAtomic),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, commercialTotalMinor),
			},
		},
	}
}

func postReportingAdjustment(
	t *testing.T,
	store *Store,
	idempotencyKey string,
	occurredOn string,
	postedAtMS int64,
	reason domain.DocumentReason,
	itemID domain.ItemID,
	direction domain.Direction,
	quantityAtomic int64,
	inventoryValue domain.Option[domain.InventoryValue],
) PostedAdjustmentDocument {
	t.Helper()
	posted, err := store.PostAdjustment(context.Background(), PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, idempotencyKey),
		OccurredOn:     mustPurchaseDate(t, occurredOn),
		PostedAt:       mustCatalogInstant(t, postedAtMS),
		Reason:         reason,
		Lines: []PostAdjustmentLineInput{
			{
				ItemID:         itemID,
				Direction:      direction,
				Quantity:       mustPurchaseQuantity(t, quantityAtomic),
				EnteredUnit:    mustCatalogUnitCode(t, "g"),
				Conversion:     mustCatalogConversion(t, 1_000, 1),
				InventoryValue: inventoryValue,
				LotCode:        domain.Some(counterpartyText(t, idempotencyKey+"-lot")),
				ExpiresOn:      domain.Some(mustPurchaseDate(t, "2026-12-31")),
			},
		},
	})
	if err != nil {
		t.Fatalf("post reporting adjustment: %v", err)
	}
	return posted
}

func createReportingItem(
	t *testing.T,
	store *Store,
	name string,
	sellable bool,
	reorderQuantity domain.Option[domain.AtomicQuantity],
) domain.ItemID {
	t.Helper()
	created := createCatalogItem(t, store, CreateItemInput{
		Name:            mustCatalogName(t, name),
		BaseUnit:        mustCatalogUnitCode(t, "g"),
		Capabilities:    catalog.NewCapabilities(true, false, sellable),
		ReorderQuantity: reorderQuantity,
		CreatedAt:       mustCatalogInstant(t, 1_000),
		UpdatedAt:       mustCatalogInstant(t, 1_000),
	})
	return created.Item().ID()
}

func createReportingSupplier(t *testing.T, store *Store, name string) domain.CounterpartyID {
	t.Helper()
	created, err := store.CreateCounterparty(context.Background(), CreateCounterpartyInput{
		Name:      counterpartyName(t, name),
		Roles:     counterpartyRoles(t, domain.RoleSupplier),
		CreatedAt: counterpartyInstant(t, 1_000),
	})
	if err != nil {
		t.Fatalf("create reporting supplier: %v", err)
	}
	return created.ID()
}

func postReportingPurchase(
	t *testing.T,
	store *Store,
	itemID domain.ItemID,
	idempotencyKey string,
	occurredOn string,
	postedAtMS int64,
	counterpartyID domain.Option[domain.CounterpartyID],
	reason domain.Option[domain.DocumentReason],
	quantityAtomic int64,
	commercialTotalMinor int64,
) PostedPurchaseDocument {
	t.Helper()
	posted, err := store.PostPurchase(context.Background(), PostPurchaseInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, idempotencyKey),
		CounterpartyID: counterpartyID,
		OccurredOn:     mustPurchaseDate(t, occurredOn),
		PostedAt:       mustCatalogInstant(t, postedAtMS),
		Reason:         reason,
		Lines: []PostPurchaseLineInput{
			{
				ItemID:          itemID,
				Quantity:        mustPurchaseQuantity(t, quantityAtomic),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, commercialTotalMinor),
				LotCode:         domain.Some(counterpartyText(t, idempotencyKey+"-lot")),
				ExpiresOn:       domain.Some(mustPurchaseDate(t, "2026-12-31")),
			},
		},
	})
	if err != nil {
		t.Fatalf("post reporting purchase: %v", err)
	}
	return posted
}

func productionReportInput(
	t *testing.T,
	idempotencyKey string,
	revisionID domain.RecipeRevisionID,
	componentID domain.ItemID,
	componentQuantity int64,
	outputQuantity int64,
	directCostMicro int64,
) PostProductionInput {
	t.Helper()
	return PostProductionInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, idempotencyKey),
		RecipeRevisionID: revisionID,
		OccurredOn:       mustPurchaseDate(t, "2026-07-15"),
		PostedAt:         recipeInstant(t, outputQuantity*10),
		DirectCost:       mustInventoryValue(t, directCostMicro),
		Output: PostProductionOutputInput{
			Quantity:    recipeQuantity(t, outputQuantity),
			EnteredUnit: recipeUnit(t, "g"),
			Conversion:  recipeConversion(t, 1_000, 1),
			LotCode:     domain.Some(recipeText(t, idempotencyKey+"-out")),
			ExpiresOn:   domain.Some(mustPurchaseDate(t, "2026-12-31")),
		},
		Inputs: []PostProductionComponentInput{
			{
				ItemID:      componentID,
				Quantity:    recipeQuantity(t, componentQuantity),
				EnteredUnit: recipeUnit(t, "g"),
				Conversion:  recipeConversion(t, 1_000, 1),
			},
		},
	}
}
