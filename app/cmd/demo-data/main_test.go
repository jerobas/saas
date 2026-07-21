package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
	presentationwails "github.com/jerobas/saas/internal/presentation/wails"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

func TestSeedDemoDatabasePopulatesOperationalScreensAndReports(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "app.db")
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	if err := seedDemoDatabase(databasePath, 3, now); err != nil {
		t.Fatalf("seed demo database: %v", err)
	}

	db, err := database.NewDatabase(databasePath)
	if err != nil {
		t.Fatalf("reopen demo database: %v", err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)
	clock := application.SystemClock{}
	catalog := presentationwails.NewCatalogHandler(application.NewCatalogService(
		application.NewSQLiteCatalogStore(store), clock,
	))
	counterparty := presentationwails.NewCounterpartyHandler(application.NewCounterpartyService(
		application.NewSQLiteCounterpartyStore(store), clock,
	))
	purchase := presentationwails.NewPurchaseHandler(application.NewPurchaseService(
		application.NewSQLitePurchaseStore(store), clock,
	))
	sale := presentationwails.NewSaleHandler(application.NewSaleService(
		application.NewSQLiteSaleStore(store), clock,
	))
	reporting := presentationwails.NewReportingHandler(application.NewReportingService(
		application.NewSQLiteReportingStore(store),
	))

	items, err := catalog.ListItems(dto.ItemListRequest{PageSize: 100})
	if err != nil {
		t.Fatalf("list demo items: %v", err)
	}
	if len(items.Items) != 20 {
		t.Fatalf("demo items = %d, want 20", len(items.Items))
	}
	people, err := counterparty.ListCounterparties(dto.CounterpartyListRequest{PageSize: 100})
	if err != nil {
		t.Fatalf("list demo counterparties: %v", err)
	}
	if len(people.Items) != 16 {
		t.Fatalf("demo counterparties = %d, want 16", len(people.Items))
	}
	purchases, err := purchase.ListPurchases(dto.PurchaseListRequest{PageSize: 100})
	if err != nil {
		t.Fatalf("list demo purchases: %v", err)
	}
	if len(purchases.Items) != 6 {
		t.Fatalf("demo purchases = %d, want 6", len(purchases.Items))
	}
	sales, err := sale.ListSales(dto.SaleListRequest{PageSize: 100})
	if err != nil {
		t.Fatalf("list demo sales: %v", err)
	}
	if len(sales.Items) != 18 {
		t.Fatalf("demo sales = %d, want 18", len(sales.Items))
	}

	report, err := reporting.GetSalesReport(dto.ReportingPeriodRequest{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-20",
		Granularity:    "DAY",
	})
	if err != nil {
		t.Fatalf("get demo sales report: %v", err)
	}
	if report.TotalSalesCount != 3 || report.CommercialTotalMinor <= 0 {
		t.Fatalf(
			"demo sales report = count %d revenue %d",
			report.TotalSalesCount,
			report.CommercialTotalMinor,
		)
	}
	if len(report.SalesRevenueSeries) == 0 || len(report.TopProductsByQuantity) == 0 {
		t.Fatalf("demo report should contain chart series and top products")
	}

	inventory, err := reporting.GetInventoryReport(dto.ReportingPeriodRequest{
		FromOccurredOn: "2026-07-01",
		ToOccurredOn:   "2026-07-20",
		Granularity:    "DAY",
	})
	if err != nil {
		t.Fatalf("get demo inventory report: %v", err)
	}
	if inventory.TotalInventoryValueMicro <= 0 || len(inventory.InventoryValueByItem) == 0 {
		t.Fatalf("demo inventory report should contain valued stock")
	}
}

func TestSeedDemoDatabaseRefusesExistingFile(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "app.db")
	if err := os.WriteFile(databasePath, []byte("keep me"), 0o600); err != nil {
		t.Fatalf("create existing file: %v", err)
	}

	err := seedDemoDatabase(databasePath, 1, time.Now())
	if err == nil || !strings.Contains(err.Error(), "database already exists") {
		t.Fatalf("existing database error = %v", err)
	}
	contents, readErr := os.ReadFile(databasePath)
	if readErr != nil {
		t.Fatalf("read existing file: %v", readErr)
	}
	if string(contents) != "keep me" {
		t.Fatalf("existing database was modified: %q", contents)
	}
}
