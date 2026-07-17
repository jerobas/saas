package application

import (
	"context"
	"errors"
	"strings"

	"github.com/jerobas/saas/internal/domain"
)

var ErrReportingEndpointNotImplemented = errors.New("reporting endpoint not implemented")

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
	return ReportingPeriodInput{FromOccurredOn: from, ToOccurredOn: to, Granularity: granularity}, nil
}

type DashboardReport struct{}

type SalesReport struct{}

type InventoryReport struct{}

type PurchaseReport struct{}

type ProductionReport struct{}

type AdjustmentReport struct{}

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

type ReportingService struct{}

func NewReportingService() *ReportingService {
	return &ReportingService{}
}

func (s *ReportingService) GetDashboardReport(context.Context, ReportingPeriodInput) (DashboardReport, error) {
	return DashboardReport{}, ErrReportingEndpointNotImplemented
}

func (s *ReportingService) GetSalesReport(context.Context, ReportingPeriodInput) (SalesReport, error) {
	return SalesReport{}, ErrReportingEndpointNotImplemented
}

func (s *ReportingService) GetInventoryReport(context.Context, ReportingPeriodInput) (InventoryReport, error) {
	return InventoryReport{}, ErrReportingEndpointNotImplemented
}

func (s *ReportingService) GetPurchaseReport(context.Context, ReportingPeriodInput) (PurchaseReport, error) {
	return PurchaseReport{}, ErrReportingEndpointNotImplemented
}

func (s *ReportingService) GetProductionReport(context.Context, ReportingPeriodInput) (ProductionReport, error) {
	return ProductionReport{}, ErrReportingEndpointNotImplemented
}

func (s *ReportingService) GetAdjustmentReport(context.Context, ReportingPeriodInput) (AdjustmentReport, error) {
	return AdjustmentReport{}, ErrReportingEndpointNotImplemented
}

func (s *ReportingService) GetCategoryMixReport(_ context.Context, input ReportingPeriodInput) (CategoryMixReport, error) {
	return CategoryMixReport{
		Period:            input,
		Available:         false,
		UnavailableReason: "Catalog categories/tags are not modeled in V2 yet.",
		Rows:              []CategoryMixRow{},
	}, nil
}
