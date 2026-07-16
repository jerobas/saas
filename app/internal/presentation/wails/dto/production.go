package dto

type ProductionPostRequest struct {
	IdempotencyKey   string                       `json:"idempotencyKey"`
	RecipeRevisionID int64                        `json:"recipeRevisionId"`
	OccurredOn       string                       `json:"occurredOn"`
	DirectCostMicro  int64                        `json:"directCostMicro"`
	Notes            *string                      `json:"notes,omitempty"`
	Output           ProductionOutputRequest      `json:"output"`
	Inputs           []ProductionComponentRequest `json:"inputs"`
}

type ProductionOutputRequest struct {
	QuantityAtomic            int64   `json:"quantityAtomic"`
	EnteredUnitCode           string  `json:"enteredUnitCode"`
	EnteredPackagingName      *string `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64   `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64   `json:"conversionDenominator"`
	LotCode                   *string `json:"lotCode,omitempty"`
	ExpiresOn                 *string `json:"expiresOn,omitempty"`
}

type ProductionComponentRequest struct {
	ItemID                    int64   `json:"itemId"`
	QuantityAtomic            int64   `json:"quantityAtomic"`
	EnteredUnitCode           string  `json:"enteredUnitCode"`
	EnteredPackagingName      *string `json:"enteredPackagingName,omitempty"`
	ConversionNumeratorAtomic int64   `json:"conversionNumeratorAtomic"`
	ConversionDenominator     int64   `json:"conversionDenominator"`
	LotID                     *int64  `json:"lotId,omitempty"`
}

type ProductionDocumentResponse struct {
	ID                  int64                    `json:"id"`
	IdempotencyKey      string                   `json:"idempotencyKey"`
	PostingSequence     int64                    `json:"postingSequence"`
	RecipeRevisionID    int64                    `json:"recipeRevisionId"`
	OutputItemID        int64                    `json:"outputItemId"`
	OccurredOn          string                   `json:"occurredOn"`
	PostedAtMs          int64                    `json:"postedAtMs"`
	CurrencyCode        string                   `json:"currencyCode"`
	CurrencyMinorDigits int64                    `json:"currencyMinorDigits"`
	DirectCostMicro     int64                    `json:"directCostMicro"`
	Notes               *string                  `json:"notes,omitempty"`
	OutputLine          ProductionLineResponse   `json:"outputLine"`
	InputLines          []ProductionLineResponse `json:"inputLines"`
}

type ProductionLineResponse struct {
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
	Allocations               []ProductionAllocationResponse `json:"allocations"`
}

type ProductionAllocationResponse struct {
	ID             int64 `json:"id"`
	LotID          int64 `json:"lotId"`
	QuantityAtomic int64 `json:"quantityAtomic"`
}
