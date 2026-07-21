package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
)

type PostReversalInput struct {
	IdempotencyKey   domain.IdempotencyKey
	TargetDocumentID domain.StockDocumentID
	OccurredOn       domain.BusinessDate
	PostedAt         domain.UTCInstant
	Notes            domain.Option[domain.NonEmptyText]
}

type PostedReversalDocument struct {
	id               domain.StockDocumentID
	idempotencyKey   domain.IdempotencyKey
	postingSequence  domain.PostingSequence
	targetDocumentID domain.StockDocumentID
	occurredOn       domain.BusinessDate
	postedAt         domain.UTCInstant
	currency         domain.Currency
	notes            domain.Option[domain.NonEmptyText]
	lines            []PostedReversalLine
}

func NewPostedReversalDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	targetDocumentID domain.StockDocumentID,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	notes domain.Option[domain.NonEmptyText],
	lines []PostedReversalLine,
) PostedReversalDocument {
	cloned := make([]PostedReversalLine, len(lines))
	copy(cloned, lines)
	return PostedReversalDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		targetDocumentID: targetDocumentID, occurredOn: occurredOn, postedAt: postedAt,
		currency: currency, notes: notes, lines: cloned,
	}
}

func (d PostedReversalDocument) ID() domain.StockDocumentID              { return d.id }
func (d PostedReversalDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d PostedReversalDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d PostedReversalDocument) TargetDocumentID() domain.StockDocumentID {
	return d.targetDocumentID
}
func (d PostedReversalDocument) OccurredOn() domain.BusinessDate { return d.occurredOn }
func (d PostedReversalDocument) PostedAt() domain.UTCInstant     { return d.postedAt }
func (d PostedReversalDocument) Currency() domain.Currency       { return d.currency }
func (d PostedReversalDocument) Notes() domain.Option[domain.NonEmptyText] {
	return d.notes
}
func (d PostedReversalDocument) Lines() []PostedReversalLine {
	lines := make([]PostedReversalLine, len(d.lines))
	copy(lines, d.lines)
	return lines
}

type PostedReversalLine struct {
	id                   domain.StockDocumentLineID
	lineOrder            domain.LineOrder
	itemID               domain.ItemID
	direction            domain.Direction
	quantity             domain.AtomicQuantity
	enteredUnit          domain.UnitCode
	enteredPackagingName domain.Option[domain.NonEmptyText]
	conversion           domain.UnitConversion
	inventoryValue       domain.InventoryValue
	commercialTotal      domain.Option[domain.MinorAmount]
	reversesLineID       domain.StockDocumentLineID
	allocations          []ReversalAllocation
}

