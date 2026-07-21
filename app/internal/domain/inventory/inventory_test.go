package inventory_test

import (
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	"github.com/jerobas/saas/internal/domain/inventory"
)

func TestBalanceProjectionAndCatalogViews(t *testing.T) {
	instant := must(domain.UTCInstantFromUnixMilli(1000))
	itemID := must(domain.NewItemID(1))
	quantity := must(domain.NewAtomicQuantity(2))
	value := must(domain.NewInventoryValue(5))
	balance := must(inventory.NewBalance(inventory.BalanceParams{
		ItemID: itemID, Quantity: quantity, Value: value, UpdatedAt: instant,
	}))
	average, ok := balance.AverageValuePerAtomicUnit()
	if !ok || average.Numerator() != 5 || average.Denominator() != 2 {
		t.Fatalf("average = %v, %t", average, ok)
	}
	snapshot := must(inventory.NewBalanceSnapshot(inventory.BalanceSnapshotParams{
		Balance: balance, ItemName: must(domain.NewUniqueName("Flour")),
		BaseUnit: must(domain.NewUnitCode("g")),
	}))
	listItem := must(inventory.NewBalanceListItem(inventory.BalanceListItemParams{
		Snapshot: snapshot, Capabilities: catalog.NewCapabilities(true, false, false),
		ReorderQuantity: domain.Some(must(domain.NewAtomicQuantity(10))),
	}))
	if listItem.Snapshot().ItemName().Display() != "Flour" || !listItem.Capabilities().Purchasable() {
		t.Fatalf("balance list item = %#v", listItem)
	}

	zero, _ := domain.NewAtomicQuantity(0)
	_, err := inventory.NewBalance(inventory.BalanceParams{
		ItemID: itemID, Quantity: zero, Value: value, UpdatedAt: instant,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("zero quantity/nonzero value error = %v", err)
	}
}

func TestLotTotalsSourceSnapshotAndInclusiveExpiry(t *testing.T) {
	origin := must(domain.ParseBusinessDate("2026-07-01"))
	expiry := must(domain.ParseBusinessDate("2026-07-14"))
	nextDay := must(domain.ParseBusinessDate("2026-07-15"))
	instant := must(domain.UTCInstantFromUnixMilli(1000))
	lot := must(inventory.NewLot(inventory.LotParams{
		ID: must(domain.NewInventoryLotID(1)), ItemID: must(domain.NewItemID(2)),
		SourceLineID:          must(domain.NewStockDocumentLineID(3)),
		SourcePostingSequence: must(domain.NewPostingSequence(4)),
		InitialQuantity:       must(domain.NewPositiveAtomicQuantity(10)),
		ConsumedQuantity:      must(domain.NewAtomicQuantity(7)),
		RestoredQuantity:      must(domain.NewAtomicQuantity(2)),
		AvailableQuantity:     must(domain.NewAtomicQuantity(5)),
		LotCode:               domain.Some(must(domain.NewNonEmptyText("LOT-1"))), OriginatedOn: origin,
		ExpiresOn: domain.Some(expiry), CreatedAt: instant,
	}))
	view := must(inventory.NewLotView(inventory.LotViewParams{
		Lot: lot, SourceDocumentID: must(domain.NewStockDocumentID(5)),
		SourceKind: domain.DocumentPurchase, SourceOccurredOn: origin,
	}))
	if view.SourceKind() != domain.DocumentPurchase || view.Lot().ConsumedQuantity().Int64() != 7 {
		t.Fatalf("lot view = %#v", view)
	}
	if lot.IsExpired(expiry) || lot.State(expiry) != domain.LotAvailable {
		t.Fatal("lot was not usable through inclusive expiry date")
	}
	if !lot.IsExpired(nextDay) || lot.State(nextDay) != domain.LotExpired {
		t.Fatal("lot did not expire after its inclusive date")
	}
	if lot.Cursor().PostingSequence().Int64() != 4 {
		t.Fatalf("lot cursor = %#v", lot.Cursor())
	}

	_, err := inventory.NewLot(inventory.LotParams{
		ID: must(domain.NewInventoryLotID(1)), ItemID: must(domain.NewItemID(2)),
		SourceLineID:          must(domain.NewStockDocumentLineID(3)),
		SourcePostingSequence: must(domain.NewPostingSequence(4)),
		InitialQuantity:       must(domain.NewPositiveAtomicQuantity(10)),
		ConsumedQuantity:      must(domain.NewAtomicQuantity(7)),
		RestoredQuantity:      must(domain.NewAtomicQuantity(2)),
		AvailableQuantity:     must(domain.NewAtomicQuantity(6)), OriginatedOn: origin, CreatedAt: instant,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("inconsistent lot totals error = %v", err)
	}
}

func TestAllocationAndLotMetadataView(t *testing.T) {
	instant := must(domain.UTCInstantFromUnixMilli(1000))
	allocation := must(inventory.NewAllocation(inventory.AllocationParams{
		ID: must(domain.NewLotAllocationID(1)), LineID: must(domain.NewStockDocumentLineID(2)),
		LotID: must(domain.NewInventoryLotID(3)), Quantity: must(domain.NewPositiveAtomicQuantity(4)),
		Effect: domain.AllocationConsumption, CreatedAt: instant,
	}))
	origin := must(domain.ParseBusinessDate("2026-07-01"))
	view := must(inventory.NewAllocationView(inventory.AllocationViewParams{
		Allocation: allocation, SourceLineID: must(domain.NewStockDocumentLineID(10)),
		LotInitialQuantity: must(domain.NewPositiveAtomicQuantity(20)), OriginatedOn: origin,
	}))
	if view.Allocation().Effect() != domain.AllocationConsumption || view.SourceLineID().Int64() != 10 {
		t.Fatalf("allocation view = %#v", view)
	}

	_, err := inventory.NewAllocation(inventory.AllocationParams{
		ID: must(domain.NewLotAllocationID(5)), LineID: must(domain.NewStockDocumentLineID(2)),
		LotID: must(domain.NewInventoryLotID(3)), Quantity: must(domain.NewPositiveAtomicQuantity(4)),
		Effect: domain.AllocationRestoration, CreatedAt: instant,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unlinked restoration error = %v", err)
	}
}

func TestLedgerViewPreservesEveryHistoricalSnapshot(t *testing.T) {
	instant := must(domain.UTCInstantFromUnixMilli(1000))
	date := must(domain.ParseBusinessDate("2026-07-14"))
	currency := must(domain.RestoreCurrency("BRL", 2))
	entry := must(inventory.NewLedgerEntry(inventory.LedgerEntryParams{
		LineID: must(domain.NewStockDocumentLineID(1)), DocumentID: must(domain.NewStockDocumentID(2)),
		PostingSequence: must(domain.NewPostingSequence(3)), LineOrder: must(domain.NewLineOrder(1)),
		Kind: domain.DocumentPurchase, OccurredOn: date, PostedAt: instant,
		ItemID: must(domain.NewItemID(4)), Direction: domain.DirectionIn,
		Quantity: must(domain.NewPositiveAtomicQuantity(5)), Value: must(domain.NewInventoryValue(600)),
		CommercialTotal: domain.Some(must(domain.NewMinorAmount(100))), Currency: currency,
	}))
	view := must(inventory.NewLedgerEntryView(inventory.LedgerEntryViewParams{
		Entry: entry, EnteredUnit: must(domain.NewUnitCode("kg")),
		EnteredPackaging: domain.Some(must(domain.NewNonEmptyText("bag"))),
		Conversion:       must(domain.NewUnitConversion(1_000_000, 1)),
		IdempotencyKey:   must(domain.NewIdempotencyKey("command-1")),
		CounterpartyID:   domain.Some(must(domain.NewCounterpartyID(9))),
		CounterpartyName: domain.Some(must(domain.NewDisplayName("Supplier"))),
	}))
	if view.Entry().Cursor().PostingSequence().Int64() != 3 || view.EnteredUnit().String() != "kg" || view.IdempotencyKey().String() != "command-1" {
		t.Fatalf("ledger view = %#v", view)
	}
	if total, ok := view.Entry().CommercialTotal().Get(); !ok || total.Int64() != 100 {
		t.Fatalf("commercial snapshot = %#v, %t", total, ok)
	}

	_, err := inventory.NewLedgerEntryView(inventory.LedgerEntryViewParams{
		Entry: entry, EnteredUnit: must(domain.NewUnitCode("kg")),
		Conversion:     must(domain.NewUnitConversion(1_000_000, 1)),
		IdempotencyKey: must(domain.NewIdempotencyKey("command-1")),
		CounterpartyID: domain.Some(must(domain.NewCounterpartyID(9))),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unpaired counterparty snapshot error = %v", err)
	}

	adjustment := must(inventory.NewLedgerEntry(inventory.LedgerEntryParams{
		LineID: must(domain.NewStockDocumentLineID(10)), DocumentID: must(domain.NewStockDocumentID(11)),
		PostingSequence: must(domain.NewPostingSequence(12)), LineOrder: must(domain.NewLineOrder(1)),
		Kind: domain.DocumentAdjustment, OccurredOn: date, PostedAt: instant,
		ItemID: must(domain.NewItemID(4)), Direction: domain.DirectionOut,
		Quantity: must(domain.NewPositiveAtomicQuantity(1)), Value: must(domain.NewInventoryValue(1)), Currency: currency,
	}))
	_, err = inventory.NewLedgerEntryView(inventory.LedgerEntryViewParams{
		Entry: adjustment, EnteredUnit: must(domain.NewUnitCode("g")),
		Conversion:     must(domain.NewUnitConversion(1000, 1)),
		IdempotencyKey: must(domain.NewIdempotencyKey("adjust-1")),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing adjustment reason error = %v", err)
	}
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
