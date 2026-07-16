package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
)

type PurchaseStore interface {
	GetPurchase(ctx context.Context, id domain.StockDocumentID) (PurchaseDocument, error)
	ListPurchases(ctx context.Context, input PurchaseListInput) (PurchasePage, error)
	PostPurchase(ctx context.Context, input purchasePostStoreInput) (PurchaseDocument, error)
}

type PurchaseCursor struct {
	PostingSequence domain.PostingSequence
	ID              domain.StockDocumentID
}

type PurchaseListInput struct {
	After    domain.Option[PurchaseCursor]
	PageSize int
}

type PurchasePage struct {
	items []PurchaseDocument
	next  domain.Option[PurchaseCursor]
}

func NewPurchasePage(items []PurchaseDocument, next domain.Option[PurchaseCursor]) PurchasePage {
	cloned := make([]PurchaseDocument, len(items))
	copy(cloned, items)
	return PurchasePage{items: cloned, next: next}
}

func (p PurchasePage) Items() []PurchaseDocument {
	items := make([]PurchaseDocument, len(p.items))
	copy(items, p.items)
	return items
}

func (p PurchasePage) Next() domain.Option[PurchaseCursor] { return p.next }

type PurchaseLineInput struct {
	ItemID               domain.ItemID
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	CommercialTotal      domain.MinorAmount
	LotCode              domain.Option[domain.NonEmptyText]
	ExpiresOn            domain.Option[domain.BusinessDate]
}

type PurchasePostInput struct {
	IdempotencyKey domain.IdempotencyKey
	CounterpartyID domain.Option[domain.CounterpartyID]
	OccurredOn     domain.BusinessDate
	Reason         domain.Option[domain.DocumentReason]
	Notes          domain.Option[domain.NonEmptyText]
	Lines          []PurchaseLineInput
}

type purchasePostStoreInput struct {
	PurchasePostInput
	PostedAt domain.UTCInstant
}

type PurchaseDocument struct {
	id              domain.StockDocumentID
	idempotencyKey  domain.IdempotencyKey
	postingSequence domain.PostingSequence
	counterpartyID  domain.Option[domain.CounterpartyID]
	occurredOn      domain.BusinessDate
	postedAt        domain.UTCInstant
	currency        domain.Currency
	reason          domain.Option[domain.DocumentReason]
	notes           domain.Option[domain.NonEmptyText]
	lines           []PostedPurchaseLine
}

func NewPurchaseDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	counterpartyID domain.Option[domain.CounterpartyID],
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	reason domain.Option[domain.DocumentReason],
	notes domain.Option[domain.NonEmptyText],
	lines []PostedPurchaseLine,
) (PurchaseDocument, error) {
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
		return PurchaseDocument{}, err
	}
	cloned := make([]PostedPurchaseLine, len(lines))
	copy(cloned, lines)
	return PurchaseDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		counterpartyID: counterpartyID, occurredOn: occurredOn, postedAt: postedAt,
		currency: currency, reason: reason, notes: notes, lines: cloned,
	}, nil
}

func (d PurchaseDocument) ID() domain.StockDocumentID              { return d.id }
func (d PurchaseDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d PurchaseDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d PurchaseDocument) CounterpartyID() domain.Option[domain.CounterpartyID] {
	return d.counterpartyID
}
func (d PurchaseDocument) OccurredOn() domain.BusinessDate              { return d.occurredOn }
func (d PurchaseDocument) PostedAt() domain.UTCInstant                  { return d.postedAt }
func (d PurchaseDocument) Currency() domain.Currency                    { return d.currency }
func (d PurchaseDocument) Reason() domain.Option[domain.DocumentReason] { return d.reason }
func (d PurchaseDocument) Notes() domain.Option[domain.NonEmptyText]    { return d.notes }
func (d PurchaseDocument) Lines() []PostedPurchaseLine {
	lines := make([]PostedPurchaseLine, len(d.lines))
	copy(lines, d.lines)
	return lines
}

type PostedPurchaseLine struct {
	id                   domain.StockDocumentLineID
	lineOrder            domain.LineOrder
	itemID               domain.ItemID
	quantity             domain.AtomicQuantity
	enteredUnit          domain.UnitCode
	enteredPackagingName domain.Option[domain.NonEmptyText]
	conversion           domain.UnitConversion
	inventoryValue       domain.InventoryValue
	commercialTotal      domain.MinorAmount
	lotID                domain.InventoryLotID
	lotCode              domain.Option[domain.NonEmptyText]
	originatedOn         domain.BusinessDate
	expiresOn            domain.Option[domain.BusinessDate]
}

