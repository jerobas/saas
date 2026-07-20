package dto

type ReportingPeriodRequest struct {
	FromOccurredOn string `json:"fromOccurredOn"`
	ToOccurredOn   string `json:"toOccurredOn"`
	Granularity    string `json:"granularity,omitempty"`
}

type ReportingPeriodResponse struct {
	FromOccurredOn string `json:"fromOccurredOn"`
	ToOccurredOn   string `json:"toOccurredOn"`
	Granularity    string `json:"granularity"`
}

type SalesReportResponse struct {
	Period                         ReportingPeriodResponse               `json:"period"`
	CurrencyCode                   string                                `json:"currencyCode"`
	CurrencyMinorDigits            int64                                 `json:"currencyMinorDigits"`
	TotalSalesCount                int64                                 `json:"totalSalesCount"`
	CommercialTotalMinor           int64                                 `json:"commercialTotalMinor"`
	COGSInventoryValueMicro        int64                                 `json:"cogsInventoryValueMicro"`
	GrossMarginInventoryValueMicro int64                                 `json:"grossMarginInventoryValueMicro"`
	GrossMarginBasisPoints         *int64                                `json:"grossMarginBasisPoints,omitempty"`
	AverageCommercialTotalMinor    *int64                                `json:"averageCommercialTotalMinor,omitempty"`
	GrowthBasisPoints              *int64                                `json:"growthBasisPoints,omitempty"`
	SalesRevenueSeries             []ReportingSeriesResponse             `json:"salesRevenueSeries"`
	MonthlyRevenueSeries           []ReportingSeriesResponse             `json:"monthlyRevenueSeries"`
	MonthlySalesSeries             []ReportingSeriesResponse             `json:"monthlySalesSeries"`
	TopProductsByQuantity          []ReportingItemMetricResponse         `json:"topProductsByQuantity"`
	TopProductsByRevenue           []ReportingItemMetricResponse         `json:"topProductsByRevenue"`
	FreeSales                      ReportingReasonMetricResponse         `json:"freeSales"`
	SalesByCustomer                []ReportingCounterpartyMetricResponse `json:"salesByCustomer"`
	AnonymousSales                 ReportingCounterpartyMetricResponse   `json:"anonymousSales"`
}

type InventoryReportResponse struct {
	Period                   ReportingPeriodResponse       `json:"period"`
	CurrencyCode             string                        `json:"currencyCode"`
	CurrencyMinorDigits      int64                         `json:"currencyMinorDigits"`
	TotalInventoryValueMicro int64                         `json:"totalInventoryValueMicro"`
	LowStockItemCount        int64                         `json:"lowStockItemCount"`
	ZeroStockSellableCount   int64                         `json:"zeroStockSellableCount"`
	LowStockItems            []ReportingItemMetricResponse `json:"lowStockItems"`
	ExpiringLots7Days        []ReportingLotMetricResponse  `json:"expiringLots7Days"`
	ExpiringLots30Days       []ReportingLotMetricResponse  `json:"expiringLots30Days"`
	ExpiredLotsWithStock     []ReportingLotMetricResponse  `json:"expiredLotsWithStock"`
	InventoryValueByItem     []ReportingItemMetricResponse `json:"inventoryValueByItem"`
}

type PurchaseReportResponse struct {
	Period              ReportingPeriodResponse               `json:"period"`
	CurrencyCode        string                                `json:"currencyCode"`
	CurrencyMinorDigits int64                                 `json:"currencyMinorDigits"`
	PurchaseSpendSeries []ReportingSeriesResponse             `json:"purchaseSpendSeries"`
	TopSuppliersBySpend []ReportingCounterpartyMetricResponse `json:"topSuppliersBySpend"`
	FreeStockEntries    []ReportingSeriesResponse             `json:"freeStockEntries"`
}

type ProductionReportResponse struct {
	Period                    ReportingPeriodResponse       `json:"period"`
	CurrencyCode              string                        `json:"currencyCode"`
	CurrencyMinorDigits       int64                         `json:"currencyMinorDigits"`
	ProductionByRecipeProduct []ReportingItemMetricResponse `json:"productionByRecipeProduct"`
	DirectCostSeries          []ReportingSeriesResponse     `json:"directCostSeries"`
	YieldVariance             []ReportingItemMetricResponse `json:"yieldVariance"`
}

