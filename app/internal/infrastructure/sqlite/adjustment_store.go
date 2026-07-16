package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
)

type PostAdjustmentInput struct {
	IdempotencyKey domain.IdempotencyKey
	OccurredOn     domain.BusinessDate
	PostedAt       domain.UTCInstant
	Reason         domain.DocumentReason
	Notes          domain.Option[domain.NonEmptyText]
	Lines          []PostAdjustmentLineInput
}

type PostAdjustmentLineInput struct {
	ItemID               domain.ItemID
	Direction            domain.Direction
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	InventoryValue       domain.Option[domain.InventoryValue]
	LotCode              domain.Option[domain.NonEmptyText]
	ExpiresOn            domain.Option[domain.BusinessDate]
}

type PostedAdjustmentDocument struct {
	id              domain.StockDocumentID
	idempotencyKey  domain.IdempotencyKey
	postingSequence domain.PostingSequence
	occurredOn      domain.BusinessDate
	postedAt        domain.UTCInstant
	currency        domain.Currency
	reason          domain.DocumentReason
	notes           domain.Option[domain.NonEmptyText]
	lines           []PostedAdjustmentLine
}

func NewPostedAdjustmentDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	reason domain.DocumentReason,
	notes domain.Option[domain.NonEmptyText],
	lines []PostedAdjustmentLine,
) PostedAdjustmentDocument {
	cloned := make([]PostedAdjustmentLine, len(lines))
	copy(cloned, lines)
	return PostedAdjustmentDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		occurredOn: occurredOn, postedAt: postedAt, currency: currency, reason: reason,
		notes: notes, lines: cloned,
	}
}

func (d PostedAdjustmentDocument) ID() domain.StockDocumentID              { return d.id }
func (d PostedAdjustmentDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d PostedAdjustmentDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d PostedAdjustmentDocument) OccurredOn() domain.BusinessDate         { return d.occurredOn }
func (d PostedAdjustmentDocument) PostedAt() domain.UTCInstant             { return d.postedAt }
func (d PostedAdjustmentDocument) Currency() domain.Currency               { return d.currency }
func (d PostedAdjustmentDocument) Reason() domain.DocumentReason           { return d.reason }
func (d PostedAdjustmentDocument) Notes() domain.Option[domain.NonEmptyText] {
	return d.notes
}
func (d PostedAdjustmentDocument) Lines() []PostedAdjustmentLine {
	lines := make([]PostedAdjustmentLine, len(d.lines))
	copy(lines, d.lines)
	return lines
}

type PostedAdjustmentLine struct {
	id                   domain.StockDocumentLineID
	lineOrder            domain.LineOrder
	itemID               domain.ItemID
	direction            domain.Direction
	quantity             domain.AtomicQuantity
	enteredUnit          domain.UnitCode
	enteredPackagingName domain.Option[domain.NonEmptyText]
	conversion           domain.UnitConversion
	inventoryValue       domain.InventoryValue
	lotID                domain.Option[domain.InventoryLotID]
	lotCode              domain.Option[domain.NonEmptyText]
	originatedOn         domain.Option[domain.BusinessDate]
	expiresOn            domain.Option[domain.BusinessDate]
	allocations          []AdjustmentAllocation
}

