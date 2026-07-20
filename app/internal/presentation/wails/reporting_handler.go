package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type ReportingHandler struct {
	service *application.ReportingService
}

func NewReportingHandler(service *application.ReportingService) *ReportingHandler {
	if service == nil {
		panic("reporting handler requires a service")
	}
	return &ReportingHandler{service: service}
}

func (h *ReportingHandler) GetSalesReport(req dto.ReportingPeriodRequest) (dto.SalesReportResponse, error) {
	input, err := parseReportingPeriodRequest(req)
	if err != nil {
		return dto.SalesReportResponse{}, err
	}
	report, err := h.service.GetSalesReport(handlerContext(), input)
	if err != nil {
		return dto.SalesReportResponse{}, fmt.Errorf("get sales report: %w", err)
	}
	return mapSalesReport(report), nil
}

func (h *ReportingHandler) GetInventoryReport(req dto.ReportingPeriodRequest) (dto.InventoryReportResponse, error) {
	input, err := parseReportingPeriodRequest(req)
	if err != nil {
		return dto.InventoryReportResponse{}, err
	}
	report, err := h.service.GetInventoryReport(handlerContext(), input)
	if err != nil {
		return dto.InventoryReportResponse{}, fmt.Errorf("get inventory report: %w", err)
	}
	return mapInventoryReport(report), nil
}

func (h *ReportingHandler) GetPurchaseReport(req dto.ReportingPeriodRequest) (dto.PurchaseReportResponse, error) {
	input, err := parseReportingPeriodRequest(req)
	if err != nil {
		return dto.PurchaseReportResponse{}, err
	}
	report, err := h.service.GetPurchaseReport(handlerContext(), input)
	if err != nil {
		return dto.PurchaseReportResponse{}, fmt.Errorf("get purchase report: %w", err)
	}
	return mapPurchaseReport(report), nil
}

func (h *ReportingHandler) GetProductionReport(req dto.ReportingPeriodRequest) (dto.ProductionReportResponse, error) {
	input, err := parseReportingPeriodRequest(req)
	if err != nil {
		return dto.ProductionReportResponse{}, err
	}
	report, err := h.service.GetProductionReport(handlerContext(), input)
	if err != nil {
		return dto.ProductionReportResponse{}, fmt.Errorf("get production report: %w", err)
	}
	return mapProductionReport(report), nil
}

func (h *ReportingHandler) GetAdjustmentReport(req dto.ReportingPeriodRequest) (dto.AdjustmentReportResponse, error) {
	input, err := parseReportingPeriodRequest(req)
	if err != nil {
		return dto.AdjustmentReportResponse{}, err
	}
	report, err := h.service.GetAdjustmentReport(handlerContext(), input)
	if err != nil {
		return dto.AdjustmentReportResponse{}, fmt.Errorf("get adjustment report: %w", err)
	}
	return mapAdjustmentReport(report), nil
}

func (h *ReportingHandler) GetCategoryMixReport(req dto.ReportingPeriodRequest) (dto.CategoryMixReportResponse, error) {
	input, err := parseReportingPeriodRequest(req)
	if err != nil {
		return dto.CategoryMixReportResponse{}, err
	}
	report, err := h.service.GetCategoryMixReport(handlerContext(), input)
	if err != nil {
		return dto.CategoryMixReportResponse{}, fmt.Errorf("get category mix report: %w", err)
	}
	return mapCategoryMixReport(report), nil
}

func parseReportingPeriodRequest(req dto.ReportingPeriodRequest) (application.ReportingPeriodInput, error) {
	from, err := domain.ParseBusinessDate(req.FromOccurredOn)
	if err != nil {
		return application.ReportingPeriodInput{}, fmt.Errorf("from occurred on: %w", err)
	}
	to, err := domain.ParseBusinessDate(req.ToOccurredOn)
	if err != nil {
		return application.ReportingPeriodInput{}, fmt.Errorf("to occurred on: %w", err)
	}
	granularity, err := application.NewReportingGranularity(req.Granularity)
	if err != nil {
		return application.ReportingPeriodInput{}, fmt.Errorf("granularity: %w", err)
	}
	input, err := application.NewReportingPeriodInput(from, to, granularity)
	if err != nil {
		return application.ReportingPeriodInput{}, fmt.Errorf("reporting period: %w", err)
	}
	return input, nil
}

func mapReportingPeriod(input application.ReportingPeriodInput) dto.ReportingPeriodResponse {
	return dto.ReportingPeriodResponse{
		FromOccurredOn: input.FromOccurredOn.String(),
		ToOccurredOn:   input.ToOccurredOn.String(),
		Granularity:    string(input.Granularity),
	}
}

