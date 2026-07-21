package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

func TestPurchaseStorePostsInboundPurchaseLotAndBalance(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "purchase-post.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Flour"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})
	supplier, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name:      counterpartyName(t, "Supplier One"),
		Roles:     counterpartyRoles(t, domain.RoleSupplier),
		CreatedAt: counterpartyInstant(t, 1_100),
	})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}

	postedAt := mustCatalogInstant(t, 2_000)
	posted, err := store.PostPurchase(ctx, PostPurchaseInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "purchase-1"),
		CounterpartyID: domain.Some(supplier.ID()),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       postedAt,
		Lines: []PostPurchaseLineInput{
			{
				ItemID:          item.Item().ID(),
				Quantity:        mustPurchaseQuantity(t, 1_000),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, 500),
				LotCode:         domain.Some(counterpartyText(t, "LOT-A")),
				ExpiresOn:       domain.Some(mustPurchaseDate(t, "2026-12-31")),
			},
		},
	})
	if err != nil {
		t.Fatalf("post purchase: %v", err)
	}
	if posted.ID().IsZero() || posted.PostingSequence().Int64() != 1 || !posted.PostedAt().Equal(postedAt) {
		t.Fatalf("posted purchase = %#v", posted)
	}
	lines := posted.Lines()
	if len(lines) != 1 || lines[0].LotID().IsZero() || lines[0].InventoryValue().Int64() != 5_000_000 {
		t.Fatalf("posted lines = %#v", lines)
	}

	var quantity, value, lastDocumentID, updatedAt int64
	err = store.database.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro, last_document_id, updated_at_ms
		FROM inventory_balances WHERE item_id = ?`, item.Item().ID().Int64()).
		Scan(&quantity, &value, &lastDocumentID, &updatedAt)
	if err != nil {
		t.Fatalf("read balance: %v", err)
	}
	if quantity != 1_000 || value != 5_000_000 || lastDocumentID != posted.ID().Int64() || updatedAt != postedAt.UnixMilli() {
		t.Fatalf("balance = %d/%d/%d/%d", quantity, value, lastDocumentID, updatedAt)
	}

	replayed, err := store.PostPurchase(ctx, PostPurchaseInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "purchase-1"),
		CounterpartyID: domain.Some(supplier.ID()),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 3_000),
		Lines: []PostPurchaseLineInput{
			{
				ItemID:          item.Item().ID(),
				Quantity:        mustPurchaseQuantity(t, 1_000),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, 500),
			},
		},
	})
	if err != nil {
		t.Fatalf("replay purchase: %v", err)
	}
	if replayed.ID() != posted.ID() || !replayed.PostedAt().Equal(posted.PostedAt()) {
		t.Fatalf("replayed = %#v, want original %#v", replayed, posted)
	}

	loaded, err := store.GetPostedPurchase(ctx, posted.ID())
	if err != nil {
		t.Fatalf("get posted purchase: %v", err)
	}
	if loaded.ID() != posted.ID() || len(loaded.Lines()) != 1 {
		t.Fatalf("loaded purchase = %#v", loaded)
	}

	second, err := store.PostPurchase(ctx, PostPurchaseInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "purchase-2"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-16"),
		PostedAt:       mustCatalogInstant(t, 4_000),
		Lines: []PostPurchaseLineInput{
			{
				ItemID:          item.Item().ID(),
				Quantity:        mustPurchaseQuantity(t, 500),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 500, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, 250),
			},
		},
	})
	if err != nil {
		t.Fatalf("post second purchase: %v", err)
	}

	firstPage, err := store.ListPostedPurchases(ctx, PurchaseListFilter{PageSize: 1})
	if err != nil {
		t.Fatalf("list first purchase page: %v", err)
	}
	firstItems := firstPage.Items()
	if len(firstItems) != 1 || firstItems[0].ID() != second.ID() {
		t.Fatalf("first page = %#v, want second purchase first", firstItems)
	}
	cursor, ok := firstPage.Next().Get()
	if !ok || cursor.ID != second.ID() || cursor.PostingSequence != second.PostingSequence() {
		t.Fatalf("first page cursor = %#v", firstPage.Next())
	}

	secondPage, err := store.ListPostedPurchases(ctx, PurchaseListFilter{
		After:    firstPage.Next(),
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("list second purchase page: %v", err)
	}
	secondItems := secondPage.Items()
	if len(secondItems) != 1 || secondItems[0].ID() != posted.ID() {
		t.Fatalf("second page = %#v, want original purchase", secondItems)
	}
	if _, ok := secondPage.Next().Get(); ok {
		t.Fatalf("second page next = %#v, want none", secondPage.Next())
	}
}

func mustPurchaseIdempotencyKey(t *testing.T, raw string) domain.IdempotencyKey {
	t.Helper()
	value, err := domain.NewIdempotencyKey(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustPurchaseDate(t *testing.T, raw string) domain.BusinessDate {
	t.Helper()
	value, err := domain.ParseBusinessDate(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustPurchaseQuantity(t *testing.T, raw int64) domain.AtomicQuantity {
	t.Helper()
	value, err := domain.NewPositiveAtomicQuantity(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustPurchaseMinorAmount(t *testing.T, raw int64) domain.MinorAmount {
	t.Helper()
	value, err := domain.NewMinorAmount(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
