package inventory

import "github.com/jerobas/saas/internal/domain"

type BalanceParams struct {
	ItemID         domain.ItemID
	Quantity       domain.AtomicQuantity
	Value          domain.InventoryValue
	LastDocumentID domain.Option[domain.StockDocumentID]
	UpdatedAt      domain.UTCInstant
}

// Balance is the rebuildable, read-only inventory projection for one item.
type Balance struct {
	itemID         domain.ItemID
	quantity       domain.AtomicQuantity
	value          domain.InventoryValue
	lastDocumentID domain.Option[domain.StockDocumentID]
	updatedAt      domain.UTCInstant
}

func NewBalance(params BalanceParams) (Balance, error) {
	violations := make([]domain.Violation, 0, 4)
	if params.ItemID.IsZero() {
		violations = append(violations, required("item_id"))
	}
	if params.Quantity.IsZero() && !params.Value.IsZero() {
		violations = append(violations, domain.Violation{Field: "inventory_value_micro", Code: domain.ViolationInvariant, InvariantID: "INV-003"})
	}
	if documentID, ok := params.LastDocumentID.Get(); ok && documentID.IsZero() {
		violations = append(violations, required("last_document_id"))
	}
	if params.UpdatedAt.IsZero() {
		violations = append(violations, required("updated_at"))
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Balance{}, err
	}
	return Balance{
		itemID: params.ItemID, quantity: params.Quantity, value: params.Value,
		lastDocumentID: params.LastDocumentID, updatedAt: params.UpdatedAt,
	}, nil
}

func (b Balance) ItemID() domain.ItemID                                 { return b.itemID }
func (b Balance) Quantity() domain.AtomicQuantity                       { return b.quantity }
func (b Balance) Value() domain.InventoryValue                          { return b.value }
func (b Balance) LastDocumentID() domain.Option[domain.StockDocumentID] { return b.lastDocumentID }
func (b Balance) UpdatedAt() domain.UTCInstant                          { return b.updatedAt }
func (b Balance) AverageValuePerAtomicUnit() (domain.Fraction, bool)    { return b.value.Per(b.quantity) }

type LotParams struct {
	ID                    domain.InventoryLotID
	ItemID                domain.ItemID
	SourceLineID          domain.StockDocumentLineID
	SourcePostingSequence domain.PostingSequence
	InitialQuantity       domain.AtomicQuantity
	ConsumedQuantity      domain.AtomicQuantity
	RestoredQuantity      domain.AtomicQuantity
	AvailableQuantity     domain.AtomicQuantity
	LotCode               domain.Option[domain.NonEmptyText]
	OriginatedOn          domain.BusinessDate
	ExpiresOn             domain.Option[domain.BusinessDate]
	CreatedAt             domain.UTCInstant
}

// Lot combines its immutable source with allocation totals calculated by the
// query. The constructor verifies those derived totals before exposing them.
type Lot struct {
	id                    domain.InventoryLotID
	itemID                domain.ItemID
	sourceLineID          domain.StockDocumentLineID
	sourcePostingSequence domain.PostingSequence
	initialQuantity       domain.AtomicQuantity
	consumedQuantity      domain.AtomicQuantity
	restoredQuantity      domain.AtomicQuantity
	availableQuantity     domain.AtomicQuantity
	lotCode               domain.Option[domain.NonEmptyText]
	originatedOn          domain.BusinessDate
	expiresOn             domain.Option[domain.BusinessDate]
	createdAt             domain.UTCInstant
}

