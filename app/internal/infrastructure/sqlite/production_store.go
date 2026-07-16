package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
)

type PostProductionInput struct {
	IdempotencyKey   domain.IdempotencyKey
	RecipeRevisionID domain.RecipeRevisionID
	OccurredOn       domain.BusinessDate
	PostedAt         domain.UTCInstant
	DirectCost       domain.InventoryValue
	Notes            domain.Option[domain.NonEmptyText]
	Output           PostProductionOutputInput
	Inputs           []PostProductionComponentInput
}

type PostProductionOutputInput struct {
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	LotCode              domain.Option[domain.NonEmptyText]
	ExpiresOn            domain.Option[domain.BusinessDate]
}

type PostProductionComponentInput struct {
	ItemID               domain.ItemID
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	LotID                domain.Option[domain.InventoryLotID]
}

type PostedProductionDocument struct {
	id               domain.StockDocumentID
	idempotencyKey   domain.IdempotencyKey
	postingSequence  domain.PostingSequence
	recipeRevisionID domain.RecipeRevisionID
	outputItemID     domain.ItemID
	occurredOn       domain.BusinessDate
	postedAt         domain.UTCInstant
	currency         domain.Currency
	directCost       domain.InventoryValue
	notes            domain.Option[domain.NonEmptyText]
	outputLine       PostedProductionLine
	inputLines       []PostedProductionLine
}

func NewPostedProductionDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	recipeRevisionID domain.RecipeRevisionID,
	outputItemID domain.ItemID,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	directCost domain.InventoryValue,
	notes domain.Option[domain.NonEmptyText],
	outputLine PostedProductionLine,
	inputLines []PostedProductionLine,
) PostedProductionDocument {
	cloned := make([]PostedProductionLine, len(inputLines))
	copy(cloned, inputLines)
	return PostedProductionDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		recipeRevisionID: recipeRevisionID, outputItemID: outputItemID,
		occurredOn: occurredOn, postedAt: postedAt, currency: currency,
		directCost: directCost, notes: notes, outputLine: outputLine,
		inputLines: cloned,
	}
}

func (d PostedProductionDocument) ID() domain.StockDocumentID              { return d.id }
func (d PostedProductionDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d PostedProductionDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d PostedProductionDocument) RecipeRevisionID() domain.RecipeRevisionID {
	return d.recipeRevisionID
}
func (d PostedProductionDocument) OutputItemID() domain.ItemID               { return d.outputItemID }
func (d PostedProductionDocument) OccurredOn() domain.BusinessDate           { return d.occurredOn }
func (d PostedProductionDocument) PostedAt() domain.UTCInstant               { return d.postedAt }
func (d PostedProductionDocument) Currency() domain.Currency                 { return d.currency }
func (d PostedProductionDocument) DirectCost() domain.InventoryValue         { return d.directCost }
func (d PostedProductionDocument) Notes() domain.Option[domain.NonEmptyText] { return d.notes }
func (d PostedProductionDocument) OutputLine() PostedProductionLine          { return d.outputLine }
func (d PostedProductionDocument) InputLines() []PostedProductionLine {
	lines := make([]PostedProductionLine, len(d.inputLines))
	copy(lines, d.inputLines)
	return lines
}

type PostedProductionLine struct {
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
	allocations          []ProductionAllocation
}