func NewPostedReversalLine(
	id domain.StockDocumentLineID,
	lineOrder domain.LineOrder,
	itemID domain.ItemID,
	direction domain.Direction,
	quantity domain.AtomicQuantity,
	enteredUnit domain.UnitCode,
	enteredPackagingName domain.Option[domain.NonEmptyText],
	conversion domain.UnitConversion,
	inventoryValue domain.InventoryValue,
	commercialTotal domain.Option[domain.MinorAmount],
	reversesLineID domain.StockDocumentLineID,
	allocations []ReversalAllocation,
) PostedReversalLine {
	cloned := make([]ReversalAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedReversalLine{
		id: id, lineOrder: lineOrder, itemID: itemID, direction: direction,
		quantity: quantity, enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, commercialTotal: commercialTotal,
		reversesLineID: reversesLineID, allocations: cloned,
	}
}

func (l PostedReversalLine) ID() domain.StockDocumentLineID  { return l.id }
func (l PostedReversalLine) LineOrder() domain.LineOrder     { return l.lineOrder }
func (l PostedReversalLine) ItemID() domain.ItemID           { return l.itemID }
func (l PostedReversalLine) Direction() domain.Direction     { return l.direction }
func (l PostedReversalLine) Quantity() domain.AtomicQuantity { return l.quantity }
func (l PostedReversalLine) EnteredUnit() domain.UnitCode    { return l.enteredUnit }
func (l PostedReversalLine) EnteredPackagingName() domain.Option[domain.NonEmptyText] {
	return l.enteredPackagingName
}
func (l PostedReversalLine) Conversion() domain.UnitConversion     { return l.conversion }
func (l PostedReversalLine) InventoryValue() domain.InventoryValue { return l.inventoryValue }
func (l PostedReversalLine) CommercialTotal() domain.Option[domain.MinorAmount] {
	return l.commercialTotal
}
func (l PostedReversalLine) ReversesLineID() domain.StockDocumentLineID { return l.reversesLineID }
func (l PostedReversalLine) Allocations() []ReversalAllocation {
	allocations := make([]ReversalAllocation, len(l.allocations))
	copy(allocations, l.allocations)
	return allocations
}

type ReversalAllocation struct {
	id                   domain.LotAllocationID
	lotID                domain.InventoryLotID
	quantity             domain.AtomicQuantity
	restoresAllocationID domain.Option[domain.LotAllocationID]
}

func NewReversalAllocation(
	id domain.LotAllocationID,
	lotID domain.InventoryLotID,
	quantity domain.AtomicQuantity,
	restoresAllocationID domain.Option[domain.LotAllocationID],
) ReversalAllocation {
	return ReversalAllocation{
		id: id, lotID: lotID, quantity: quantity, restoresAllocationID: restoresAllocationID,
	}
}

func (a ReversalAllocation) ID() domain.LotAllocationID      { return a.id }
func (a ReversalAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a ReversalAllocation) Quantity() domain.AtomicQuantity { return a.quantity }
func (a ReversalAllocation) RestoresAllocationID() domain.Option[domain.LotAllocationID] {
	return a.restoresAllocationID
}

func (s *Store) PostReversal(ctx context.Context, input PostReversalInput) (PostedReversalDocument, error) {
	var posted PostedReversalDocument
	err := s.database.Write(ctx, func(tx *database.WriteTx) error {
		value, err := postReversalTx(ctx, tx, input)
		if err != nil {
			return err
		}
		posted = value
		return nil
	})
	if err != nil {
		return PostedReversalDocument{}, classifyError("post reversal", err)
	}
	return posted, nil
}

func postReversalTx(ctx context.Context, tx databaseWriteTx, input PostReversalInput) (PostedReversalDocument, error) {
	if input.IdempotencyKey.String() == "" {
		return PostedReversalDocument{}, domain.Invalid("idempotency_key", domain.ViolationRequired, "DOC-003")
	}
	if input.TargetDocumentID.IsZero() {
		return PostedReversalDocument{}, domain.Invalid("target_document_id", domain.ViolationRequired, "REV-001")
	}
	if input.OccurredOn.IsZero() {
		return PostedReversalDocument{}, domain.Invalid("occurred_on", domain.ViolationRequired, "DOC-004")
	}
	if input.PostedAt.IsZero() {
		return PostedReversalDocument{}, domain.Invalid("posted_at", domain.ViolationRequired, "DOC-004")
	}

	var existingID int64
	var existingKind string
	err := tx.QueryRowContext(ctx, `
		SELECT id, kind FROM stock_documents WHERE idempotency_key = ?
	`, input.IdempotencyKey.String()).Scan(&existingID, &existingKind)
	if err == nil {
		if existingKind != domain.DocumentReversal.String() {
			return PostedReversalDocument{}, fmt.Errorf("%w: idempotency key belongs to %s", domain.ErrConflict, existingKind)
		}
		return loadPostedReversalDocument(ctx, tx, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return PostedReversalDocument{}, err
	}

	target, err := loadReversalTarget(ctx, tx, input.TargetDocumentID.Int64())
	if err != nil {
		return PostedReversalDocument{}, err
	}
	if err := validateReversalEligibility(ctx, tx, target); err != nil {
		return PostedReversalDocument{}, err
	}

	currency, err := loadDocumentCurrency(ctx, tx)
	if err != nil {
		return PostedReversalDocument{}, err
	}

	var postingSequence int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(posting_sequence), 0) + 1 FROM stock_documents
	`).Scan(&postingSequence); err != nil {
		return PostedReversalDocument{}, err
	}

	var documentID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, occurred_on, posted_at_ms,
			currency_code, currency_minor_digits, reason_code, notes, reverses_document_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		domain.DocumentReversal.String(),
		input.IdempotencyKey.String(),
		postingSequence,
		input.OccurredOn.String(),
		input.PostedAt.UnixMilli(),
		currency.Code().String(),
		int64(currency.MinorDigits().Int()),
		domain.ReasonExactReversal.String(),
		nullableText(input.Notes),
		target.id,
	).Scan(&documentID); err != nil {
		return PostedReversalDocument{}, err
	}

	for _, line := range target.lines {
		if err := insertReversalLine(ctx, tx, documentID, input.PostedAt, line); err != nil {
			return PostedReversalDocument{}, fmt.Errorf("line %d: %w", line.lineOrder, err)
		}
	}

	return loadPostedReversalDocument(ctx, tx, documentID)
}

type reversalTarget struct {
	id    int64
	kind  string
	lines []reversalTargetLine
}

type reversalTargetLine struct {
	id, lineOrder, itemID, quantityAtomic            int64
	conversionNumeratorAtomic, conversionDenominator int64
	inventoryValueMicro                              int64
	commercialTotalMinor                             sql.NullInt64
	direction, enteredUnitCode                       string
	enteredPackagingName                             sql.NullString
	sourceLotID                                      sql.NullInt64
	allocations                                      []reversalTargetAllocation
}

type reversalTargetAllocation struct {
	id, lotID, quantityAtomic int64
}

func loadReversalTarget(ctx context.Context, tx databaseWriteTx, targetDocumentID int64) (reversalTarget, error) {
	var target reversalTarget
	err := tx.QueryRowContext(ctx, `
		SELECT id, kind
		FROM stock_documents
		WHERE id = ?
	`, targetDocumentID).Scan(&target.id, &target.kind)
	if err != nil {
		return reversalTarget{}, err
	}
	if target.kind == domain.DocumentReversal.String() {
		return reversalTarget{}, domain.Invalid("target_document_id", domain.ViolationInvalidEnum, "REV-006")
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT line.id, line.line_order, line.item_id, line.direction, line.quantity_atomic,
		       line.entered_unit_code, line.entered_packaging_name,
		       line.conversion_numerator_atomic, line.conversion_denominator,
		       line.inventory_value_micro, line.commercial_total_minor,
		       lot.id
		FROM stock_document_lines line
		LEFT JOIN inventory_lots lot ON lot.source_line_id = line.id
		WHERE line.document_id = ?
		ORDER BY line.line_order, line.id
	`, targetDocumentID)
	if err != nil {
		return reversalTarget{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var line reversalTargetLine
		if err := rows.Scan(
			&line.id,
			&line.lineOrder,
			&line.itemID,
			&line.direction,
			&line.quantityAtomic,
			&line.enteredUnitCode,
			&line.enteredPackagingName,
			&line.conversionNumeratorAtomic,
			&line.conversionDenominator,
			&line.inventoryValueMicro,
			&line.commercialTotalMinor,
			&line.sourceLotID,
		); err != nil {
			return reversalTarget{}, err
		}
		allocations, err := loadReversalTargetAllocations(ctx, tx, line.id)
		if err != nil {
			return reversalTarget{}, err
		}
		line.allocations = allocations
		target.lines = append(target.lines, line)
	}
	if err := rows.Err(); err != nil {
		return reversalTarget{}, err
	}
	if len(target.lines) == 0 {
		return reversalTarget{}, domain.Invalid("target_document_id", domain.ViolationRequired, "REV-001")
	}
	return target, nil
}

func loadReversalTargetAllocations(ctx context.Context, tx databaseWriteTx, lineID int64) ([]reversalTargetAllocation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, lot_id, quantity_atomic
		FROM lot_allocations
		WHERE line_id = ? AND restores_allocation_id IS NULL
		ORDER BY id
	`, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocations []reversalTargetAllocation
	for rows.Next() {
		var allocation reversalTargetAllocation
		if err := rows.Scan(&allocation.id, &allocation.lotID, &allocation.quantityAtomic); err != nil {
			return nil, err
		}
		allocations = append(allocations, allocation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return allocations, nil
}

func validateReversalEligibility(ctx context.Context, tx databaseWriteTx, target reversalTarget) error {
	var alreadyReversed int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM stock_documents WHERE reverses_document_id = ?
	`, target.id).Scan(&alreadyReversed); err != nil {
		return err
	}
	if alreadyReversed != 0 {
		return domain.Invalid("target_document_id", domain.ViolationInvariant, "REV-006")
	}

	for _, line := range target.lines {
		var lastDocumentID sql.NullInt64
		if err := tx.QueryRowContext(ctx, `
			SELECT last_document_id FROM inventory_balances WHERE item_id = ?
		`, line.itemID).Scan(&lastDocumentID); err != nil {
			return err
		}
		if !lastDocumentID.Valid || lastDocumentID.Int64 != target.id {
			return domain.Invalid("target_document_id", domain.ViolationInvariant, "REV-002")
		}
		if line.direction == domain.DirectionIn.String() {
			if !line.sourceLotID.Valid {
				return domain.Invalid("target_document_id", domain.ViolationInvariant, "REV-003")
			}
			available, err := lotAvailableQuantity(ctx, tx, line.sourceLotID.Int64)
			if err != nil {
				return err
			}
			if available != line.quantityAtomic {
				return domain.Invalid("target_document_id", domain.ViolationInvariant, "REV-003")
			}
		}
	}
	return nil
}

