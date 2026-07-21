package dto

type ReversalPostRequest struct {
	IdempotencyKey   string  `json:"idempotencyKey"`
	TargetDocumentID int64   `json:"targetDocumentId"`
	OccurredOn       string  `json:"occurredOn"`
	Notes            *string `json:"notes,omitempty"`
}

type ReversalDocumentResponse struct {
	ID                  int64                  `json:"id"`
	IdempotencyKey      string                 `json:"idempotencyKey"`
	PostingSequence     int64                  `json:"postingSequence"`
	TargetDocumentID    int64                  `json:"targetDocumentId"`
	OccurredOn          string                 `json:"occurredOn"`
	PostedAtMs          int64                  `json:"postedAtMs"`
	CurrencyCode        string                 `json:"currencyCode"`
	CurrencyMinorDigits int64                  `json:"currencyMinorDigits"`
	ReasonCode          string                 `json:"reasonCode"`
	Notes               *string                `json:"notes,omitempty"`
	Lines               []ReversalLineResponse `json:"lines"`
}

type ReversalLineResponse struct {
	ID                        int64                        `json:"id"`
	LineOrder                 int64                        `json:"lineOrder"`
	ItemID                    int64                        `json:"itemId"`
	Direction                 string                       `json:"direction"`
	QuantityAtomic            int64                        `json:"quantityAtomic"`
	EnteredUnitCode           string                       `json:"enteredUnitCode"`
	EnteredPackagingName      *string                      `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64                        `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64                        `json:"conversionDenominator"`
	InventoryValueMicro       int64                        `json:"inventoryValueMicro"`
	CommercialTotalMinor      *int64                       `json:"commercialTotalMinor,omitempty"`
	ReversesLineID            int64                        `json:"reversesLineId"`
	Allocations               []ReversalAllocationResponse `json:"allocations"`
}

type ReversalAllocationResponse struct {
	ID                   int64  `json:"id"`
	LotID                int64  `json:"lotId"`
	QuantityAtomic       int64  `json:"quantityAtomic"`
	RestoresAllocationID *int64 `json:"restoresAllocationId,omitempty"`
}