func NewPostedProductionLine(
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
	allocations []ProductionAllocation,
) PostedProductionLine {
	cloned := make([]ProductionAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedProductionLine{
		id: id, lineOrder: lineOrder, itemID: itemID, direction: direction,
		quantity: quantity, enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, lotID: lotID,
		lotCode: lotCode, originatedOn: originatedOn, expiresOn: expiresOn,
		allocations: cloned,
	}
}

func (l PostedProductionLine) ID() domain.StockDocumentLineID  { return l.id }
func (l PostedProductionLine) LineOrder() domain.LineOrder     { return l.lineOrder }
func (l PostedProductionLine) ItemID() domain.ItemID           { return l.itemID }
func (l PostedProductionLine) Direction() domain.Direction     { return l.direction }
func (l PostedProductionLine) Quantity() domain.AtomicQuantity { return l.quantity }
func (l PostedProductionLine) EnteredUnit() domain.UnitCode    { return l.enteredUnit }
func (l PostedProductionLine) EnteredPackagingName() domain.Option[domain.NonEmptyText] {
	return l.enteredPackagingName
}
func (l PostedProductionLine) Conversion() domain.UnitConversion           { return l.conversion }
func (l PostedProductionLine) InventoryValue() domain.InventoryValue       { return l.inventoryValue }
func (l PostedProductionLine) LotID() domain.Option[domain.InventoryLotID] { return l.lotID }
func (l PostedProductionLine) LotCode() domain.Option[domain.NonEmptyText] { return l.lotCode }
func (l PostedProductionLine) OriginatedOn() domain.Option[domain.BusinessDate] {
	return l.originatedOn
}
func (l PostedProductionLine) ExpiresOn() domain.Option[domain.BusinessDate] {
	return l.expiresOn
}
func (l PostedProductionLine) Allocations() []ProductionAllocation {
	allocations := make([]ProductionAllocation, len(l.allocations))
	copy(allocations, l.allocations)
	return allocations
}

type ProductionAllocation struct {
	id       domain.LotAllocationID
	lotID    domain.InventoryLotID
	quantity domain.AtomicQuantity
}

func NewProductionAllocation(id domain.LotAllocationID, lotID domain.InventoryLotID, quantity domain.AtomicQuantity) ProductionAllocation {
	return ProductionAllocation{id: id, lotID: lotID, quantity: quantity}
}

func (a ProductionAllocation) ID() domain.LotAllocationID      { return a.id }
func (a ProductionAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a ProductionAllocation) Quantity() domain.AtomicQuantity { return a.quantity }

func (s *Store) PostProduction(ctx context.Context, input PostProductionInput) (PostedProductionDocument, error) {
	var posted PostedProductionDocument
	err := s.database.Write(ctx, func(tx *database.WriteTx) error {
		value, err := postProductionTx(ctx, tx, input)
		if err != nil {
			return err
		}
		posted = value
		return nil
	})
	if err != nil {
		return PostedProductionDocument{}, classifyError("post production", err)
	}
	return posted, nil
}

func postProductionTx(ctx context.Context, tx databaseWriteTx, input PostProductionInput) (PostedProductionDocument, error) {
	if err := validateProductionInput(input); err != nil {
		return PostedProductionDocument{}, err
	}

	var existingID int64
	var existingKind string
	err := tx.QueryRowContext(ctx, `
		SELECT id, kind FROM stock_documents WHERE idempotency_key = ?
	`, input.IdempotencyKey.String()).Scan(&existingID, &existingKind)
	if err == nil {
		if existingKind != domain.DocumentProduction.String() {
			return PostedProductionDocument{}, fmt.Errorf("%w: idempotency key belongs to %s", domain.ErrConflict, existingKind)
		}
		return loadPostedProductionDocument(ctx, tx, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return PostedProductionDocument{}, err
	}

	revision, err := loadProductionRecipeRevision(ctx, tx, input.RecipeRevisionID)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	if err := validateProductionComponents(revision.outputItemID, input.Inputs); err != nil {
		return PostedProductionDocument{}, err
	}

	currency, err := loadDocumentCurrency(ctx, tx)
	if err != nil {
		return PostedProductionDocument{}, err
	}

	var postingSequence int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(posting_sequence), 0) + 1 FROM stock_documents
	`).Scan(&postingSequence); err != nil {
		return PostedProductionDocument{}, err
	}

	var documentID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, occurred_on,
			posted_at_ms, currency_code, currency_minor_digits, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		domain.DocumentProduction.String(),
		input.IdempotencyKey.String(),
		postingSequence,
		input.OccurredOn.String(),
		input.PostedAt.UnixMilli(),
		currency.Code().String(),
		int64(currency.MinorDigits().Int()),
		nullableText(input.Notes),
	).Scan(&documentID); err != nil {
		return PostedProductionDocument{}, err
	}

	totalConsumedValue, err := insertProductionInputLines(ctx, tx, documentID, input)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	outputValue, err := totalConsumedValue.Add(input.DirectCost)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	outputLineID, err := insertProductionOutputLine(ctx, tx, documentID, int64(len(input.Inputs)+1), revision.outputItemID, input, outputValue)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO production_runs (
			document_id, recipe_revision_id, output_line_id, direct_production_cost_micro
		) VALUES (?, ?, ?, ?)
	`, documentID, input.RecipeRevisionID.Int64(), outputLineID, input.DirectCost.Int64()); err != nil {
		return PostedProductionDocument{}, err
	}

	return loadPostedProductionDocument(ctx, tx, documentID)
}