func NewLot(params LotParams) (Lot, error) {
	violations := make([]domain.Violation, 0, 10)
	if params.ID.IsZero() {
		violations = append(violations, required("lot_id"))
	}
	if params.ItemID.IsZero() {
		violations = append(violations, required("item_id"))
	}
	if params.SourceLineID.IsZero() {
		violations = append(violations, required("source_line_id"))
	}
	if params.SourcePostingSequence.IsZero() {
		violations = append(violations, required("source_posting_sequence"))
	}
	if params.InitialQuantity.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "initial_quantity_atomic", Code: domain.ViolationNotPositive, InvariantID: "LOT-001"})
	}
	if code, ok := params.LotCode.Get(); ok && code.String() == "" {
		violations = append(violations, required("lot_code"))
	}
	if params.OriginatedOn.IsZero() {
		violations = append(violations, required("originated_on"))
	}
	if expiry, ok := params.ExpiresOn.Get(); ok && expiry.IsZero() {
		violations = append(violations, required("expires_on"))
	}
	if params.CreatedAt.IsZero() {
		violations = append(violations, required("created_at"))
	}
	consumed := params.ConsumedQuantity.Int64()
	restored := params.RestoredQuantity.Int64()
	initial := params.InitialQuantity.Int64()
	available := params.AvailableQuantity.Int64()
	if restored > consumed || consumed-restored > initial || initial-(consumed-restored) != available {
		violations = append(violations, domain.Violation{Field: "available_quantity_atomic", Code: domain.ViolationInvariant, InvariantID: "LOT-004"})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Lot{}, err
	}
	return Lot{
		id: params.ID, itemID: params.ItemID, sourceLineID: params.SourceLineID,
		sourcePostingSequence: params.SourcePostingSequence,
		initialQuantity:       params.InitialQuantity, consumedQuantity: params.ConsumedQuantity,
		restoredQuantity: params.RestoredQuantity, availableQuantity: params.AvailableQuantity,
		lotCode: params.LotCode, originatedOn: params.OriginatedOn,
		expiresOn: params.ExpiresOn, createdAt: params.CreatedAt,
	}, nil
}

func (l Lot) ID() domain.InventoryLotID                     { return l.id }
func (l Lot) ItemID() domain.ItemID                         { return l.itemID }
func (l Lot) SourceLineID() domain.StockDocumentLineID      { return l.sourceLineID }
func (l Lot) SourcePostingSequence() domain.PostingSequence { return l.sourcePostingSequence }
func (l Lot) InitialQuantity() domain.AtomicQuantity        { return l.initialQuantity }
func (l Lot) ConsumedQuantity() domain.AtomicQuantity       { return l.consumedQuantity }
func (l Lot) RestoredQuantity() domain.AtomicQuantity       { return l.restoredQuantity }
func (l Lot) AvailableQuantity() domain.AtomicQuantity      { return l.availableQuantity }
func (l Lot) LotCode() domain.Option[domain.NonEmptyText]   { return l.lotCode }
func (l Lot) OriginatedOn() domain.BusinessDate             { return l.originatedOn }
func (l Lot) ExpiresOn() domain.Option[domain.BusinessDate] { return l.expiresOn }
func (l Lot) CreatedAt() domain.UTCInstant                  { return l.createdAt }

// IsExpired treats expires_on as usable through that date: expiry begins on
// the following business date.
func (l Lot) IsExpired(on domain.BusinessDate) bool {
	expires, ok := l.expiresOn.Get()
	return ok && expires.Before(on)
}

func (l Lot) State(on domain.BusinessDate) domain.LotState {
	if l.availableQuantity.IsZero() {
		return domain.LotDepleted
	}
	if l.IsExpired(on) {
		return domain.LotExpired
	}
	return domain.LotAvailable
}

type LotCursor struct {
	postingSequence domain.PostingSequence
	lotID           domain.InventoryLotID
}

func NewLotCursor(postingSequence domain.PostingSequence, lotID domain.InventoryLotID) (LotCursor, error) {
	if postingSequence.IsZero() || lotID.IsZero() {
		return LotCursor{}, domain.Invalid("lot_cursor", domain.ViolationInvalidFormat, "LOT-005")
	}
	return LotCursor{postingSequence: postingSequence, lotID: lotID}, nil
}

func (c LotCursor) PostingSequence() domain.PostingSequence { return c.postingSequence }
func (c LotCursor) LotID() domain.InventoryLotID            { return c.lotID }
func (l Lot) Cursor() LotCursor {
	return LotCursor{postingSequence: l.sourcePostingSequence, lotID: l.id}
}

type AllocationParams struct {
	ID                   domain.LotAllocationID
	LineID               domain.StockDocumentLineID
	LotID                domain.InventoryLotID
	Quantity             domain.AtomicQuantity
	Effect               domain.AllocationEffect
	RestoresAllocationID domain.Option[domain.LotAllocationID]
	CreatedAt            domain.UTCInstant
}

