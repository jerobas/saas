package inventory

import (
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

type BalanceSnapshotParams struct {
	Balance        Balance
	ItemName       domain.UniqueName
	BaseUnit       domain.UnitCode
	ItemArchivedAt domain.Option[domain.UTCInstant]
}

// BalanceSnapshot is the common projection returned by GetInventoryBalance.
type BalanceSnapshot struct {
	balance        Balance
	itemName       domain.UniqueName
	baseUnit       domain.UnitCode
	itemArchivedAt domain.Option[domain.UTCInstant]
}

func NewBalanceSnapshot(params BalanceSnapshotParams) (BalanceSnapshot, error) {
	violations := make([]domain.Violation, 0, 4)
	if params.Balance.ItemID().IsZero() {
		violations = append(violations, required("balance"))
	}
	if params.ItemName.Display() == "" || params.ItemName.Key() == "" {
		violations = append(violations, required("item_name"))
	}
	if params.BaseUnit.String() == "" {
		violations = append(violations, required("base_unit_code"))
	}
	if archived, ok := params.ItemArchivedAt.Get(); ok && archived.IsZero() {
		violations = append(violations, required("item_archived_at"))
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return BalanceSnapshot{}, err
	}
	return BalanceSnapshot{
		balance: params.Balance, itemName: params.ItemName,
		baseUnit: params.BaseUnit, itemArchivedAt: params.ItemArchivedAt,
	}, nil
}

func (s BalanceSnapshot) Balance() Balance                                 { return s.balance }
func (s BalanceSnapshot) ItemName() domain.UniqueName                      { return s.itemName }
func (s BalanceSnapshot) BaseUnit() domain.UnitCode                        { return s.baseUnit }
func (s BalanceSnapshot) ItemArchivedAt() domain.Option[domain.UTCInstant] { return s.itemArchivedAt }
func (s BalanceSnapshot) ItemIsArchived() bool                             { return s.itemArchivedAt.IsSome() }

type BalanceListItemParams struct {
	Snapshot        BalanceSnapshot
	Capabilities    catalog.Capabilities
	ReorderQuantity domain.Option[domain.AtomicQuantity]
}

// BalanceListItem adds the catalog facts selected by ListInventoryBalances.
type BalanceListItem struct {
	snapshot        BalanceSnapshot
	capabilities    catalog.Capabilities
	reorderQuantity domain.Option[domain.AtomicQuantity]
}

func NewBalanceListItem(params BalanceListItemParams) (BalanceListItem, error) {
	if params.Snapshot.Balance().ItemID().IsZero() {
		return BalanceListItem{}, domain.Invalid("balance_snapshot", domain.ViolationRequired, "INV-002")
	}
	if !params.Snapshot.ItemIsArchived() && !params.Capabilities.Any() {
		return BalanceListItem{}, domain.Invalid("capabilities", domain.ViolationRequired, "CAT-002")
	}
	return BalanceListItem{
		snapshot: params.Snapshot, capabilities: params.Capabilities,
		reorderQuantity: params.ReorderQuantity,
	}, nil
}

func (i BalanceListItem) Snapshot() BalanceSnapshot          { return i.snapshot }
func (i BalanceListItem) Capabilities() catalog.Capabilities { return i.capabilities }
func (i BalanceListItem) ReorderQuantity() domain.Option[domain.AtomicQuantity] {
	return i.reorderQuantity
}

type LotViewParams struct {
	Lot              Lot
	SourceDocumentID domain.StockDocumentID
	SourceKind       domain.DocumentKind
	SourceOccurredOn domain.BusinessDate
}

// LotView is the query-facing lot fact including its immutable source
// document snapshot. Lot holds the allocation-derived quantity totals.
type LotView struct {
	lot              Lot
	sourceDocumentID domain.StockDocumentID
	sourceKind       domain.DocumentKind
	sourceOccurredOn domain.BusinessDate
}

func NewLotView(params LotViewParams) (LotView, error) {
	violations := make([]domain.Violation, 0, 4)
	if params.Lot.ID().IsZero() {
		violations = append(violations, required("lot"))
	}
	if params.SourceDocumentID.IsZero() {
		violations = append(violations, required("source_document_id"))
	}
	if _, err := domain.ParseDocumentKind(params.SourceKind.String()); err != nil {
		violations = append(violations, domain.Violation{Field: "source_document_kind", Code: domain.ViolationInvalidEnum, InvariantID: "DOC-001"})
	}
	if params.SourceOccurredOn.IsZero() {
		violations = append(violations, required("source_occurred_on"))
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return LotView{}, err
	}
	return LotView{
		lot: params.Lot, sourceDocumentID: params.SourceDocumentID,
		sourceKind: params.SourceKind, sourceOccurredOn: params.SourceOccurredOn,
	}, nil
}

func (v LotView) Lot() Lot                                 { return v.lot }
func (v LotView) SourceDocumentID() domain.StockDocumentID { return v.sourceDocumentID }
func (v LotView) SourceKind() domain.DocumentKind          { return v.sourceKind }
func (v LotView) SourceOccurredOn() domain.BusinessDate    { return v.sourceOccurredOn }

type AllocationViewParams struct {
	Allocation         Allocation
	SourceLineID       domain.StockDocumentLineID
	LotInitialQuantity domain.AtomicQuantity
	LotCode            domain.Option[domain.NonEmptyText]
	OriginatedOn       domain.BusinessDate
	ExpiresOn          domain.Option[domain.BusinessDate]
}

type AllocationView struct {
	allocation         Allocation
	sourceLineID       domain.StockDocumentLineID
	lotInitialQuantity domain.AtomicQuantity
	lotCode            domain.Option[domain.NonEmptyText]
	originatedOn       domain.BusinessDate
	expiresOn          domain.Option[domain.BusinessDate]
}

func NewAllocationView(params AllocationViewParams) (AllocationView, error) {
	violations := make([]domain.Violation, 0, 6)
	if params.Allocation.ID().IsZero() {
		violations = append(violations, required("allocation"))
	}
	if params.SourceLineID.IsZero() {
		violations = append(violations, required("source_line_id"))
	}
	if params.LotInitialQuantity.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "lot_initial_quantity_atomic", Code: domain.ViolationNotPositive, InvariantID: "LOT-001"})
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
	if err := domain.NewValidationError(violations...); err != nil {
		return AllocationView{}, err
	}
	return AllocationView{
		allocation: params.Allocation, sourceLineID: params.SourceLineID,
		lotInitialQuantity: params.LotInitialQuantity, lotCode: params.LotCode,
		originatedOn: params.OriginatedOn, expiresOn: params.ExpiresOn,
	}, nil
}