func NewPostedAdjustmentLine(
	id domain.StockDocumentLineID,
	lineOrder domain.LineOrder,
	itemID domain.ItemID,
	direction domain.Direction,
	quantity domain.AtomicQuantity,
	enteredUnit domain.UnitCode,
	enteredPackagingName domain.Option[domain.NonEmptyText],
	conversion domain.UnitConversion,
	inventoryValue domain.InventoryValue,
	lotID domain.Option[domain.InventoryLotID],
	lotCode domain.Option[domain.NonEmptyText],
	originatedOn domain.Option[domain.BusinessDate],
	expiresOn domain.Option[domain.BusinessDate],
	allocations []AdjustmentAllocation,
) PostedAdjustmentLine {
	cloned := make([]AdjustmentAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedAdjustmentLine{
		id: id, lineOrder: lineOrder, itemID: itemID, direction: direction,
		quantity: quantity, enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, lotID: lotID, lotCode: lotCode,
		originatedOn: originatedOn, expiresOn: expiresOn, allocations: cloned,
	}
}

func (l PostedAdjustmentLine) ID() domain.StockDocumentLineID  { return l.id }
func (l PostedAdjustmentLine) LineOrder() domain.LineOrder     { return l.lineOrder }
func (l PostedAdjustmentLine) ItemID() domain.ItemID           { return l.itemID }
func (l PostedAdjustmentLine) Direction() domain.Direction     { return l.direction }
func (l PostedAdjustmentLine) Quantity() domain.AtomicQuantity { return l.quantity }
func (l PostedAdjustmentLine) EnteredUnit() domain.UnitCode    { return l.enteredUnit }
func (l PostedAdjustmentLine) EnteredPackagingName() domain.Option[domain.NonEmptyText] {
	return l.enteredPackagingName
}
func (l PostedAdjustmentLine) Conversion() domain.UnitConversion           { return l.conversion }
func (l PostedAdjustmentLine) InventoryValue() domain.InventoryValue       { return l.inventoryValue }
func (l PostedAdjustmentLine) LotID() domain.Option[domain.InventoryLotID] { return l.lotID }
func (l PostedAdjustmentLine) LotCode() domain.Option[domain.NonEmptyText] { return l.lotCode }
func (l PostedAdjustmentLine) OriginatedOn() domain.Option[domain.BusinessDate] {
	return l.originatedOn
}
func (l PostedAdjustmentLine) ExpiresOn() domain.Option[domain.BusinessDate] {
	return l.expiresOn
}
func (l PostedAdjustmentLine) Allocations() []AdjustmentAllocation {
	allocations := make([]AdjustmentAllocation, len(l.allocations))
	copy(allocations, l.allocations)
	return allocations
}

type AdjustmentAllocation struct {
	id       domain.LotAllocationID
	lotID    domain.InventoryLotID
	quantity domain.AtomicQuantity
}

func NewAdjustmentAllocation(id domain.LotAllocationID, lotID domain.InventoryLotID, quantity domain.AtomicQuantity) AdjustmentAllocation {
	return AdjustmentAllocation{id: id, lotID: lotID, quantity: quantity}
}

func (a AdjustmentAllocation) ID() domain.LotAllocationID      { return a.id }
func (a AdjustmentAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a AdjustmentAllocation) Quantity() domain.AtomicQuantity { return a.quantity }

func (s *Store) PostAdjustment(ctx context.Context, input PostAdjustmentInput) (PostedAdjustmentDocument, error) {
	var posted PostedAdjustmentDocument
	err := s.database.Write(ctx, func(tx *database.WriteTx) error {
		value, err := postAdjustmentTx(ctx, tx, input)
		if err != nil {
			return err
		}
		posted = value
		return nil
	})
	if err != nil {
		return PostedAdjustmentDocument{}, classifyError("post adjustment", err)
	}
	return posted, nil
}

func postAdjustmentTx(ctx context.Context, tx databaseWriteTx, input PostAdjustmentInput) (PostedAdjustmentDocument, error) {
	if input.IdempotencyKey.String() == "" {
		return PostedAdjustmentDocument{}, domain.Invalid("idempotency_key", domain.ViolationRequired, "DOC-003")
	}
	if input.OccurredOn.IsZero() {
		return PostedAdjustmentDocument{}, domain.Invalid("occurred_on", domain.ViolationRequired, "DOC-004")
	}
	if input.PostedAt.IsZero() {
		return PostedAdjustmentDocument{}, domain.Invalid("posted_at", domain.ViolationRequired, "DOC-004")
	}
	if len(input.Lines) == 0 {
		return PostedAdjustmentDocument{}, domain.Invalid("lines", domain.ViolationRequired, "DOC-002")
	}
	if _, err := domain.ParseDocumentReason(domain.DocumentAdjustment, input.Reason.String()); err != nil {
		return PostedAdjustmentDocument{}, err
	}

	var existingID int64
	var existingKind string
	err := tx.QueryRowContext(ctx, `
		SELECT id, kind FROM stock_documents WHERE idempotency_key = ?
	`, input.IdempotencyKey.String()).Scan(&existingID, &existingKind)
	if err == nil {
		if existingKind != domain.DocumentAdjustment.String() {
			return PostedAdjustmentDocument{}, fmt.Errorf("%w: idempotency key belongs to %s", domain.ErrConflict, existingKind)
		}
		return loadPostedAdjustmentDocument(ctx, tx, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return PostedAdjustmentDocument{}, err
	}

	currency, err := loadDocumentCurrency(ctx, tx)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}

	var postingSequence int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(posting_sequence), 0) + 1 FROM stock_documents
	`).Scan(&postingSequence); err != nil {
		return PostedAdjustmentDocument{}, err
	}

	var documentID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, occurred_on,
			posted_at_ms, currency_code, currency_minor_digits, reason_code, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		domain.DocumentAdjustment.String(),
		input.IdempotencyKey.String(),
		postingSequence,
		input.OccurredOn.String(),
		input.PostedAt.UnixMilli(),
		currency.Code().String(),
		int64(currency.MinorDigits().Int()),
		input.Reason.String(),
		nullableText(input.Notes),
	).Scan(&documentID); err != nil {
		return PostedAdjustmentDocument{}, err
	}

	for index, line := range input.Lines {
		if err := insertPostedAdjustmentLine(ctx, tx, documentID, int64(index+1), input.OccurredOn, input.PostedAt, input.Reason, line); err != nil {
			return PostedAdjustmentDocument{}, fmt.Errorf("line %d: %w", index+1, err)
		}
	}

	return loadPostedAdjustmentDocument(ctx, tx, documentID)
}

func insertPostedAdjustmentLine(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	lineOrder int64,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	reason domain.DocumentReason,
	line PostAdjustmentLineInput,
) error {
	if err := validateAdjustmentLineInput(reason, line); err != nil {
		return err
	}

	balance, err := readAdjustmentBalance(ctx, tx, line.ItemID)
	if err != nil {
		return err
	}
	inventoryValue, err := adjustmentLineValue(balance, reason, line)
	if err != nil {
		return err
	}

	var lineID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, entered_packaging_name, conversion_numerator_atomic,
			conversion_denominator, inventory_value_micro, commercial_total_minor
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)
		RETURNING id
	`,
		documentID,
		lineOrder,
		line.ItemID.Int64(),
		line.Direction.String(),
		line.Quantity.Int64(),
		line.EnteredUnit.String(),
		nullableText(line.EnteredPackagingName),
		line.Conversion.NumeratorAtomic(),
		line.Conversion.Denominator(),
		inventoryValue.Int64(),
	).Scan(&lineID); err != nil {
		return err
	}

	switch line.Direction {
	case domain.DirectionIn:
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inventory_lots (
				item_id, source_line_id, initial_quantity_atomic, lot_code,
				originated_on, expires_on, created_at_ms
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
			line.ItemID.Int64(),
			lineID,
			line.Quantity.Int64(),
			nullableText(line.LotCode),
			occurredOn.String(),
			nullableBusinessDate(line.ExpiresOn),
			postedAt.UnixMilli(),
		); err != nil {
			return err
		}
		return updateAdjustmentBalance(ctx, tx, documentID, postedAt, line.ItemID, line.Quantity.Int64(), inventoryValue.Int64())
	case domain.DirectionOut:
		if err := allocateAdjustmentFEFO(ctx, tx, lineID, line.ItemID, line.Quantity.Int64(), occurredOn, postedAt); err != nil {
			return err
		}
		return updateAdjustmentBalance(ctx, tx, documentID, postedAt, line.ItemID, -line.Quantity.Int64(), -inventoryValue.Int64())
	default:
		return domain.Invalid("direction", domain.ViolationInvalidEnum, "DOC-008")
	}
}

