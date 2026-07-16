package dto

type SalePostRequest struct {
	IdempotencyKey string            `json:"idempotencyKey"`
	CounterpartyID *int64            `json:"counterpartyId,omitempty"`
	OccurredOn     string            `json:"occurredOn"`
	ReasonCode     *string           `json:"reasonCode,omitempty"`
	Notes          *string           `json:"notes,omitempty"`
	Lines          []SaleLineRequest `json:"lines"`
}

type SaleLineRequest struct {
	ItemID                    int64   `json:"itemId"`
	QuantityAtomic            int64   `json:"quantityAtomic"`
	EnteredUnitCode           string  `json:"enteredUnitCode"`
	EnteredPackagingName      *string `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64   `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64   `json:"conversionDenominator"`
	CommercialTotalMinor      int64   `json:"commercialTotalMinor"`
	LotID                     *int64  `json:"lotId,omitempty"`
}

type SaleDocumentResponse struct {
	ID                  int64              `json:"id"`
	IdempotencyKey      string             `json:"idempotencyKey"`
	PostingSequence     int64              `json:"postingSequence"`
	CounterpartyID      *int64             `json:"counterpartyId,omitempty"`
	OccurredOn          string             `json:"occurredOn"`
	PostedAtMs          int64              `json:"postedAtMs"`
	CurrencyCode        string             `json:"currencyCode"`
	CurrencyMinorDigits int64              `json:"currencyMinorDigits"`
	ReasonCode          *string            `json:"reasonCode,omitempty"`
	Notes               *string            `json:"notes,omitempty"`
	Lines               []SaleLineResponse `json:"lines"`
}

type SaleLineResponse struct {
	ID                        int64                    `json:"id"`
	LineOrder                 int64                    `json:"lineOrder"`
	ItemID                    int64                    `json:"itemId"`
	Direction                 string                   `json:"direction"`
	QuantityAtomic            int64                    `json:"quantityAtomic"`
	EnteredUnitCode           string                   `json:"enteredUnitCode"`
	EnteredPackagingName      *string                  `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64                    `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64                    `json:"conversionDenominator"`
	InventoryValueMicro       int64                    `json:"inventoryValueMicro"`
	CommercialTotalMinor      int64                    `json:"commercialTotalMinor"`
	Allocations               []SaleAllocationResponse `json:"allocations"`
}

type SaleAllocationResponse struct {
	ID             int64 `json:"id"`
	LotID          int64 `json:"lotId"`
	QuantityAtomic int64 `json:"quantityAtomic"`
}