func lotAvailableQuantity(ctx context.Context, tx databaseWriteTx, lotID int64) (int64, error) {
	var available int64
	err := tx.QueryRowContext(ctx, `
		SELECT lot.initial_quantity_atomic
			- COALESCE(SUM(
				CASE WHEN allocation.restores_allocation_id IS NULL
					THEN allocation.quantity_atomic ELSE 0 END
			), 0)
			+ COALESCE(SUM(
				CASE WHEN allocation.restores_allocation_id IS NOT NULL
					THEN allocation.quantity_atomic ELSE 0 END
			), 0) AS remaining_quantity_atomic
		FROM inventory_lots lot
		LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
		WHERE lot.id = ?
		GROUP BY lot.id, lot.initial_quantity_atomic
	`, lotID).Scan(&available)
	return available, err
}

func insertReversalLine(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	postedAt domain.UTCInstant,
	target reversalTargetLine,
) error {
	reversalDirection := domain.DirectionIn.String()
	quantityDelta := target.quantityAtomic
	valueDelta := target.inventoryValueMicro
	if target.direction == domain.DirectionIn.String() {
		reversalDirection = domain.DirectionOut.String()
		quantityDelta = -quantityDelta
		valueDelta = -valueDelta
	}

	var lineID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, entered_packaging_name, conversion_numerator_atomic,
			conversion_denominator, inventory_value_micro, commercial_total_minor,
			reverses_line_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		documentID,
		target.lineOrder,
		target.itemID,
		reversalDirection,
		target.quantityAtomic,
		target.enteredUnitCode,
		nullableSQLString(target.enteredPackagingName),
		target.conversionNumeratorAtomic,
		target.conversionDenominator,
		target.inventoryValueMicro,
		nullableSQLInt64(target.commercialTotalMinor),
		target.id,
	).Scan(&lineID); err != nil {
		return err
	}

	if target.direction == domain.DirectionIn.String() {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lot_allocations (
				line_id, lot_id, quantity_atomic, restores_allocation_id, created_at_ms
			) VALUES (?, ?, ?, NULL, ?)
		`, lineID, target.sourceLotID.Int64, target.quantityAtomic, postedAt.UnixMilli()); err != nil {
			return err
		}
	} else {
		for _, allocation := range target.allocations {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO lot_allocations (
					line_id, lot_id, quantity_atomic, restores_allocation_id, created_at_ms
				) VALUES (?, ?, ?, ?, ?)
			`, lineID, allocation.lotID, allocation.quantityAtomic, allocation.id, postedAt.UnixMilli()); err != nil {
				return err
			}
		}
	}

	return updateReversalBalance(ctx, tx, documentID, postedAt, target.itemID, quantityDelta, valueDelta)
}

