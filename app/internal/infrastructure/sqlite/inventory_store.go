package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	"github.com/jerobas/saas/internal/domain/inventory"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

const (
	getInventoryBalanceOperation   = "get inventory balance"
	listInventoryBalancesOperation = "list inventory balances"
	listItemLotFactsOperation      = "list item lot facts"
	listEligibleFEFOLotsOperation  = "list eligible FEFO lots"
	listItemLedgerPageOperation    = "list item ledger page"
	listLineAllocationsOperation   = "list line allocations"
	inventoryDefaultPageSize       = int64(50)
	inventoryMaximumPageSize       = int64(200)
)

// InventoryBalanceCursor is the stable keyset boundary for the balance list.
// It can be recreated directly from the last returned BalanceListItem.
type InventoryBalanceCursor struct {
	itemName domain.UniqueName
	itemID   domain.ItemID
}

func NewInventoryBalanceCursor(itemName domain.UniqueName, itemID domain.ItemID) (InventoryBalanceCursor, error) {
	if itemName.Key() == "" {
		return InventoryBalanceCursor{}, domain.Invalid("item_name", domain.ViolationRequired, "")
	}
	if itemID.IsZero() {
		return InventoryBalanceCursor{}, domain.Invalid("item_id", domain.ViolationRequired, "")
	}
	return InventoryBalanceCursor{itemName: itemName, itemID: itemID}, nil
}

func InventoryBalanceCursorFor(item inventory.BalanceListItem) (InventoryBalanceCursor, error) {
	snapshot := item.Snapshot()
	return NewInventoryBalanceCursor(snapshot.ItemName(), snapshot.Balance().ItemID())
}

func (c InventoryBalanceCursor) ItemName() domain.UniqueName { return c.itemName }
func (c InventoryBalanceCursor) ItemID() domain.ItemID       { return c.itemID }

type InventoryBalanceListParams struct {
	IncludeArchived bool
	Search          domain.Option[domain.NonEmptyText]
	After           domain.Option[InventoryBalanceCursor]
	Limit           int64
}

type ItemLedgerPageParams struct {
	ItemID domain.ItemID
	After  domain.Option[inventory.LedgerCursor]
	Limit  int64
}

func (s *Store) GetInventoryBalance(ctx context.Context, itemID domain.ItemID) (inventory.BalanceSnapshot, error) {
	if itemID.IsZero() {
		return inventory.BalanceSnapshot{}, domain.Invalid("item_id", domain.ViolationRequired, "")
	}
	row, err := s.queries.GetInventoryBalance(ctx, itemID.Int64())
	if err != nil {
		return inventory.BalanceSnapshot{}, classifyError(getInventoryBalanceOperation, err)
	}
	snapshot, err := mapInventoryBalanceSnapshot(inventoryBalanceFields{
		itemID: row.ItemID, itemName: row.ItemName, itemNormalizedName: row.ItemNormalizedName,
		baseUnitCode: row.BaseUnitCode, itemArchivedAtMS: row.ItemArchivedAtMs,
		quantityAtomic: row.QuantityAtomic, inventoryValueMicro: row.InventoryValueMicro,
		lastDocumentID: row.LastDocumentID, updatedAtMS: row.UpdatedAtMs,
	})
	if err != nil {
		return inventory.BalanceSnapshot{}, corruptDataError(getInventoryBalanceOperation, err)
	}
	return snapshot, nil
}

