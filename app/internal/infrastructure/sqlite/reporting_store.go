package sqlite

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

type ReportingPeriodFilter struct {
	FromOccurredOn string
	ToOccurredOn   string
	Granularity    string
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

type ReportingSeries struct {
	Bucket         string
	Label          string
	SalesCount     int64
	QuantityAtomic int64
	RevenueMinor   int64
	COGSMicro      int64
}

type ReportingItemMetric struct {
	ItemID         domain.Option[domain.ItemID]
	ItemName       string
	BaseUnitCode   domain.Option[domain.UnitCode]
	QuantityAtomic int64
	RevenueMinor   int64
	COGSMicro      int64
}

type ReportingCounterpartyMetric struct {
	CounterpartyID   domain.Option[domain.CounterpartyID]
	CounterpartyName domain.Option[string]
	DocumentCount    int64
	RevenueMinor     int64
	SpendMinor       int64
}

type ReportingReasonMetric struct {
	ReasonCode     string
	DocumentCount  int64
	QuantityAtomic int64
	RevenueMinor   int64
	COGSMicro      int64
}

func (s *Store) GetSalesReportData(
	ctx context.Context,
	current ReportingPeriodFilter,
	previous ReportingPeriodFilter,
	topLimit int,
) (SalesReportData, error) {
	if topLimit <= 0 {
		topLimit = 5
	}
	var data SalesReportData
	err := s.withReadQueries(ctx, "get sales report data", func(queries *sqlcgen.Queries) error {
		currencyRow, err := queries.GetReportingCurrency(ctx)
		if err != nil {
			return err
		}
		currency, err := domain.RestoreCurrency(currencyRow.CurrencyCode, int(currencyRow.CurrencyMinorDigits))
		if err != nil {
			return err
		}

		currentTotals, err := queries.GetSalesReportTotals(ctx, salesTotalsParams(current))
		if err != nil {
			return err
		}
		previousTotals, err := queries.GetSalesReportTotals(ctx, salesTotalsParams(previous))
		if err != nil {
			return err
		}
		salesRevenueSeries, err := queries.ListSalesRevenueSeries(ctx, salesSeriesParams(current))
		if err != nil {
			return err
		}
		monthly := current
		monthly.Granularity = "MONTH"
		monthlySeries, err := queries.ListSalesRevenueSeries(ctx, salesSeriesParams(monthly))
		if err != nil {
			return err
		}
		topByQuantity, err := queries.ListTopSalesProductsByQuantity(ctx, topProductsByQuantityParams(current, topLimit))
		if err != nil {
			return err
		}
		topByRevenue, err := queries.ListTopSalesProductsByRevenue(ctx, topProductsByRevenueParams(current, topLimit))
		if err != nil {
			return err
		}
		freeSales, err := queries.GetFreeSalesTotals(ctx, freeSalesParams(current))
		if err != nil {
			return err
		}
		byCustomer, err := queries.ListSalesByCustomer(ctx, salesByCustomerParams(current, topLimit))
		if err != nil {
			return err
		}
		anonymous, err := queries.GetAnonymousSalesTotals(ctx, anonymousSalesParams(current))
		if err != nil {
			return err
		}

		data = SalesReportData{
			Currency:              currency,
			CurrentTotals:         mapSalesTotalsRow(currentTotals),
			PreviousTotals:        mapSalesTotalsRow(previousTotals),
			SalesRevenueSeries:    mapSalesSeriesRows(salesRevenueSeries),
			MonthlySeries:         mapSalesSeriesRows(monthlySeries),
			TopProductsByQuantity: mapTopProductsByQuantityRows(topByQuantity),
			TopProductsByRevenue:  mapTopProductsByRevenueRows(topByRevenue),
			FreeSales:             mapFreeSalesTotals(freeSales),
			SalesByCustomer:       mapSalesByCustomerRows(byCustomer),
			AnonymousSales:        mapAnonymousSalesTotals(anonymous),
		}
		return nil
	})
	if err != nil {
		return SalesReportData{}, err
	}
	return data, nil
}

func salesTotalsParams(filter ReportingPeriodFilter) sqlcgen.GetSalesReportTotalsParams {
	return sqlcgen.GetSalesReportTotalsParams{
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func salesSeriesParams(filter ReportingPeriodFilter) sqlcgen.ListSalesRevenueSeriesParams {
	return sqlcgen.ListSalesRevenueSeriesParams{
		Granularity:    filter.Granularity,
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func topProductsByQuantityParams(filter ReportingPeriodFilter, limit int) sqlcgen.ListTopSalesProductsByQuantityParams {
	return sqlcgen.ListTopSalesProductsByQuantityParams{
		LimitCount:     int64(limit),
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func topProductsByRevenueParams(filter ReportingPeriodFilter, limit int) sqlcgen.ListTopSalesProductsByRevenueParams {
	return sqlcgen.ListTopSalesProductsByRevenueParams{
		LimitCount:     int64(limit),
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func freeSalesParams(filter ReportingPeriodFilter) sqlcgen.GetFreeSalesTotalsParams {
	return sqlcgen.GetFreeSalesTotalsParams{
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func salesByCustomerParams(filter ReportingPeriodFilter, limit int) sqlcgen.ListSalesByCustomerParams {
	return sqlcgen.ListSalesByCustomerParams{
		LimitCount:     int64(limit),
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func anonymousSalesParams(filter ReportingPeriodFilter) sqlcgen.GetAnonymousSalesTotalsParams {
	return sqlcgen.GetAnonymousSalesTotalsParams{
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func mapSalesTotalsRow(row sqlcgen.GetSalesReportTotalsRow) SalesReportTotals {
	return SalesReportTotals{
		SalesCount:     row.SalesCount,
		QuantityAtomic: row.QuantityAtomic,
		RevenueMinor:   row.RevenueMinor,
		COGSMicro:      row.CogsMicro,
	}
}

func mapSalesSeriesRows(rows []sqlcgen.ListSalesRevenueSeriesRow) []ReportingSeries {
	items := make([]ReportingSeries, 0, len(rows))
	for _, row := range rows {
		items = append(items, ReportingSeries{
			Bucket:         row.Bucket,
			Label:          row.Label,
			SalesCount:     row.SalesCount,
			QuantityAtomic: row.QuantityAtomic,
			RevenueMinor:   row.RevenueMinor,
			COGSMicro:      row.CogsMicro,
		})
	}
	return items
}

func mapTopProductsByQuantityRows(rows []sqlcgen.ListTopSalesProductsByQuantityRow) []ReportingItemMetric {
	items := make([]ReportingItemMetric, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTopProduct(row.ItemID, row.ItemName, row.BaseUnitCode, row.QuantityAtomic, row.RevenueMinor, row.CogsMicro))
	}
	return items
}

func mapTopProductsByRevenueRows(rows []sqlcgen.ListTopSalesProductsByRevenueRow) []ReportingItemMetric {
	items := make([]ReportingItemMetric, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapTopProduct(row.ItemID, row.ItemName, row.BaseUnitCode, row.QuantityAtomic, row.RevenueMinor, row.CogsMicro))
	}
	return items
}

func mapTopProduct(itemIDValue int64, itemName string, baseUnitCodeValue string, quantityAtomic, revenueMinor, cogsMicro int64) ReportingItemMetric {
	itemID, itemIDErr := domain.NewItemID(itemIDValue)
	baseUnitCode, baseUnitErr := domain.NewUnitCode(baseUnitCodeValue)
	return ReportingItemMetric{
		ItemID:         optionWhenValid(itemID, itemIDErr),
		ItemName:       itemName,
		BaseUnitCode:   optionWhenValid(baseUnitCode, baseUnitErr),
		QuantityAtomic: quantityAtomic,
		RevenueMinor:   revenueMinor,
		COGSMicro:      cogsMicro,
	}
}

func mapFreeSalesTotals(row sqlcgen.GetFreeSalesTotalsRow) ReportingReasonMetric {
	return ReportingReasonMetric{
		ReasonCode:     "FREE_SALES",
		DocumentCount:  row.DocumentCount,
		QuantityAtomic: row.QuantityAtomic,
		RevenueMinor:   row.RevenueMinor,
		COGSMicro:      row.CogsMicro,
	}
}

func mapSalesByCustomerRows(rows []sqlcgen.ListSalesByCustomerRow) []ReportingCounterpartyMetric {
	items := make([]ReportingCounterpartyMetric, 0, len(rows))
	for _, row := range rows {
		counterpartyID := domain.None[domain.CounterpartyID]()
		if row.CounterpartyID.Valid {
			if id, err := domain.NewCounterpartyID(row.CounterpartyID.Int64); err == nil {
				counterpartyID = domain.Some(id)
			}
		}
		items = append(items, ReportingCounterpartyMetric{
			CounterpartyID:   counterpartyID,
			CounterpartyName: domain.Some(row.CounterpartyName),
			DocumentCount:    row.DocumentCount,
			RevenueMinor:     row.RevenueMinor,
		})
	}
	return items
}

func mapAnonymousSalesTotals(row sqlcgen.GetAnonymousSalesTotalsRow) ReportingCounterpartyMetric {
	return ReportingCounterpartyMetric{
		CounterpartyID:   domain.None[domain.CounterpartyID](),
		CounterpartyName: domain.None[string](),
		DocumentCount:    row.DocumentCount,
		RevenueMinor:     row.RevenueMinor,
	}
}

func optionWhenValid[T any](value T, err error) domain.Option[T] {
	if err != nil {
		return domain.None[T]()
	}
	return domain.Some(value)
}
