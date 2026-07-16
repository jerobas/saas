package dto

type AdjustmentPostRequest struct {
	IdempotencyKey string                  `json:"idempotencyKey"`
	OccurredOn     string                  `json:"occurredOn"`
	ReasonCode     string                  `json:"reasonCode"`
	Notes          *string                 `json:"notes,omitempty"`
	Lines          []AdjustmentLineRequest `json:"lines"`
}

type AdjustmentLineRequest struct {
	ItemID                    int64   `json:"itemId"`
	Direction                 string  `json:"direction"`
	QuantityAtomic            int64   `json:"quantityAtomic"`
	EnteredUnitCode           string  `json:"enteredUnitCode"`
	EnteredPackagingName      *string `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64   `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64   `json:"conversionDenominator"`
	InventoryValueMicro       *int64  `json:"inventoryValueMicro,omitempty"`
	LotCode                   *string `json:"lotCode,omitempty"`
	ExpiresOn                 *string `json:"expiresOn,omitempty"`
}

type AdjustmentDocumentResponse struct {
	ID                  int64                    `json:"id"`
	IdempotencyKey      string                   `json:"idempotencyKey"`
	PostingSequence     int64                    `json:"postingSequence"`
	OccurredOn          string                   `json:"occurredOn"`
	PostedAtMs          int64                    `json:"postedAtMs"`
	CurrencyCode        string                   `json:"currencyCode"`
	CurrencyMinorDigits int64                    `json:"currencyMinorDigits"`
	ReasonCode          string                   `json:"reasonCode"`
	Notes               *string                  `json:"notes,omitempty"`
	Lines               []AdjustmentLineResponse `json:"lines"`
}

type AdjustmentLineResponse struct {
	ID                        int64                          `json:"id"`
	LineOrder                 int64                          `json:"lineOrder"`
	ItemID                    int64                          `json:"itemId"`
	Direction                 string                         `json:"direction"`
	QuantityAtomic            int64                          `json:"quantityAtomic"`
	EnteredUnitCode           string                         `json:"enteredUnitCode"`
	EnteredPackagingName      *string                        `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64                          `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64                          `json:"conversionDenominator"`
	InventoryValueMicro       int64                          `json:"inventoryValueMicro"`
	LotID                     *int64                         `json:"lotId,omitempty"`
	LotCode                   *string                        `json:"lotCode,omitempty"`
	OriginatedOn              *string                        `json:"originatedOn,omitempty"`
	ExpiresOn                 *string                        `json:"expiresOn,omitempty"`
	Allocations               []AdjustmentAllocationResponse `json:"allocations"`
}

type AdjustmentAllocationResponse struct {
	ID             int64 `json:"id"`
	LotID          int64 `json:"lotId"`
	QuantityAtomic int64 `json:"quantityAtomic"`
}
