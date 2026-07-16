package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
)

type ProductionStore interface {
	PostProduction(ctx context.Context, input productionPostStoreInput) (ProductionDocument, error)
}

type ProductionOutputInput struct {
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	LotCode              domain.Option[domain.NonEmptyText]
	ExpiresOn            domain.Option[domain.BusinessDate]
}

type ProductionComponentInput struct {
	ItemID               domain.ItemID
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	LotID                domain.Option[domain.InventoryLotID]
}

type ProductionPostInput struct {
	IdempotencyKey   domain.IdempotencyKey
	RecipeRevisionID domain.RecipeRevisionID
	OccurredOn       domain.BusinessDate
	DirectCost       domain.InventoryValue
	Notes            domain.Option[domain.NonEmptyText]
	Output           ProductionOutputInput
	Inputs           []ProductionComponentInput
}

type productionPostStoreInput struct {
	ProductionPostInput
	PostedAt domain.UTCInstant
}

type ProductionDocument struct {
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

func NewProductionDocument(
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
) (ProductionDocument, error) {
	violations := make([]domain.Violation, 0, 10)
	if id.IsZero() {
		violations = append(violations, domain.Violation{Field: "document_id", Code: domain.ViolationRequired})
	}
	if idempotencyKey.String() == "" {
		violations = append(violations, domain.Violation{Field: "idempotency_key", Code: domain.ViolationRequired})
	}
	if postingSequence.IsZero() {
		violations = append(violations, domain.Violation{Field: "posting_sequence", Code: domain.ViolationRequired})
	}
	if recipeRevisionID.IsZero() {
		violations = append(violations, domain.Violation{Field: "recipe_revision_id", Code: domain.ViolationRequired})
	}
	if outputItemID.IsZero() {
		violations = append(violations, domain.Violation{Field: "output_item_id", Code: domain.ViolationRequired})
	}
	if occurredOn.IsZero() {
		violations = append(violations, domain.Violation{Field: "occurred_on", Code: domain.ViolationRequired})
	}
	if postedAt.IsZero() {
		violations = append(violations, domain.Violation{Field: "posted_at", Code: domain.ViolationRequired})
	}
	if currency.IsZero() {
		violations = append(violations, domain.Violation{Field: "currency", Code: domain.ViolationRequired})
	}
	if outputLine.ID().IsZero() || outputLine.Direction() != domain.DirectionIn {
		violations = append(violations, domain.Violation{Field: "output_line", Code: domain.ViolationInvariant, InvariantID: "PRO-001"})
	}
	if len(inputLines) == 0 {
		violations = append(violations, domain.Violation{Field: "input_lines", Code: domain.ViolationRequired, InvariantID: "PRO-001"})
	}
	for _, line := range inputLines {
		if line.Direction() != domain.DirectionOut {
			violations = append(violations, domain.Violation{Field: "input_lines.direction", Code: domain.ViolationInvariant, InvariantID: "PRO-001"})
			break
		}
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return ProductionDocument{}, err
	}
	cloned := make([]PostedProductionLine, len(inputLines))
	copy(cloned, inputLines)
	return ProductionDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		recipeRevisionID: recipeRevisionID, outputItemID: outputItemID,
		occurredOn: occurredOn, postedAt: postedAt, currency: currency,
		directCost: directCost, notes: notes, outputLine: outputLine,
		inputLines: cloned,
	}, nil
}

func (d ProductionDocument) ID() domain.StockDocumentID              { return d.id }
func (d ProductionDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d ProductionDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d ProductionDocument) RecipeRevisionID() domain.RecipeRevisionID {
	return d.recipeRevisionID
}
func (d ProductionDocument) OutputItemID() domain.ItemID               { return d.outputItemID }
func (d ProductionDocument) OccurredOn() domain.BusinessDate           { return d.occurredOn }
func (d ProductionDocument) PostedAt() domain.UTCInstant               { return d.postedAt }
func (d ProductionDocument) Currency() domain.Currency                 { return d.currency }
func (d ProductionDocument) DirectCost() domain.InventoryValue         { return d.directCost }
func (d ProductionDocument) Notes() domain.Option[domain.NonEmptyText] { return d.notes }
func (d ProductionDocument) OutputLine() PostedProductionLine          { return d.outputLine }
func (d ProductionDocument) InputLines() []PostedProductionLine {
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
) (PostedProductionLine, error) {
	violations := make([]domain.Violation, 0, 7)
	if id.IsZero() {
		violations = append(violations, domain.Violation{Field: "line_id", Code: domain.ViolationRequired})
	}
	if lineOrder.IsZero() {
		violations = append(violations, domain.Violation{Field: "line_order", Code: domain.ViolationRequired})
	}
	if itemID.IsZero() {
		violations = append(violations, domain.Violation{Field: "item_id", Code: domain.ViolationRequired})
	}
	if direction != domain.DirectionIn && direction != domain.DirectionOut {
		violations = append(violations, domain.Violation{Field: "direction", Code: domain.ViolationInvalidEnum})
	}
	if quantity.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "quantity_atomic", Code: domain.ViolationNotPositive})
	}
	if enteredUnit.String() == "" {
		violations = append(violations, domain.Violation{Field: "entered_unit_code", Code: domain.ViolationRequired})
	}
	if conversion.IsZero() {
		violations = append(violations, domain.Violation{Field: "conversion", Code: domain.ViolationRequired})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return PostedProductionLine{}, err
	}
	cloned := make([]ProductionAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedProductionLine{
		id: id, lineOrder: lineOrder, itemID: itemID, direction: direction,
		quantity: quantity, enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, lotID: lotID,
		lotCode: lotCode, originatedOn: originatedOn, expiresOn: expiresOn,
		allocations: cloned,
	}, nil
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

func NewProductionAllocation(id domain.LotAllocationID, lotID domain.InventoryLotID, quantity domain.AtomicQuantity) (ProductionAllocation, error) {
	if id.IsZero() || lotID.IsZero() || quantity.Int64() <= 0 {
		return ProductionAllocation{}, domain.ErrInvariant
	}
	return ProductionAllocation{id: id, lotID: lotID, quantity: quantity}, nil
}

func (a ProductionAllocation) ID() domain.LotAllocationID      { return a.id }
func (a ProductionAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a ProductionAllocation) Quantity() domain.AtomicQuantity { return a.quantity }

type ProductionService struct {
	store ProductionStore
	clock Clock
}

func NewProductionService(store ProductionStore, clock Clock) *ProductionService {
	if store == nil {
		panic("production service requires a store")
	}
	if clock == nil {
		panic("production service requires a clock")
	}
	return &ProductionService{store: store, clock: clock}
}

func (s *ProductionService) PostProduction(ctx context.Context, input ProductionPostInput) (ProductionDocument, error) {
	if len(input.Inputs) == 0 {
		return ProductionDocument{}, domain.Invalid("inputs", domain.ViolationRequired, "PRO-001")
	}
	postedAt, err := s.clock.Now()
	if err != nil {
		return ProductionDocument{}, fmt.Errorf("read clock: %w", err)
	}
	document, err := s.store.PostProduction(ctx, productionPostStoreInput{
		ProductionPostInput: input,
		PostedAt:            postedAt,
	})
	if err != nil {
		return ProductionDocument{}, fmt.Errorf("post production: %w", err)
	}
	if !document.PostedAt().Equal(postedAt) {
		return ProductionDocument{}, domain.ErrInvariant
	}
	return document, nil
}
