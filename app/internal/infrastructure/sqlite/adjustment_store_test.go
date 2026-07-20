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

func TestAdjustmentStorePostsPositiveAdjustmentLotAndBalance(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "adjustment-in.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Sugar"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})
	value := mustInventoryValue(t, 2_500_000)
	postedAt := mustCatalogInstant(t, 2_000)
	posted, err := store.PostAdjustment(ctx, PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "adjustment-in-1"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       postedAt,
		Reason:         domain.ReasonOpeningBalance,
		Lines: []PostAdjustmentLineInput{
			{
				ItemID:         item.Item().ID(),
				Direction:      domain.DirectionIn,
				Quantity:       mustPurchaseQuantity(t, 500),
				EnteredUnit:    mustCatalogUnitCode(t, "g"),
				Conversion:     mustCatalogConversion(t, 1_000, 1),
				InventoryValue: domain.Some(value),
				LotCode:        domain.Some(counterpartyText(t, "OPENING-SUGAR")),
				ExpiresOn:      domain.Some(mustPurchaseDate(t, "2026-12-31")),
			},
		},
	})
	if err != nil {
		t.Fatalf("post adjustment: %v", err)
	}
	if posted.ID().IsZero() || posted.PostingSequence().Int64() != 1 || posted.Reason() != domain.ReasonOpeningBalance {
		t.Fatalf("posted adjustment = %#v", posted)
	}
	lines := posted.Lines()
	lotID, hasLot := lines[0].LotID().Get()
	if len(lines) != 1 || lines[0].Direction() != domain.DirectionIn || !hasLot || lotID.IsZero() {
		t.Fatalf("posted lines = %#v", lines)
	}

	var quantity, inventoryValue, lastDocumentID, updatedAt int64
	err = store.database.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro, last_document_id, updated_at_ms
		FROM inventory_balances WHERE item_id = ?`, item.Item().ID().Int64()).
		Scan(&quantity, &inventoryValue, &lastDocumentID, &updatedAt)
	if err != nil {
		t.Fatalf("read balance: %v", err)
	}
	if quantity != 500 || inventoryValue != value.Int64() || lastDocumentID != posted.ID().Int64() || updatedAt != postedAt.UnixMilli() {
		t.Fatalf("balance = %d/%d/%d/%d", quantity, inventoryValue, lastDocumentID, updatedAt)
	}

	replayed, err := store.PostAdjustment(ctx, PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "adjustment-in-1"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 3_000),
		Reason:         domain.ReasonOpeningBalance,
		Lines: []PostAdjustmentLineInput{
			{
				ItemID:         item.Item().ID(),
				Direction:      domain.DirectionIn,
				Quantity:       mustPurchaseQuantity(t, 500),
				EnteredUnit:    mustCatalogUnitCode(t, "g"),
				Conversion:     mustCatalogConversion(t, 1_000, 1),
				InventoryValue: domain.Some(value),
			},
		},
	})
	if err != nil {
		t.Fatalf("replay adjustment: %v", err)
	}
	if replayed.ID() != posted.ID() || !replayed.PostedAt().Equal(posted.PostedAt()) {
		t.Fatalf("replayed = %#v, want original %#v", replayed, posted)
	}

	second, err := store.PostAdjustment(ctx, PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "adjustment-in-2"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-16"),
		PostedAt:       mustCatalogInstant(t, 4_000),
		Reason:         domain.ReasonOpeningBalance,
		Lines: []PostAdjustmentLineInput{
			{
				ItemID:         item.Item().ID(),
				Direction:      domain.DirectionIn,
				Quantity:       mustPurchaseQuantity(t, 100),
				EnteredUnit:    mustCatalogUnitCode(t, "g"),
				Conversion:     mustCatalogConversion(t, 1_000, 1),
				InventoryValue: domain.Some(mustInventoryValue(t, 500_000)),
			},
		},
	})
	if err != nil {
		t.Fatalf("post second adjustment: %v", err)
	}

	firstPage, err := store.ListPostedAdjustments(ctx, AdjustmentListFilter{PageSize: 1})
	if err != nil {
		t.Fatalf("list first adjustment page: %v", err)
	}
	firstItems := firstPage.Items()
	if len(firstItems) != 1 || firstItems[0].ID() != second.ID() {
		t.Fatalf("first page = %#v, want second adjustment first", firstItems)
	}
	if _, ok := firstPage.Next().Get(); !ok {
		t.Fatal("first adjustment page should have a next cursor")
	}

	secondPage, err := store.ListPostedAdjustments(ctx, AdjustmentListFilter{
		After:    firstPage.Next(),
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("list second adjustment page: %v", err)
	}
	secondItems := secondPage.Items()
	if len(secondItems) != 1 || secondItems[0].ID() != posted.ID() {
		t.Fatalf("second page = %#v, want original adjustment", secondItems)
	}
	if _, ok := secondPage.Next().Get(); ok {
		t.Fatalf("second page next = %#v, want none", secondPage.Next())
	}
}

func TestAdjustmentStorePostsNegativeAdjustmentWithFEFOAllocations(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "adjustment-out.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Butter"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})
	firstPurchase := postAdjustmentTestPurchase(t, store, item.Item().ID(), "purchase-first", "LOT-LATE", "2026-12-31", 100, 1_000)
	secondPurchase := postAdjustmentTestPurchase(t, store, item.Item().ID(), "purchase-second", "LOT-EARLY", "2026-08-01", 100, 1_000)

	postedAt := mustCatalogInstant(t, 4_000)
	posted, err := store.PostAdjustment(ctx, PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "adjustment-out-1"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       postedAt,
		Reason:         domain.ReasonWaste,
		Lines: []PostAdjustmentLineInput{
			{
				ItemID:      item.Item().ID(),
				Direction:   domain.DirectionOut,
				Quantity:    mustPurchaseQuantity(t, 150),
				EnteredUnit: mustCatalogUnitCode(t, "g"),
				Conversion:  mustCatalogConversion(t, 1_000, 1),
			},
		},
	})
	if err != nil {
		t.Fatalf("post negative adjustment: %v", err)
	}
	lines := posted.Lines()
	if len(lines) != 1 || lines[0].InventoryValue().Int64() != 15_000_000 {
		t.Fatalf("posted negative lines = %#v", lines)
	}
	allocations := lines[0].Allocations()
	if len(allocations) != 2 {
		t.Fatalf("allocations = %#v, want 2", allocations)
	}
	secondLotID := secondPurchase.Lines()[0].LotID()
	firstLotID := firstPurchase.Lines()[0].LotID()
	if allocations[0].LotID() != secondLotID || allocations[0].Quantity().Int64() != 100 ||
		allocations[1].LotID() != firstLotID || allocations[1].Quantity().Int64() != 50 {
		t.Fatalf("allocations = %#v, want early lot then late lot", allocations)
	}

	var quantity, inventoryValue int64
	err = store.database.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro
		FROM inventory_balances WHERE item_id = ?`, item.Item().ID().Int64()).
		Scan(&quantity, &inventoryValue)
	if err != nil {
		t.Fatalf("read balance: %v", err)
	}
	if quantity != 50 || inventoryValue != 5_000_000 {
		t.Fatalf("balance = %d/%d, want 50/5000000", quantity, inventoryValue)
	}
}