func (v AllocationView) Allocation() Allocation                        { return v.allocation }
func (v AllocationView) SourceLineID() domain.StockDocumentLineID      { return v.sourceLineID }
func (v AllocationView) LotInitialQuantity() domain.AtomicQuantity     { return v.lotInitialQuantity }
func (v AllocationView) LotCode() domain.Option[domain.NonEmptyText]   { return v.lotCode }
func (v AllocationView) OriginatedOn() domain.BusinessDate             { return v.originatedOn }
func (v AllocationView) ExpiresOn() domain.Option[domain.BusinessDate] { return v.expiresOn }

type LedgerEntryViewParams struct {
	Entry              LedgerEntry
	EnteredUnit        domain.UnitCode
	EnteredPackaging   domain.Option[domain.NonEmptyText]
	Conversion         domain.UnitConversion
	ReversesLineID     domain.Option[domain.StockDocumentLineID]
	IdempotencyKey     domain.IdempotencyKey
	CounterpartyID     domain.Option[domain.CounterpartyID]
	CounterpartyName   domain.Option[domain.DisplayName]
	Reason             domain.Option[domain.DocumentReason]
	Notes              domain.Option[domain.NonEmptyText]
	ReversesDocumentID domain.Option[domain.StockDocumentID]
}

// LedgerEntryView contains every historical line/document snapshot selected
// for the ledger page, without leaking sqlc nullable/generated types.
type LedgerEntryView struct {
	entry              LedgerEntry
	enteredUnit        domain.UnitCode
	enteredPackaging   domain.Option[domain.NonEmptyText]
	conversion         domain.UnitConversion
	reversesLineID     domain.Option[domain.StockDocumentLineID]
	idempotencyKey     domain.IdempotencyKey
	counterpartyID     domain.Option[domain.CounterpartyID]
	counterpartyName   domain.Option[domain.DisplayName]
	reason             domain.Option[domain.DocumentReason]
	notes              domain.Option[domain.NonEmptyText]
	reversesDocumentID domain.Option[domain.StockDocumentID]
}