func mapSalesReport(report application.SalesReport) dto.SalesReportResponse {
	return dto.SalesReportResponse{
		Period:                 mapReportingPeriod(report.Period),
		CurrencyCode:           report.Currency.Code().String(),
		CurrencyMinorDigits:    int64(report.Currency.MinorDigits().Int()),
		TotalSalesCount:        report.TotalSalesCount,
		TotalRevenueMinor:      report.TotalRevenueMinor,
		COGSMicro:              report.COGSMicro,
		GrossMarginMicro:       report.GrossMarginMicro,
		GrossMarginBasisPoints: optionalInt64(report.GrossMarginBasisPoints),
		AverageTicketMinor:     optionalInt64(report.AverageTicketMinor),
		GrowthBasisPoints:      optionalInt64(report.GrowthBasisPoints),
		SalesRevenueSeries:     mapReportingSeries(report.SalesRevenueSeries),
		MonthlyRevenueSeries:   mapReportingSeries(report.MonthlyRevenueSeries),
		MonthlySalesSeries:     mapReportingSeries(report.MonthlySalesSeries),
		TopProductsByQuantity:  mapReportingItemMetrics(report.TopProductsByQuantity),
		TopProductsByRevenue:   mapReportingItemMetrics(report.TopProductsByRevenue),
		FreeSales:              mapReportingReasonMetric(report.FreeSales),
		SalesByCustomer:        mapReportingCounterpartyMetrics(report.SalesByCustomer),
		AnonymousSales:         mapReportingCounterpartyMetric(report.AnonymousSales),
	}
}

func mapInventoryReport(report application.InventoryReport) dto.InventoryReportResponse {
	return dto.InventoryReportResponse{
		Period:                   mapReportingPeriod(report.Period),
		CurrencyCode:             report.Currency.Code().String(),
		CurrencyMinorDigits:      int64(report.Currency.MinorDigits().Int()),
		TotalInventoryValueMicro: report.TotalInventoryValueMicro,
		LowStockItemCount:        report.LowStockItemCount,
		ZeroStockSellableCount:   report.ZeroStockSellableCount,
		LowStockItems:            mapReportingItemMetrics(report.LowStockItems),
		ExpiringLots7Days:        mapReportingLotMetrics(report.ExpiringLots7Days),
		ExpiringLots30Days:       mapReportingLotMetrics(report.ExpiringLots30Days),
		ExpiredLotsWithStock:     mapReportingLotMetrics(report.ExpiredLotsWithStock),
		InventoryValueByItem:     mapReportingItemMetrics(report.InventoryValueByItem),
	}
}

func mapPurchaseReport(report application.PurchaseReport) dto.PurchaseReportResponse {
	return dto.PurchaseReportResponse{
		Period:              mapReportingPeriod(report.Period),
		CurrencyCode:        report.Currency.Code().String(),
		CurrencyMinorDigits: int64(report.Currency.MinorDigits().Int()),
		PurchaseSpendSeries: mapReportingSeries(report.PurchaseSpendSeries),
		TopSuppliersBySpend: mapReportingCounterpartyMetrics(report.TopSuppliersBySpend),
		FreeStockEntries:    mapReportingSeries(report.FreeStockEntries),
	}
}

func mapProductionReport(report application.ProductionReport) dto.ProductionReportResponse {
	return dto.ProductionReportResponse{
		Period:                    mapReportingPeriod(report.Period),
		CurrencyCode:              report.Currency.Code().String(),
		CurrencyMinorDigits:       int64(report.Currency.MinorDigits().Int()),
		ProductionByRecipeProduct: mapReportingItemMetrics(report.ProductionByRecipeProduct),
		DirectCostSeries:          mapReportingSeries(report.DirectCostSeries),
		YieldVariance:             mapReportingItemMetrics(report.YieldVariance),
	}
}

func mapAdjustmentReport(report application.AdjustmentReport) dto.AdjustmentReportResponse {
	return dto.AdjustmentReportResponse{
		Period:              mapReportingPeriod(report.Period),
		CurrencyCode:        report.Currency.Code().String(),
		CurrencyMinorDigits: int64(report.Currency.MinorDigits().Int()),
		NegativeByReason:    mapReportingReasonMetrics(report.NegativeByReason),
		PositiveByReason:    mapReportingReasonMetrics(report.PositiveByReason),
		ExactReversals:      mapReportingSeries(report.ExactReversals),
	}
}

func mapCategoryMixReport(report application.CategoryMixReport) dto.CategoryMixReportResponse {
	rows := make([]dto.CategoryMixRowResponse, 0, len(report.Rows))
	for _, row := range report.Rows {
		rows = append(rows, dto.CategoryMixRowResponse{
			CategoryName:     row.CategoryName,
			QuantityAtomic:   row.QuantityAtomic,
			RevenueMinor:     row.RevenueMinor,
			ShareBasisPoints: row.ShareBasisPoints,
		})
	}
	return dto.CategoryMixReportResponse{
		Period:            mapReportingPeriod(report.Period),
		Available:         report.Available,
		UnavailableReason: optionalString(report.UnavailableReason),
		Rows:              rows,
	}
}

