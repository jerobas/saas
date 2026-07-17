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

func (s *sqliteReportingStore) GetInventoryReportData(
	ctx context.Context,
	input ReportingPeriodInput,
	rowLimit int,
) (InventoryReportData, error) {
	data, err := s.store.GetInventoryReportData(ctx, sqlite.ReportingPeriodFilter{
		FromOccurredOn: input.FromOccurredOn.String(),
		ToOccurredOn:   input.ToOccurredOn.String(),
		Granularity:    string(input.Granularity),
	}, rowLimit)
	if err != nil {
		return InventoryReportData{}, err
	}
	return InventoryReportData{
		Currency:                 data.Currency,
		TotalInventoryValueMicro: data.TotalInventoryValueMicro,
		LowStockItemCount:        data.LowStockItemCount,
		ZeroStockSellableCount:   data.ZeroStockSellableCount,
		LowStockItems:            mapReportingItemMetrics(data.LowStockItems),
		ExpiringLots7Days:        mapReportingLotMetrics(data.ExpiringLots7Days),
		ExpiringLots30Days:       mapReportingLotMetrics(data.ExpiringLots30Days),
		ExpiredLotsWithStock:     mapReportingLotMetrics(data.ExpiredLotsWithStock),
		InventoryValueByItem:     mapReportingItemMetrics(data.InventoryValueByItem),
	}, nil
}

func (s *sqliteReportingStore) GetPurchaseReportData(
	ctx context.Context,
	input ReportingPeriodInput,
	rowLimit int,
) (PurchaseReportData, error) {
	data, err := s.store.GetPurchaseReportData(ctx, sqlite.ReportingPeriodFilter{
		FromOccurredOn: input.FromOccurredOn.String(),
		ToOccurredOn:   input.ToOccurredOn.String(),
		Granularity:    string(input.Granularity),
	}, rowLimit)
	if err != nil {
		return PurchaseReportData{}, err
	}
	return PurchaseReportData{
		Currency:             data.Currency,
		PurchaseSpendSeries:  mapReportingSeries(data.PurchaseSpendSeries),
		TopSuppliersBySpend:  mapReportingCounterpartyMetrics(data.TopSuppliersBySpend),
		FreeStockEntrySeries: mapReportingSeries(data.FreeStockEntrySeries),
	}, nil
}

func (s *sqliteReportingStore) GetProductionReportData(
	ctx context.Context,
	input ReportingPeriodInput,
	rowLimit int,
) (ProductionReportData, error) {
	data, err := s.store.GetProductionReportData(ctx, sqlite.ReportingPeriodFilter{
		FromOccurredOn: input.FromOccurredOn.String(),
		ToOccurredOn:   input.ToOccurredOn.String(),
		Granularity:    string(input.Granularity),
	}, rowLimit)
	if err != nil {
		return ProductionReportData{}, err
	}
	return ProductionReportData{
		Currency:                  data.Currency,
		ProductionByRecipeProduct: mapReportingItemMetrics(data.ProductionByRecipeProduct),
		DirectCostSeries:          mapReportingSeries(data.DirectCostSeries),
		YieldVariance:             mapReportingItemMetrics(data.YieldVariance),
	}, nil
}

func (s *sqliteReportingStore) GetAdjustmentReportData(
	ctx context.Context,
	input ReportingPeriodInput,
) (AdjustmentReportData, error) {
	data, err := s.store.GetAdjustmentReportData(ctx, sqlite.ReportingPeriodFilter{
		FromOccurredOn: input.FromOccurredOn.String(),
		ToOccurredOn:   input.ToOccurredOn.String(),
		Granularity:    string(input.Granularity),
	})
	if err != nil {
		return AdjustmentReportData{}, err
	}
	return AdjustmentReportData{
		Currency:         data.Currency,
		NegativeByReason: mapReportingReasonMetrics(data.NegativeByReason),
		PositiveByReason: mapReportingReasonMetrics(data.PositiveByReason),
		ExactReversals:   mapReportingSeries(data.ExactReversals),
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
			Bucket:              item.Bucket,
			Label:               item.Label,
			DocumentCount:       item.DocumentCount,
			SalesCount:          item.SalesCount,
			QuantityAtomic:      item.QuantityAtomic,
			RevenueMinor:        item.RevenueMinor,
			SpendMinor:          item.SpendMinor,
			InventoryValueMicro: item.InventoryValueMicro,
			DirectCostMicro:     item.DirectCostMicro,
			COGSMicro:           item.COGSMicro,
		})
	}
	return mapped
}

func mapReportingItemMetrics(items []sqlite.ReportingItemMetric) []ReportingItemMetric {
	mapped := make([]ReportingItemMetric, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, ReportingItemMetric{
			ItemID:                item.ItemID,
			ItemName:              item.ItemName,
			RecipeID:              item.RecipeID,
			RecipeName:            item.RecipeName,
			BaseUnitCode:          item.BaseUnitCode,
			DocumentCount:         item.DocumentCount,
			QuantityAtomic:        item.QuantityAtomic,
			RevenueMinor:          item.RevenueMinor,
			InventoryValueMicro:   item.InventoryValueMicro,
			DirectCostMicro:       reportingDirectCostMicro(item),
			ReorderQuantityAtomic: item.ReorderQuantityAtomic,
			StandardYieldAtomic:   item.StandardYieldAtomic,
			ActualYieldAtomic:     item.ActualYieldAtomic,
			VarianceAtomic:        item.VarianceAtomic,
		})
	}
	return mapped
}

func reportingDirectCostMicro(item sqlite.ReportingItemMetric) int64 {
	if item.DirectCostMicro != 0 {
		return item.DirectCostMicro
	}
	return item.COGSMicro
}

func mapReportingLotMetrics(items []sqlite.ReportingLotMetric) []ReportingLotMetric {
	mapped := make([]ReportingLotMetric, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, ReportingLotMetric{
			LotID:               item.LotID,
			ItemID:              item.ItemID,
			ItemName:            item.ItemName,
			LotCode:             item.LotCode,
			ExpiresOn:           item.ExpiresOn,
			AvailableQuantity:   item.AvailableQuantity,
			InventoryValueMicro: item.InventoryValueMicro,
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
		InventoryValueMicro: reportingReasonInventoryValueMicro(item),
	}
}

func mapReportingReasonMetrics(items []sqlite.ReportingReasonMetric) []ReportingReasonMetric {
	mapped := make([]ReportingReasonMetric, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapReportingReasonMetric(item))
	}
	return mapped
}

func reportingReasonInventoryValueMicro(item sqlite.ReportingReasonMetric) int64 {
	if item.InventoryValueMicro != 0 {
		return item.InventoryValueMicro
	}
	return item.COGSMicro
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
