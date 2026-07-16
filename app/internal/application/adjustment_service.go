package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
)

type AdjustmentStore interface {
	PostAdjustment(ctx context.Context, input adjustmentPostStoreInput) (AdjustmentDocument, error)
}

type AdjustmentLineInput struct {
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

type AdjustmentPostInput struct {
	IdempotencyKey domain.IdempotencyKey
	OccurredOn     domain.BusinessDate
	Reason         domain.DocumentReason
	Notes          domain.Option[domain.NonEmptyText]
	Lines          []AdjustmentLineInput
}

type adjustmentPostStoreInput struct {
	AdjustmentPostInput
	PostedAt domain.UTCInstant
}

type AdjustmentDocument struct {
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

func NewAdjustmentDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	reason domain.DocumentReason,
	notes domain.Option[domain.NonEmptyText],
	lines []PostedAdjustmentLine,
) (AdjustmentDocument, error) {
	violations := make([]domain.Violation, 0, 7)
	if id.IsZero() {
		violations = append(violations, domain.Violation{Field: "document_id", Code: domain.ViolationRequired})
	}
	if idempotencyKey.String() == "" {
		violations = append(violations, domain.Violation{Field: "idempotency_key", Code: domain.ViolationRequired})
	}
	if postingSequence.IsZero() {
		violations = append(violations, domain.Violation{Field: "posting_sequence", Code: domain.ViolationRequired})
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
	if len(lines) == 0 {
		violations = append(violations, domain.Violation{Field: "lines", Code: domain.ViolationRequired})
	}
	if _, err := domain.ParseDocumentReason(domain.DocumentAdjustment, reason.String()); err != nil {
		violations = append(violations, domain.Violation{Field: "reason", Code: domain.ViolationInvalidEnum})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return AdjustmentDocument{}, err
	}
	cloned := make([]PostedAdjustmentLine, len(lines))
	copy(cloned, lines)
	return AdjustmentDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		occurredOn: occurredOn, postedAt: postedAt, currency: currency, reason: reason,
		notes: notes, lines: cloned,
	}, nil
}

func (d AdjustmentDocument) ID() domain.StockDocumentID              { return d.id }
func (d AdjustmentDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d AdjustmentDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d AdjustmentDocument) OccurredOn() domain.BusinessDate         { return d.occurredOn }
func (d AdjustmentDocument) PostedAt() domain.UTCInstant             { return d.postedAt }
func (d AdjustmentDocument) Currency() domain.Currency               { return d.currency }
func (d AdjustmentDocument) Reason() domain.DocumentReason           { return d.reason }
func (d AdjustmentDocument) Notes() domain.Option[domain.NonEmptyText] {
	return d.notes
}
func (d AdjustmentDocument) Lines() []PostedAdjustmentLine {
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
) (PostedAdjustmentLine, error) {
	violations := make([]domain.Violation, 0, 8)
	if id.IsZero() {
		violations = append(violations, domain.Violation{Field: "line_id", Code: domain.ViolationRequired})
	}
	if lineOrder.IsZero() {
		violations = append(violations, domain.Violation{Field: "line_order", Code: domain.ViolationRequired})
	}
	if itemID.IsZero() {
		violations = append(violations, domain.Violation{Field: "item_id", Code: domain.ViolationRequired})
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
	if direction != domain.DirectionIn && direction != domain.DirectionOut {
		violations = append(violations, domain.Violation{Field: "direction", Code: domain.ViolationInvalidEnum})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return PostedAdjustmentLine{}, err
	}
	cloned := make([]AdjustmentAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedAdjustmentLine{
		id: id, lineOrder: lineOrder, itemID: itemID, direction: direction,
		quantity: quantity, enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, lotID: lotID, lotCode: lotCode,
		originatedOn: originatedOn, expiresOn: expiresOn, allocations: cloned,
	}, nil
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

func NewAdjustmentAllocation(id domain.LotAllocationID, lotID domain.InventoryLotID, quantity domain.AtomicQuantity) (AdjustmentAllocation, error) {
	if id.IsZero() || lotID.IsZero() || quantity.Int64() <= 0 {
		return AdjustmentAllocation{}, domain.ErrInvariant
	}
	return AdjustmentAllocation{id: id, lotID: lotID, quantity: quantity}, nil
}

func (a AdjustmentAllocation) ID() domain.LotAllocationID      { return a.id }
func (a AdjustmentAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a AdjustmentAllocation) Quantity() domain.AtomicQuantity { return a.quantity }

type AdjustmentService struct {
	store AdjustmentStore
	clock Clock
}

func NewAdjustmentService(store AdjustmentStore, clock Clock) *AdjustmentService {
	if store == nil {
		panic("adjustment service requires a store")
	}
	if clock == nil {
		panic("adjustment service requires a clock")
	}
	return &AdjustmentService{store: store, clock: clock}
}

func (s *AdjustmentService) PostAdjustment(ctx context.Context, input AdjustmentPostInput) (AdjustmentDocument, error) {
	if len(input.Lines) == 0 {
		return AdjustmentDocument{}, domain.Invalid("lines", domain.ViolationRequired, "DOC-002")
	}
	postedAt, err := s.clock.Now()
	if err != nil {
		return AdjustmentDocument{}, fmt.Errorf("read clock: %w", err)
	}
	document, err := s.store.PostAdjustment(ctx, adjustmentPostStoreInput{
		AdjustmentPostInput: input,
		PostedAt:            postedAt,
	})
	if err != nil {
		return AdjustmentDocument{}, fmt.Errorf("post adjustment: %w", err)
	}
	if !document.PostedAt().Equal(postedAt) {
		return AdjustmentDocument{}, domain.ErrInvariant
	}
	return document, nil
}