func updateReversalBalance(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	postedAt domain.UTCInstant,
	itemID int64,
	quantityDelta int64,
	valueDelta int64,
) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE inventory_balances
		SET quantity_atomic = quantity_atomic + ?,
		    inventory_value_micro = inventory_value_micro + ?,
		    last_document_id = ?,
		    updated_at_ms = ?
		WHERE item_id = ?
	`, quantityDelta, valueDelta, documentID, postedAt.UnixMilli(), itemID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("%w: inventory balance missing", domain.ErrInvalidReference)
	}
	return nil
}

func loadPostedReversalDocument(ctx context.Context, tx databaseWriteTx, id int64) (PostedReversalDocument, error) {
	var row postedReversalDocumentRow
	err := tx.QueryRowContext(ctx, `
		SELECT id, idempotency_key, posting_sequence, reverses_document_id, occurred_on,
		       posted_at_ms, currency_code, currency_minor_digits, notes
		FROM stock_documents
		WHERE id = ? AND kind = 'REVERSAL'
	`, id).Scan(
		&row.id,
		&row.idempotencyKey,
		&row.postingSequence,
		&row.targetDocumentID,
		&row.occurredOn,
		&row.postedAtMS,
		&row.currencyCode,
		&row.currencyMinorDigits,
		&row.notes,
	)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	lines, err := loadPostedReversalLines(ctx, tx, id)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	return mapPostedReversalDocument(row, lines)
}

type postedReversalDocumentRow struct {
	id, postingSequence, targetDocumentID, postedAtMS, currencyMinorDigits int64
	idempotencyKey, occurredOn, currencyCode                               string
	notes                                                                  sql.NullString
}

func loadPostedReversalLines(ctx context.Context, tx databaseWriteTx, documentID int64) ([]PostedReversalLine, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, line_order, item_id, direction, quantity_atomic,
		       entered_unit_code, entered_packaging_name,
		       conversion_numerator_atomic, conversion_denominator,
		       inventory_value_micro, commercial_total_minor, reverses_line_id
		FROM stock_document_lines
		WHERE document_id = ?
		ORDER BY line_order, id
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []PostedReversalLine
	for rows.Next() {
		var row postedReversalLineRow
		if err := rows.Scan(
			&row.id,
			&row.lineOrder,
			&row.itemID,
			&row.direction,
			&row.quantityAtomic,
			&row.enteredUnitCode,
			&row.enteredPackagingName,
			&row.conversionNumeratorAtomic,
			&row.conversionDenominator,
			&row.inventoryValueMicro,
			&row.commercialTotalMinor,
			&row.reversesLineID,
		); err != nil {
			return nil, err
		}
		allocations, err := loadReversalAllocations(ctx, tx, row.id)
		if err != nil {
			return nil, err
		}
		line, err := mapPostedReversalLine(row, allocations)
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

type postedReversalLineRow struct {
	id, lineOrder, itemID, quantityAtomic            int64
	conversionNumeratorAtomic, conversionDenominator int64
	inventoryValueMicro, reversesLineID              int64
	commercialTotalMinor                             sql.NullInt64
	direction, enteredUnitCode                       string
	enteredPackagingName                             sql.NullString
}

func loadReversalAllocations(ctx context.Context, tx databaseWriteTx, lineID int64) ([]ReversalAllocation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, lot_id, quantity_atomic, restores_allocation_id
		FROM lot_allocations
		WHERE line_id = ?
		ORDER BY id
	`, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocations []ReversalAllocation
	for rows.Next() {
		var rawID, rawLotID, rawQuantity int64
		var rawRestores sql.NullInt64
		if err := rows.Scan(&rawID, &rawLotID, &rawQuantity, &rawRestores); err != nil {
			return nil, err
		}
		id, err := domain.NewLotAllocationID(rawID)
		if err != nil {
			return nil, err
		}
		lotID, err := domain.NewInventoryLotID(rawLotID)
		if err != nil {
			return nil, err
		}
		quantity, err := domain.NewPositiveAtomicQuantity(rawQuantity)
		if err != nil {
			return nil, err
		}
		restores := domain.None[domain.LotAllocationID]()
		if rawRestores.Valid {
			parsed, err := domain.NewLotAllocationID(rawRestores.Int64)
			if err != nil {
				return nil, err
			}
			restores = domain.Some(parsed)
		}
		allocations = append(allocations, NewReversalAllocation(id, lotID, quantity, restores))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return allocations, nil
}