func NewPostedPurchaseLine(
	id domain.StockDocumentLineID,
	lineOrder domain.LineOrder,
	itemID domain.ItemID,
	quantity domain.AtomicQuantity,
	enteredUnit domain.UnitCode,
	enteredPackagingName domain.Option[domain.NonEmptyText],
	conversion domain.UnitConversion,
	inventoryValue domain.InventoryValue,
	commercialTotal domain.MinorAmount,
	lotID domain.InventoryLotID,
	lotCode domain.Option[domain.NonEmptyText],
	originatedOn domain.BusinessDate,
	expiresOn domain.Option[domain.BusinessDate],
) (PostedPurchaseLine, error) {
	violations := make([]domain.Violation, 0, 9)
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
	if lotID.IsZero() {
		violations = append(violations, domain.Violation{Field: "lot_id", Code: domain.ViolationRequired})
	}
	if originatedOn.IsZero() {
		violations = append(violations, domain.Violation{Field: "originated_on", Code: domain.ViolationRequired})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return PostedPurchaseLine{}, err
	}
	return PostedPurchaseLine{
		id: id, lineOrder: lineOrder, itemID: itemID, quantity: quantity,
		enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, commercialTotal: commercialTotal,
		lotID: lotID, lotCode: lotCode, originatedOn: originatedOn, expiresOn: expiresOn,
	}, nil
}

func (l PostedPurchaseLine) ID() domain.StockDocumentLineID  { return l.id }
func (l PostedPurchaseLine) LineOrder() domain.LineOrder     { return l.lineOrder }
func (l PostedPurchaseLine) ItemID() domain.ItemID           { return l.itemID }
func (l PostedPurchaseLine) Quantity() domain.AtomicQuantity { return l.quantity }
func (l PostedPurchaseLine) EnteredUnit() domain.UnitCode    { return l.enteredUnit }
func (l PostedPurchaseLine) EnteredPackagingName() domain.Option[domain.NonEmptyText] {
	return l.enteredPackagingName
}
func (l PostedPurchaseLine) Conversion() domain.UnitConversion           { return l.conversion }
func (l PostedPurchaseLine) InventoryValue() domain.InventoryValue       { return l.inventoryValue }
func (l PostedPurchaseLine) CommercialTotal() domain.MinorAmount         { return l.commercialTotal }
func (l PostedPurchaseLine) LotID() domain.InventoryLotID                { return l.lotID }
func (l PostedPurchaseLine) LotCode() domain.Option[domain.NonEmptyText] { return l.lotCode }
func (l PostedPurchaseLine) OriginatedOn() domain.BusinessDate           { return l.originatedOn }
func (l PostedPurchaseLine) ExpiresOn() domain.Option[domain.BusinessDate] {
	return l.expiresOn
}

type PurchaseService struct {
	store PurchaseStore
	clock Clock
}

func NewPurchaseService(store PurchaseStore, clock Clock) *PurchaseService {
	if store == nil {
		panic("purchase service requires a store")
	}
	if clock == nil {
		panic("purchase service requires a clock")
	}
	return &PurchaseService{store: store, clock: clock}
}

func (s *PurchaseService) GetPurchase(ctx context.Context, id domain.StockDocumentID) (PurchaseDocument, error) {
	document, err := s.store.GetPurchase(ctx, id)
	if err != nil {
		return PurchaseDocument{}, fmt.Errorf("get purchase: %w", err)
	}
	return document, nil
}

func (s *PurchaseService) ListPurchases(ctx context.Context, input PurchaseListInput) (PurchasePage, error) {
	page, err := s.store.ListPurchases(ctx, input)
	if err != nil {
		return PurchasePage{}, fmt.Errorf("list purchases: %w", err)
	}
	return page, nil
}

func (s *PurchaseService) PostPurchase(ctx context.Context, input PurchasePostInput) (PurchaseDocument, error) {
	if len(input.Lines) == 0 {
		return PurchaseDocument{}, domain.Invalid("lines", domain.ViolationRequired, "DOC-002")
	}
	postedAt, err := s.clock.Now()
	if err != nil {
		return PurchaseDocument{}, fmt.Errorf("read clock: %w", err)
	}
	document, err := s.store.PostPurchase(ctx, purchasePostStoreInput{
		PurchasePostInput: input,
		PostedAt:          postedAt,
	})
	if err != nil {
		return PurchaseDocument{}, fmt.Errorf("post purchase: %w", err)
	}
	if !document.PostedAt().Equal(postedAt) {
		return PurchaseDocument{}, domain.ErrInvariant
	}
	return document, nil
}