type productionRecipeRevision struct {
	outputItemID domain.ItemID
}

func loadProductionRecipeRevision(ctx context.Context, tx databaseWriteTx, revisionID domain.RecipeRevisionID) (productionRecipeRevision, error) {
	var outputItemIDValue int64
	err := tx.QueryRowContext(ctx, `
		SELECT recipe.output_item_id
		FROM recipe_revisions revision
		JOIN recipes recipe ON recipe.id = revision.recipe_id
		WHERE revision.id = ?
		  AND recipe.archived_at_ms IS NULL
	`, revisionID.Int64()).Scan(&outputItemIDValue)
	if err != nil {
		return productionRecipeRevision{}, err
	}
	outputItemID, err := domain.NewItemID(outputItemIDValue)
	if err != nil {
		return productionRecipeRevision{}, err
	}
	return productionRecipeRevision{outputItemID: outputItemID}, nil
}

func validateProductionInput(input PostProductionInput) error {
	if input.IdempotencyKey.String() == "" {
		return domain.Invalid("idempotency_key", domain.ViolationRequired, "DOC-003")
	}
	if input.RecipeRevisionID.IsZero() {
		return domain.Invalid("recipe_revision_id", domain.ViolationRequired, "PRO-001")
	}
	if input.OccurredOn.IsZero() {
		return domain.Invalid("occurred_on", domain.ViolationRequired, "DOC-004")
	}
	if input.PostedAt.IsZero() {
		return domain.Invalid("posted_at", domain.ViolationRequired, "DOC-004")
	}
	if input.Output.Quantity.Int64() <= 0 {
		return domain.Invalid("output.quantity_atomic", domain.ViolationNotPositive, "DOC-005")
	}
	if input.Output.EnteredUnit.String() == "" {
		return domain.Invalid("output.entered_unit_code", domain.ViolationRequired, "DOC-005")
	}
	if input.Output.Conversion.IsZero() {
		return domain.Invalid("output.conversion", domain.ViolationRequired, "DOC-005")
	}
	if expiresOn, ok := input.Output.ExpiresOn.Get(); ok && expiresOn.Before(input.OccurredOn) {
		return domain.Invalid("output.expires_on", domain.ViolationOutOfRange, "LOT-009")
	}
	if len(input.Inputs) == 0 {
		return domain.Invalid("inputs", domain.ViolationRequired, "PRO-001")
	}
	for index, line := range input.Inputs {
		if line.ItemID.IsZero() {
			return domain.Invalid(fmt.Sprintf("inputs[%d].item_id", index), domain.ViolationRequired, "DOC-005")
		}
		if line.Quantity.Int64() <= 0 {
			return domain.Invalid(fmt.Sprintf("inputs[%d].quantity_atomic", index), domain.ViolationNotPositive, "DOC-005")
		}
		if line.EnteredUnit.String() == "" {
			return domain.Invalid(fmt.Sprintf("inputs[%d].entered_unit_code", index), domain.ViolationRequired, "DOC-005")
		}
		if line.Conversion.IsZero() {
			return domain.Invalid(fmt.Sprintf("inputs[%d].conversion", index), domain.ViolationRequired, "DOC-005")
		}
	}
	return nil
}

func validateProductionComponents(outputItemID domain.ItemID, inputs []PostProductionComponentInput) error {
	seen := make(map[int64]struct{}, len(inputs))
	for _, line := range inputs {
		if line.ItemID == outputItemID {
			return domain.Invalid("inputs.item_id", domain.ViolationInvariant, "PRO-002")
		}
		if _, ok := seen[line.ItemID.Int64()]; ok {
			return domain.Invalid("inputs.item_id", domain.ViolationDuplicate, "PRO-002")
		}
		seen[line.ItemID.Int64()] = struct{}{}
	}
	return nil
}

