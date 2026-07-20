package application

import (
	"context"
	"strings"
	"time"

	"github.com/jerobas/saas/internal/domain"
)

type ReportingGranularity string

const (
	ReportingGranularityDay   ReportingGranularity = "DAY"
	ReportingGranularityMonth ReportingGranularity = "MONTH"
)

func NewReportingGranularity(raw string) (ReportingGranularity, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if normalized == "" {
		return ReportingGranularityMonth, nil
	}
	switch ReportingGranularity(normalized) {
	case ReportingGranularityDay, ReportingGranularityMonth:
		return ReportingGranularity(normalized), nil
	default:
		return "", domain.Invalid("granularity", domain.ViolationInvalidEnum, "RPT-001")
	}
}

type ReportingPeriodInput struct {
	FromOccurredOn domain.BusinessDate
	ToOccurredOn   domain.BusinessDate
	Granularity    ReportingGranularity
}

func NewReportingPeriodInput(from, to domain.BusinessDate, granularity ReportingGranularity) (ReportingPeriodInput, error) {
	if from.IsZero() {
		return ReportingPeriodInput{}, domain.Invalid("from_occurred_on", domain.ViolationRequired, "RPT-002")
	}
	if to.IsZero() {
		return ReportingPeriodInput{}, domain.Invalid("to_occurred_on", domain.ViolationRequired, "RPT-002")
	}
	if to.Before(from) {
		return ReportingPeriodInput{}, domain.Invalid("to_occurred_on", domain.ViolationOutOfRange, "RPT-003")
	}
	if granularity == "" {
		granularity = ReportingGranularityMonth
	}
	switch granularity {
	case ReportingGranularityDay, ReportingGranularityMonth:
	default:
		return ReportingPeriodInput{}, domain.Invalid("granularity", domain.ViolationInvalidEnum, "RPT-001")
	}
	return ReportingPeriodInput{FromOccurredOn: from, ToOccurredOn: to, Granularity: granularity}, nil
}

type SalesReport struct {
	Period                 ReportingPeriodInput
	Currency               domain.Currency
	TotalSalesCount        int64
	TotalRevenueMinor      int64
	COGSMicro              int64
	GrossMarginMicro       int64
	GrossMarginBasisPoints domain.Option[int64]
	AverageTicketMinor     domain.Option[int64]
	GrowthBasisPoints      domain.Option[int64]
	SalesRevenueSeries     []ReportingSeries
	MonthlyRevenueSeries   []ReportingSeries
	MonthlySalesSeries     []ReportingSeries
	TopProductsByQuantity  []ReportingItemMetric
	TopProductsByRevenue   []ReportingItemMetric
	FreeSales              ReportingReasonMetric
	SalesByCustomer        []ReportingCounterpartyMetric
	AnonymousSales         ReportingCounterpartyMetric
}

type InventoryReport struct {
	Period                   ReportingPeriodInput
	Currency                 domain.Currency
	TotalInventoryValueMicro int64
	LowStockItemCount        int64
	ZeroStockSellableCount   int64
	LowStockItems            []ReportingItemMetric
	ExpiringLots7Days        []ReportingLotMetric
	ExpiringLots30Days       []ReportingLotMetric
	ExpiredLotsWithStock     []ReportingLotMetric
	InventoryValueByItem     []ReportingItemMetric
}

type PurchaseReport struct {
	Period              ReportingPeriodInput
	Currency            domain.Currency
	PurchaseSpendSeries []ReportingSeries
	TopSuppliersBySpend []ReportingCounterpartyMetric
	FreeStockEntries    []ReportingSeries
}

type ProductionReport struct {
	Period                    ReportingPeriodInput
	Currency                  domain.Currency
	ProductionByRecipeProduct []ReportingItemMetric
	DirectCostSeries          []ReportingSeries
	YieldVariance             []ReportingItemMetric
}

type AdjustmentReport struct {
	Period           ReportingPeriodInput
	Currency         domain.Currency
	NegativeByReason []ReportingReasonMetric
	PositiveByReason []ReportingReasonMetric
	ExactReversals   []ReportingSeries
}

