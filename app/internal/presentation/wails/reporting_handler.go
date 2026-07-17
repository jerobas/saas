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

func (h *ReportingHandler) GetDashboardReport(req dto.ReportingPeriodRequest) (dto.DashboardReportResponse, error) {
	input, err := parseReportingPeriodRequest(req)
	if err != nil {
		return dto.DashboardReportResponse{}, err
	}
	report, err := h.service.GetDashboardReport(handlerContext(), input)
	if err != nil {
		return dto.DashboardReportResponse{}, fmt.Errorf("get dashboard report: %w", err)
	}
	return mapDashboardReport(report), nil
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

func mapDashboardReport(application.DashboardReport) dto.DashboardReportResponse {
	return dto.DashboardReportResponse{}
}

func mapSalesReport(application.SalesReport) dto.SalesReportResponse {
	return dto.SalesReportResponse{}
}

func mapInventoryReport(application.InventoryReport) dto.InventoryReportResponse {
	return dto.InventoryReportResponse{}
}

func mapPurchaseReport(application.PurchaseReport) dto.PurchaseReportResponse {
	return dto.PurchaseReportResponse{}
}

func mapProductionReport(application.ProductionReport) dto.ProductionReportResponse {
	return dto.ProductionReportResponse{}
}

func mapAdjustmentReport(application.AdjustmentReport) dto.AdjustmentReportResponse {
	return dto.AdjustmentReportResponse{}
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

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
