package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
)

type SaleStore interface {
	GetSale(ctx context.Context, id domain.StockDocumentID) (SaleDocument, error)
	ListSales(ctx context.Context, input SaleListInput) (SalePage, error)
	PostSale(ctx context.Context, input salePostStoreInput) (SaleDocument, error)
}

type SaleCursor struct {
	PostingSequence domain.PostingSequence
	ID              domain.StockDocumentID
}

type SaleListInput struct {
	After    domain.Option[SaleCursor]
	PageSize int
}

type SalePage struct {
	items []SaleDocument
	next  domain.Option[SaleCursor]
}

func NewSalePage(items []SaleDocument, next domain.Option[SaleCursor]) SalePage {
	cloned := make([]SaleDocument, len(items))
	copy(cloned, items)
	return SalePage{items: cloned, next: next}
}

func (p SalePage) Items() []SaleDocument {
	items := make([]SaleDocument, len(p.items))
	copy(items, p.items)
	return items
}

func (p SalePage) Next() domain.Option[SaleCursor] { return p.next }

type SaleLineInput struct {
	ItemID               domain.ItemID
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	CommercialTotal      domain.MinorAmount
	LotID                domain.Option[domain.InventoryLotID]
}

type SalePostInput struct {
	IdempotencyKey domain.IdempotencyKey
	CounterpartyID domain.Option[domain.CounterpartyID]
	OccurredOn     domain.BusinessDate
	Reason         domain.Option[domain.DocumentReason]
	Notes          domain.Option[domain.NonEmptyText]
	Lines          []SaleLineInput
}

type salePostStoreInput struct {
	SalePostInput
	PostedAt domain.UTCInstant
}

type SaleDocument struct {
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

func NewSaleDocument(
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
) (SaleDocument, error) {
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
	if err := domain.NewValidationError(violations...); err != nil {
		return SaleDocument{}, err
	}
	cloned := make([]PostedSaleLine, len(lines))
	copy(cloned, lines)
	return SaleDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		counterpartyID: counterpartyID, occurredOn: occurredOn, postedAt: postedAt,
		currency: currency, reason: reason, notes: notes, lines: cloned,
	}, nil
}

func (d SaleDocument) ID() domain.StockDocumentID              { return d.id }
func (d SaleDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d SaleDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d SaleDocument) CounterpartyID() domain.Option[domain.CounterpartyID] {
	return d.counterpartyID
}
func (d SaleDocument) OccurredOn() domain.BusinessDate              { return d.occurredOn }
func (d SaleDocument) PostedAt() domain.UTCInstant                  { return d.postedAt }
func (d SaleDocument) Currency() domain.Currency                    { return d.currency }
func (d SaleDocument) Reason() domain.Option[domain.DocumentReason] { return d.reason }
func (d SaleDocument) Notes() domain.Option[domain.NonEmptyText]    { return d.notes }
func (d SaleDocument) Lines() []PostedSaleLine {
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
) (PostedSaleLine, error) {
	violations := make([]domain.Violation, 0, 6)
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
	if err := domain.NewValidationError(violations...); err != nil {
		return PostedSaleLine{}, err
	}
	cloned := make([]SaleAllocation, len(allocations))
	copy(cloned, allocations)
	return PostedSaleLine{
		id: id, lineOrder: lineOrder, itemID: itemID, quantity: quantity,
		enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue,
		commercialTotal: commercialTotal, allocations: cloned,
	}, nil
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

func NewSaleAllocation(id domain.LotAllocationID, lotID domain.InventoryLotID, quantity domain.AtomicQuantity) (SaleAllocation, error) {
	if id.IsZero() || lotID.IsZero() || quantity.Int64() <= 0 {
		return SaleAllocation{}, domain.ErrInvariant
	}
	return SaleAllocation{id: id, lotID: lotID, quantity: quantity}, nil
}

func (a SaleAllocation) ID() domain.LotAllocationID      { return a.id }
func (a SaleAllocation) LotID() domain.InventoryLotID    { return a.lotID }
func (a SaleAllocation) Quantity() domain.AtomicQuantity { return a.quantity }

type SaleService struct {
	store SaleStore
	clock Clock
}

func NewSaleService(store SaleStore, clock Clock) *SaleService {
	if store == nil {
		panic("sale service requires a store")
	}
	if clock == nil {
		panic("sale service requires a clock")
	}
	return &SaleService{store: store, clock: clock}
}

func (s *SaleService) GetSale(ctx context.Context, id domain.StockDocumentID) (SaleDocument, error) {
	document, err := s.store.GetSale(ctx, id)
	if err != nil {
		return SaleDocument{}, fmt.Errorf("get sale: %w", err)
	}
	return document, nil
}

func (s *SaleService) ListSales(ctx context.Context, input SaleListInput) (SalePage, error) {
	page, err := s.store.ListSales(ctx, input)
	if err != nil {
		return SalePage{}, fmt.Errorf("list sales: %w", err)
	}
	return page, nil
}

func (s *SaleService) PostSale(ctx context.Context, input SalePostInput) (SaleDocument, error) {
	if len(input.Lines) == 0 {
		return SaleDocument{}, domain.Invalid("lines", domain.ViolationRequired, "DOC-002")
	}
	postedAt, err := s.clock.Now()
	if err != nil {
		return SaleDocument{}, fmt.Errorf("read clock: %w", err)
	}
	document, err := s.store.PostSale(ctx, salePostStoreInput{
		SalePostInput: input,
		PostedAt:      postedAt,
	})
	if err != nil {
		return SaleDocument{}, fmt.Errorf("post sale: %w", err)
	}
	if !document.PostedAt().Equal(postedAt) {
		return SaleDocument{}, domain.ErrInvariant
	}
	return document, nil
}