func validateAdjustmentLineInput(reason domain.DocumentReason, line PostAdjustmentLineInput) error {
	if line.ItemID.IsZero() {
		return domain.Invalid("item_id", domain.ViolationRequired, "DOC-005")
	}
	if line.Quantity.Int64() <= 0 {
		return domain.Invalid("quantity_atomic", domain.ViolationNotPositive, "DOC-005")
	}
	if line.EnteredUnit.String() == "" {
		return domain.Invalid("entered_unit_code", domain.ViolationRequired, "DOC-005")
	}
	if line.Conversion.IsZero() {
		return domain.Invalid("conversion", domain.ViolationRequired, "DOC-005")
	}
	if reason == domain.ReasonOpeningBalance || reason == domain.ReasonFreeStock {
		if line.Direction != domain.DirectionIn {
			return domain.Invalid("direction", domain.ViolationInvalidEnum, "ADJ-002")
		}
	}
	switch reason {
	case domain.ReasonWaste, domain.ReasonExpiry, domain.ReasonDamage, domain.ReasonSample:
		if line.Direction != domain.DirectionOut {
			return domain.Invalid("direction", domain.ViolationInvalidEnum, "ADJ-002")
		}
	}
	if line.Direction != domain.DirectionIn && line.Direction != domain.DirectionOut {
		return domain.Invalid("direction", domain.ViolationInvalidEnum, "DOC-008")
	}
	return nil
}