func (s *Store) ListInventoryBalances(ctx context.Context, params InventoryBalanceListParams) ([]inventory.BalanceListItem, error) {
	limit, err := inventoryPageSize(params.Limit)
	if err != nil {
		return nil, err
	}

	searchKey := ""
	if search, ok := params.Search.Get(); ok {
		_, normalized, err := domain.NormalizeDisplayAndKey(search.String())
		if err != nil {
			return nil, err
		}
		searchKey = normalized
	}

	afterName, afterItemID := "", int64(0)
	if cursor, ok := params.After.Get(); ok {
		if cursor.ItemName().Key() == "" || cursor.ItemID().IsZero() {
			return nil, domain.Invalid("after", domain.ViolationInvalidFormat, "")
		}
		afterName, afterItemID = cursor.ItemName().Key(), cursor.ItemID().Int64()
	}
	includeArchived := int64(0)
	if params.IncludeArchived {
		includeArchived = 1
	}

	rows, err := s.queries.ListInventoryBalances(ctx, sqlcgen.ListInventoryBalancesParams{
		IncludeArchived: includeArchived, SearchKey: searchKey,
		AfterNormalizedName: afterName, AfterItemID: afterItemID, LimitCount: limit,
	})
	if err != nil {
		return nil, classifyError(listInventoryBalancesOperation, err)
	}

	items := make([]inventory.BalanceListItem, 0, len(rows))
	for index, row := range rows {
		snapshot, mapErr := mapInventoryBalanceSnapshot(inventoryBalanceFields{
			itemID: row.ItemID, itemName: row.ItemName, itemNormalizedName: row.ItemNormalizedName,
			baseUnitCode: row.BaseUnitCode, itemArchivedAtMS: row.ItemArchivedAtMs,
			quantityAtomic: row.QuantityAtomic, inventoryValueMicro: row.InventoryValueMicro,
			lastDocumentID: row.LastDocumentID, updatedAtMS: row.UpdatedAtMs,
		})
		if mapErr != nil {
			return nil, corruptInventoryRow(listInventoryBalancesOperation, index, mapErr)
		}
		purchasable, mapErr := sqliteBool(row.IsPurchasable)
		if mapErr != nil {
			return nil, corruptInventoryRow(listInventoryBalancesOperation, index, mapErr)
		}
		producible, mapErr := sqliteBool(row.IsProducible)
		if mapErr != nil {
			return nil, corruptInventoryRow(listInventoryBalancesOperation, index, mapErr)
		}
		sellable, mapErr := sqliteBool(row.IsSellable)
		if mapErr != nil {
			return nil, corruptInventoryRow(listInventoryBalancesOperation, index, mapErr)
		}
		reorderQuantity, mapErr := optionalAtomicQuantity(row.ReorderQuantityAtomic)
		if mapErr != nil {
			return nil, corruptInventoryRow(listInventoryBalancesOperation, index, mapErr)
		}
		item, mapErr := inventory.NewBalanceListItem(inventory.BalanceListItemParams{
			Snapshot:        snapshot,
			Capabilities:    catalog.NewCapabilities(purchasable, producible, sellable),
			ReorderQuantity: reorderQuantity,
		})
		if mapErr != nil {
			return nil, corruptInventoryRow(listInventoryBalancesOperation, index, mapErr)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Store) ListItemLotFacts(ctx context.Context, itemID domain.ItemID) ([]inventory.LotView, error) {
	if itemID.IsZero() {
		return nil, domain.Invalid("item_id", domain.ViolationRequired, "")
	}
	rows, err := s.queries.ListItemLotFacts(ctx, itemID.Int64())
	if err != nil {
		return nil, classifyError(listItemLotFactsOperation, err)
	}
	lots := make([]inventory.LotView, 0, len(rows))
	for index, row := range rows {
		lot, mapErr := mapLotView(lotViewFields{
			id: row.ID, itemID: row.ItemID, sourceLineID: row.SourceLineID,
			initialQuantity: row.InitialQuantityAtomic, consumedQuantity: row.ConsumedQuantityAtomic,
			restoredQuantity: row.RestoredQuantityAtomic, availableQuantity: row.RemainingQuantityAtomic,
			lotCode: row.LotCode, originatedOn: row.OriginatedOn, expiresOn: row.ExpiresOn,
			createdAtMS: row.CreatedAtMs, sourceDocumentID: row.SourceDocumentID,
			sourceKind: row.SourceDocumentKind, sourcePostingSequence: row.SourcePostingSequence,
			sourceOccurredOn: row.SourceOccurredOn,
		})
		if mapErr != nil {
			return nil, corruptInventoryRow(listItemLotFactsOperation, index, mapErr)
		}
		lots = append(lots, lot)
	}
	return lots, nil
}

func (s *Store) ListEligibleFEFOLots(ctx context.Context, itemID domain.ItemID, on domain.BusinessDate) ([]inventory.LotView, error) {
	if itemID.IsZero() {
		return nil, domain.Invalid("item_id", domain.ViolationRequired, "")
	}
	if on.IsZero() {
		return nil, domain.Invalid("business_date", domain.ViolationRequired, "")
	}
	rows, err := s.queries.ListEligibleFEFOLots(ctx, sqlcgen.ListEligibleFEFOLotsParams{
		BusinessDate: on.String(), ItemID: itemID.Int64(),
	})
	if err != nil {
		return nil, classifyError(listEligibleFEFOLotsOperation, err)
	}
	lots := make([]inventory.LotView, 0, len(rows))
	for index, row := range rows {
		lot, mapErr := mapLotView(lotViewFields{
			id: row.ID, itemID: row.ItemID, sourceLineID: row.SourceLineID,
			initialQuantity: row.InitialQuantityAtomic, consumedQuantity: row.ConsumedQuantityAtomic,
			restoredQuantity: row.RestoredQuantityAtomic, availableQuantity: row.RemainingQuantityAtomic,
			lotCode: row.LotCode, originatedOn: row.OriginatedOn, expiresOn: row.ExpiresOn,
			createdAtMS: row.CreatedAtMs, sourceDocumentID: row.SourceDocumentID,
			sourceKind: row.SourceDocumentKind, sourcePostingSequence: row.SourcePostingSequence,
			sourceOccurredOn: row.SourceOccurredOn,
		})
		if mapErr != nil {
			return nil, corruptInventoryRow(listEligibleFEFOLotsOperation, index, mapErr)
		}
		lots = append(lots, lot)
	}
	return lots, nil
}

func (s *Store) ListItemLedgerPage(ctx context.Context, params ItemLedgerPageParams) ([]inventory.LedgerEntryView, error) {
	if params.ItemID.IsZero() {
		return nil, domain.Invalid("item_id", domain.ViolationRequired, "")
	}
	limit, err := inventoryPageSize(params.Limit)
	if err != nil {
		return nil, err
	}
	beforeSequence, afterLineOrder, afterLineID := int64(0), int64(0), int64(0)
	if cursor, ok := params.After.Get(); ok {
		if cursor.PostingSequence().IsZero() || cursor.LineOrder().IsZero() || cursor.LineID().IsZero() {
			return nil, domain.Invalid("after", domain.ViolationInvalidFormat, "INV-008")
		}
		beforeSequence = cursor.PostingSequence().Int64()
		afterLineOrder = cursor.LineOrder().Int64()
		afterLineID = cursor.LineID().Int64()
	}
	rows, err := s.queries.ListItemLedgerPage(ctx, sqlcgen.ListItemLedgerPageParams{
		ItemID: params.ItemID.Int64(), BeforePostingSequence: beforeSequence,
		AfterLineOrder: afterLineOrder, AfterLineID: afterLineID, LimitCount: limit,
	})
	if err != nil {
		return nil, classifyError(listItemLedgerPageOperation, err)
	}
	entries := make([]inventory.LedgerEntryView, 0, len(rows))
	for index, row := range rows {
		entry, mapErr := mapLedgerEntryView(row)
		if mapErr != nil {
			return nil, corruptInventoryRow(listItemLedgerPageOperation, index, mapErr)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (s *Store) ListLineAllocations(ctx context.Context, lineID domain.StockDocumentLineID) ([]inventory.AllocationView, error) {
	if lineID.IsZero() {
		return nil, domain.Invalid("line_id", domain.ViolationRequired, "")
	}
	rows, err := s.queries.ListLineAllocations(ctx, lineID.Int64())
	if err != nil {
		return nil, classifyError(listLineAllocationsOperation, err)
	}
	allocations := make([]inventory.AllocationView, 0, len(rows))
	for index, row := range rows {
		allocation, mapErr := mapAllocationView(row)
		if mapErr != nil {
			return nil, corruptInventoryRow(listLineAllocationsOperation, index, mapErr)
		}
		allocations = append(allocations, allocation)
	}
	return allocations, nil
}

type inventoryBalanceFields struct {
	itemID, quantityAtomic, inventoryValueMicro, updatedAtMS int64
	itemName, itemNormalizedName, baseUnitCode               string
	itemArchivedAtMS, lastDocumentID                         sql.NullInt64
}

func mapInventoryBalanceSnapshot(fields inventoryBalanceFields) (inventory.BalanceSnapshot, error) {
	itemID, err := domain.NewItemID(fields.itemID)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	quantity, err := domain.NewAtomicQuantity(fields.quantityAtomic)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	value, err := domain.NewInventoryValue(fields.inventoryValueMicro)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	lastDocumentID, err := optionalStockDocumentID(fields.lastDocumentID)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(fields.updatedAtMS)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	balance, err := inventory.NewBalance(inventory.BalanceParams{
		ItemID: itemID, Quantity: quantity, Value: value,
		LastDocumentID: lastDocumentID, UpdatedAt: updatedAt,
	})
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	itemName, err := domain.RestoreUniqueName(fields.itemName, fields.itemNormalizedName)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	if itemName.Display() != fields.itemName {
		return inventory.BalanceSnapshot{}, domain.ErrInvariant
	}
	baseUnit, err := domain.NewUnitCode(fields.baseUnitCode)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	if baseUnit.String() != fields.baseUnitCode {
		return inventory.BalanceSnapshot{}, domain.ErrInvariant
	}
	archivedAt, err := optionalInstant(fields.itemArchivedAtMS)
	if err != nil {
		return inventory.BalanceSnapshot{}, err
	}
	return inventory.NewBalanceSnapshot(inventory.BalanceSnapshotParams{
		Balance: balance, ItemName: itemName, BaseUnit: baseUnit, ItemArchivedAt: archivedAt,
	})
}

type lotViewFields struct {
	id, itemID, sourceLineID, initialQuantity, consumedQuantity int64
	restoredQuantity, availableQuantity, createdAtMS            int64
	sourceDocumentID, sourcePostingSequence                     int64
	lotCode, expiresOn                                          sql.NullString
	originatedOn, sourceKind, sourceOccurredOn                  string
}

func mapLotView(fields lotViewFields) (inventory.LotView, error) {
	id, err := domain.NewInventoryLotID(fields.id)
	if err != nil {
		return inventory.LotView{}, err
	}
	itemID, err := domain.NewItemID(fields.itemID)
	if err != nil {
		return inventory.LotView{}, err
	}
	sourceLineID, err := domain.NewStockDocumentLineID(fields.sourceLineID)
	if err != nil {
		return inventory.LotView{}, err
	}
	postingSequence, err := domain.NewPostingSequence(fields.sourcePostingSequence)
	if err != nil {
		return inventory.LotView{}, err
	}
	initial, err := domain.NewAtomicQuantity(fields.initialQuantity)
	if err != nil {
		return inventory.LotView{}, err
	}
	consumed, err := domain.NewAtomicQuantity(fields.consumedQuantity)
	if err != nil {
		return inventory.LotView{}, err
	}
	restored, err := domain.NewAtomicQuantity(fields.restoredQuantity)
	if err != nil {
		return inventory.LotView{}, err
	}
	available, err := domain.NewAtomicQuantity(fields.availableQuantity)
	if err != nil {
		return inventory.LotView{}, err
	}
	lotCode, err := optionalNonEmptyText(fields.lotCode)
	if err != nil {
		return inventory.LotView{}, err
	}
	originatedOn, err := domain.ParseBusinessDate(fields.originatedOn)
	if err != nil {
		return inventory.LotView{}, err
	}
	expiresOn, err := optionalBusinessDate(fields.expiresOn)
	if err != nil {
		return inventory.LotView{}, err
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(fields.createdAtMS)
	if err != nil {
		return inventory.LotView{}, err
	}
	lot, err := inventory.NewLot(inventory.LotParams{
		ID: id, ItemID: itemID, SourceLineID: sourceLineID, SourcePostingSequence: postingSequence,
		InitialQuantity: initial, ConsumedQuantity: consumed, RestoredQuantity: restored,
		AvailableQuantity: available, LotCode: lotCode, OriginatedOn: originatedOn,
		ExpiresOn: expiresOn, CreatedAt: createdAt,
	})
	if err != nil {
		return inventory.LotView{}, err
	}
	sourceDocumentID, err := domain.NewStockDocumentID(fields.sourceDocumentID)
	if err != nil {
		return inventory.LotView{}, err
	}
	sourceKind, err := domain.ParseDocumentKind(fields.sourceKind)
	if err != nil {
		return inventory.LotView{}, err
	}
	sourceOccurredOn, err := domain.ParseBusinessDate(fields.sourceOccurredOn)
	if err != nil {
		return inventory.LotView{}, err
	}
	return inventory.NewLotView(inventory.LotViewParams{
		Lot: lot, SourceDocumentID: sourceDocumentID,
		SourceKind: sourceKind, SourceOccurredOn: sourceOccurredOn,
	})
}

func mapAllocationView(row sqlcgen.ListLineAllocationsRow) (inventory.AllocationView, error) {
	id, err := domain.NewLotAllocationID(row.ID)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	lineID, err := domain.NewStockDocumentLineID(row.LineID)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	lotID, err := domain.NewInventoryLotID(row.LotID)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	quantity, err := domain.NewAtomicQuantity(row.QuantityAtomic)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	effect := domain.AllocationConsumption
	restoresAllocationID := domain.None[domain.LotAllocationID]()
	if row.RestoresAllocationID.Valid {
		restoredID, restoreErr := domain.NewLotAllocationID(row.RestoresAllocationID.Int64)
		if restoreErr != nil {
			return inventory.AllocationView{}, restoreErr
		}
		effect = domain.AllocationRestoration
		restoresAllocationID = domain.Some(restoredID)
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.CreatedAtMs)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	allocation, err := inventory.NewAllocation(inventory.AllocationParams{
		ID: id, LineID: lineID, LotID: lotID, Quantity: quantity, Effect: effect,
		RestoresAllocationID: restoresAllocationID, CreatedAt: createdAt,
	})
	if err != nil {
		return inventory.AllocationView{}, err
	}
	sourceLineID, err := domain.NewStockDocumentLineID(row.SourceLineID)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	initialQuantity, err := domain.NewAtomicQuantity(row.LotInitialQuantityAtomic)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	lotCode, err := optionalNonEmptyText(row.LotCode)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	originatedOn, err := domain.ParseBusinessDate(row.OriginatedOn)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	expiresOn, err := optionalBusinessDate(row.ExpiresOn)
	if err != nil {
		return inventory.AllocationView{}, err
	}
	return inventory.NewAllocationView(inventory.AllocationViewParams{
		Allocation: allocation, SourceLineID: sourceLineID, LotInitialQuantity: initialQuantity,
		LotCode: lotCode, OriginatedOn: originatedOn, ExpiresOn: expiresOn,
	})
}

func mapLedgerEntryView(row sqlcgen.ListItemLedgerPageRow) (inventory.LedgerEntryView, error) {
	lineID, err := domain.NewStockDocumentLineID(row.LineID)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	documentID, err := domain.NewStockDocumentID(row.DocumentID)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	postingSequence, err := domain.NewPostingSequence(row.PostingSequence)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	lineOrder, err := domain.NewLineOrder(row.LineOrder)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	kind, err := domain.ParseDocumentKind(row.DocumentKind)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	occurredOn, err := domain.ParseBusinessDate(row.OccurredOn)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	postedAt, err := domain.UTCInstantFromUnixMilli(row.PostedAtMs)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	itemID, err := domain.NewItemID(row.ItemID)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	direction, err := domain.ParseDirection(row.Direction)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	quantity, err := domain.NewAtomicQuantity(row.QuantityAtomic)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	value, err := domain.NewInventoryValue(row.InventoryValueMicro)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	commercialTotal, err := optionalMinorAmount(row.CommercialTotalMinor)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	currency, err := domain.RestoreCurrency(row.CurrencyCode, int(row.CurrencyMinorDigits))
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	entry, err := inventory.NewLedgerEntry(inventory.LedgerEntryParams{
		LineID: lineID, DocumentID: documentID, PostingSequence: postingSequence,
		LineOrder: lineOrder, Kind: kind, OccurredOn: occurredOn, PostedAt: postedAt,
		ItemID: itemID, Direction: direction, Quantity: quantity, Value: value,
		CommercialTotal: commercialTotal, Currency: currency,
	})
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.EnteredUnitCode)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	if enteredUnit.String() != row.EnteredUnitCode {
		return inventory.LedgerEntryView{}, domain.ErrInvariant
	}
	enteredPackaging, err := optionalNonEmptyText(row.EnteredPackagingName)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	conversion, err := domain.NewUnitConversion(row.ConversionNumeratorAtomic, row.ConversionDenominator)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	if conversion.NumeratorAtomic() != row.ConversionNumeratorAtomic ||
		conversion.Denominator() != row.ConversionDenominator {
		return inventory.LedgerEntryView{}, domain.ErrInvariant
	}
	reversesLineID, err := optionalStockDocumentLineID(row.ReversesLineID)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	idempotencyKey, err := domain.NewIdempotencyKey(row.IdempotencyKey)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	if idempotencyKey.String() != row.IdempotencyKey {
		return inventory.LedgerEntryView{}, domain.ErrInvariant
	}
	counterpartyID, err := optionalCounterpartyID(row.CounterpartyID)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	counterpartyName, err := optionalDisplayName(row.CounterpartyName)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	reason, err := optionalDocumentReason(kind, row.ReasonCode)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	notes, err := optionalNonEmptyText(row.Notes)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	reversesDocumentID, err := optionalStockDocumentID(row.ReversesDocumentID)
	if err != nil {
		return inventory.LedgerEntryView{}, err
	}
	return inventory.NewLedgerEntryView(inventory.LedgerEntryViewParams{
		Entry: entry, EnteredUnit: enteredUnit, EnteredPackaging: enteredPackaging,
		Conversion: conversion, ReversesLineID: reversesLineID, IdempotencyKey: idempotencyKey,
		CounterpartyID: counterpartyID, CounterpartyName: counterpartyName,
		Reason: reason, Notes: notes, ReversesDocumentID: reversesDocumentID,
	})
}

func sqliteBool(value int64) (bool, error) {
	switch value {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("invalid SQLite boolean %d", value)
	}
}

func inventoryPageSize(requested int64) (int64, error) {
	if requested == 0 {
		return inventoryDefaultPageSize, nil
	}
	if requested < 0 || requested > inventoryMaximumPageSize {
		return 0, domain.Invalid("limit", domain.ViolationOutOfRange, "")
	}
	return requested, nil
}

func optionalAtomicQuantity(value sql.NullInt64) (domain.Option[domain.AtomicQuantity], error) {
	if !value.Valid {
		return domain.None[domain.AtomicQuantity](), nil
	}
	mapped, err := domain.NewAtomicQuantity(value.Int64)
	if err != nil {
		return domain.None[domain.AtomicQuantity](), err
	}
	return domain.Some(mapped), nil
}

func optionalMinorAmount(value sql.NullInt64) (domain.Option[domain.MinorAmount], error) {
	if !value.Valid {
		return domain.None[domain.MinorAmount](), nil
	}
	mapped, err := domain.NewMinorAmount(value.Int64)
	if err != nil {
		return domain.None[domain.MinorAmount](), err
	}
	return domain.Some(mapped), nil
}

func optionalInstant(value sql.NullInt64) (domain.Option[domain.UTCInstant], error) {
	if !value.Valid {
		return domain.None[domain.UTCInstant](), nil
	}
	mapped, err := domain.UTCInstantFromUnixMilli(value.Int64)
	if err != nil {
		return domain.None[domain.UTCInstant](), err
	}
	return domain.Some(mapped), nil
}

func optionalStockDocumentID(value sql.NullInt64) (domain.Option[domain.StockDocumentID], error) {
	if !value.Valid {
		return domain.None[domain.StockDocumentID](), nil
	}
	mapped, err := domain.NewStockDocumentID(value.Int64)
	if err != nil {
		return domain.None[domain.StockDocumentID](), err
	}
	return domain.Some(mapped), nil
}

func optionalStockDocumentLineID(value sql.NullInt64) (domain.Option[domain.StockDocumentLineID], error) {
	if !value.Valid {
		return domain.None[domain.StockDocumentLineID](), nil
	}
	mapped, err := domain.NewStockDocumentLineID(value.Int64)
	if err != nil {
		return domain.None[domain.StockDocumentLineID](), err
	}
	return domain.Some(mapped), nil
}

func optionalCounterpartyID(value sql.NullInt64) (domain.Option[domain.CounterpartyID], error) {
	if !value.Valid {
		return domain.None[domain.CounterpartyID](), nil
	}
	mapped, err := domain.NewCounterpartyID(value.Int64)
	if err != nil {
		return domain.None[domain.CounterpartyID](), err
	}
	return domain.Some(mapped), nil
}

func optionalNonEmptyText(value sql.NullString) (domain.Option[domain.NonEmptyText], error) {
	if !value.Valid {
		return domain.None[domain.NonEmptyText](), nil
	}
	mapped, err := domain.NewNonEmptyText(value.String)
	if err != nil {
		return domain.None[domain.NonEmptyText](), err
	}
	if mapped.String() != value.String {
		return domain.None[domain.NonEmptyText](), domain.ErrInvariant
	}
	return domain.Some(mapped), nil
}

func optionalDisplayName(value sql.NullString) (domain.Option[domain.DisplayName], error) {
	if !value.Valid {
		return domain.None[domain.DisplayName](), nil
	}
	mapped, err := domain.NewDisplayName(value.String)
	if err != nil {
		return domain.None[domain.DisplayName](), err
	}
	if mapped.String() != value.String {
		return domain.None[domain.DisplayName](), domain.ErrInvariant
	}
	return domain.Some(mapped), nil
}

func optionalBusinessDate(value sql.NullString) (domain.Option[domain.BusinessDate], error) {
	if !value.Valid {
		return domain.None[domain.BusinessDate](), nil
	}
	mapped, err := domain.ParseBusinessDate(value.String)
	if err != nil {
		return domain.None[domain.BusinessDate](), err
	}
	return domain.Some(mapped), nil
}

func optionalDocumentReason(kind domain.DocumentKind, value sql.NullString) (domain.Option[domain.DocumentReason], error) {
	if !value.Valid {
		return domain.ParseDocumentReason(kind, "")
	}
	return domain.ParseDocumentReason(kind, value.String)
}

func corruptInventoryRow(operation string, index int, err error) error {
	return corruptDataError(operation, fmt.Errorf("row %d: %w", index, err))
}
