package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
)

const (
	saleDefaultPageSize = 50
	saleMaximumPageSize = 100
)

type SaleCursor struct {
	PostingSequence domain.PostingSequence
	ID              domain.StockDocumentID
}

type SaleListFilter struct {
	After    domain.Option[SaleCursor]
	PageSize int
}

type SalePage struct {
	items []PostedSaleDocument
	next  domain.Option[SaleCursor]
}

func (p SalePage) Items() []PostedSaleDocument {
	items := make([]PostedSaleDocument, len(p.items))
	copy(items, p.items)
	return items
}

func (p SalePage) Next() domain.Option[SaleCursor] { return p.next }

type PostSaleInput struct {
	IdempotencyKey domain.IdempotencyKey
	CounterpartyID domain.Option[domain.CounterpartyID]
	OccurredOn     domain.BusinessDate
	PostedAt       domain.UTCInstant
	Reason         domain.Option[domain.DocumentReason]
	Notes          domain.Option[domain.NonEmptyText]
	Lines          []PostSaleLineInput
}

type PostSaleLineInput struct {
	ItemID               domain.ItemID
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	CommercialTotal      domain.MinorAmount
	LotID                domain.Option[domain.InventoryLotID]
}

type PostedSaleDocument struct {
	id              domain.StockDocumentID
	idempotencyKey  domain.IdempotencyKey
	postingSequence domain.PostingSequence
	counterpartyID  domain.Option[domain.CounterpartyID]
	occurredOn      domain.BusinessDate
	postedAt        domain.UTCInstant
	currency        domain.Currency
	reason          domain.Option[domain.DocumentReason]
	notes           domain.Option[domain.NonEmptyText]
	lines           []PostedSaleLine
}

func NewPostedSaleDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	counterpartyID domain.Option[domain.CounterpartyID],
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	reason domain.Option[domain.DocumentReason],
	notes domain.Option[domain.NonEmptyText],
	lines []PostedSaleLine,
) PostedSaleDocument {
	cloned := make([]PostedSaleLine, len(lines))
	copy(cloned, lines)
	return PostedSaleDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		counterpartyID: counterpartyID, occurredOn: occurredOn, postedAt: postedAt,
		currency: currency, reason: reason, notes: notes, lines: cloned,
	}
}

func (d PostedSaleDocument) ID() domain.StockDocumentID              { return d.id }
func (d PostedSaleDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d PostedSaleDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d PostedSaleDocument) CounterpartyID() domain.Option[domain.CounterpartyID] {
	return d.counterpartyID
}
func (d PostedSaleDocument) OccurredOn() domain.BusinessDate              { return d.occurredOn }
func (d PostedSaleDocument) PostedAt() domain.UTCInstant                  { return d.postedAt }
func (d PostedSaleDocument) Currency() domain.Currency                    { return d.currency }
func (d PostedSaleDocument) Reason() domain.Option[domain.DocumentReason] { return d.reason }
func (d PostedSaleDocument) Notes() domain.Option[domain.NonEmptyText]    { return d.notes }
func (d PostedSaleDocument) Lines() []PostedSaleLine {
	lines := make([]PostedSaleLine, len(d.lines))
	copy(lines, d.lines)
	return lines
}

type PostedSaleLine struct {
	id                   domain.StockDocumentLineID
	lineOrder            domain.LineOrder
	itemID               domain.ItemID
	quantity             domain.AtomicQuantity
	enteredUnit          domain.UnitCode
	enteredPackagingName domain.Option[domain.NonEmptyText]
	conversion           domain.UnitConversion
	inventoryValue       domain.InventoryValue
	commercialTotal      domain.MinorAmount
	allocations          []SaleAllocation
}