type CategoryMixReport struct {
	Period            ReportingPeriodInput
	Available         bool
	UnavailableReason string
	Rows              []CategoryMixRow
}

type CategoryMixRow struct {
	CategoryName     string
	QuantityAtomic   int64
	RevenueMinor     int64
	ShareBasisPoints int64
}

type ReportingStore interface {
	GetSalesReportData(ctx context.Context, current ReportingPeriodInput, previous ReportingPeriodInput, topLimit int) (SalesReportData, error)
	GetInventoryReportData(ctx context.Context, input ReportingPeriodInput, rowLimit int) (InventoryReportData, error)
	GetPurchaseReportData(ctx context.Context, input ReportingPeriodInput, rowLimit int) (PurchaseReportData, error)
	GetProductionReportData(ctx context.Context, input ReportingPeriodInput, rowLimit int) (ProductionReportData, error)
	GetAdjustmentReportData(ctx context.Context, input ReportingPeriodInput) (AdjustmentReportData, error)
}

type SalesReportData struct {
	Currency              domain.Currency
	CurrentTotals         SalesReportTotals
	PreviousTotals        SalesReportTotals
	SalesRevenueSeries    []ReportingSeries
	MonthlySeries         []ReportingSeries
	TopProductsByQuantity []ReportingItemMetric
	TopProductsByRevenue  []ReportingItemMetric
	FreeSales             ReportingReasonMetric
	SalesByCustomer       []ReportingCounterpartyMetric
	AnonymousSales        ReportingCounterpartyMetric
}

type SalesReportTotals struct {
	SalesCount     int64
	QuantityAtomic int64
	RevenueMinor   int64
	COGSMicro      int64
}

type InventoryReportData struct {
	Currency                 domain.Currency
	TotalInventoryValueMicro int64
	LowStockItemCount        int64
	ZeroStockSellableCount   int64
	LowStockItems            []ReportingItemMetric
	ExpiringLots7Days        []ReportingLotMetric
	ExpiringLots30Days       []ReportingLotMetric
	ExpiredLotsWithStock     []ReportingLotMetric
	InventoryValueByItem     []ReportingItemMetric
}

type PurchaseReportData struct {
	Currency             domain.Currency
	PurchaseSpendSeries  []ReportingSeries
	TopSuppliersBySpend  []ReportingCounterpartyMetric
	FreeStockEntrySeries []ReportingSeries
}

type ProductionReportData struct {
	Currency                  domain.Currency
	ProductionByRecipeProduct []ReportingItemMetric
	DirectCostSeries          []ReportingSeries
	YieldVariance             []ReportingItemMetric
}

type AdjustmentReportData struct {
	Currency         domain.Currency
	NegativeByReason []ReportingReasonMetric
	PositiveByReason []ReportingReasonMetric
	ExactReversals   []ReportingSeries
}

type ReportingSeries struct {
	Bucket              string
	Label               string
	DocumentCount       int64
	SalesCount          int64
	QuantityAtomic      int64
	RevenueMinor        int64
	SpendMinor          int64
	InventoryValueMicro int64
	DirectCostMicro     int64
	COGSMicro           int64
	GrossMarginMicro    int64
}

type ReportingItemMetric struct {
	ItemID                domain.Option[domain.ItemID]
	ItemName              string
	RecipeID              domain.Option[domain.RecipeID]
	RecipeName            domain.Option[string]
	BaseUnitCode          domain.Option[domain.UnitCode]
	DocumentCount         int64
	QuantityAtomic        int64
	RevenueMinor          int64
	InventoryValueMicro   int64
	DirectCostMicro       int64
	ReorderQuantityAtomic domain.Option[int64]
	StandardYieldAtomic   domain.Option[int64]
	ActualYieldAtomic     domain.Option[int64]
	VarianceAtomic        domain.Option[int64]
}