func TestAdjustmentStoreRejectsInsufficientStock(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "adjustment-insufficient.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Chocolate"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})

	_, err := store.PostAdjustment(ctx, PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "adjustment-too-large"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 2_000),
		Reason:         domain.ReasonDamage,
		Lines: []PostAdjustmentLineInput{
			{
				ItemID:      item.Item().ID(),
				Direction:   domain.DirectionOut,
				Quantity:    mustPurchaseQuantity(t, 1),
				EnteredUnit: mustCatalogUnitCode(t, "g"),
				Conversion:  mustCatalogConversion(t, 1_000, 1),
			},
		},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("insufficient stock error = %v, want validation", err)
	}
}

func postAdjustmentTestPurchase(
	t *testing.T,
	store *Store,
	itemID domain.ItemID,
	idempotencyKey string,
	lotCode string,
	expiresOn string,
	quantity int64,
	totalMinor int64,
) PostedPurchaseDocument {
	t.Helper()
	posted, err := store.PostPurchase(context.Background(), PostPurchaseInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, idempotencyKey),
		OccurredOn:     mustPurchaseDate(t, "2026-07-01"),
		PostedAt:       mustCatalogInstant(t, quantity*10),
		Lines: []PostPurchaseLineInput{
			{
				ItemID:          itemID,
				Quantity:        mustPurchaseQuantity(t, quantity),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, totalMinor),
				LotCode:         domain.Some(counterpartyText(t, lotCode)),
				ExpiresOn:       domain.Some(mustPurchaseDate(t, expiresOn)),
			},
		},
	})
	if err != nil {
		t.Fatalf("post fixture purchase: %v", err)
	}
	return posted
}

func mustInventoryValue(t *testing.T, raw int64) domain.InventoryValue {
	t.Helper()
	value, err := domain.NewInventoryValue(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