func NewPostedSaleLine(
	id domain.StockDocumentLineID,
	lineOrder domain.LineOrder,
	itemID domain.ItemID,
	quantity domain.AtomicQuantity,
	enteredUnit domain.UnitCode,
	enteredPackagingName domain.Option[domain.NonEmptyText],
	conversion domain.UnitConversion,
	inventoryValue domain.InventoryValue,
	commercialTotal domain.MinorAmount,
	allocations []SaleAllocation,
) PostedSaleLine {
	cloned := make([]SaleAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedSaleLine{
		id: id, lineOrder: lineOrder, itemID: itemID, quantity: quantity,
		enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue,
		commercialTotal: commercialTotal, allocations: cloned,
	}
}

func (l PostedSaleLine) ID() domain.StockDocumentLineID  { return l.id }
func (l PostedSaleLine) LineOrder() domain.LineOrder     { return l.lineOrder }
func (l PostedSaleLine) ItemID() domain.ItemID           { return l.itemID }
func (l PostedSaleLine) Quantity() domain.AtomicQuantity { return l.quantity }
func (l PostedSaleLine) EnteredUnit() domain.UnitCode    { return l.enteredUnit }
func (l PostedSaleLine) EnteredPackagingName() domain.Option[domain.NonEmptyText] {
	return l.enteredPackagingName
}
func (l PostedSaleLine) Conversion() domain.UnitConversion     { return l.conversion }
func (l PostedSaleLine) InventoryValue() domain.InventoryValue { return l.inventoryValue }
func (l PostedSaleLine) CommercialTotal() domain.MinorAmount   { return l.commercialTotal }
func (l PostedSaleLine) Allocations() []SaleAllocation {
	allocations := make([]SaleAllocation, len(l.allocations))
	copy(allocations, l.allocations)
	return allocations
}

type SaleAllocation struct {
	id       domain.LotAllocationID
	lotID    domain.InventoryLotID
	quantity domain.AtomicQuantity
}

func NewSaleAllocation(id domain.LotAllocationID, lotID domain.InventoryLotID, quantity domain.AtomicQuantity) SaleAllocation {
	return SaleAllocation{id: id, lotID: lotID, quantity: quantity}
}

func (a SaleAllocation) ID() domain.LotAllocationID      { return a.id }
func (a SaleAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a SaleAllocation) Quantity() domain.AtomicQuantity { return a.quantity }

func (s *Store) GetPostedSale(ctx context.Context, id domain.StockDocumentID) (PostedSaleDocument, error) {
	if id.IsZero() {
		return PostedSaleDocument{}, domain.Invalid("document_id", domain.ViolationRequired, "DOC-001")
	}
	var document PostedSaleDocument
	err := s.database.Read(ctx, func(tx *database.ReadTx) error {
		value, err := loadPostedSaleDocument(ctx, tx, id.Int64())
		if err != nil {
			return err
		}
		document = value
		return nil
	})
	if err != nil {
		return PostedSaleDocument{}, classifyError("get posted sale", err)
	}
	return document, nil
}

func (s *Store) ListPostedSales(ctx context.Context, filter SaleListFilter) (SalePage, error) {
	pageSize, err := salePageSize(filter.PageSize)
	if err != nil {
		return SalePage{}, err
	}
	var page SalePage
	err = s.database.Read(ctx, func(tx *database.ReadTx) error {
		documentIDs, err := listPostedSaleIDs(ctx, tx, filter.After, pageSize+1)
		if err != nil {
			return err
		}
		hasMore := len(documentIDs) > pageSize
		if hasMore {
			documentIDs = documentIDs[:pageSize]
		}
		items := make([]PostedSaleDocument, 0, len(documentIDs))
		for _, id := range documentIDs {
			document, err := loadPostedSaleDocument(ctx, tx, id)
			if err != nil {
				return err
			}
			items = append(items, document)
		}
		next := domain.None[SaleCursor]()
		if hasMore && len(items) > 0 {
			last := items[len(items)-1]
			next = domain.Some(SaleCursor{
				PostingSequence: last.PostingSequence(),
				ID:              last.ID(),
			})
		}
		page = SalePage{items: items, next: next}
		return nil
	})
	if err != nil {
		return SalePage{}, classifyError("list posted sales", err)
	}
	return page, nil
}

func (s *Store) PostSale(ctx context.Context, input PostSaleInput) (PostedSaleDocument, error) {
	var posted PostedSaleDocument
	err := s.database.Write(ctx, func(tx *database.WriteTx) error {
		value, err := postSaleTx(ctx, tx, input)
		if err != nil {
			return err
		}
		posted = value
		return nil
	})
	if err != nil {
		return PostedSaleDocument{}, classifyError("post sale", err)
	}
	return posted, nil
}

func salePageSize(requested int) (int, error) {
	if requested == 0 {
		return saleDefaultPageSize, nil
	}
	if requested < 1 || requested > saleMaximumPageSize {
		return 0, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	return requested, nil
}

func listPostedSaleIDs(
	ctx context.Context,
	tx databaseWriteTx,
	after domain.Option[SaleCursor],
	limit int,
) ([]int64, error) {
	args := []any{limit}
	query := `
		SELECT id
		FROM stock_documents
		WHERE kind = 'SALE'
		ORDER BY posting_sequence DESC, id DESC
		LIMIT ?
	`
	if cursor, ok := after.Get(); ok {
		if cursor.PostingSequence.IsZero() || cursor.ID.IsZero() {
			return nil, domain.Invalid("cursor", domain.ViolationInvalidFormat, "")
		}
		args = []any{cursor.PostingSequence.Int64(), cursor.PostingSequence.Int64(), cursor.ID.Int64(), limit}
		query = `
			SELECT id
			FROM stock_documents
			WHERE kind = 'SALE'
			  AND (posting_sequence < ? OR (posting_sequence = ? AND id < ?))
			ORDER BY posting_sequence DESC, id DESC
			LIMIT ?
		`
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func postSaleTx(ctx context.Context, tx databaseWriteTx, input PostSaleInput) (PostedSaleDocument, error) {
	if err := validateSaleInput(input); err != nil {
		return PostedSaleDocument{}, err
	}

	var existingID int64
	var existingKind string
	err := tx.QueryRowContext(ctx, `
		SELECT id, kind FROM stock_documents WHERE idempotency_key = ?
	`, input.IdempotencyKey.String()).Scan(&existingID, &existingKind)
	if err == nil {
		if existingKind != domain.DocumentSale.String() {
			return PostedSaleDocument{}, fmt.Errorf("%w: idempotency key belongs to %s", domain.ErrConflict, existingKind)
		}
		return loadPostedSaleDocument(ctx, tx, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return PostedSaleDocument{}, err
	}

	currency, err := loadDocumentCurrency(ctx, tx)
	if err != nil {
		return PostedSaleDocument{}, err
	}

	var postingSequence int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(posting_sequence), 0) + 1 FROM stock_documents
	`).Scan(&postingSequence); err != nil {
		return PostedSaleDocument{}, err
	}

	var documentID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, counterparty_id, occurred_on,
			posted_at_ms, currency_code, currency_minor_digits, reason_code, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		domain.DocumentSale.String(),
		input.IdempotencyKey.String(),
		postingSequence,
		nullableCounterpartyID(input.CounterpartyID),
		input.OccurredOn.String(),
		input.PostedAt.UnixMilli(),
		currency.Code().String(),
		int64(currency.MinorDigits().Int()),
		nullableDocumentReason(input.Reason),
		nullableText(input.Notes),
	).Scan(&documentID); err != nil {
		return PostedSaleDocument{}, err
	}

	for index, line := range input.Lines {
		if err := insertPostedSaleLine(ctx, tx, documentID, int64(index+1), input.OccurredOn, input.PostedAt, line); err != nil {
			return PostedSaleDocument{}, fmt.Errorf("line %d: %w", index+1, err)
		}
	}

	return loadPostedSaleDocument(ctx, tx, documentID)
}

func validateSaleInput(input PostSaleInput) error {
	if input.IdempotencyKey.String() == "" {
		return domain.Invalid("idempotency_key", domain.ViolationRequired, "DOC-003")
	}
	if input.OccurredOn.IsZero() {
		return domain.Invalid("occurred_on", domain.ViolationRequired, "DOC-004")
	}
	if input.PostedAt.IsZero() {
		return domain.Invalid("posted_at", domain.ViolationRequired, "DOC-004")
	}
	if len(input.Lines) == 0 {
		return domain.Invalid("lines", domain.ViolationRequired, "DOC-002")
	}
	if reason, ok := input.Reason.Get(); ok {
		if _, err := domain.ParseDocumentReason(domain.DocumentSale, reason.String()); err != nil {
			return err
		}
	}
	for index, line := range input.Lines {
		if line.ItemID.IsZero() {
			return domain.Invalid(fmt.Sprintf("lines[%d].item_id", index), domain.ViolationRequired, "DOC-005")
		}
		if line.Quantity.Int64() <= 0 {
			return domain.Invalid(fmt.Sprintf("lines[%d].quantity_atomic", index), domain.ViolationNotPositive, "DOC-005")
		}
		if line.EnteredUnit.String() == "" {
			return domain.Invalid(fmt.Sprintf("lines[%d].entered_unit_code", index), domain.ViolationRequired, "DOC-005")
		}
		if line.Conversion.IsZero() {
			return domain.Invalid(fmt.Sprintf("lines[%d].conversion", index), domain.ViolationRequired, "DOC-005")
		}
		if line.CommercialTotal.IsZero() && input.Reason.IsNone() {
			return domain.Invalid("reason_code", domain.ViolationRequired, "SAL-002")
		}
	}
	return nil
}

func insertPostedSaleLine(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	lineOrder int64,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	line PostSaleLineInput,
) error {
	balance, err := readAdjustmentBalance(ctx, tx, line.ItemID)
	if err != nil {
		return err
	}
	if line.Quantity.Int64() > balance.quantityAtomic {
		return domain.Invalid("quantity_atomic", domain.ViolationOutOfRange, "INV-004")
	}
	inventoryValue, err := weightedAverageValue(balance.inventoryValueMicro, balance.quantityAtomic, line.Quantity.Int64())
	if err != nil {
		return err
	}

	var lineID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, entered_packaging_name, conversion_numerator_atomic,
			conversion_denominator, inventory_value_micro, commercial_total_minor
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		documentID,
		lineOrder,
		line.ItemID.Int64(),
		domain.DirectionOut.String(),
		line.Quantity.Int64(),
		line.EnteredUnit.String(),
		nullableText(line.EnteredPackagingName),
		line.Conversion.NumeratorAtomic(),
		line.Conversion.Denominator(),
		inventoryValue.Int64(),
		line.CommercialTotal.Int64(),
	).Scan(&lineID); err != nil {
		return err
	}

	if err := allocateProductionLots(ctx, tx, lineID, line.ItemID, line.Quantity.Int64(), occurredOn, postedAt, line.LotID); err != nil {
		return err
	}
	return updateAdjustmentBalance(ctx, tx, documentID, postedAt, line.ItemID, -line.Quantity.Int64(), -inventoryValue.Int64())
}

func loadPostedSaleDocument(ctx context.Context, tx databaseWriteTx, id int64) (PostedSaleDocument, error) {
	var row postedSaleDocumentRow
	err := tx.QueryRowContext(ctx, `
		SELECT id, idempotency_key, posting_sequence, counterparty_id, occurred_on,
		       posted_at_ms, currency_code, currency_minor_digits, reason_code, notes
		FROM stock_documents
		WHERE id = ? AND kind = 'SALE'
	`, id).Scan(
		&row.id,
		&row.idempotencyKey,
		&row.postingSequence,
		&row.counterpartyID,
		&row.occurredOn,
		&row.postedAtMS,
		&row.currencyCode,
		&row.currencyMinorDigits,
		&row.reasonCode,
		&row.notes,
	)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	lines, err := loadPostedSaleLines(ctx, tx, id)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	return mapPostedSaleDocument(row, lines)
}

type postedSaleDocumentRow struct {
	id, postingSequence, postedAtMS, currencyMinorDigits int64
	idempotencyKey, occurredOn, currencyCode             string
	counterpartyID                                       sql.NullInt64
	reasonCode, notes                                    sql.NullString
}

func loadPostedSaleLines(ctx context.Context, tx databaseWriteTx, documentID int64) ([]PostedSaleLine, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, line_order, item_id, quantity_atomic,
		       entered_unit_code, entered_packaging_name,
		       conversion_numerator_atomic, conversion_denominator,
		       inventory_value_micro, commercial_total_minor
		FROM stock_document_lines
		WHERE document_id = ?
		ORDER BY line_order, id
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []PostedSaleLine
	for rows.Next() {
		var row postedSaleLineRow
		if err := rows.Scan(
			&row.id,
			&row.lineOrder,
			&row.itemID,
			&row.quantityAtomic,
			&row.enteredUnitCode,
			&row.enteredPackagingName,
			&row.conversionNumeratorAtomic,
			&row.conversionDenominator,
			&row.inventoryValueMicro,
			&row.commercialTotalMinor,
		); err != nil {
			return nil, err
		}
		allocations, err := loadSaleAllocations(ctx, tx, row.id)
		if err != nil {
			return nil, err
		}
		line, err := mapPostedSaleLine(row, allocations)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

type postedSaleLineRow struct {
	id, lineOrder, itemID, quantityAtomic            int64
	conversionNumeratorAtomic, conversionDenominator int64
	inventoryValueMicro, commercialTotalMinor        int64
	enteredUnitCode                                  string
	enteredPackagingName                             sql.NullString
}

func loadSaleAllocations(ctx context.Context, tx databaseWriteTx, lineID int64) ([]SaleAllocation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, lot_id, quantity_atomic
		FROM lot_allocations
		WHERE line_id = ?
		ORDER BY id
	`, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocations []SaleAllocation
	for rows.Next() {
		var idValue, lotIDValue, quantityValue int64
		if err := rows.Scan(&idValue, &lotIDValue, &quantityValue); err != nil {
			return nil, err
		}
		id, err := domain.NewLotAllocationID(idValue)
		if err != nil {
			return nil, err
		}
		lotID, err := domain.NewInventoryLotID(lotIDValue)
		if err != nil {
			return nil, err
		}
		quantity, err := domain.NewAtomicQuantity(quantityValue)
		if err != nil {
			return nil, err
		}
		allocations = append(allocations, NewSaleAllocation(id, lotID, quantity))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return allocations, nil
}

func mapPostedSaleDocument(row postedSaleDocumentRow, lines []PostedSaleLine) (PostedSaleDocument, error) {
	id, err := domain.NewStockDocumentID(row.id)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	idempotencyKey, err := domain.NewIdempotencyKey(row.idempotencyKey)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	postingSequence, err := domain.NewPostingSequence(row.postingSequence)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	counterpartyID, err := optionalCounterpartyID(row.counterpartyID)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	occurredOn, err := domain.ParseBusinessDate(row.occurredOn)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	postedAt, err := domain.UTCInstantFromUnixMilli(row.postedAtMS)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	currency, err := domain.RestoreCurrency(row.currencyCode, int(row.currencyMinorDigits))
	if err != nil {
		return PostedSaleDocument{}, err
	}
	reason, err := optionalDocumentReason(domain.DocumentSale, row.reasonCode)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	notes, err := optionalNonEmptyText(row.notes)
	if err != nil {
		return PostedSaleDocument{}, err
	}
	return NewPostedSaleDocument(
		id, idempotencyKey, postingSequence, counterpartyID, occurredOn, postedAt,
		currency, reason, notes, lines,
	), nil
}

func mapPostedSaleLine(row postedSaleLineRow, allocations []SaleAllocation) (PostedSaleLine, error) {
	id, err := domain.NewStockDocumentLineID(row.id)
	if err != nil {
		return PostedSaleLine{}, err
	}
	lineOrder, err := domain.NewLineOrder(row.lineOrder)
	if err != nil {
		return PostedSaleLine{}, err
	}
	itemID, err := domain.NewItemID(row.itemID)
	if err != nil {
		return PostedSaleLine{}, err
	}
	quantity, err := domain.NewAtomicQuantity(row.quantityAtomic)
	if err != nil {
		return PostedSaleLine{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.enteredUnitCode)
	if err != nil {
		return PostedSaleLine{}, err
	}
	enteredPackagingName, err := optionalNonEmptyText(row.enteredPackagingName)
	if err != nil {
		return PostedSaleLine{}, err
	}
	conversion, err := domain.NewUnitConversion(row.conversionNumeratorAtomic, row.conversionDenominator)
	if err != nil {
		return PostedSaleLine{}, err
	}
	inventoryValue, err := domain.NewInventoryValue(row.inventoryValueMicro)
	if err != nil {
		return PostedSaleLine{}, err
	}
	commercialTotal, err := domain.NewMinorAmount(row.commercialTotalMinor)
	if err != nil {
		return PostedSaleLine{}, err
	}
	return NewPostedSaleLine(
		id, lineOrder, itemID, quantity, enteredUnit, enteredPackagingName,
		conversion, inventoryValue, commercialTotal, allocations,
	), nil
}