type ReportingLotMetric struct {
	LotID               domain.InventoryLotID
	ItemID              domain.ItemID
	ItemName            string
	LotCode             domain.Option[string]
	ExpiresOn           domain.Option[domain.BusinessDate]
	AvailableQuantity   int64
	InventoryValueMicro int64
}

type ReportingCounterpartyMetric struct {
	CounterpartyID   domain.Option[domain.CounterpartyID]
	CounterpartyName domain.Option[string]
	DocumentCount    int64
	RevenueMinor     int64
	SpendMinor       int64
}

type ReportingReasonMetric struct {
	ReasonCode          string
	DocumentCount       int64
	QuantityAtomic      int64
	RevenueMinor        int64
	InventoryValueMicro int64
}

type ReportingService struct {
	store ReportingStore
}

func NewReportingService(store ReportingStore) *ReportingService {
	if store == nil {
		panic("reporting service requires a store")
	}
	return &ReportingService{store: store}
}

func (s *ReportingService) GetSalesReport(ctx context.Context, input ReportingPeriodInput) (SalesReport, error) {
	previous, err := previousReportingPeriod(input)
	if err != nil {
		return SalesReport{}, err
	}
	data, err := s.store.GetSalesReportData(ctx, input, previous, 5)
	if err != nil {
		return SalesReport{}, err
	}
	revenueMicro, err := minorToMicro(data.CurrentTotals.RevenueMinor, data.Currency)
	if err != nil {
		return SalesReport{}, err
	}
	grossMarginMicro := revenueMicro - data.CurrentTotals.COGSMicro
	return SalesReport{
		Period:                 input,
		Currency:               data.Currency,
		TotalSalesCount:        data.CurrentTotals.SalesCount,
		TotalRevenueMinor:      data.CurrentTotals.RevenueMinor,
		COGSMicro:              data.CurrentTotals.COGSMicro,
		GrossMarginMicro:       grossMarginMicro,
		GrossMarginBasisPoints: ratioBasisPoints(grossMarginMicro, revenueMicro),
		AverageTicketMinor:     averageMinor(data.CurrentTotals.RevenueMinor, data.CurrentTotals.SalesCount),
		GrowthBasisPoints:      growthBasisPoints(data.CurrentTotals.RevenueMinor, data.PreviousTotals.RevenueMinor),
		SalesRevenueSeries:     enrichSeries(data.SalesRevenueSeries, data.Currency),
		MonthlyRevenueSeries:   enrichSeries(data.MonthlySeries, data.Currency),
		MonthlySalesSeries:     enrichSeries(data.MonthlySeries, data.Currency),
		TopProductsByQuantity:  data.TopProductsByQuantity,
		TopProductsByRevenue:   data.TopProductsByRevenue,
		FreeSales:              data.FreeSales,
		SalesByCustomer:        data.SalesByCustomer,
		AnonymousSales:         data.AnonymousSales,
	}, nil
}

func (s *ReportingService) GetInventoryReport(ctx context.Context, input ReportingPeriodInput) (InventoryReport, error) {
	data, err := s.store.GetInventoryReportData(ctx, input, 10)
	if err != nil {
		return InventoryReport{}, err
	}
	return InventoryReport{
		Period:                   input,
		Currency:                 data.Currency,
		TotalInventoryValueMicro: data.TotalInventoryValueMicro,
		LowStockItemCount:        data.LowStockItemCount,
		ZeroStockSellableCount:   data.ZeroStockSellableCount,
		LowStockItems:            data.LowStockItems,
		ExpiringLots7Days:        data.ExpiringLots7Days,
		ExpiringLots30Days:       data.ExpiringLots30Days,
		ExpiredLotsWithStock:     data.ExpiredLotsWithStock,
		InventoryValueByItem:     data.InventoryValueByItem,
	}, nil
}

func (s *ReportingService) GetPurchaseReport(ctx context.Context, input ReportingPeriodInput) (PurchaseReport, error) {
	data, err := s.store.GetPurchaseReportData(ctx, input, 10)
	if err != nil {
		return PurchaseReport{}, err
	}
	return PurchaseReport{
		Period:              input,
		Currency:            data.Currency,
		PurchaseSpendSeries: data.PurchaseSpendSeries,
		TopSuppliersBySpend: data.TopSuppliersBySpend,
		FreeStockEntries:    data.FreeStockEntrySeries,
	}, nil
}

