package dto

type CapabilitiesRequest struct {
	Purchasable bool `json:"purchasable"`
	Producible  bool `json:"producible"`
	Sellable    bool `json:"sellable"`
}

type CapabilitiesResponse struct {
	Purchasable bool `json:"purchasable"`
	Producible  bool `json:"producible"`
	Sellable    bool `json:"sellable"`
}

type ItemCursorRequest struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type ItemCursorResponse struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type ItemListRequest struct {
	ArchiveFilter       string              `json:"archiveFilter,omitempty"`
	RequireCapabilities CapabilitiesRequest `json:"requireCapabilities"`
	Search              *string             `json:"search,omitempty"`
	After               *ItemCursorRequest  `json:"after,omitempty"`
	PageSize            int                 `json:"pageSize,omitempty"`
}

type ItemPageResponse struct {
	Items []ItemSummaryResponse `json:"items"`
	Next  *ItemCursorResponse   `json:"next,omitempty"`
}

type ItemSummaryResponse struct {
	ID               int64                `json:"id"`
	Name             string               `json:"name"`
	SKU              *string              `json:"sku,omitempty"`
	Description      *string              `json:"description,omitempty"`
	BaseUnitCode     string               `json:"baseUnitCode"`
	Capabilities     CapabilitiesResponse `json:"capabilities"`
	DefaultSalePrice *int64               `json:"defaultSalePrice,omitempty"`
	ReorderQuantity  *int64               `json:"reorderQuantityAtomic,omitempty"`
	CreatedAtMs      int64                `json:"createdAtMs"`
	UpdatedAtMs      int64                `json:"updatedAtMs"`
	ArchivedAtMs     *int64               `json:"archivedAtMs,omitempty"`
}

type ItemResponse struct {
	ItemSummaryResponse
	BaseUnit   MeasurementUnitResponse `json:"baseUnit"`
	Packagings []PackagingResponse     `json:"packagings"`
}

type ItemWriteRequest struct {
	Name             string              `json:"name"`
	SKU              *string             `json:"sku,omitempty"`
	Description      *string             `json:"description,omitempty"`
	BaseUnitCode     string              `json:"baseUnitCode"`
	Capabilities     CapabilitiesRequest `json:"capabilities"`
	DefaultSalePrice *int64              `json:"defaultSalePrice,omitempty"`
	ReorderQuantity  *int64              `json:"reorderQuantityAtomic,omitempty"`
}

type ItemUpdateRequest struct {
	ItemWriteRequest
	ExpectedUpdatedAtMs int64 `json:"expectedUpdatedAtMs"`
}

type VersionedRequest struct {
	ExpectedUpdatedAtMs int64 `json:"expectedUpdatedAtMs"`
}

type PackagingResponse struct {
	ID                    int64                   `json:"id"`
	ItemID                int64                   `json:"itemId"`
	Name                  string                  `json:"name"`
	EnteredUnitCode       string                  `json:"enteredUnitCode"`
	ConversionNumerator   int64                   `json:"conversionNumeratorAtomic"`
	ConversionDenominator int64                   `json:"conversionDenominator"`
	BaseUnit              MeasurementUnitResponse `json:"baseUnit"`
	EnteredUnit           MeasurementUnitResponse `json:"enteredUnit"`
	CreatedAtMs           int64                   `json:"createdAtMs"`
	UpdatedAtMs           int64                   `json:"updatedAtMs"`
	ArchivedAtMs          *int64                  `json:"archivedAtMs,omitempty"`
}

type PackagingCreateRequest struct {
	ItemID int64 `json:"itemId"`
	PackagingWriteRequest
}

type PackagingWriteRequest struct {
	Name                  string `json:"name"`
	EnteredUnitCode       string `json:"enteredUnitCode"`
	ConversionNumerator   int64  `json:"conversionNumeratorAtomic"`
	ConversionDenominator int64  `json:"conversionDenominator"`
}

type PackagingUpdateRequest struct {
	PackagingWriteRequest
	ExpectedUpdatedAtMs int64 `json:"expectedUpdatedAtMs"`
}