func mapReportingSeries(items []application.ReportingSeries) []dto.ReportingSeriesResponse {
	response := make([]dto.ReportingSeriesResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.ReportingSeriesResponse{
			Bucket:              item.Bucket,
			Label:               item.Label,
			DocumentCount:       item.DocumentCount,
			SalesCount:          item.SalesCount,
			QuantityAtomic:      item.QuantityAtomic,
			RevenueMinor:        item.RevenueMinor,
			SpendMinor:          item.SpendMinor,
			InventoryValueMicro: item.InventoryValueMicro,
			DirectCostMicro:     item.DirectCostMicro,
			GrossMarginMicro:    item.GrossMarginMicro,
		})
	}
	return response
}

func mapReportingItemMetrics(items []application.ReportingItemMetric) []dto.ReportingItemMetricResponse {
	response := make([]dto.ReportingItemMetricResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.ReportingItemMetricResponse{
			ItemID:                optionalItemID(item.ItemID),
			ItemName:              item.ItemName,
			RecipeID:              optionalRecipeID(item.RecipeID),
			RecipeName:            optionalStringOption(item.RecipeName),
			BaseUnitCode:          optionalUnitCode(item.BaseUnitCode),
			DocumentCount:         item.DocumentCount,
			QuantityAtomic:        item.QuantityAtomic,
			RevenueMinor:          item.RevenueMinor,
			InventoryValueMicro:   item.InventoryValueMicro,
			DirectCostMicro:       item.DirectCostMicro,
			ReorderQuantityAtomic: optionalInt64(item.ReorderQuantityAtomic),
			StandardYieldAtomic:   optionalInt64(item.StandardYieldAtomic),
			ActualYieldAtomic:     optionalInt64(item.ActualYieldAtomic),
			VarianceAtomic:        optionalInt64(item.VarianceAtomic),
		})
	}
	return response
}

func mapReportingLotMetrics(items []application.ReportingLotMetric) []dto.ReportingLotMetricResponse {
	response := make([]dto.ReportingLotMetricResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.ReportingLotMetricResponse{
			LotID:               item.LotID.Int64(),
			ItemID:              item.ItemID.Int64(),
			ItemName:            item.ItemName,
			LotCode:             optionalStringOption(item.LotCode),
			ExpiresOn:           optionalReportingBusinessDate(item.ExpiresOn),
			AvailableQuantity:   item.AvailableQuantity,
			InventoryValueMicro: item.InventoryValueMicro,
		})
	}
	return response
}

func mapReportingCounterpartyMetrics(items []application.ReportingCounterpartyMetric) []dto.ReportingCounterpartyMetricResponse {
	response := make([]dto.ReportingCounterpartyMetricResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapReportingCounterpartyMetric(item))
	}
	return response
}

func mapReportingCounterpartyMetric(item application.ReportingCounterpartyMetric) dto.ReportingCounterpartyMetricResponse {
	return dto.ReportingCounterpartyMetricResponse{
		CounterpartyID:   optionalCounterpartyIDValue(item.CounterpartyID),
		CounterpartyName: optionalStringOption(item.CounterpartyName),
		DocumentCount:    item.DocumentCount,
		RevenueMinor:     item.RevenueMinor,
		SpendMinor:       item.SpendMinor,
	}
}

func mapReportingReasonMetric(item application.ReportingReasonMetric) dto.ReportingReasonMetricResponse {
	return dto.ReportingReasonMetricResponse{
		ReasonCode:          item.ReasonCode,
		DocumentCount:       item.DocumentCount,
		QuantityAtomic:      item.QuantityAtomic,
		RevenueMinor:        item.RevenueMinor,
		InventoryValueMicro: item.InventoryValueMicro,
	}
}

func mapReportingReasonMetrics(items []application.ReportingReasonMetric) []dto.ReportingReasonMetricResponse {
	response := make([]dto.ReportingReasonMetricResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapReportingReasonMetric(item))
	}
	return response
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalStringOption(value domain.Option[string]) *string {
	raw, ok := value.Get()
	if !ok {
		return nil
	}
	return &raw
}

func optionalReportingBusinessDate(value domain.Option[domain.BusinessDate]) *string {
	date, ok := value.Get()
	if !ok {
		return nil
	}
	raw := date.String()
	return &raw
}

func optionalInt64(value domain.Option[int64]) *int64 {
	raw, ok := value.Get()
	if !ok {
		return nil
	}
	return &raw
}

func optionalItemID(value domain.Option[domain.ItemID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}

func optionalRecipeID(value domain.Option[domain.RecipeID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}

func optionalUnitCode(value domain.Option[domain.UnitCode]) *string {
	code, ok := value.Get()
	if !ok {
		return nil
	}
	raw := code.String()
	return &raw
}
