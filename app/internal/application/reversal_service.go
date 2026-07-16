package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
)

type ReversalStore interface {
	PostReversal(ctx context.Context, input reversalPostStoreInput) (ReversalDocument, error)
}

type ReversalPostInput struct {
	IdempotencyKey   domain.IdempotencyKey
	TargetDocumentID domain.StockDocumentID
	OccurredOn       domain.BusinessDate
	Notes            domain.Option[domain.NonEmptyText]
}

type reversalPostStoreInput struct {
	ReversalPostInput
	PostedAt domain.UTCInstant
}

type ReversalDocument struct {
	id               domain.StockDocumentID
	idempotencyKey   domain.IdempotencyKey
	postingSequence  domain.PostingSequence
	targetDocumentID domain.StockDocumentID
	occurredOn       domain.BusinessDate
	postedAt         domain.UTCInstant
	currency         domain.Currency
	reason           domain.DocumentReason
	notes            domain.Option[domain.NonEmptyText]
	lines            []PostedReversalLine
}

func NewReversalDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	targetDocumentID domain.StockDocumentID,
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	reason domain.DocumentReason,
	notes domain.Option[domain.NonEmptyText],
	lines []PostedReversalLine,
) (ReversalDocument, error) {
	violations := make([]domain.Violation, 0, 8)
	if id.IsZero() {
		violations = append(violations, domain.Violation{Field: "document_id", Code: domain.ViolationRequired})
	}
	if idempotencyKey.String() == "" {
		violations = append(violations, domain.Violation{Field: "idempotency_key", Code: domain.ViolationRequired})
	}
	if postingSequence.IsZero() {
		violations = append(violations, domain.Violation{Field: "posting_sequence", Code: domain.ViolationRequired})
	}
	if targetDocumentID.IsZero() {
		violations = append(violations, domain.Violation{Field: "target_document_id", Code: domain.ViolationRequired})
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
	if reason != domain.ReasonExactReversal {
		violations = append(violations, domain.Violation{Field: "reason", Code: domain.ViolationInvalidEnum})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return ReversalDocument{}, err
	}
	cloned := make([]PostedReversalLine, len(lines))
	copy(cloned, lines)
	return ReversalDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		targetDocumentID: targetDocumentID, occurredOn: occurredOn, postedAt: postedAt,
		currency: currency, reason: reason, notes: notes, lines: cloned,
	}, nil
}

func (d ReversalDocument) ID() domain.StockDocumentID              { return d.id }
func (d ReversalDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d ReversalDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d ReversalDocument) TargetDocumentID() domain.StockDocumentID {
	return d.targetDocumentID
}
func (d ReversalDocument) OccurredOn() domain.BusinessDate           { return d.occurredOn }
func (d ReversalDocument) PostedAt() domain.UTCInstant               { return d.postedAt }
func (d ReversalDocument) Currency() domain.Currency                 { return d.currency }
func (d ReversalDocument) Reason() domain.DocumentReason             { return d.reason }
func (d ReversalDocument) Notes() domain.Option[domain.NonEmptyText] { return d.notes }
func (d ReversalDocument) Lines() []PostedReversalLine {
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
) (PostedReversalLine, error) {
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
	if reversesLineID.IsZero() {
		violations = append(violations, domain.Violation{Field: "reverses_line_id", Code: domain.ViolationRequired})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return PostedReversalLine{}, err
	}
	cloned := make([]ReversalAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedReversalLine{
		id: id, lineOrder: lineOrder, itemID: itemID, direction: direction,
		quantity: quantity, enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, commercialTotal: commercialTotal,
		reversesLineID: reversesLineID, allocations: cloned,
	}, nil
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
) (ReversalAllocation, error) {
	if id.IsZero() || lotID.IsZero() || quantity.Int64() <= 0 {
		return ReversalAllocation{}, domain.ErrInvariant
	}
	return ReversalAllocation{
		id: id, lotID: lotID, quantity: quantity, restoresAllocationID: restoresAllocationID,
	}, nil
}

func (a ReversalAllocation) ID() domain.LotAllocationID      { return a.id }
func (a ReversalAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a ReversalAllocation) Quantity() domain.AtomicQuantity { return a.quantity }
func (a ReversalAllocation) RestoresAllocationID() domain.Option[domain.LotAllocationID] {
	return a.restoresAllocationID
}

type ReversalService struct {
	store ReversalStore
	clock Clock
}

func NewReversalService(store ReversalStore, clock Clock) *ReversalService {
	if store == nil {
		panic("reversal service requires a store")
	}
	if clock == nil {
		panic("reversal service requires a clock")
	}
	return &ReversalService{store: store, clock: clock}
}

func (s *ReversalService) PostReversal(ctx context.Context, input ReversalPostInput) (ReversalDocument, error) {
	postedAt, err := s.clock.Now()
	if err != nil {
		return ReversalDocument{}, fmt.Errorf("read clock: %w", err)
	}
	document, err := s.store.PostReversal(ctx, reversalPostStoreInput{
		ReversalPostInput: input,
		PostedAt:          postedAt,
	})
	if err != nil {
		return ReversalDocument{}, fmt.Errorf("post reversal: %w", err)
	}
	if !document.PostedAt().Equal(postedAt) {
		return ReversalDocument{}, domain.ErrInvariant
	}
	return document, nil
}