type Allocation struct {
	id                   domain.LotAllocationID
	lineID               domain.StockDocumentLineID
	lotID                domain.InventoryLotID
	quantity             domain.AtomicQuantity
	effect               domain.AllocationEffect
	restoresAllocationID domain.Option[domain.LotAllocationID]
	createdAt            domain.UTCInstant
}

func NewAllocation(params AllocationParams) (Allocation, error) {
	violations := make([]domain.Violation, 0, 7)
	if params.ID.IsZero() {
		violations = append(violations, required("allocation_id"))
	}
	if params.LineID.IsZero() {
		violations = append(violations, required("line_id"))
	}
	if params.LotID.IsZero() {
		violations = append(violations, required("lot_id"))
	}
	if params.Quantity.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "quantity_atomic", Code: domain.ViolationNotPositive, InvariantID: "LOT-003"})
	}
	if _, err := domain.ParseAllocationEffect(params.Effect.String()); err != nil {
		violations = append(violations, domain.Violation{Field: "effect", Code: domain.ViolationInvalidEnum, InvariantID: "LOT-004"})
	}
	restoredID, restores := params.RestoresAllocationID.Get()
	if params.Effect == domain.AllocationRestoration {
		if !restores || restoredID.IsZero() || restoredID == params.ID {
			violations = append(violations, domain.Violation{Field: "restores_allocation_id", Code: domain.ViolationInvariant, InvariantID: "LOT-004"})
		}
	} else if restores {
		violations = append(violations, domain.Violation{Field: "restores_allocation_id", Code: domain.ViolationInvariant, InvariantID: "LOT-004"})
	}
	if params.CreatedAt.IsZero() {
		violations = append(violations, required("created_at"))
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Allocation{}, err
	}
	return Allocation{
		id: params.ID, lineID: params.LineID, lotID: params.LotID,
		quantity: params.Quantity, effect: params.Effect,
		restoresAllocationID: params.RestoresAllocationID, createdAt: params.CreatedAt,
	}, nil
}

func (a Allocation) ID() domain.LotAllocationID         { return a.id }
func (a Allocation) LineID() domain.StockDocumentLineID { return a.lineID }
func (a Allocation) LotID() domain.InventoryLotID       { return a.lotID }
func (a Allocation) Quantity() domain.AtomicQuantity    { return a.quantity }
func (a Allocation) Effect() domain.AllocationEffect    { return a.effect }
func (a Allocation) RestoresAllocationID() domain.Option[domain.LotAllocationID] {
	return a.restoresAllocationID
}
func (a Allocation) CreatedAt() domain.UTCInstant { return a.createdAt }

type LedgerEntryParams struct {
	LineID          domain.StockDocumentLineID
	DocumentID      domain.StockDocumentID
	PostingSequence domain.PostingSequence
	LineOrder       domain.LineOrder
	Kind            domain.DocumentKind
	OccurredOn      domain.BusinessDate
	PostedAt        domain.UTCInstant
	ItemID          domain.ItemID
	Direction       domain.Direction
	Quantity        domain.AtomicQuantity
	Value           domain.InventoryValue
	CommercialTotal domain.Option[domain.MinorAmount]
	Currency        domain.Currency
}

type LedgerEntry struct {
	lineID          domain.StockDocumentLineID
	documentID      domain.StockDocumentID
	postingSequence domain.PostingSequence
	lineOrder       domain.LineOrder
	kind            domain.DocumentKind
	occurredOn      domain.BusinessDate
	postedAt        domain.UTCInstant
	itemID          domain.ItemID
	direction       domain.Direction
	quantity        domain.AtomicQuantity
	value           domain.InventoryValue
	commercialTotal domain.Option[domain.MinorAmount]
	currency        domain.Currency
}