type AdjustmentReportResponse struct {
	Period              ReportingPeriodResponse         `json:"period"`
	CurrencyCode        string                          `json:"currencyCode"`
	CurrencyMinorDigits int64                           `json:"currencyMinorDigits"`
	NegativeByReason    []ReportingReasonMetricResponse `json:"negativeByReason"`
	PositiveByReason    []ReportingReasonMetricResponse `json:"positiveByReason"`
	ExactReversals      []ReportingSeriesResponse       `json:"exactReversals"`
}

type CategoryMixReportResponse struct {
	Period            ReportingPeriodResponse  `json:"period"`
	Available         bool                     `json:"available"`
	UnavailableReason *string                  `json:"unavailableReason,omitempty"`
	Rows              []CategoryMixRowResponse `json:"rows"`
}

type CategoryMixRowResponse struct {
	CategoryName         string `json:"categoryName"`
	QuantityAtomic       int64  `json:"quantityAtomic"`
	CommercialTotalMinor int64  `json:"commercialTotalMinor"`
	ShareBasisPoints     int64  `json:"shareBasisPoints"`
}

type ReportingSeriesResponse struct {
	Bucket                         string `json:"bucket"`
	Label                          string `json:"label"`
	DocumentCount                  int64  `json:"documentCount"`
	SalesCount                     int64  `json:"salesCount"`
	QuantityAtomic                 int64  `json:"quantityAtomic"`
	CommercialTotalMinor           int64  `json:"commercialTotalMinor"`
	InventoryValueMicro            int64  `json:"inventoryValueMicro"`
	DirectCostInventoryValueMicro  int64  `json:"directCostInventoryValueMicro"`
	GrossMarginInventoryValueMicro int64  `json:"grossMarginInventoryValueMicro"`
}

type ReportingItemMetricResponse struct {
	ItemID                        *int64  `json:"itemId,omitempty"`
	ItemName                      string  `json:"itemName"`
	RecipeID                      *int64  `json:"recipeId,omitempty"`
	RecipeName                    *string `json:"recipeName,omitempty"`
	BaseUnitCode                  *string `json:"baseUnitCode,omitempty"`
	DocumentCount                 int64   `json:"documentCount"`
	QuantityAtomic                int64   `json:"quantityAtomic"`
	CommercialTotalMinor          int64   `json:"commercialTotalMinor"`
	InventoryValueMicro           int64   `json:"inventoryValueMicro"`
	DirectCostInventoryValueMicro int64   `json:"directCostInventoryValueMicro"`
	ReorderQuantityAtomic         *int64  `json:"reorderQuantityAtomic,omitempty"`
	StandardYieldAtomic           *int64  `json:"standardYieldAtomic,omitempty"`
	ActualYieldAtomic             *int64  `json:"actualYieldAtomic,omitempty"`
	VarianceAtomic                *int64  `json:"varianceAtomic,omitempty"`
}

type ReportingCounterpartyMetricResponse struct {
	CounterpartyID       *int64  `json:"counterpartyId,omitempty"`
	CounterpartyName     *string `json:"counterpartyName,omitempty"`
	DocumentCount        int64   `json:"documentCount"`
	CommercialTotalMinor int64   `json:"commercialTotalMinor"`
}

type ReportingReasonMetricResponse struct {
	ReasonCode           string `json:"reasonCode"`
	DocumentCount        int64  `json:"documentCount"`
	QuantityAtomic       int64  `json:"quantityAtomic"`
	CommercialTotalMinor int64  `json:"commercialTotalMinor"`
	InventoryValueMicro  int64  `json:"inventoryValueMicro"`
}

type ReportingLotMetricResponse struct {
	LotID               int64   `json:"lotId"`
	ItemID              int64   `json:"itemId"`
	ItemName            string  `json:"itemName"`
	LotCode             *string `json:"lotCode,omitempty"`
	ExpiresOn           *string `json:"expiresOn,omitempty"`
	AvailableQuantity   int64   `json:"availableQuantityAtomic"`
	InventoryValueMicro int64   `json:"inventoryValueMicro"`
}
