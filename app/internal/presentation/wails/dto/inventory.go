package dto

type InventoryBalanceCursorRequest struct {
	ItemName string `json:"itemName"`
	ItemID   int64  `json:"itemId"`
}

type InventoryBalanceCursorResponse struct {
	ItemName string `json:"itemName"`
	ItemID   int64  `json:"itemId"`
}

type InventoryBalanceListRequest struct {
	IncludeArchived bool                           `json:"includeArchived,omitempty"`
	Search          *string                        `json:"search,omitempty"`
	After           *InventoryBalanceCursorRequest `json:"after,omitempty"`
	PageSize        int                            `json:"pageSize,omitempty"`
}

type InventoryBalancePageResponse struct {
	Items []InventoryBalanceResponse      `json:"items"`
	Next  *InventoryBalanceCursorResponse `json:"next,omitempty"`
}

type InventoryBalanceResponse struct {
	ItemID              int64                `json:"itemId"`
	ItemName            string               `json:"itemName"`
	BaseUnitCode        string               `json:"baseUnitCode"`
	ItemArchivedAtMs    *int64               `json:"itemArchivedAtMs,omitempty"`
	QuantityAtomic      int64                `json:"quantityAtomic"`
	InventoryValueMicro int64                `json:"inventoryValueMicro"`
	LastDocumentID      *int64               `json:"lastDocumentId,omitempty"`
	UpdatedAtMs         int64                `json:"updatedAtMs"`
	Capabilities        CapabilitiesResponse `json:"capabilities"`
	ReorderQuantity     *int64               `json:"reorderQuantityAtomic,omitempty"`
}

type LotResponse struct {
	ID                    int64   `json:"id"`
	ItemID                int64   `json:"itemId"`
	SourceLineID          int64   `json:"sourceLineId"`
	SourcePostingSequence int64   `json:"sourcePostingSequence"`
	InitialQuantity       int64   `json:"initialQuantityAtomic"`
	ConsumedQuantity      int64   `json:"consumedQuantityAtomic"`
	RestoredQuantity      int64   `json:"restoredQuantityAtomic"`
	AvailableQuantity     int64   `json:"availableQuantityAtomic"`
	LotCode               *string `json:"lotCode,omitempty"`
	OriginatedOn          string  `json:"originatedOn"`
	ExpiresOn             *string `json:"expiresOn,omitempty"`
	CreatedAtMs           int64   `json:"createdAtMs"`
	SourceDocumentID      int64   `json:"sourceDocumentId"`
	SourceKind            string  `json:"sourceKind"`
	SourceOccurredOn      string  `json:"sourceOccurredOn"`
}

type LedgerCursorRequest struct {
	PostingSequence int64 `json:"postingSequence"`
	LineOrder       int64 `json:"lineOrder"`
	LineID          int64 `json:"lineId"`
}

type LedgerCursorResponse struct {
	PostingSequence int64 `json:"postingSequence"`
	LineOrder       int64 `json:"lineOrder"`
	LineID          int64 `json:"lineId"`
}

type ItemLedgerPageRequest struct {
	ItemID   int64                `json:"itemId"`
	After    *LedgerCursorRequest `json:"after,omitempty"`
	PageSize int                  `json:"pageSize,omitempty"`
}

type LedgerEntryPageResponse struct {
	Items []LedgerEntryResponse `json:"items"`
	Next  *LedgerCursorResponse `json:"next,omitempty"`
}

type LedgerEntryResponse struct {
	LineID                    int64   `json:"lineId"`
	DocumentID                int64   `json:"documentId"`
	PostingSequence           int64   `json:"postingSequence"`
	LineOrder                 int64   `json:"lineOrder"`
	DocumentKind              string  `json:"documentKind"`
	OccurredOn                string  `json:"occurredOn"`
	PostedAtMs                int64   `json:"postedAtMs"`
	ItemID                    int64   `json:"itemId"`
	Direction                 string  `json:"direction"`
	QuantityAtomic            int64   `json:"quantityAtomic"`
	InventoryValueMicro       int64   `json:"inventoryValueMicro"`
	CommercialTotalMinor      *int64  `json:"commercialTotalMinor,omitempty"`
	CurrencyCode              string  `json:"currencyCode"`
	CurrencyMinorDigits       int64   `json:"currencyMinorDigits"`
	EnteredUnitCode           string  `json:"enteredUnitCode"`
	EnteredPackagingName      *string `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64   `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64   `json:"conversionDenominator"`
	ReversesLineID            *int64  `json:"reversesLineId,omitempty"`
	IdempotencyKey            string  `json:"idempotencyKey"`
	CounterpartyID            *int64  `json:"counterpartyId,omitempty"`
	CounterpartyName          *string `json:"counterpartyName,omitempty"`
	ReasonCode                *string `json:"reasonCode,omitempty"`
	Notes                     *string `json:"notes,omitempty"`
	ReversesDocumentID        *int64  `json:"reversesDocumentId,omitempty"`
}

type AllocationResponse struct {
	ID                   int64   `json:"id"`
	LineID               int64   `json:"lineId"`
	LotID                int64   `json:"lotId"`
	QuantityAtomic       int64   `json:"quantityAtomic"`
	Effect               string  `json:"effect"`
	RestoresAllocationID *int64  `json:"restoresAllocationId,omitempty"`
	CreatedAtMs          int64   `json:"createdAtMs"`
	SourceLineID         int64   `json:"sourceLineId"`
	LotInitialQuantity   int64   `json:"lotInitialQuantityAtomic"`
	LotCode              *string `json:"lotCode,omitempty"`
	OriginatedOn         string  `json:"originatedOn"`
	ExpiresOn            *string `json:"expiresOn,omitempty"`
}
