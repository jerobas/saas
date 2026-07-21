package application

import (
	"context"
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
)

func TestReportingPeriodInputRejectsInvalidRange(t *testing.T) {
	from := mustReportingBusinessDate(t, "2026-07-31")
	to := mustReportingBusinessDate(t, "2026-07-01")

	_, err := NewReportingPeriodInput(from, to, ReportingGranularityDay)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid range error = %v, want validation", err)
	}

	var validation *domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("invalid range error did not expose validation metadata: %v", err)
	}
	violations := validation.Violations()
	if len(violations) != 1 ||
		violations[0].Field != "to_occurred_on" ||
		violations[0].Code != domain.ViolationOutOfRange ||
		violations[0].InvariantID != "RPT-003" {
		t.Fatalf("invalid range violations = %#v", violations)
	}
}

func TestReportingServiceUsesDefaultMonthGranularityAndPreviousPeriod(t *testing.T) {
	from := mustReportingBusinessDate(t, "2026-07-10")
	to := mustReportingBusinessDate(t, "2026-07-12")
	input, err := NewReportingPeriodInput(from, to, "")
	if err != nil {
		t.Fatalf("new reporting period: %v", err)
	}

	store := &recordingReportingStore{currency: mustReportingCurrency(t)}
	service := NewReportingService(store)
	if _, err := service.GetSalesReport(context.Background(), input); err != nil {
		t.Fatalf("get sales report: %v", err)
	}

	if store.salesCalls != 1 {
		t.Fatalf("sales report calls = %d, want 1", store.salesCalls)
	}
	if store.current.Granularity != ReportingGranularityMonth {
		t.Fatalf("current granularity = %q, want month", store.current.Granularity)
	}
	if store.previous.FromOccurredOn.String() != "2026-07-07" ||
		store.previous.ToOccurredOn.String() != "2026-07-09" ||
		store.previous.Granularity != ReportingGranularityMonth {
		t.Fatalf("previous period = %#v", store.previous)
	}
}

type recordingReportingStore struct {
	currency   domain.Currency
	salesCalls int
	current    ReportingPeriodInput
	previous   ReportingPeriodInput
}

func (s *recordingReportingStore) GetSalesReportData(
	_ context.Context,
	current ReportingPeriodInput,
	previous ReportingPeriodInput,
	_ int,
) (SalesReportData, error) {
	s.salesCalls++
	s.current = current
	s.previous = previous
	return SalesReportData{
		Currency:       s.currency,
		CurrentTotals:  SalesReportTotals{},
		PreviousTotals: SalesReportTotals{},
	}, nil
}

func (s *recordingReportingStore) GetInventoryReportData(
	context.Context,
	ReportingPeriodInput,
	int,
) (InventoryReportData, error) {
	return InventoryReportData{Currency: s.currency}, nil
}

func (s *recordingReportingStore) GetPurchaseReportData(
	context.Context,
	ReportingPeriodInput,
	int,
) (PurchaseReportData, error) {
	return PurchaseReportData{Currency: s.currency}, nil
}

func (s *recordingReportingStore) GetProductionReportData(
	context.Context,
	ReportingPeriodInput,
	int,
) (ProductionReportData, error) {
	return ProductionReportData{Currency: s.currency}, nil
}

func (s *recordingReportingStore) GetAdjustmentReportData(
	context.Context,
	ReportingPeriodInput,
) (AdjustmentReportData, error) {
	return AdjustmentReportData{Currency: s.currency}, nil
}

func mustReportingBusinessDate(t *testing.T, raw string) domain.BusinessDate {
	t.Helper()
	value, err := domain.ParseBusinessDate(raw)
	if err != nil {
		t.Fatalf("parse business date %q: %v", raw, err)
	}
	return value
}

func mustReportingCurrency(t *testing.T) domain.Currency {
	t.Helper()
	value, err := domain.NewCurrency("BRL")
	if err != nil {
		t.Fatalf("new reporting currency: %v", err)
	}
	return value
}
