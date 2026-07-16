package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

func TestSaleStorePostsSaleWithCustomerFEFOCOGSAndBalance(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "sale.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	itemID := createSaleTestItem(t, store, "Brigadeiro", true)
	customer, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name: counterpartyName(t, "Walk-in customer"), Roles: counterpartyRoles(t, domain.RoleCustomer),
		CreatedAt: counterpartyInstant(t, 1_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	firstPurchase := postAdjustmentTestPurchase(t, store, itemID, "sale-late", "BRIG-LATE", "2026-12-31", 100, 1_000)
	secondPurchase := postAdjustmentTestPurchase(t, store, itemID, "sale-early", "BRIG-EARLY", "2026-08-01", 100, 1_000)

	posted, err := store.PostSale(ctx, PostSaleInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "sale-1"),
		CounterpartyID: domain.Some(customer.ID()),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 5_000),
		Lines: []PostSaleLineInput{
			{
				ItemID:          itemID,
				Quantity:        mustPurchaseQuantity(t, 150),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, 4_500),
			},
		},
	})
	if err != nil {
		t.Fatalf("post sale: %v", err)
	}
	if posted.ID().IsZero() || posted.PostingSequence().Int64() != 3 {
		t.Fatalf("posted sale = %#v", posted)
	}
	customerID, hasCustomer := posted.CounterpartyID().Get()
	if !hasCustomer || customerID != customer.ID() {
		t.Fatalf("posted customer = %#v, want %s", posted.CounterpartyID(), customer.ID())
	}

	lines := posted.Lines()
	if len(lines) != 1 || lines[0].InventoryValue().Int64() != 15_000_000 ||
		lines[0].CommercialTotal().Int64() != 4_500 {
		t.Fatalf("sale lines = %#v", lines)
	}
	allocations := lines[0].Allocations()
	if len(allocations) != 2 {
		t.Fatalf("allocations = %#v, want 2", allocations)
	}
	if allocations[0].LotID() != secondPurchase.Lines()[0].LotID() || allocations[0].Quantity().Int64() != 100 ||
		allocations[1].LotID() != firstPurchase.Lines()[0].LotID() || allocations[1].Quantity().Int64() != 50 {
		t.Fatalf("allocations = %#v, want early lot then late lot", allocations)
	}

	var quantity, inventoryValue int64
	if err := store.database.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro
		FROM inventory_balances WHERE item_id = ?`, itemID.Int64()).Scan(&quantity, &inventoryValue); err != nil {
		t.Fatal(err)
	}
	if quantity != 50 || inventoryValue != 5_000_000 {
		t.Fatalf("balance = %d/%d, want 50/5000000", quantity, inventoryValue)
	}

	replayed, err := store.PostSale(ctx, PostSaleInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "sale-1"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 9_000),
		Lines: []PostSaleLineInput{
			{
				ItemID:          itemID,
				Quantity:        mustPurchaseQuantity(t, 150),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, 4_500),
			},
		},
	})
	if err != nil {
		t.Fatalf("replay sale: %v", err)
	}
	if replayed.ID() != posted.ID() || !replayed.PostedAt().Equal(posted.PostedAt()) {
		t.Fatalf("replayed = %#v, want original %#v", replayed, posted)
	}
}

func TestSaleStorePostsPromotionWithoutCustomer(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "sale-promotion.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	itemID := createSaleTestItem(t, store, "Sample cake", true)
	purchase := postAdjustmentTestPurchase(t, store, itemID, "sale-promotion-stock", "SAMPLE-LOT", "2026-12-31", 20, 200)

	posted, err := store.PostSale(ctx, PostSaleInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "sale-promotion"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 3_000),
		Reason:         domain.Some(domain.ReasonSample),
		Lines: []PostSaleLineInput{
			{
				ItemID:          itemID,
				Quantity:        mustPurchaseQuantity(t, 5),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, 0),
				LotID:           domain.Some(purchase.Lines()[0].LotID()),
			},
		},
	})
	if err != nil {
		t.Fatalf("post promotional sale: %v", err)
	}
	if posted.CounterpartyID().IsSome() {
		t.Fatalf("counterparty = %#v, want none", posted.CounterpartyID())
	}
	reason, ok := posted.Reason().Get()
	if !ok || reason != domain.ReasonSample {
		t.Fatalf("reason = %#v, want SAMPLE", posted.Reason())
	}
	lines := posted.Lines()
	if len(lines) != 1 || lines[0].CommercialTotal().Int64() != 0 ||
		lines[0].InventoryValue().Int64() != 500_000 {
		t.Fatalf("promotional sale lines = %#v", lines)
	}
	if allocations := lines[0].Allocations(); len(allocations) != 1 ||
		allocations[0].LotID() != purchase.Lines()[0].LotID() ||
		allocations[0].Quantity().Int64() != 5 {
		t.Fatalf("promotional sale allocations = %#v", allocations)
	}
}

func TestSaleStoreRejectsZeroTotalWithoutReason(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "sale-zero-no-reason.db"), database.DefaultOpenOptions())
	itemID := createSaleTestItem(t, store, "Free cookie", true)
	postAdjustmentTestPurchase(t, store, itemID, "sale-zero-stock", "FREE-LOT", "2026-12-31", 10, 100)

	_, err := store.PostSale(context.Background(), saleInputFixture(t, itemID, "sale-zero-no-reason", 1, 0))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("zero sale without reason error = %v, want validation", err)
	}
}

func TestSaleStoreRejectsNonSellableItem(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "sale-non-sellable.db"), database.DefaultOpenOptions())
	itemID := createSaleTestItem(t, store, "Raw flour", false)
	postAdjustmentTestPurchase(t, store, itemID, "sale-non-sellable-stock", "RAW-LOT", "2026-12-31", 10, 100)

	_, err := store.PostSale(context.Background(), saleInputFixture(t, itemID, "sale-non-sellable", 1, 100))
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("non-sellable sale error = %v, want conflict", err)
	}
}

func TestSaleStoreRejectsInsufficientStock(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "sale-insufficient.db"), database.DefaultOpenOptions())
	itemID := createSaleTestItem(t, store, "Insufficient cupcake", true)

	_, err := store.PostSale(context.Background(), saleInputFixture(t, itemID, "sale-insufficient", 1, 100))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("insufficient stock error = %v, want validation", err)
	}
}

func TestSaleStoreRejectsExpiredLots(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "sale-expired.db"), database.DefaultOpenOptions())
	itemID := createSaleTestItem(t, store, "Expired pie", true)
	postAdjustmentTestPurchase(t, store, itemID, "sale-expired-stock", "OLD-LOT", "2026-07-01", 10, 100)

	_, err := store.PostSale(context.Background(), saleInputFixture(t, itemID, "sale-expired", 1, 100))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expired lot error = %v, want validation", err)
	}
}

func createSaleTestItem(t *testing.T, store *Store, name string, sellable bool) domain.ItemID {
	t.Helper()
	created := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, name),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, sellable),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})
	return created.Item().ID()
}

func saleInputFixture(
	t *testing.T,
	itemID domain.ItemID,
	idempotencyKey string,
	quantityAtomic int64,
	commercialTotalMinor int64,
) PostSaleInput {
	t.Helper()
	return PostSaleInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, idempotencyKey),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 5_000),
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