func (s *ReportingService) GetProductionReport(ctx context.Context, input ReportingPeriodInput) (ProductionReport, error) {
	data, err := s.store.GetProductionReportData(ctx, input, 10)
	if err != nil {
		return ProductionReport{}, err
	}
	return ProductionReport{
		Period:                    input,
		Currency:                  data.Currency,
		ProductionByRecipeProduct: data.ProductionByRecipeProduct,
		DirectCostSeries:          data.DirectCostSeries,
		YieldVariance:             data.YieldVariance,
	}, nil
}

func (s *ReportingService) GetAdjustmentReport(ctx context.Context, input ReportingPeriodInput) (AdjustmentReport, error) {
	data, err := s.store.GetAdjustmentReportData(ctx, input)
	if err != nil {
		return AdjustmentReport{}, err
	}
	return AdjustmentReport{
		Period:           input,
		Currency:         data.Currency,
		NegativeByReason: data.NegativeByReason,
		PositiveByReason: data.PositiveByReason,
		ExactReversals:   data.ExactReversals,
	}, nil
}

func (s *ReportingService) GetCategoryMixReport(_ context.Context, input ReportingPeriodInput) (CategoryMixReport, error) {
	return CategoryMixReport{
		Period:            input,
		Available:         false,
		UnavailableReason: "Catalog categories/tags are not modeled in V2 yet.",
		Rows:              []CategoryMixRow{},
	}, nil
}

func previousReportingPeriod(input ReportingPeriodInput) (ReportingPeriodInput, error) {
	from, err := time.Parse("2006-01-02", input.FromOccurredOn.String())
	if err != nil {
		return ReportingPeriodInput{}, err
	}
	to, err := time.Parse("2006-01-02", input.ToOccurredOn.String())
	if err != nil {
		return ReportingPeriodInput{}, err
	}
	days := int(to.Sub(from).Hours()/24) + 1
	previousTo := from.AddDate(0, 0, -1)
	previousFrom := previousTo.AddDate(0, 0, -days+1)
	parsedFrom, err := domain.ParseBusinessDate(previousFrom.Format("2006-01-02"))
	if err != nil {
		return ReportingPeriodInput{}, err
	}
	parsedTo, err := domain.ParseBusinessDate(previousTo.Format("2006-01-02"))
	if err != nil {
		return ReportingPeriodInput{}, err
	}
	return NewReportingPeriodInput(parsedFrom, parsedTo, input.Granularity)
}

func minorToMicro(value int64, currency domain.Currency) (int64, error) {
	amount, err := domain.NewMinorAmount(value)
	if err != nil {
		return 0, err
	}
	converted, err := amount.ToInventoryValue(currency)
	if err != nil {
		return 0, err
	}
	return converted.Int64(), nil
}

func enrichSeries(items []ReportingSeries, currency domain.Currency) []ReportingSeries {
	enriched := make([]ReportingSeries, 0, len(items))
	for _, item := range items {
		revenueMicro, err := minorToMicro(item.RevenueMinor, currency)
		if err == nil {
			item.GrossMarginMicro = revenueMicro - item.COGSMicro
		}
		enriched = append(enriched, item)
	}
	return enriched
}

func averageMinor(totalMinor, count int64) domain.Option[int64] {
	if count <= 0 {
		return domain.None[int64]()
	}
	return domain.Some(totalMinor / count)
}

func growthBasisPoints(current, previous int64) domain.Option[int64] {
	if previous == 0 {
		return domain.None[int64]()
	}
	return domain.Some(((current - previous) * 10_000) / previous)
}

func ratioBasisPoints(numerator, denominator int64) domain.Option[int64] {
	if denominator == 0 {
		return domain.None[int64]()
	}
	return domain.Some((numerator * 10_000) / denominator)
}
