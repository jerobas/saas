package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/inventory"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

func TestInventoryStoreGetsAndListsBalances(t *testing.T) {
	fixture := seedInventoryStoreFixture(t)
	ctx := context.Background()

	snapshot, err := fixture.store.GetInventoryBalance(ctx, mustItemID(t, fixture.itemID))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Balance().Quantity().Int64() != 220 {
		t.Fatalf("quantity = %d, want 220", snapshot.Balance().Quantity().Int64())
	}
	if snapshot.Balance().Value().Int64() != 2200 {
		t.Fatalf("value = %d, want 2200", snapshot.Balance().Value().Int64())
	}
	if snapshot.ItemName().Display() != "Butter" || snapshot.ItemName().Key() != "butter" {
		t.Fatalf("item name = %q/%q", snapshot.ItemName().Display(), snapshot.ItemName().Key())
	}
	if snapshot.BaseUnit().String() != "g" {
		t.Fatalf("base unit = %q, want g", snapshot.BaseUnit().String())
	}
	lastDocumentID, ok := snapshot.Balance().LastDocumentID().Get()
	if !ok || lastDocumentID.Int64() != fixture.lastDocumentID {
		t.Fatalf("last document = %v/%v, want %d", lastDocumentID, ok, fixture.lastDocumentID)
	}

	fixture.insertItem(t, "Zinc", "zinc", false)
	fixture.insertItem(t, "Archived", "archived", true)
	firstPage, err := fixture.store.ListInventoryBalances(ctx, InventoryBalanceListParams{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(firstPage) != 1 || firstPage[0].Snapshot().ItemName().Display() != "Butter" {
		t.Fatalf("first page = %v", balanceNames(firstPage))
	}
	cursor, err := InventoryBalanceCursorFor(firstPage[0])
	if err != nil {
		t.Fatal(err)
	}
	secondPage, err := fixture.store.ListInventoryBalances(ctx, InventoryBalanceListParams{
		After: domain.Some(cursor), Limit: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(secondPage) != 1 || secondPage[0].Snapshot().ItemName().Display() != "Zinc" {
		t.Fatalf("second page = %v", balanceNames(secondPage))
	}

	searched, err := fixture.store.ListInventoryBalances(ctx, InventoryBalanceListParams{
		Search: someSearchText(t, "BUT"), Limit: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(searched) != 1 || searched[0].Snapshot().Balance().ItemID().Int64() != fixture.itemID {
		t.Fatalf("search results = %v", balanceNames(searched))
	}
	if quantity, ok := searched[0].ReorderQuantity().Get(); !ok || quantity.Int64() != 25 {
		t.Fatalf("reorder quantity = %v/%v, want 25", quantity, ok)
	}
	capabilities := searched[0].Capabilities()
	if !capabilities.Purchasable() || !capabilities.Sellable() || capabilities.Producible() {
		t.Fatalf("capabilities = purchase:%v produce:%v sell:%v",
			capabilities.Purchasable(), capabilities.Producible(), capabilities.Sellable())
	}

	withArchived, err := fixture.store.ListInventoryBalances(ctx, InventoryBalanceListParams{
		IncludeArchived: true, Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := balanceNames(withArchived); fmt.Sprint(got) != "[Archived Butter Zinc]" {
		t.Fatalf("all balance names = %v", got)
	}

	empty, err := fixture.store.ListInventoryBalances(ctx, InventoryBalanceListParams{
		Search: someSearchText(t, "does not exist"), Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty result = %#v, want non-nil empty slice", empty)
	}
}

func TestInventoryStoreLotFactsPreserveGrossRestorationMath(t *testing.T) {
	fixture := seedInventoryStoreFixture(t)
	lots, err := fixture.store.ListItemLotFacts(context.Background(), mustItemID(t, fixture.itemID))
	if err != nil {
		t.Fatal(err)
	}
	if lots == nil {
		t.Fatal("lot facts must be a non-nil slice")
	}

	var restoredLotFound bool
	for _, view := range lots {
		lot := view.Lot()
		if lot.ID().Int64() != fixture.restoredLotID {
			continue
		}
		restoredLotFound = true
		if lot.InitialQuantity().Int64() != 100 || lot.ConsumedQuantity().Int64() != 90 ||
			lot.RestoredQuantity().Int64() != 60 || lot.AvailableQuantity().Int64() != 70 {
			t.Fatalf("lot quantities initial/consumed/restored/available = %d/%d/%d/%d",
				lot.InitialQuantity().Int64(), lot.ConsumedQuantity().Int64(),
				lot.RestoredQuantity().Int64(), lot.AvailableQuantity().Int64())
		}
		if view.SourceKind() != domain.DocumentPurchase || view.SourceDocumentID().Int64() != fixture.firstPurchaseID {
			t.Fatalf("source = %s/%d", view.SourceKind(), view.SourceDocumentID().Int64())
		}
	}
	if !restoredLotFound {
		t.Fatalf("restored lot %d not found", fixture.restoredLotID)
	}
}

func TestInventoryStoreEligibleLotsUseInclusiveFEFOOrdering(t *testing.T) {
	fixture := seedInventoryStoreFixture(t)
	businessDate, err := domain.ParseBusinessDate("2026-07-14")
	if err != nil {
		t.Fatal(err)
	}
	lots, err := fixture.store.ListEligibleFEFOLots(
		context.Background(), mustItemID(t, fixture.itemID), businessDate,
	)
	if err != nil {
		t.Fatal(err)
	}
	if lots == nil {
		t.Fatal("eligible lots must be a non-nil slice")
	}
	got := make([]int64, 0, len(lots))
	for _, view := range lots {
		got = append(got, view.Lot().ID().Int64())
	}
	want := []int64{
		fixture.restoredLotID,
		fixture.todayLotID,
		fixture.futureLotID,
		fixture.noExpiryLotIDs[0],
		fixture.noExpiryLotIDs[1],
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("FEFO lot ids = %v, want %v", got, want)
	}
	for _, id := range got {
		if id == fixture.expiredLotID {
			t.Fatalf("expired lot %d was eligible", id)
		}
	}
	firstExpiry, ok := lots[0].Lot().ExpiresOn().Get()
	if !ok || firstExpiry.String() != businessDate.String() {
		t.Fatalf("first expiry = %v/%v, want inclusive %s", firstExpiry, ok, businessDate)
	}
}

func TestInventoryStoreLedgerUsesCompleteKeysetCursor(t *testing.T) {
	fixture := seedInventoryStoreFixture(t)
	itemID := mustItemID(t, fixture.itemID)
	ctx := context.Background()

	first, err := fixture.store.ListItemLedgerPage(ctx, ItemLedgerPageParams{ItemID: itemID, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || first[0].Entry().LineID().Int64() != fixture.lastDocumentLineIDs[0] {
		t.Fatalf("first ledger line = %v, want %d", ledgerLineIDs(first), fixture.lastDocumentLineIDs[0])
	}
	second, err := fixture.store.ListItemLedgerPage(ctx, ItemLedgerPageParams{
		ItemID: itemID, After: domain.Some(first[0].Entry().Cursor()), Limit: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 1 || second[0].Entry().LineID().Int64() != fixture.lastDocumentLineIDs[1] {
		t.Fatalf("second ledger line = %v, want %d", ledgerLineIDs(second), fixture.lastDocumentLineIDs[1])
	}
	if second[0].Entry().PostingSequence().Int64() != first[0].Entry().PostingSequence().Int64() ||
		second[0].Entry().LineOrder().Int64() != 2 {
		t.Fatalf("same-document keyset did not advance by line order")
	}
	third, err := fixture.store.ListItemLedgerPage(ctx, ItemLedgerPageParams{
		ItemID: itemID, After: domain.Some(second[0].Entry().Cursor()), Limit: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(third) != 1 || third[0].Entry().PostingSequence().Int64() != 7 {
		t.Fatalf("third ledger posting sequence = %v, want 7", ledgerSequences(third))
	}
	if first[0].EnteredUnit().String() != "g" || first[0].IdempotencyKey().String() != "purchase-8" {
		t.Fatalf("ledger snapshot unit/key = %q/%q", first[0].EnteredUnit(), first[0].IdempotencyKey())
	}
}

func TestInventoryStoreMapsConsumptionAndRestorationAllocations(t *testing.T) {
	fixture := seedInventoryStoreFixture(t)
	ctx := context.Background()

	consumptions, err := fixture.store.ListLineAllocations(ctx, mustLineID(t, fixture.firstSaleLineID))
	if err != nil {
		t.Fatal(err)
	}
	if len(consumptions) != 1 {
		t.Fatalf("consumption allocations = %d, want 1", len(consumptions))
	}
	consumption := consumptions[0]
	if consumption.Allocation().Effect() != domain.AllocationConsumption ||
		consumption.Allocation().RestoresAllocationID().IsSome() ||
		consumption.Allocation().Quantity().Int64() != 60 {
		t.Fatalf("consumption allocation mapped incorrectly")
	}
	if consumption.SourceLineID().Int64() != fixture.firstPurchaseLineID ||
		consumption.LotInitialQuantity().Int64() != 100 {
		t.Fatalf("consumption lot snapshot mapped incorrectly")
	}

	restorations, err := fixture.store.ListLineAllocations(ctx, mustLineID(t, fixture.reversalLineID))
	if err != nil {
		t.Fatal(err)
	}
	if len(restorations) != 1 {
		t.Fatalf("restoration allocations = %d, want 1", len(restorations))
	}
	restoration := restorations[0].Allocation()
	restoredID, ok := restoration.RestoresAllocationID().Get()
	if restoration.Effect() != domain.AllocationRestoration || !ok ||
		restoredID.Int64() != fixture.firstAllocationID {
		t.Fatalf("restoration effect/reference = %s/%v/%d",
			restoration.Effect(), ok, restoredID.Int64())
	}

	empty, err := fixture.store.ListLineAllocations(ctx, mustLineID(t, fixture.unallocatedLineID))
	if err != nil {
		t.Fatal(err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty allocations = %#v, want non-nil empty slice", empty)
	}
}

func TestInventoryStoreMapsNotFoundCancellationAndCorruptData(t *testing.T) {
	fixture := seedInventoryStoreFixture(t)
	validDate, err := domain.ParseBusinessDate("2026-07-14")
	if err != nil {
		t.Fatal(err)
	}
	var zeroItemID domain.ItemID
	var zeroLineID domain.StockDocumentLineID
	invalidIDReads := []func() error{
		func() error {
			_, readErr := fixture.store.GetInventoryBalance(context.Background(), zeroItemID)
			return readErr
		},
		func() error {
			_, readErr := fixture.store.ListItemLotFacts(context.Background(), zeroItemID)
			return readErr
		},
		func() error {
			_, readErr := fixture.store.ListEligibleFEFOLots(context.Background(), zeroItemID, validDate)
			return readErr
		},
		func() error {
			_, readErr := fixture.store.ListItemLedgerPage(context.Background(), ItemLedgerPageParams{ItemID: zeroItemID})
			return readErr
		},
		func() error {
			_, readErr := fixture.store.ListLineAllocations(context.Background(), zeroLineID)
			return readErr
		},
	}
	for index, read := range invalidIDReads {
		if readErr := read(); !errors.Is(readErr, domain.ErrValidation) {
			t.Fatalf("zero-id read %d error = %v, want ErrValidation", index, readErr)
		}
	}
	if _, err := fixture.store.ListInventoryBalances(context.Background(), InventoryBalanceListParams{
		Limit: inventoryMaximumPageSize + 1,
	}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("oversized balance page error = %v, want ErrValidation", err)
	}
	if _, err := fixture.store.ListItemLedgerPage(context.Background(), ItemLedgerPageParams{
		ItemID: mustItemID(t, fixture.itemID), Limit: inventoryMaximumPageSize + 1,
	}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("oversized ledger page error = %v, want ErrValidation", err)
	}
	if defaultPage, err := fixture.store.ListInventoryBalances(context.Background(), InventoryBalanceListParams{}); err != nil || defaultPage == nil {
		t.Fatalf("default balance page = %#v, error = %v", defaultPage, err)
	}

	missing := mustItemID(t, 999999)
	if _, err := fixture.store.GetInventoryBalance(context.Background(), missing); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing balance error = %v, want ErrNotFound", err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fixture.store.GetInventoryBalance(canceled, mustItemID(t, fixture.itemID)); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled read error = %v, want context.Canceled", err)
	}

	fixture.exec(t, `UPDATE items SET normalized_name = 'not-butter' WHERE id = ?`, fixture.itemID)
	if _, err := fixture.store.GetInventoryBalance(context.Background(), mustItemID(t, fixture.itemID)); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("corrupt mapping error = %v, want ErrCorruptData", err)
	}
}

func TestInventoryMappingRejectsNoncanonicalPersistedSnapshots(t *testing.T) {
	fixture := seedInventoryStoreFixture(t)
	ctx := context.Background()
	itemID := mustItemID(t, fixture.itemID)

	balanceRow, err := fixture.store.queries.GetInventoryBalance(ctx, fixture.itemID)
	if err != nil {
		t.Fatal(err)
	}
	validBalance := inventoryBalanceFields{
		itemID: balanceRow.ItemID, itemName: balanceRow.ItemName,
		itemNormalizedName: balanceRow.ItemNormalizedName, baseUnitCode: balanceRow.BaseUnitCode,
		itemArchivedAtMS: balanceRow.ItemArchivedAtMs, quantityAtomic: balanceRow.QuantityAtomic,
		inventoryValueMicro: balanceRow.InventoryValueMicro, lastDocumentID: balanceRow.LastDocumentID,
		updatedAtMS: balanceRow.UpdatedAtMs,
	}
	for name, mutate := range map[string]func(*inventoryBalanceFields){
		"display name whitespace": func(fields *inventoryBalanceFields) { fields.itemName = " Butter " },
		"unit whitespace":         func(fields *inventoryBalanceFields) { fields.baseUnitCode = " g " },
	} {
		t.Run("balance "+name, func(t *testing.T) {
			fields := validBalance
			mutate(&fields)
			if _, err := mapInventoryBalanceSnapshot(fields); err == nil {
				t.Fatal("noncanonical balance snapshot was accepted")
			}
		})
	}

	ledgerRows, err := fixture.store.queries.ListItemLedgerPage(ctx, sqlcgen.ListItemLedgerPageParams{
		ItemID: itemID.Int64(), LimitCount: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ledgerRows) != 1 {
		t.Fatalf("ledger fixture rows = %d, want 1", len(ledgerRows))
	}
	validLedger := ledgerRows[0]
	for name, mutate := range map[string]func(*sqlcgen.ListItemLedgerPageRow){
		"entered unit whitespace": func(row *sqlcgen.ListItemLedgerPageRow) { row.EnteredUnitCode = " g " },
		"unreduced conversion": func(row *sqlcgen.ListItemLedgerPageRow) {
			row.ConversionNumeratorAtomic = 2_000
			row.ConversionDenominator = 2
		},
		"idempotency whitespace": func(row *sqlcgen.ListItemLedgerPageRow) { row.IdempotencyKey = " purchase-8 " },
		"packaging whitespace": func(row *sqlcgen.ListItemLedgerPageRow) {
			row.EnteredPackagingName = sql.NullString{String: " Sack ", Valid: true}
		},
		"counterparty whitespace": func(row *sqlcgen.ListItemLedgerPageRow) {
			row.CounterpartyID = sql.NullInt64{Int64: 1, Valid: true}
			row.CounterpartyName = sql.NullString{String: " Acme ", Valid: true}
		},
	} {
		t.Run("ledger "+name, func(t *testing.T) {
			row := validLedger
			mutate(&row)
			if _, err := mapLedgerEntryView(row); err == nil {
				t.Fatal("noncanonical ledger snapshot was accepted")
			}
		})
	}
}

type inventoryStoreFixture struct {
	db                  *database.Database
	store               *Store
	itemID              int64
	firstPurchaseID     int64
	firstPurchaseLineID int64
	restoredLotID       int64
	firstSaleLineID     int64
	firstAllocationID   int64
	reversalLineID      int64
	unallocatedLineID   int64
	expiredLotID        int64
	todayLotID          int64
	futureLotID         int64
	noExpiryLotIDs      [2]int64
	lastDocumentID      int64
	lastDocumentLineIDs [2]int64
}

func seedInventoryStoreFixture(t *testing.T) *inventoryStoreFixture {
	t.Helper()
	db, err := database.NewDatabase(filepath.Join(t.TempDir(), "inventory.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("close database: %v", closeErr)
		}
	})
	fixture := &inventoryStoreFixture{db: db, store: NewStore(db)}
	fixture.itemID = fixture.insertItem(t, "Butter", "butter", false)

	fixture.firstPurchaseID = fixture.insertDocument(t, "PURCHASE", 1, nil, nil, "purchase-1")
	fixture.firstPurchaseLineID = fixture.insertLine(
		t, fixture.firstPurchaseID, 1, "IN", 100, 1000, 100, nil,
	)
	fixture.restoredLotID = fixture.insertLot(
		t, fixture.firstPurchaseLineID, 100, "LOT-TODAY-A", "2026-07-14", 1,
	)

	firstSaleID := fixture.insertDocument(t, "SALE", 2, nil, nil, "sale-2")
	fixture.firstSaleLineID = fixture.insertLine(t, firstSaleID, 1, "OUT", 60, 600, 90, nil)
	fixture.firstAllocationID = fixture.insertAllocation(t, fixture.firstSaleLineID, fixture.restoredLotID, 60, nil, 2)

	reversalID := fixture.insertDocument(t, "REVERSAL", 3, "EXACT_REVERSAL", firstSaleID, "reversal-3")
	fixture.reversalLineID = fixture.insertLine(
		t, reversalID, 1, "IN", 60, 600, 90, fixture.firstSaleLineID,
	)
	fixture.insertAllocation(t, fixture.reversalLineID, fixture.restoredLotID, 60, fixture.firstAllocationID, 3)

	secondSaleID := fixture.insertDocument(t, "SALE", 4, nil, nil, "sale-4")
	secondSaleLineID := fixture.insertLine(t, secondSaleID, 1, "OUT", 30, 300, 45, nil)
	fixture.insertAllocation(t, secondSaleLineID, fixture.restoredLotID, 30, nil, 4)

	expiredDocumentID := fixture.insertDocument(t, "PURCHASE", 5, nil, nil, "purchase-5")
	expiredLineID := fixture.insertLine(t, expiredDocumentID, 1, "IN", 10, 100, 10, nil)
	fixture.expiredLotID = fixture.insertLot(t, expiredLineID, 10, "LOT-EXPIRED", "2026-07-13", 5)

	todayDocumentID := fixture.insertDocument(t, "PURCHASE", 6, nil, nil, "purchase-6")
	todayLineID := fixture.insertLine(t, todayDocumentID, 1, "IN", 20, 200, 20, nil)
	fixture.todayLotID = fixture.insertLot(t, todayLineID, 20, "LOT-TODAY-B", "2026-07-14", 6)

	futureDocumentID := fixture.insertDocument(t, "PURCHASE", 7, nil, nil, "purchase-7")
	futureLineID := fixture.insertLine(t, futureDocumentID, 1, "IN", 30, 300, 30, nil)
	fixture.futureLotID = fixture.insertLot(t, futureLineID, 30, "LOT-FUTURE", "2026-07-15", 7)

	fixture.lastDocumentID = fixture.insertDocument(t, "PURCHASE", 8, nil, nil, "purchase-8")
	fixture.lastDocumentLineIDs[0] = fixture.insertLine(t, fixture.lastDocumentID, 1, "IN", 40, 400, 40, nil)
	fixture.noExpiryLotIDs[0] = fixture.insertLot(t, fixture.lastDocumentLineIDs[0], 40, "LOT-OPEN-A", "", 8)
	fixture.lastDocumentLineIDs[1] = fixture.insertLine(t, fixture.lastDocumentID, 2, "IN", 50, 500, 50, nil)
	fixture.noExpiryLotIDs[1] = fixture.insertLot(t, fixture.lastDocumentLineIDs[1], 50, "LOT-OPEN-B", "", 8)
	fixture.unallocatedLineID = fixture.lastDocumentLineIDs[0]

	fixture.exec(t, `
		UPDATE inventory_balances
		SET quantity_atomic = 220,
			inventory_value_micro = 2200,
			last_document_id = ?,
			updated_at_ms = 900
		WHERE item_id = ?
	`, fixture.lastDocumentID, fixture.itemID)
	return fixture
}

func (f *inventoryStoreFixture) insertItem(t *testing.T, name, normalizedName string, archived bool) int64 {
	t.Helper()
	updatedAt := int64(1)
	if archived {
		updatedAt = 2
	}
	result := f.exec(t, `
		INSERT INTO items (
			name, normalized_name, base_unit_code,
			is_purchasable, is_producible, is_sellable,
			reorder_quantity_atomic, created_at_ms, updated_at_ms, archived_at_ms
		) VALUES (?, ?, 'g', 1, 0, 1, 25, 1, ?, ?)
	`, name, normalizedName, updatedAt, nullableArchivedAt(archived))
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *inventoryStoreFixture) insertDocument(
	t *testing.T,
	kind string,
	postingSequence int64,
	reason any,
	reversesDocument any,
	idempotencyKey string,
) int64 {
	t.Helper()
	result := f.exec(t, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, occurred_on, posted_at_ms,
			currency_code, currency_minor_digits, reason_code, reverses_document_id
		) VALUES (?, ?, ?, '2026-07-14', ?, 'BRL', 2, ?, ?)
	`, kind, idempotencyKey, postingSequence, postingSequence*100, reason, reversesDocument)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *inventoryStoreFixture) insertLine(
	t *testing.T,
	documentID, lineOrder int64,
	direction string,
	quantity, inventoryValue int64,
	commercialTotal any,
	reversesLine any,
) int64 {
	t.Helper()
	result := f.exec(t, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, conversion_numerator_atomic, conversion_denominator,
			inventory_value_micro, commercial_total_minor, reverses_line_id
		) VALUES (?, ?, ?, ?, ?, 'g', 1000, 1, ?, ?, ?)
	`, documentID, lineOrder, f.itemID, direction, quantity, inventoryValue, commercialTotal, reversesLine)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *inventoryStoreFixture) insertLot(
	t *testing.T,
	sourceLineID, quantity int64,
	lotCode, expiresOn string,
	createdAt int64,
) int64 {
	t.Helper()
	var expiry any
	if expiresOn != "" {
		expiry = expiresOn
	}
	result := f.exec(t, `
		INSERT INTO inventory_lots (
			item_id, source_line_id, initial_quantity_atomic,
			lot_code, originated_on, expires_on, created_at_ms
		) VALUES (?, ?, ?, ?, '2026-07-01', ?, ?)
	`, f.itemID, sourceLineID, quantity, lotCode, expiry, createdAt)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *inventoryStoreFixture) insertAllocation(
	t *testing.T,
	lineID, lotID, quantity int64,
	restoresAllocation any,
	createdAt int64,
) int64 {
	t.Helper()
	result := f.exec(t, `
		INSERT INTO lot_allocations (
			line_id, lot_id, quantity_atomic, restores_allocation_id, created_at_ms
		) VALUES (?, ?, ?, ?, ?)
	`, lineID, lotID, quantity, restoresAllocation, createdAt)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *inventoryStoreFixture) exec(t *testing.T, statement string, args ...any) interface {
	LastInsertId() (int64, error)
} {
	t.Helper()
	result, err := f.db.ExecContext(context.Background(), statement, args...)
	if err != nil {
		t.Fatalf("execute fixture SQL: %v\n%s", err, statement)
	}
	return result
}

func nullableArchivedAt(archived bool) any {
	if archived {
		return int64(2)
	}
	return nil
}

func mustItemID(t *testing.T, raw int64) domain.ItemID {
	t.Helper()
	id, err := domain.NewItemID(raw)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustLineID(t *testing.T, raw int64) domain.StockDocumentLineID {
	t.Helper()
	id, err := domain.NewStockDocumentLineID(raw)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func someSearchText(t *testing.T, raw string) domain.Option[domain.NonEmptyText] {
	t.Helper()
	value, err := domain.NewNonEmptyText(raw)
	if err != nil {
		t.Fatal(err)
	}
	return domain.Some(value)
}

func balanceNames(items []inventory.BalanceListItem) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Snapshot().ItemName().Display())
	}
	return names
}

func ledgerLineIDs(entries []inventory.LedgerEntryView) []int64 {
	ids := make([]int64, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.Entry().LineID().Int64())
	}
	return ids
}

func ledgerSequences(entries []inventory.LedgerEntryView) []int64 {
	sequences := make([]int64, 0, len(entries))
	for _, entry := range entries {
		sequences = append(sequences, entry.Entry().PostingSequence().Int64())
	}
	return sequences
}