type adjustmentBalance struct {
	quantityAtomic      int64
	inventoryValueMicro int64
}

func readAdjustmentBalance(ctx context.Context, tx databaseWriteTx, itemID domain.ItemID) (adjustmentBalance, error) {
	var balance adjustmentBalance
	err := tx.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro
		FROM inventory_balances
		WHERE item_id = ?
	`, itemID.Int64()).Scan(&balance.quantityAtomic, &balance.inventoryValueMicro)
	if err != nil {
		return adjustmentBalance{}, err
	}
	return balance, nil
}

func adjustmentLineValue(balance adjustmentBalance, reason domain.DocumentReason, line PostAdjustmentLineInput) (domain.InventoryValue, error) {
	if line.Direction == domain.DirectionIn {
		if value, ok := line.InventoryValue.Get(); ok {
			if value.IsZero() && reason != domain.ReasonFreeStock {
				return domain.InventoryValue{}, domain.Invalid("inventory_value_micro", domain.ViolationNotPositive, "ADJ-003")
			}
			return value, nil
		}
		if balance.quantityAtomic <= 0 {
			return domain.InventoryValue{}, domain.Invalid("inventory_value_micro", domain.ViolationRequired, "ADJ-003")
		}
		return weightedAverageValue(balance.inventoryValueMicro, balance.quantityAtomic, line.Quantity.Int64())
	}
	if line.InventoryValue.IsSome() {
		return domain.InventoryValue{}, domain.Invalid("inventory_value_micro", domain.ViolationInvariant, "ADJ-002")
	}
	if line.Quantity.Int64() > balance.quantityAtomic {
		return domain.InventoryValue{}, domain.Invalid("quantity_atomic", domain.ViolationOutOfRange, "INV-004")
	}
	return weightedAverageValue(balance.inventoryValueMicro, balance.quantityAtomic, line.Quantity.Int64())
}

func weightedAverageValue(totalValueMicro, totalQuantityAtomic, quantityAtomic int64) (domain.InventoryValue, error) {
	if quantityAtomic < 0 || totalQuantityAtomic <= 0 || quantityAtomic > totalQuantityAtomic {
		return domain.InventoryValue{}, domain.Invalid("quantity_atomic", domain.ViolationOutOfRange, "INV-004")
	}
	if quantityAtomic == totalQuantityAtomic {
		return domain.NewInventoryValue(totalValueMicro)
	}
	numerator := big.NewInt(totalValueMicro)
	numerator.Mul(numerator, big.NewInt(quantityAtomic))
	denominator := big.NewInt(totalQuantityAtomic)
	quotient, remainder := new(big.Int).QuoRem(numerator, denominator, new(big.Int))
	remainder.Mul(remainder, big.NewInt(2))
	if remainder.Cmp(denominator) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if !quotient.IsInt64() || quotient.Int64() > math.MaxInt64 {
		return domain.InventoryValue{}, domain.ErrOverflow
	}
	return domain.NewInventoryValue(quotient.Int64())
}

func allocateAdjustmentFEFO(
	ctx context.Context,
	tx databaseWriteTx,
	lineID int64,
	itemID domain.ItemID,
	quantityAtomic int64,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
) error {
	rows, err := tx.QueryContext(ctx, `
		WITH lot_facts AS (
			SELECT
				lot.id,
				lot.initial_quantity_atomic
					- COALESCE(SUM(
						CASE WHEN allocation.restores_allocation_id IS NULL
							THEN allocation.quantity_atomic ELSE 0 END
					), 0)
					+ COALESCE(SUM(
						CASE WHEN allocation.restores_allocation_id IS NOT NULL
							THEN allocation.quantity_atomic ELSE 0 END
					), 0) AS remaining_quantity_atomic,
				lot.expires_on,
				source_document.posting_sequence
			FROM inventory_lots lot
			JOIN stock_document_lines source_line ON source_line.id = lot.source_line_id
			JOIN stock_documents source_document ON source_document.id = source_line.document_id
			LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
			WHERE lot.item_id = ?
			GROUP BY lot.id, lot.initial_quantity_atomic, lot.expires_on, source_document.posting_sequence
		)
		SELECT id, remaining_quantity_atomic
		FROM lot_facts
		WHERE remaining_quantity_atomic > 0
		  AND (expires_on IS NULL OR expires_on >= ?)
		ORDER BY expires_on IS NULL, expires_on, posting_sequence, id
	`, itemID.Int64(), occurredOn.String())
	if err != nil {
		return err
	}
	defer rows.Close()

	remaining := quantityAtomic
	for rows.Next() {
		var lotID, available int64
		if err := rows.Scan(&lotID, &available); err != nil {
			return err
		}
		if remaining <= 0 {
			continue
		}
		consume := available
		if consume > remaining {
			consume = remaining
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lot_allocations (
				line_id, lot_id, quantity_atomic, restores_allocation_id, created_at_ms
			) VALUES (?, ?, ?, NULL, ?)
		`, lineID, lotID, consume, postedAt.UnixMilli()); err != nil {
			return err
		}
		remaining -= consume
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if remaining != 0 {
		return domain.Invalid("quantity_atomic", domain.ViolationOutOfRange, "INV-004")
	}
	return nil
}