func NewLedgerEntryView(params LedgerEntryViewParams) (LedgerEntryView, error) {
	violations := make([]domain.Violation, 0, 10)
	if params.Entry.LineID().IsZero() {
		violations = append(violations, required("entry"))
	}
	if params.EnteredUnit.String() == "" {
		violations = append(violations, required("entered_unit_code"))
	}
	if packaging, ok := params.EnteredPackaging.Get(); ok && packaging.String() == "" {
		violations = append(violations, required("entered_packaging_name"))
	}
	if params.Conversion.IsZero() {
		violations = append(violations, required("conversion"))
	}
	if params.IdempotencyKey.String() == "" {
		violations = append(violations, required("idempotency_key"))
	}
	counterpartyID, hasCounterparty := params.CounterpartyID.Get()
	counterpartyName, hasCounterpartyName := params.CounterpartyName.Get()
	if hasCounterparty != hasCounterpartyName || (hasCounterparty && (counterpartyID.IsZero() || counterpartyName.String() == "")) {
		violations = append(violations, domain.Violation{Field: "counterparty", Code: domain.ViolationInvariant, InvariantID: "CPY-002"})
	}
	reason := ""
	if value, ok := params.Reason.Get(); ok {
		reason = value.String()
	}
	if _, err := domain.ParseDocumentReason(params.Entry.Kind(), reason); err != nil {
		violations = append(violations, domain.Violation{Field: "reason_code", Code: domain.ViolationInvalidEnum, InvariantID: "ADJ-001"})
	}
	if notes, ok := params.Notes.Get(); ok && notes.String() == "" {
		violations = append(violations, required("notes"))
	}
	reversesLine, hasReversesLine := params.ReversesLineID.Get()
	reversesDocument, hasReversesDocument := params.ReversesDocumentID.Get()
	if params.Entry.Kind() == domain.DocumentReversal {
		if !hasReversesLine || !hasReversesDocument || reversesLine.IsZero() || reversesDocument.IsZero() {
			violations = append(violations, domain.Violation{Field: "reversal", Code: domain.ViolationInvariant, InvariantID: "REV-004"})
		}
	} else if hasReversesLine || hasReversesDocument {
		violations = append(violations, domain.Violation{Field: "reversal", Code: domain.ViolationInvariant, InvariantID: "REV-004"})
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return LedgerEntryView{}, err
	}
	return LedgerEntryView{
		entry: params.Entry, enteredUnit: params.EnteredUnit,
		enteredPackaging: params.EnteredPackaging, conversion: params.Conversion,
		reversesLineID: params.ReversesLineID, idempotencyKey: params.IdempotencyKey,
		counterpartyID: params.CounterpartyID, counterpartyName: params.CounterpartyName,
		reason: params.Reason, notes: params.Notes,
		reversesDocumentID: params.ReversesDocumentID,
	}, nil
}

func (v LedgerEntryView) Entry() LedgerEntry           { return v.entry }
func (v LedgerEntryView) EnteredUnit() domain.UnitCode { return v.enteredUnit }
func (v LedgerEntryView) EnteredPackaging() domain.Option[domain.NonEmptyText] {
	return v.enteredPackaging
}
func (v LedgerEntryView) Conversion() domain.UnitConversion { return v.conversion }
func (v LedgerEntryView) ReversesLineID() domain.Option[domain.StockDocumentLineID] {
	return v.reversesLineID
}
func (v LedgerEntryView) IdempotencyKey() domain.IdempotencyKey { return v.idempotencyKey }
func (v LedgerEntryView) CounterpartyID() domain.Option[domain.CounterpartyID] {
	return v.counterpartyID
}
func (v LedgerEntryView) CounterpartyName() domain.Option[domain.DisplayName] {
	return v.counterpartyName
}
func (v LedgerEntryView) Reason() domain.Option[domain.DocumentReason] { return v.reason }
func (v LedgerEntryView) Notes() domain.Option[domain.NonEmptyText]    { return v.notes }
func (v LedgerEntryView) ReversesDocumentID() domain.Option[domain.StockDocumentID] {
	return v.reversesDocumentID
}
