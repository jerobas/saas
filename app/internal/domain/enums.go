package domain

type Dimension string

const (
	DimensionMass   Dimension = "MASS"
	DimensionVolume Dimension = "VOLUME"
	DimensionCount  Dimension = "COUNT"
)

func ParseDimension(raw string) (Dimension, error) {
	value := Dimension(raw)
	switch value {
	case DimensionMass, DimensionVolume, DimensionCount:
		return value, nil
	default:
		return "", Invalid("dimension", ViolationInvalidEnum, "UNIT-002")
	}
}

func (d Dimension) String() string { return string(d) }

type CounterpartyRole string

const (
	RoleSupplier CounterpartyRole = "SUPPLIER"
	RoleCustomer CounterpartyRole = "CUSTOMER"
)

func ParseCounterpartyRole(raw string) (CounterpartyRole, error) {
	value := CounterpartyRole(raw)
	if value != RoleSupplier && value != RoleCustomer {
		return "", Invalid("counterparty_role", ViolationInvalidEnum, "CPY-001")
	}
	return value, nil
}

func (r CounterpartyRole) String() string { return string(r) }

type DocumentKind string

const (
	DocumentPurchase   DocumentKind = "PURCHASE"
	DocumentSale       DocumentKind = "SALE"
	DocumentProduction DocumentKind = "PRODUCTION"
	DocumentAdjustment DocumentKind = "ADJUSTMENT"
	DocumentReversal   DocumentKind = "REVERSAL"
)

func ParseDocumentKind(raw string) (DocumentKind, error) {
	value := DocumentKind(raw)
	switch value {
	case DocumentPurchase, DocumentSale, DocumentProduction, DocumentAdjustment, DocumentReversal:
		return value, nil
	default:
		return "", Invalid("document_kind", ViolationInvalidEnum, "DOC-001")
	}
}

func (k DocumentKind) String() string { return string(k) }

type Direction string

const (
	DirectionIn  Direction = "IN"
	DirectionOut Direction = "OUT"
)

func ParseDirection(raw string) (Direction, error) {
	value := Direction(raw)
	if value != DirectionIn && value != DirectionOut {
		return "", Invalid("direction", ViolationInvalidEnum, "DOC-008")
	}
	return value, nil
}

func (d Direction) String() string { return string(d) }

type DocumentReason string

const (
	ReasonFreeStock            DocumentReason = "FREE_STOCK"
	ReasonPromotion            DocumentReason = "PROMOTION"
	ReasonSample               DocumentReason = "SAMPLE"
	ReasonOpeningBalance       DocumentReason = "OPENING_BALANCE"
	ReasonPhysicalCount        DocumentReason = "PHYSICAL_COUNT"
	ReasonWaste                DocumentReason = "WASTE"
	ReasonExpiry               DocumentReason = "EXPIRY"
	ReasonDamage               DocumentReason = "DAMAGE"
	ReasonDocumentedCorrection DocumentReason = "DOCUMENTED_CORRECTION"
	ReasonExactReversal        DocumentReason = "EXACT_REVERSAL"
)

func ParseDocumentReason(kind DocumentKind, raw string) (Option[DocumentReason], error) {
	if _, err := ParseDocumentKind(kind.String()); err != nil {
		return None[DocumentReason](), err
	}
	if raw == "" {
		if kind == DocumentAdjustment || kind == DocumentReversal {
			return None[DocumentReason](), Invalid("document_reason", ViolationRequired, "ADJ-001")
		}
		return None[DocumentReason](), nil
	}
	reason := DocumentReason(raw)
	valid := false
	switch kind {
	case DocumentPurchase:
		valid = reason == ReasonFreeStock
	case DocumentSale:
		valid = reason == ReasonPromotion || reason == ReasonSample
	case DocumentProduction:
		valid = false
	case DocumentAdjustment:
		switch reason {
		case ReasonOpeningBalance, ReasonFreeStock, ReasonPhysicalCount, ReasonWaste,
			ReasonExpiry, ReasonDamage, ReasonSample, ReasonDocumentedCorrection:
			valid = true
		}
	case DocumentReversal:
		valid = reason == ReasonExactReversal
	}
	if !valid {
		return None[DocumentReason](), Invalid("document_reason", ViolationInvalidEnum, "ADJ-001")
	}
	return Some(reason), nil
}

func (r DocumentReason) String() string { return string(r) }

type AllocationEffect string

const (
	AllocationConsumption AllocationEffect = "CONSUMPTION"
	AllocationRestoration AllocationEffect = "RESTORATION"
)

func ParseAllocationEffect(raw string) (AllocationEffect, error) {
	value := AllocationEffect(raw)
	if value != AllocationConsumption && value != AllocationRestoration {
		return "", Invalid("allocation_effect", ViolationInvalidEnum, "LOT-004")
	}
	return value, nil
}

func (e AllocationEffect) String() string { return string(e) }

type ArchiveFilter string

const (
	ArchiveActive   ArchiveFilter = "ACTIVE"
	ArchiveArchived ArchiveFilter = "ARCHIVED"
	ArchiveAll      ArchiveFilter = "ALL"
)

func ParseArchiveFilter(raw string) (ArchiveFilter, error) {
	value := ArchiveFilter(raw)
	switch value {
	case ArchiveActive, ArchiveArchived, ArchiveAll:
		return value, nil
	default:
		return "", Invalid("archive_filter", ViolationInvalidEnum, "")
	}
}

func (f ArchiveFilter) String() string { return string(f) }

type LotState string

const (
	LotAvailable LotState = "AVAILABLE"
	LotDepleted  LotState = "DEPLETED"
	LotExpired   LotState = "EXPIRED"
)

func ParseLotState(raw string) (LotState, error) {
	value := LotState(raw)
	switch value {
	case LotAvailable, LotDepleted, LotExpired:
		return value, nil
	default:
		return "", Invalid("lot_state", ViolationInvalidEnum, "")
	}
}

func (s LotState) String() string { return string(s) }