func insertProductionInputLines(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	input PostProductionInput,
) (domain.InventoryValue, error) {
	total, err := domain.NewInventoryValue(0)
	if err != nil {
		return domain.InventoryValue{}, err
	}
	for index, line := range input.Inputs {
		balance, err := readAdjustmentBalance(ctx, tx, line.ItemID)
		if err != nil {
			return domain.InventoryValue{}, err
		}
		if line.Quantity.Int64() > balance.quantityAtomic {
			return domain.InventoryValue{}, domain.Invalid("quantity_atomic", domain.ViolationOutOfRange, "INV-004")
		}
		inventoryValue, err := weightedAverageValue(balance.inventoryValueMicro, balance.quantityAtomic, line.Quantity.Int64())
		if err != nil {
			return domain.InventoryValue{}, err
		}

		lineID, err := insertProductionLine(ctx, tx, documentID, int64(index+1), line.ItemID, domain.DirectionOut, line.Quantity,
			line.EnteredUnit, line.EnteredPackagingName, line.Conversion, inventoryValue)
		if err != nil {
			return domain.InventoryValue{}, fmt.Errorf("input line %d: %w", index+1, err)
		}
		if err := allocateProductionLots(ctx, tx, lineID, line.ItemID, line.Quantity.Int64(), input.OccurredOn, input.PostedAt, line.LotID); err != nil {
			return domain.InventoryValue{}, fmt.Errorf("input line %d: %w", index+1, err)
		}
		if err := updateAdjustmentBalance(ctx, tx, documentID, input.PostedAt, line.ItemID, -line.Quantity.Int64(), -inventoryValue.Int64()); err != nil {
			return domain.InventoryValue{}, fmt.Errorf("input line %d: %w", index+1, err)
		}
		total, err = total.Add(inventoryValue)
		if err != nil {
			return domain.InventoryValue{}, err
		}
	}
	return total, nil
}

func insertProductionOutputLine(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	lineOrder int64,
	outputItemID domain.ItemID,
	input PostProductionInput,
	inventoryValue domain.InventoryValue,
) (int64, error) {
	lineID, err := insertProductionLine(ctx, tx, documentID, lineOrder, outputItemID, domain.DirectionIn, input.Output.Quantity,
		input.Output.EnteredUnit, input.Output.EnteredPackagingName, input.Output.Conversion, inventoryValue)
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO inventory_lots (
			item_id, source_line_id, initial_quantity_atomic, lot_code,
			originated_on, expires_on, created_at_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		outputItemID.Int64(),
		lineID,
		input.Output.Quantity.Int64(),
		nullableText(input.Output.LotCode),
		input.OccurredOn.String(),
		nullableBusinessDate(input.Output.ExpiresOn),
		input.PostedAt.UnixMilli(),
	); err != nil {
		return 0, err
	}
	if err := updateAdjustmentBalance(ctx, tx, documentID, input.PostedAt, outputItemID, input.Output.Quantity.Int64(), inventoryValue.Int64()); err != nil {
		return 0, err
	}
	return lineID, nil
}

func insertProductionLine(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	lineOrder int64,
	itemID domain.ItemID,
	direction domain.Direction,
	quantity domain.AtomicQuantity,
	enteredUnit domain.UnitCode,
	enteredPackagingName domain.Option[domain.NonEmptyText],
	conversion domain.UnitConversion,
	inventoryValue domain.InventoryValue,
) (int64, error) {
	var lineID int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, entered_packaging_name, conversion_numerator_atomic,
			conversion_denominator, inventory_value_micro, commercial_total_minor
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)
		RETURNING id
	`,
		documentID,
		lineOrder,
		itemID.Int64(),
		direction.String(),
		quantity.Int64(),
		enteredUnit.String(),
		nullableText(enteredPackagingName),
		conversion.NumeratorAtomic(),
		conversion.Denominator(),
		inventoryValue.Int64(),
	).Scan(&lineID)
	return lineID, err
}

func allocateProductionLots(
	ctx context.Context,
	tx databaseWriteTx,
	lineID int64,
	itemID domain.ItemID,
	quantityAtomic int64,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	lotID domain.Option[domain.InventoryLotID],
) error {
	if override, ok := lotID.Get(); ok {
		available, err := availableProductionLotQuantity(ctx, tx, override, itemID, occurredOn)
		if err != nil {
			return err
		}
		if available < quantityAtomic {
			return domain.Invalid("quantity_atomic", domain.ViolationOutOfRange, "INV-004")
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO lot_allocations (
				line_id, lot_id, quantity_atomic, restores_allocation_id, created_at_ms
			) VALUES (?, ?, ?, NULL, ?)
		`, lineID, override.Int64(), quantityAtomic, postedAt.UnixMilli())
		return err
	}
	return allocateAdjustmentFEFO(ctx, tx, lineID, itemID, quantityAtomic, occurredOn, postedAt)
}