func NewLedgerEntry(params LedgerEntryParams) (LedgerEntry, error) {
	violations := make([]domain.Violation, 0, 10)
	if params.LineID.IsZero() {
		violations = append(violations, required("line_id"))
	}
	if params.DocumentID.IsZero() {
		violations = append(violations, required("document_id"))
	}
	if params.PostingSequence.IsZero() {
		violations = append(violations, required("posting_sequence"))
	}
	if params.LineOrder.IsZero() {
		violations = append(violations, required("line_order"))
	}
	if _, err := domain.ParseDocumentKind(params.Kind.String()); err != nil {
		violations = append(violations, domain.Violation{Field: "kind", Code: domain.ViolationInvalidEnum, InvariantID: "DOC-001"})
	}
	if params.OccurredOn.IsZero() {
		violations = append(violations, required("occurred_on"))
	}
	if params.PostedAt.IsZero() {
		violations = append(violations, required("posted_at"))
	}
	if params.ItemID.IsZero() {
		violations = append(violations, required("item_id"))
	}
	if _, err := domain.ParseDirection(params.Direction.String()); err != nil {
		violations = append(violations, domain.Violation{Field: "direction", Code: domain.ViolationInvalidEnum, InvariantID: "DOC-008"})
	}
	if params.Quantity.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "quantity_atomic", Code: domain.ViolationNotPositive})
	}
	if params.Currency.IsZero() {
		violations = append(violations, required("currency"))
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return LedgerEntry{}, err
	}
	return LedgerEntry{
		lineID: params.LineID, documentID: params.DocumentID,
		postingSequence: params.PostingSequence, lineOrder: params.LineOrder,
		kind: params.Kind, occurredOn: params.OccurredOn, postedAt: params.PostedAt,
		itemID: params.ItemID, direction: params.Direction, quantity: params.Quantity,
		value: params.Value, commercialTotal: params.CommercialTotal, currency: params.Currency,
	}, nil
}

func (e LedgerEntry) LineID() domain.StockDocumentLineID                 { return e.lineID }
func (e LedgerEntry) DocumentID() domain.StockDocumentID                 { return e.documentID }
func (e LedgerEntry) PostingSequence() domain.PostingSequence            { return e.postingSequence }
func (e LedgerEntry) LineOrder() domain.LineOrder                        { return e.lineOrder }
func (e LedgerEntry) Kind() domain.DocumentKind                          { return e.kind }
func (e LedgerEntry) OccurredOn() domain.BusinessDate                    { return e.occurredOn }
func (e LedgerEntry) PostedAt() domain.UTCInstant                        { return e.postedAt }
func (e LedgerEntry) ItemID() domain.ItemID                              { return e.itemID }
func (e LedgerEntry) Direction() domain.Direction                        { return e.direction }
func (e LedgerEntry) Quantity() domain.AtomicQuantity                    { return e.quantity }
func (e LedgerEntry) Value() domain.InventoryValue                       { return e.value }
func (e LedgerEntry) CommercialTotal() domain.Option[domain.MinorAmount] { return e.commercialTotal }
func (e LedgerEntry) Currency() domain.Currency                          { return e.currency }

type LedgerCursor struct {
	postingSequence domain.PostingSequence
	lineOrder       domain.LineOrder
	lineID          domain.StockDocumentLineID
}

func NewLedgerCursor(postingSequence domain.PostingSequence, lineOrder domain.LineOrder, lineID domain.StockDocumentLineID) (LedgerCursor, error) {
	if postingSequence.IsZero() || lineOrder.IsZero() || lineID.IsZero() {
		return LedgerCursor{}, domain.Invalid("ledger_cursor", domain.ViolationInvalidFormat, "INV-008")
	}
	return LedgerCursor{postingSequence: postingSequence, lineOrder: lineOrder, lineID: lineID}, nil
}

func (c LedgerCursor) PostingSequence() domain.PostingSequence { return c.postingSequence }
func (c LedgerCursor) LineOrder() domain.LineOrder             { return c.lineOrder }
func (c LedgerCursor) LineID() domain.StockDocumentLineID      { return c.lineID }
func (e LedgerEntry) Cursor() LedgerCursor {
	return LedgerCursor{postingSequence: e.postingSequence, lineOrder: e.lineOrder, lineID: e.lineID}
}

func required(field string) domain.Violation {
	return domain.Violation{Field: field, Code: domain.ViolationRequired}
}
