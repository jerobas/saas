package application

import (
	"context"

	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteReportingStore struct {
	store *sqlite.Store
}

func NewSQLiteReportingStore(store *sqlite.Store) ReportingStore {
	if store == nil {
		panic("sqlite reporting store requires a store")
	}
	return &sqliteReportingStore{store: store}
}

func (s *sqliteReportingStore) GetSalesReportData(
	ctx context.Context,
	current ReportingPeriodInput,
	previous ReportingPeriodInput,
	topLimit int,
) (SalesReportData, error) {
	data, err := s.store.GetSalesReportData(ctx, sqlite.ReportingPeriodFilter{
		FromOccurredOn: current.FromOccurredOn.String(),
		ToOccurredOn:   current.ToOccurredOn.String(),
		Granularity:    string(current.Granularity),
	}, sqlite.ReportingPeriodFilter{
		FromOccurredOn: previous.FromOccurredOn.String(),
		ToOccurredOn:   previous.ToOccurredOn.String(),
		Granularity:    string(previous.Granularity),
	}, topLimit)
	if err != nil {
		return SalesReportData{}, err
	}
	return SalesReportData{
		Currency:              data.Currency,
		CurrentTotals:         mapSalesReportTotals(data.CurrentTotals),
		PreviousTotals:        mapSalesReportTotals(data.PreviousTotals),
		SalesRevenueSeries:    mapReportingSeries(data.SalesRevenueSeries),
		MonthlySeries:         mapReportingSeries(data.MonthlySeries),
		TopProductsByQuantity: mapReportingItemMetrics(data.TopProductsByQuantity),
		TopProductsByRevenue:  mapReportingItemMetrics(data.TopProductsByRevenue),
		FreeSales:             mapReportingReasonMetric(data.FreeSales),
		SalesByCustomer:       mapReportingCounterpartyMetrics(data.SalesByCustomer),
		AnonymousSales:        mapReportingCounterpartyMetric(data.AnonymousSales),
	}, nil
}

func mapSalesReportTotals(value sqlite.SalesReportTotals) SalesReportTotals {
	return SalesReportTotals{
		SalesCount:     value.SalesCount,
		QuantityAtomic: value.QuantityAtomic,
		RevenueMinor:   value.RevenueMinor,
		COGSMicro:      value.COGSMicro,
	}
}

func mapReportingSeries(items []sqlite.ReportingSeries) []ReportingSeries {
	mapped := make([]ReportingSeries, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, ReportingSeries{
			Bucket:         item.Bucket,
			Label:          item.Label,
			SalesCount:     item.SalesCount,
			QuantityAtomic: item.QuantityAtomic,
			RevenueMinor:   item.RevenueMinor,
			COGSMicro:      item.COGSMicro,
		})
	}
	return mapped
}

func mapReportingItemMetrics(items []sqlite.ReportingItemMetric) []ReportingItemMetric {
	mapped := make([]ReportingItemMetric, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, ReportingItemMetric{
			ItemID:              item.ItemID,
			ItemName:            item.ItemName,
			BaseUnitCode:        item.BaseUnitCode,
			QuantityAtomic:      item.QuantityAtomic,
			RevenueMinor:        item.RevenueMinor,
			InventoryValueMicro: item.COGSMicro,
		})
	}
	return mapped
}

func mapReportingReasonMetric(item sqlite.ReportingReasonMetric) ReportingReasonMetric {
	return ReportingReasonMetric{
		ReasonCode:          item.ReasonCode,
		DocumentCount:       item.DocumentCount,
		QuantityAtomic:      item.QuantityAtomic,
		RevenueMinor:        item.RevenueMinor,
		InventoryValueMicro: item.COGSMicro,
	}
}

func mapReportingCounterpartyMetrics(items []sqlite.ReportingCounterpartyMetric) []ReportingCounterpartyMetric {
	mapped := make([]ReportingCounterpartyMetric, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapReportingCounterpartyMetric(item))
	}
	return mapped
}

func mapReportingCounterpartyMetric(item sqlite.ReportingCounterpartyMetric) ReportingCounterpartyMetric {
	return ReportingCounterpartyMetric{
		CounterpartyID:   item.CounterpartyID,
		CounterpartyName: item.CounterpartyName,
		DocumentCount:    item.DocumentCount,
		RevenueMinor:     item.RevenueMinor,
		SpendMinor:       item.SpendMinor,
	}
}