func availableProductionLotQuantity(
	ctx context.Context,
	tx databaseWriteTx,
	lotID domain.InventoryLotID,
	itemID domain.ItemID,
	occurredOn domain.BusinessDate,
) (int64, error) {
	var available int64
	err := tx.QueryRowContext(ctx, `
		SELECT remaining_quantity_atomic
		FROM (
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
				lot.expires_on
			FROM inventory_lots lot
			LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
			WHERE lot.id = ? AND lot.item_id = ?
			GROUP BY lot.id, lot.initial_quantity_atomic, lot.expires_on
		)
		WHERE remaining_quantity_atomic > 0
		  AND (expires_on IS NULL OR expires_on >= ?)
	`, lotID.Int64(), itemID.Int64(), occurredOn.String()).Scan(&available)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.Invalid("lot_id", domain.ViolationInvariant, "LOT-009")
	}
	return available, err
}

func loadPostedProductionDocument(ctx context.Context, tx databaseWriteTx, id int64) (PostedProductionDocument, error) {
	var row postedProductionDocumentRow
	err := tx.QueryRowContext(ctx, `
		SELECT document.id, document.idempotency_key, document.posting_sequence,
		       run.recipe_revision_id, recipe.output_item_id,
		       document.occurred_on, document.posted_at_ms,
		       document.currency_code, document.currency_minor_digits,
		       run.direct_production_cost_micro, document.notes
		FROM stock_documents document
		JOIN production_runs run ON run.document_id = document.id
		JOIN recipe_revisions revision ON revision.id = run.recipe_revision_id
		JOIN recipes recipe ON recipe.id = revision.recipe_id
		WHERE document.id = ? AND document.kind = 'PRODUCTION'
	`, id).Scan(
		&row.id,
		&row.idempotencyKey,
		&row.postingSequence,
		&row.recipeRevisionID,
		&row.outputItemID,
		&row.occurredOn,
		&row.postedAtMS,
		&row.currencyCode,
		&row.currencyMinorDigits,
		&row.directProductionCostMicro,
		&row.notes,
	)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	outputLine, err := loadPostedProductionOutputLine(ctx, tx, id)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	inputLines, err := loadPostedProductionInputLines(ctx, tx, id)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	return mapPostedProductionDocument(row, outputLine, inputLines)
}

type postedProductionDocumentRow struct {
	id, postingSequence, recipeRevisionID, outputItemID int64
	postedAtMS, currencyMinorDigits                     int64
	directProductionCostMicro                           int64
	idempotencyKey, occurredOn, currencyCode            string
	notes                                               sql.NullString
}

func loadPostedProductionOutputLine(ctx context.Context, tx databaseWriteTx, documentID int64) (PostedProductionLine, error) {
	var row postedProductionLineRow
	err := tx.QueryRowContext(ctx, `
		SELECT line.id, line.line_order, line.item_id, line.direction, line.quantity_atomic,
		       line.entered_unit_code, line.entered_packaging_name,
		       line.conversion_numerator_atomic, line.conversion_denominator,
		       line.inventory_value_micro,
		       lot.id, lot.lot_code, lot.originated_on, lot.expires_on
		FROM production_runs run
		JOIN stock_document_lines line ON line.id = run.output_line_id
		JOIN inventory_lots lot ON lot.source_line_id = line.id
		WHERE run.document_id = ?
	`, documentID).Scan(
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
	)
	if err != nil {
		return PostedProductionLine{}, err
	}
	return mapPostedProductionLine(row, nil)
}

