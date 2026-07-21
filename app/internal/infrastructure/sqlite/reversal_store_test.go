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

func TestReversalStoreExactlyReversesPurchase(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "reverse-purchase.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Flour"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})
	purchase := postReversalTestPurchase(t, store, item.Item().ID(), "purchase-reversible", 100, 1_000)

	postedAt := mustCatalogInstant(t, 3_000)
	reversal, err := store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-purchase-1"),
		TargetDocumentID: purchase.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-16"),
		PostedAt:         postedAt,
	})
	if err != nil {
		t.Fatalf("reverse purchase: %v", err)
	}
	if reversal.ID().IsZero() || reversal.TargetDocumentID() != purchase.ID() ||
		reversal.PostingSequence().Int64() != 2 {
		t.Fatalf("reversal = %#v", reversal)
	}
	lines := reversal.Lines()
	if len(lines) != 1 || lines[0].Direction() != domain.DirectionOut ||
		lines[0].Quantity().Int64() != 100 || lines[0].InventoryValue().Int64() != 10_000_000 {
		t.Fatalf("reversal lines = %#v", lines)
	}
	if total, ok := lines[0].CommercialTotal().Get(); !ok || total.Int64() != 1_000 {
		t.Fatalf("commercial total = %v/%v, want 1000", total, ok)
	}
	if len(lines[0].Allocations()) != 1 ||
		lines[0].Allocations()[0].Quantity().Int64() != 100 ||
		lines[0].Allocations()[0].RestoresAllocationID().IsSome() {
		t.Fatalf("purchase reversal allocations = %#v", lines[0].Allocations())
	}

	assertInventoryBalance(t, store, item.Item().ID(), 0, 0, reversal.ID().Int64())
	lots, err := store.ListItemLotFacts(ctx, item.Item().ID())
	if err != nil {
		t.Fatalf("list lots: %v", err)
	}
	if len(lots) != 1 || lots[0].Lot().AvailableQuantity().Int64() != 0 {
		t.Fatalf("lots after reversal = %#v", lots)
	}

	replayed, err := store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-purchase-1"),
		TargetDocumentID: purchase.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-16"),
		PostedAt:         mustCatalogInstant(t, 4_000),
	})
	if err != nil {
		t.Fatalf("replay reversal: %v", err)
	}
	if replayed.ID() != reversal.ID() || !replayed.PostedAt().Equal(reversal.PostedAt()) {
		t.Fatalf("replayed = %#v, want original %#v", replayed, reversal)
	}
}

func TestReversalStoreExactlyReversesNegativeAdjustment(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "reverse-adjustment.db"), database.DefaultOpenOptions())
	ctx := context.Background()

	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Butter"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})
	purchase := postReversalTestPurchase(t, store, item.Item().ID(), "purchase-for-adjustment", 200, 2_000)
	adjustment, err := store.PostAdjustment(ctx, PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "waste-butter-1"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 3_000),
		Reason:         domain.ReasonWaste,
		Lines: []PostAdjustmentLineInput{
			{
				ItemID:      item.Item().ID(),
				Direction:   domain.DirectionOut,
				Quantity:    mustPurchaseQuantity(t, 50),
				EnteredUnit: mustCatalogUnitCode(t, "g"),
				Conversion:  mustCatalogConversion(t, 1_000, 1),
			},
		},
	})
	if err != nil {
		t.Fatalf("post adjustment: %v", err)
	}
	adjustmentLine := adjustment.Lines()[0]
	originalAllocation := adjustmentLine.Allocations()[0]
	assertInventoryBalance(t, store, item.Item().ID(), 150, 15_000_000, adjustment.ID().Int64())

	reversal, err := store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-adjustment-1"),
		TargetDocumentID: adjustment.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-16"),
		PostedAt:         mustCatalogInstant(t, 4_000),
	})
	if err != nil {
		t.Fatalf("reverse adjustment: %v", err)
	}
	lines := reversal.Lines()
	if len(lines) != 1 || lines[0].Direction() != domain.DirectionIn ||
		lines[0].Quantity().Int64() != 50 || lines[0].InventoryValue().Int64() != 5_000_000 ||
		lines[0].ReversesLineID() != adjustmentLine.ID() {
		t.Fatalf("reversal lines = %#v", lines)
	}
	allocations := lines[0].Allocations()
	restoresID, restores := allocations[0].RestoresAllocationID().Get()
	if len(allocations) != 1 || allocations[0].LotID() != purchase.Lines()[0].LotID() ||
		allocations[0].Quantity().Int64() != 50 || !restores || restoresID != originalAllocation.ID() {
		t.Fatalf("reversal allocations = %#v", allocations)
	}
	assertInventoryBalance(t, store, item.Item().ID(), 200, 20_000_000, reversal.ID().Int64())
}

