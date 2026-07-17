package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
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
		report.TopProductsByQuantity[0].RevenueMinor != 2_000 {
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