func loadPostedProductionInputLines(ctx context.Context, tx databaseWriteTx, documentID int64) ([]PostedProductionLine, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT line.id, line.line_order, line.item_id, line.direction, line.quantity_atomic,
		       line.entered_unit_code, line.entered_packaging_name,
		       line.conversion_numerator_atomic, line.conversion_denominator,
		       line.inventory_value_micro
		FROM stock_document_lines line
		WHERE line.document_id = ? AND line.direction = 'OUT'
		ORDER BY line.line_order, line.id
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []PostedProductionLine
	for rows.Next() {
		var row postedProductionLineRow
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
		); err != nil {
			return nil, err
		}
		allocations, err := loadProductionAllocations(ctx, tx, row.id)
		if err != nil {
			return nil, err
		}
		line, err := mapPostedProductionLine(row, allocations)
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

type postedProductionLineRow struct {
	id, lineOrder, itemID, quantityAtomic            int64
	conversionNumeratorAtomic, conversionDenominator int64
	inventoryValueMicro, lotID                       int64
	direction, enteredUnitCode, originatedOn         string
	enteredPackagingName, lotCode, expiresOn         sql.NullString
}

func loadProductionAllocations(ctx context.Context, tx databaseWriteTx, lineID int64) ([]ProductionAllocation, error) {
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

	var allocations []ProductionAllocation
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
		allocations = append(allocations, NewProductionAllocation(id, lotID, quantity))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return allocations, nil
}

func mapPostedProductionDocument(
	row postedProductionDocumentRow,
	outputLine PostedProductionLine,
	inputLines []PostedProductionLine,
) (PostedProductionDocument, error) {
	id, err := domain.NewStockDocumentID(row.id)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	idempotencyKey, err := domain.NewIdempotencyKey(row.idempotencyKey)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	postingSequence, err := domain.NewPostingSequence(row.postingSequence)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	recipeRevisionID, err := domain.NewRecipeRevisionID(row.recipeRevisionID)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	outputItemID, err := domain.NewItemID(row.outputItemID)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	occurredOn, err := domain.ParseBusinessDate(row.occurredOn)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	postedAt, err := domain.UTCInstantFromUnixMilli(row.postedAtMS)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	currency, err := domain.RestoreCurrency(row.currencyCode, int(row.currencyMinorDigits))
	if err != nil {
		return PostedProductionDocument{}, err
	}
	directCost, err := domain.NewInventoryValue(row.directProductionCostMicro)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	notes, err := optionalNonEmptyText(row.notes)
	if err != nil {
		return PostedProductionDocument{}, err
	}
	return NewPostedProductionDocument(
		id, idempotencyKey, postingSequence, recipeRevisionID, outputItemID,
		occurredOn, postedAt, currency, directCost, notes, outputLine, inputLines,
	), nil
}

func mapPostedProductionLine(row postedProductionLineRow, allocations []ProductionAllocation) (PostedProductionLine, error) {
	id, err := domain.NewStockDocumentLineID(row.id)
	if err != nil {
		return PostedProductionLine{}, err
	}
	lineOrder, err := domain.NewLineOrder(row.lineOrder)
	if err != nil {
		return PostedProductionLine{}, err
	}
	itemID, err := domain.NewItemID(row.itemID)
	if err != nil {
		return PostedProductionLine{}, err
	}
	direction, err := domain.ParseDirection(row.direction)
	if err != nil {
		return PostedProductionLine{}, err
	}
	quantity, err := domain.NewAtomicQuantity(row.quantityAtomic)
	if err != nil {
		return PostedProductionLine{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.enteredUnitCode)
	if err != nil {
		return PostedProductionLine{}, err
	}
	enteredPackagingName, err := optionalNonEmptyText(row.enteredPackagingName)
	if err != nil {
		return PostedProductionLine{}, err
	}
	conversion, err := domain.NewUnitConversion(row.conversionNumeratorAtomic, row.conversionDenominator)
	if err != nil {
		return PostedProductionLine{}, err
	}
	inventoryValue, err := domain.NewInventoryValue(row.inventoryValueMicro)
	if err != nil {
		return PostedProductionLine{}, err
	}
	lotID := domain.None[domain.InventoryLotID]()
	if row.lotID != 0 {
		id, err := domain.NewInventoryLotID(row.lotID)
		if err != nil {
			return PostedProductionLine{}, err
		}
		lotID = domain.Some(id)
	}
	lotCode, err := optionalNonEmptyText(row.lotCode)
	if err != nil {
		return PostedProductionLine{}, err
	}
	originatedOn, err := optionalBusinessDate(sql.NullString{String: row.originatedOn, Valid: row.originatedOn != ""})
	if err != nil {
		return PostedProductionLine{}, err
	}
	expiresOn, err := optionalBusinessDate(row.expiresOn)
	if err != nil {
		return PostedProductionLine{}, err
	}
	return NewPostedProductionLine(
		id, lineOrder, itemID, direction, quantity, enteredUnit, enteredPackagingName,
		conversion, inventoryValue, lotID, lotCode, originatedOn, expiresOn, allocations,
	), nil
}