func TestReversalStoreRejectsNonLatestTarget(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "reverse-ineligible.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	item := createCatalogItem(t, store, CreateItemInput{
		Name:         mustCatalogName(t, "Chocolate"),
		BaseUnit:     mustCatalogUnitCode(t, "g"),
		Capabilities: catalog.NewCapabilities(true, false, false),
		CreatedAt:    mustCatalogInstant(t, 1_000),
		UpdatedAt:    mustCatalogInstant(t, 1_000),
	})
	purchase := postReversalTestPurchase(t, store, item.Item().ID(), "purchase-not-latest", 100, 1_000)
	_, err := store.PostAdjustment(ctx, PostAdjustmentInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, "later-waste"),
		OccurredOn:     mustPurchaseDate(t, "2026-07-15"),
		PostedAt:       mustCatalogInstant(t, 3_000),
		Reason:         domain.ReasonWaste,
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
	if err != nil {
		t.Fatalf("post later adjustment: %v", err)
	}

	_, err = store.PostReversal(ctx, PostReversalInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "reverse-not-latest"),
		TargetDocumentID: purchase.ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-16"),
		PostedAt:         mustCatalogInstant(t, 4_000),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("reverse non-latest error = %v, want validation", err)
	}
}

func postReversalTestPurchase(
	t *testing.T,
	store *Store,
	itemID domain.ItemID,
	idempotencyKey string,
	quantity int64,
	totalMinor int64,
) PostedPurchaseDocument {
	t.Helper()
	posted, err := store.PostPurchase(context.Background(), PostPurchaseInput{
		IdempotencyKey: mustPurchaseIdempotencyKey(t, idempotencyKey),
		OccurredOn:     mustPurchaseDate(t, "2026-07-14"),
		PostedAt:       mustCatalogInstant(t, 2_000),
		Lines: []PostPurchaseLineInput{
			{
				ItemID:          itemID,
				Quantity:        mustPurchaseQuantity(t, quantity),
				EnteredUnit:     mustCatalogUnitCode(t, "g"),
				Conversion:      mustCatalogConversion(t, 1_000, 1),
				CommercialTotal: mustPurchaseMinorAmount(t, totalMinor),
				LotCode:         domain.Some(counterpartyText(t, "REV-LOT")),
			},
		},
	})
	if err != nil {
		t.Fatalf("post fixture purchase: %v", err)
	}
	return posted
}

func assertInventoryBalance(
	t *testing.T,
	store *Store,
	itemID domain.ItemID,
	wantQuantity int64,
	wantValue int64,
	wantLastDocumentID int64,
) {
	t.Helper()
	var quantity, value, lastDocumentID int64
	err := store.database.QueryRowContext(context.Background(), `
		SELECT quantity_atomic, inventory_value_micro, last_document_id
		FROM inventory_balances
		WHERE item_id = ?
	`, itemID.Int64()).Scan(&quantity, &value, &lastDocumentID)
	if err != nil {
		t.Fatalf("read balance: %v", err)
	}
	if quantity != wantQuantity || value != wantValue || lastDocumentID != wantLastDocumentID {
		t.Fatalf("balance = %d/%d/%d, want %d/%d/%d",
			quantity, value, lastDocumentID, wantQuantity, wantValue, wantLastDocumentID)
	}
}