func mapPostedReversalDocument(row postedReversalDocumentRow, lines []PostedReversalLine) (PostedReversalDocument, error) {
	id, err := domain.NewStockDocumentID(row.id)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	idempotencyKey, err := domain.NewIdempotencyKey(row.idempotencyKey)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	postingSequence, err := domain.NewPostingSequence(row.postingSequence)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	targetDocumentID, err := domain.NewStockDocumentID(row.targetDocumentID)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	occurredOn, err := domain.ParseBusinessDate(row.occurredOn)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	postedAt, err := domain.UTCInstantFromUnixMilli(row.postedAtMS)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	currency, err := domain.RestoreCurrency(row.currencyCode, int(row.currencyMinorDigits))
	if err != nil {
		return PostedReversalDocument{}, err
	}
	notes, err := optionalNonEmptyText(row.notes)
	if err != nil {
		return PostedReversalDocument{}, err
	}
	return NewPostedReversalDocument(
		id, idempotencyKey, postingSequence, targetDocumentID, occurredOn,
		postedAt, currency, notes, lines,
	), nil
}

func mapPostedReversalLine(row postedReversalLineRow, allocations []ReversalAllocation) (PostedReversalLine, error) {
	id, err := domain.NewStockDocumentLineID(row.id)
	if err != nil {
		return PostedReversalLine{}, err
	}
	lineOrder, err := domain.NewLineOrder(row.lineOrder)
	if err != nil {
		return PostedReversalLine{}, err
	}
	itemID, err := domain.NewItemID(row.itemID)
	if err != nil {
		return PostedReversalLine{}, err
	}
	direction, err := domain.ParseDirection(row.direction)
	if err != nil {
		return PostedReversalLine{}, err
	}
	quantity, err := domain.NewPositiveAtomicQuantity(row.quantityAtomic)
	if err != nil {
		return PostedReversalLine{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.enteredUnitCode)
	if err != nil {
		return PostedReversalLine{}, err
	}
	enteredPackagingName, err := optionalNonEmptyText(row.enteredPackagingName)
	if err != nil {
		return PostedReversalLine{}, err
	}
	conversion, err := domain.NewUnitConversion(row.conversionNumeratorAtomic, row.conversionDenominator)
	if err != nil {
		return PostedReversalLine{}, err
	}
	inventoryValue, err := domain.NewInventoryValue(row.inventoryValueMicro)
	if err != nil {
		return PostedReversalLine{}, err
	}
	commercialTotal, err := optionalMinorAmount(row.commercialTotalMinor)
	if err != nil {
		return PostedReversalLine{}, err
	}
	reversesLineID, err := domain.NewStockDocumentLineID(row.reversesLineID)
	if err != nil {
		return PostedReversalLine{}, err
	}
	return NewPostedReversalLine(
		id, lineOrder, itemID, direction, quantity, enteredUnit, enteredPackagingName,
		conversion, inventoryValue, commercialTotal, reversesLineID, allocations,
	), nil
}

func nullableSQLString(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func nullableSQLInt64(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}
