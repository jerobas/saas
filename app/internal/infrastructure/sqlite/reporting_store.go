package sqlite

import (
	"context"
	"database/sql"

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

type SalesReportTotals struct {
	SalesCount     int64
	QuantityAtomic int64
	RevenueMinor   int64
	COGSMicro      int64
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
	COGSMicro             int64
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

func (s *Store) GetInventoryReportData(
	ctx context.Context,
	filter ReportingPeriodFilter,
	rowLimit int,
) (InventoryReportData, error) {
	if rowLimit <= 0 {
		rowLimit = 10
	}
	var data InventoryReportData
	err := s.withReadQueries(ctx, "get inventory report data", func(queries *sqlcgen.Queries) error {
		currencyRow, err := queries.GetReportingCurrency(ctx)
		if err != nil {
			return err
		}
		currency, err := domain.RestoreCurrency(currencyRow.CurrencyCode, int(currencyRow.CurrencyMinorDigits))
		if err != nil {
			return err
		}
		totals, err := queries.GetInventoryReportTotals(ctx)
		if err != nil {
			return err
		}
		lowStock, err := queries.ListLowStockItems(ctx, int64(rowLimit))
		if err != nil {
			return err
		}
		valueByItem, err := queries.ListInventoryValueByItem(ctx, int64(rowLimit))
		if err != nil {
			return err
		}
		expiring7, err := queries.ListExpiringLots(ctx, sqlcgen.ListExpiringLotsParams{
			ReferenceDate: filter.ToOccurredOn,
			DaysAhead:     7,
			LimitCount:    int64(rowLimit),
		})
		if err != nil {
			return err
		}
		expiring30, err := queries.ListExpiringLots(ctx, sqlcgen.ListExpiringLotsParams{
			ReferenceDate: filter.ToOccurredOn,
			DaysAhead:     30,
			LimitCount:    int64(rowLimit),
		})
		if err != nil {
			return err
		}
		expired, err := queries.ListExpiredLotsWithStock(ctx, sqlcgen.ListExpiredLotsWithStockParams{
			ReferenceDate: filter.ToOccurredOn,
			LimitCount:    int64(rowLimit),
		})
		if err != nil {
			return err
		}

		data = InventoryReportData{
			Currency:                 currency,
			TotalInventoryValueMicro: totals.TotalInventoryValueMicro,
			LowStockItemCount:        totals.LowStockItemCount,
			ZeroStockSellableCount:   totals.ZeroStockSellableCount,
			LowStockItems:            mapLowStockRows(lowStock),
			ExpiringLots7Days:        mapExpiringLotRows(expiring7),
			ExpiringLots30Days:       mapExpiringLotRows(expiring30),
			ExpiredLotsWithStock:     mapExpiredLotRows(expired),
			InventoryValueByItem:     mapInventoryValueRows(valueByItem),
		}
		return nil
	})
	if err != nil {
		return InventoryReportData{}, err
	}
	return data, nil
}

func (s *Store) GetPurchaseReportData(
	ctx context.Context,
	filter ReportingPeriodFilter,
	rowLimit int,
) (PurchaseReportData, error) {
	if rowLimit <= 0 {
		rowLimit = 10
	}
	var data PurchaseReportData
	err := s.withReadQueries(ctx, "get purchase report data", func(queries *sqlcgen.Queries) error {
		currencyRow, err := queries.GetReportingCurrency(ctx)
		if err != nil {
			return err
		}
		currency, err := domain.RestoreCurrency(currencyRow.CurrencyCode, int(currencyRow.CurrencyMinorDigits))
		if err != nil {
			return err
		}
		spendSeries, err := queries.ListPurchaseSpendSeries(ctx, purchaseSeriesParams(filter))
		if err != nil {
			return err
		}
		topSuppliers, err := queries.ListTopSuppliersBySpend(ctx, topSuppliersBySpendParams(filter, rowLimit))
		if err != nil {
			return err
		}
		freeStock, err := queries.ListFreeStockEntrySeries(ctx, freeStockEntrySeriesParams(filter))
		if err != nil {
			return err
		}

		data = PurchaseReportData{
			Currency:             currency,
			PurchaseSpendSeries:  mapPurchaseSpendSeriesRows(spendSeries),
			TopSuppliersBySpend:  mapTopSuppliersBySpendRows(topSuppliers),
			FreeStockEntrySeries: mapFreeStockEntrySeriesRows(freeStock),
		}
		return nil
	})
	if err != nil {
		return PurchaseReportData{}, err
	}
	return data, nil
}

func (s *Store) GetProductionReportData(
	ctx context.Context,
	filter ReportingPeriodFilter,
	rowLimit int,
) (ProductionReportData, error) {
	if rowLimit <= 0 {
		rowLimit = 10
	}
	var data ProductionReportData
	err := s.withReadQueries(ctx, "get production report data", func(queries *sqlcgen.Queries) error {
		currencyRow, err := queries.GetReportingCurrency(ctx)
		if err != nil {
			return err
		}
		currency, err := domain.RestoreCurrency(currencyRow.CurrencyCode, int(currencyRow.CurrencyMinorDigits))
		if err != nil {
			return err
		}
		byProduct, err := queries.ListProductionByRecipeProduct(ctx, productionByRecipeProductParams(filter, rowLimit))
		if err != nil {
			return err
		}
		directCostSeries, err := queries.ListProductionDirectCostSeries(ctx, productionDirectCostSeriesParams(filter))
		if err != nil {
			return err
		}
		yieldVariance, err := queries.ListProductionYieldVariance(ctx, productionYieldVarianceParams(filter, rowLimit))
		if err != nil {
			return err
		}

		data = ProductionReportData{
			Currency:                  currency,
			ProductionByRecipeProduct: mapProductionByRecipeProductRows(byProduct),
			DirectCostSeries:          mapProductionDirectCostSeriesRows(directCostSeries),
			YieldVariance:             mapProductionYieldVarianceRows(yieldVariance),
		}
		return nil
	})
	if err != nil {
		return ProductionReportData{}, err
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

func purchaseSeriesParams(filter ReportingPeriodFilter) sqlcgen.ListPurchaseSpendSeriesParams {
	return sqlcgen.ListPurchaseSpendSeriesParams{
		Granularity:    filter.Granularity,
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func topSuppliersBySpendParams(filter ReportingPeriodFilter, limit int) sqlcgen.ListTopSuppliersBySpendParams {
	return sqlcgen.ListTopSuppliersBySpendParams{
		LimitCount:     int64(limit),
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func freeStockEntrySeriesParams(filter ReportingPeriodFilter) sqlcgen.ListFreeStockEntrySeriesParams {
	return sqlcgen.ListFreeStockEntrySeriesParams{
		Granularity:    filter.Granularity,
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func productionByRecipeProductParams(filter ReportingPeriodFilter, limit int) sqlcgen.ListProductionByRecipeProductParams {
	return sqlcgen.ListProductionByRecipeProductParams{
		LimitCount:     int64(limit),
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func productionDirectCostSeriesParams(filter ReportingPeriodFilter) sqlcgen.ListProductionDirectCostSeriesParams {
	return sqlcgen.ListProductionDirectCostSeriesParams{
		Granularity:    filter.Granularity,
		FromOccurredOn: filter.FromOccurredOn,
		ToOccurredOn:   filter.ToOccurredOn,
	}
}

func productionYieldVarianceParams(filter ReportingPeriodFilter, limit int) sqlcgen.ListProductionYieldVarianceParams {
	return sqlcgen.ListProductionYieldVarianceParams{
		LimitCount:     int64(limit),
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
			Bucket:              row.Bucket,
			Label:               row.Label,
			DocumentCount:       row.SalesCount,
			SalesCount:          row.SalesCount,
			QuantityAtomic:      row.QuantityAtomic,
			RevenueMinor:        row.RevenueMinor,
			InventoryValueMicro: row.CogsMicro,
			COGSMicro:           row.CogsMicro,
		})
	}
	return items
}

func mapPurchaseSpendSeriesRows(rows []sqlcgen.ListPurchaseSpendSeriesRow) []ReportingSeries {
	items := make([]ReportingSeries, 0, len(rows))
	for _, row := range rows {
		items = append(items, ReportingSeries{
			Bucket:              row.Bucket,
			Label:               row.Label,
			DocumentCount:       row.DocumentCount,
			QuantityAtomic:      row.QuantityAtomic,
			RevenueMinor:        row.SpendMinor,
			SpendMinor:          row.SpendMinor,
			InventoryValueMicro: row.InventoryValueMicro,
		})
	}
	return items
}

func mapFreeStockEntrySeriesRows(rows []sqlcgen.ListFreeStockEntrySeriesRow) []ReportingSeries {
	items := make([]ReportingSeries, 0, len(rows))
	for _, row := range rows {
		items = append(items, ReportingSeries{
			Bucket:              row.Bucket,
			Label:               row.Label,
			DocumentCount:       row.DocumentCount,
			QuantityAtomic:      row.QuantityAtomic,
			RevenueMinor:        row.SpendMinor,
			SpendMinor:          row.SpendMinor,
			InventoryValueMicro: row.InventoryValueMicro,
		})
	}
	return items
}

func mapProductionDirectCostSeriesRows(rows []sqlcgen.ListProductionDirectCostSeriesRow) []ReportingSeries {
	items := make([]ReportingSeries, 0, len(rows))
	for _, row := range rows {
		items = append(items, ReportingSeries{
			Bucket:              row.Bucket,
			Label:               row.Label,
			DocumentCount:       row.DocumentCount,
			QuantityAtomic:      row.QuantityAtomic,
			InventoryValueMicro: row.InventoryValueMicro,
			DirectCostMicro:     row.DirectCostMicro,
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

func mapLowStockRows(rows []sqlcgen.ListLowStockItemsRow) []ReportingItemMetric {
	items := make([]ReportingItemMetric, 0, len(rows))
	for _, row := range rows {
		item := mapInventoryItem(row.ItemID, row.ItemName, row.BaseUnitCode, row.QuantityAtomic, row.InventoryValueMicro)
		if row.ReorderQuantityAtomic.Valid {
			item.ReorderQuantityAtomic = domain.Some(row.ReorderQuantityAtomic.Int64)
		}
		items = append(items, item)
	}
	return items
}

func mapInventoryValueRows(rows []sqlcgen.ListInventoryValueByItemRow) []ReportingItemMetric {
	items := make([]ReportingItemMetric, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapInventoryItem(row.ItemID, row.ItemName, row.BaseUnitCode, row.QuantityAtomic, row.InventoryValueMicro))
	}
	return items
}

func mapInventoryItem(itemIDValue int64, itemName string, baseUnitCodeValue string, quantityAtomic, inventoryValueMicro int64) ReportingItemMetric {
	itemID, itemIDErr := domain.NewItemID(itemIDValue)
	baseUnitCode, baseUnitErr := domain.NewUnitCode(baseUnitCodeValue)
	return ReportingItemMetric{
		ItemID:              optionWhenValid(itemID, itemIDErr),
		ItemName:            itemName,
		BaseUnitCode:        optionWhenValid(baseUnitCode, baseUnitErr),
		QuantityAtomic:      quantityAtomic,
		InventoryValueMicro: inventoryValueMicro,
	}
}

func mapProductionByRecipeProductRows(rows []sqlcgen.ListProductionByRecipeProductRow) []ReportingItemMetric {
	items := make([]ReportingItemMetric, 0, len(rows))
	for _, row := range rows {
		item := mapProductionItem(
			row.RecipeID,
			row.RecipeName,
			row.ItemID,
			row.ItemName,
			row.BaseUnitCode,
			row.DocumentCount,
			row.ActualYieldAtomic,
			row.InventoryValueMicro,
			row.DirectCostMicro,
		)
		item.StandardYieldAtomic = domain.Some(row.StandardYieldAtomic)
		item.ActualYieldAtomic = domain.Some(row.ActualYieldAtomic)
		item.VarianceAtomic = domain.Some(row.VarianceAtomic)
		items = append(items, item)
	}
	return items
}

func mapProductionYieldVarianceRows(rows []sqlcgen.ListProductionYieldVarianceRow) []ReportingItemMetric {
	items := make([]ReportingItemMetric, 0, len(rows))
	for _, row := range rows {
		item := mapProductionItem(
			row.RecipeID,
			row.RecipeName,
			row.ItemID,
			row.ItemName,
			row.BaseUnitCode,
			row.DocumentCount,
			row.ActualYieldAtomic,
			row.InventoryValueMicro,
			row.DirectCostMicro,
		)
		item.StandardYieldAtomic = domain.Some(row.StandardYieldAtomic)
		item.ActualYieldAtomic = domain.Some(row.ActualYieldAtomic)
		item.VarianceAtomic = domain.Some(row.VarianceAtomic)
		items = append(items, item)
	}
	return items
}

func mapProductionItem(
	recipeIDValue int64,
	recipeName string,
	itemIDValue int64,
	itemName string,
	baseUnitCodeValue string,
	documentCount int64,
	actualYieldAtomic int64,
	inventoryValueMicro int64,
	directCostMicro int64,
) ReportingItemMetric {
	recipeID, recipeIDErr := domain.NewRecipeID(recipeIDValue)
	itemID, itemIDErr := domain.NewItemID(itemIDValue)
	baseUnitCode, baseUnitErr := domain.NewUnitCode(baseUnitCodeValue)
	return ReportingItemMetric{
		RecipeID:            optionWhenValid(recipeID, recipeIDErr),
		RecipeName:          domain.Some(recipeName),
		ItemID:              optionWhenValid(itemID, itemIDErr),
		ItemName:            itemName,
		BaseUnitCode:        optionWhenValid(baseUnitCode, baseUnitErr),
		DocumentCount:       documentCount,
		QuantityAtomic:      actualYieldAtomic,
		InventoryValueMicro: inventoryValueMicro,
		DirectCostMicro:     directCostMicro,
	}
}

func mapExpiringLotRows(rows []sqlcgen.ListExpiringLotsRow) []ReportingLotMetric {
	items := make([]ReportingLotMetric, 0, len(rows))
	for _, row := range rows {
		if item, ok := mapReportingLot(row.LotID, row.ItemID, row.ItemName, row.LotCode, row.ExpiresOn, row.AvailableQuantityAtomic, row.InventoryValueMicro).Get(); ok {
			items = append(items, item)
		}
	}
	return items
}

func mapExpiredLotRows(rows []sqlcgen.ListExpiredLotsWithStockRow) []ReportingLotMetric {
	items := make([]ReportingLotMetric, 0, len(rows))
	for _, row := range rows {
		if item, ok := mapReportingLot(row.LotID, row.ItemID, row.ItemName, row.LotCode, row.ExpiresOn, row.AvailableQuantityAtomic, row.InventoryValueMicro).Get(); ok {
			items = append(items, item)
		}
	}
	return items
}

func mapReportingLot(lotIDValue, itemIDValue int64, itemName string, lotCodeValue, expiresOnValue sql.NullString, availableQuantity, inventoryValueMicro int64) domain.Option[ReportingLotMetric] {
	lotID, err := domain.NewInventoryLotID(lotIDValue)
	if err != nil {
		return domain.None[ReportingLotMetric]()
	}
	itemID, err := domain.NewItemID(itemIDValue)
	if err != nil {
		return domain.None[ReportingLotMetric]()
	}
	return domain.Some(ReportingLotMetric{
		LotID:               lotID,
		ItemID:              itemID,
		ItemName:            itemName,
		LotCode:             optionSQLString(lotCodeValue),
		ExpiresOn:           optionBusinessDate(expiresOnValue),
		AvailableQuantity:   availableQuantity,
		InventoryValueMicro: inventoryValueMicro,
	})
}

func optionSQLString(value sql.NullString) domain.Option[string] {
	if !value.Valid {
		return domain.None[string]()
	}
	return domain.Some(value.String)
}

func optionBusinessDate(value sql.NullString) domain.Option[domain.BusinessDate] {
	if !value.Valid {
		return domain.None[domain.BusinessDate]()
	}
	date, err := domain.ParseBusinessDate(value.String)
	if err != nil {
		return domain.None[domain.BusinessDate]()
	}
	return domain.Some(date)
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

func mapTopSuppliersBySpendRows(rows []sqlcgen.ListTopSuppliersBySpendRow) []ReportingCounterpartyMetric {
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
			SpendMinor:       row.SpendMinor,
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