func updateAdjustmentBalance(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	postedAt domain.UTCInstant,
	itemID domain.ItemID,
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
	`,
		quantityDelta,
		valueDelta,
		documentID,
		postedAt.UnixMilli(),
		itemID.Int64(),
	)
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

func loadPostedAdjustmentDocument(ctx context.Context, tx databaseWriteTx, id int64) (PostedAdjustmentDocument, error) {
	var row postedAdjustmentDocumentRow
	err := tx.QueryRowContext(ctx, `
		SELECT id, idempotency_key, posting_sequence, occurred_on,
		       posted_at_ms, currency_code, currency_minor_digits, reason_code, notes
		FROM stock_documents
		WHERE id = ? AND kind = 'ADJUSTMENT'
	`, id).Scan(
		&row.id,
		&row.idempotencyKey,
		&row.postingSequence,
		&row.occurredOn,
		&row.postedAtMS,
		&row.currencyCode,
		&row.currencyMinorDigits,
		&row.reasonCode,
		&row.notes,
	)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	lines, err := loadPostedAdjustmentLines(ctx, tx, id)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	return mapPostedAdjustmentDocument(row, lines)
}

type postedAdjustmentDocumentRow struct {
	id, postingSequence, postedAtMS, currencyMinorDigits int64
	idempotencyKey, occurredOn, currencyCode, reasonCode string
	notes                                                sql.NullString
}

func loadPostedAdjustmentLines(ctx context.Context, tx databaseWriteTx, documentID int64) ([]PostedAdjustmentLine, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT line.id, line.line_order, line.item_id, line.direction, line.quantity_atomic,
		       line.entered_unit_code, line.entered_packaging_name,
		       line.conversion_numerator_atomic, line.conversion_denominator,
		       line.inventory_value_micro,
		       lot.id, lot.lot_code, lot.originated_on, lot.expires_on
		FROM stock_document_lines line
		LEFT JOIN inventory_lots lot ON lot.source_line_id = line.id
		WHERE line.document_id = ?
		ORDER BY line.line_order, line.id
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []PostedAdjustmentLine
	for rows.Next() {
		var row postedAdjustmentLineRow
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
			&row.lotID,
			&row.lotCode,
			&row.originatedOn,
			&row.expiresOn,
		); err != nil {
			return nil, err
		}
		allocations, err := loadAdjustmentAllocations(ctx, tx, row.id)
		if err != nil {
			return nil, err
		}
		line, err := mapPostedAdjustmentLine(row, allocations)
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

type postedAdjustmentLineRow struct {
	id, lineOrder, itemID, quantityAtomic            int64
	conversionNumeratorAtomic, conversionDenominator int64
	inventoryValueMicro                              int64
	direction, enteredUnitCode                       string
	enteredPackagingName                             sql.NullString
	lotID                                            sql.NullInt64
	lotCode, originatedOn, expiresOn                 sql.NullString
}

func loadAdjustmentAllocations(ctx context.Context, tx databaseWriteTx, lineID int64) ([]AdjustmentAllocation, error) {
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

	var allocations []AdjustmentAllocation
	for rows.Next() {
		var rawID, rawLotID, rawQuantity int64
		if err := rows.Scan(&rawID, &rawLotID, &rawQuantity); err != nil {
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
		allocations = append(allocations, NewAdjustmentAllocation(id, lotID, quantity))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return allocations, nil
}

func mapPostedAdjustmentDocument(row postedAdjustmentDocumentRow, lines []PostedAdjustmentLine) (PostedAdjustmentDocument, error) {
	id, err := domain.NewStockDocumentID(row.id)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	idempotencyKey, err := domain.NewIdempotencyKey(row.idempotencyKey)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	postingSequence, err := domain.NewPostingSequence(row.postingSequence)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	occurredOn, err := domain.ParseBusinessDate(row.occurredOn)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	postedAt, err := domain.UTCInstantFromUnixMilli(row.postedAtMS)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	currency, err := domain.RestoreCurrency(row.currencyCode, int(row.currencyMinorDigits))
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	reason, err := domain.ParseDocumentReason(domain.DocumentAdjustment, row.reasonCode)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	parsedReason, ok := reason.Get()
	if !ok {
		return PostedAdjustmentDocument{}, domain.ErrInvariant
	}
	notes, err := optionalNonEmptyText(row.notes)
	if err != nil {
		return PostedAdjustmentDocument{}, err
	}
	return NewPostedAdjustmentDocument(
		id, idempotencyKey, postingSequence, occurredOn, postedAt, currency,
		parsedReason, notes, lines,
	), nil
}

func mapPostedAdjustmentLine(row postedAdjustmentLineRow, allocations []AdjustmentAllocation) (PostedAdjustmentLine, error) {
	id, err := domain.NewStockDocumentLineID(row.id)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	lineOrder, err := domain.NewLineOrder(row.lineOrder)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	itemID, err := domain.NewItemID(row.itemID)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	direction, err := domain.ParseDirection(row.direction)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	quantity, err := domain.NewPositiveAtomicQuantity(row.quantityAtomic)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.enteredUnitCode)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	enteredPackagingName, err := optionalNonEmptyText(row.enteredPackagingName)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	conversion, err := domain.NewUnitConversion(row.conversionNumeratorAtomic, row.conversionDenominator)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	inventoryValue, err := domain.NewInventoryValue(row.inventoryValueMicro)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	lotID := domain.None[domain.InventoryLotID]()
	if row.lotID.Valid {
		parsed, err := domain.NewInventoryLotID(row.lotID.Int64)
		if err != nil {
			return PostedAdjustmentLine{}, err
		}
		lotID = domain.Some(parsed)
	}
	lotCode, err := optionalNonEmptyText(row.lotCode)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	originatedOn, err := optionalBusinessDate(row.originatedOn)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	expiresOn, err := optionalBusinessDate(row.expiresOn)
	if err != nil {
		return PostedAdjustmentLine{}, err
	}
	return NewPostedAdjustmentLine(
		id, lineOrder, itemID, direction, quantity, enteredUnit, enteredPackagingName,
		conversion, inventoryValue, lotID, lotCode, originatedOn, expiresOn, allocations,
	), nil
}
