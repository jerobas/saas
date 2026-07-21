# Reporting and dashboard read models

Reporting is a read-only surface derived from posted V2 state. It never writes
business data and never becomes stock truth. Operational stores keep owning
documents, lines, lots, allocations, recipes, and `inventory_balances`.

## Common rules

- All endpoints receive an inclusive `fromOccurredOn` / `toOccurredOn` period
  unless documented otherwise.
- Document dates use `stock_documents.occurred_on`, not posting time.
- Revenue, purchase spend, and other commercial totals use minor currency units
  and are exposed as `commercialTotalMinor`. Average ticket uses
  `averageCommercialTotalMinor`.
- COGS, inventory valuation, production direct cost, and gross margin use
  microcurrency. Shared stock valuation uses `inventoryValueMicro`; specific
  reporting fields use explicit names such as `cogsInventoryValueMicro`,
  `directCostInventoryValueMicro`, and `grossMarginInventoryValueMicro`.
- Responses include `currencyCode` and `currencyMinorDigits` when monetary
  values are present.
- Operational dashboard aggregates exclude documents that have been exactly
  reversed. Reversal documents are reported separately for audit/correction
  visibility.
- Empty databases return zero values and empty series/tables, not errors.

## Reversal policy

Operational reporting treats an exact reversal as a correction, not as a new
business event.

- Documents that were exactly reversed are excluded from visible sales,
  purchase, production, and adjustment aggregates.
- `REVERSAL` documents are also excluded from those normal business aggregates.
- `GetAdjustmentReport.ExactReversals` is the current operational summary for
  reversal volume/value. Detailed reversal audit can become a separate report
  later.
- Inventory reporting reads current lot/balance projections, so it naturally
  reflects reversal effects after the projection is updated.

## Endpoint surface

The dashboard composes the domain-specific endpoints below instead of using a
separate aggregate endpoint. This keeps each read model small and avoids a
second contract that would need to mirror sales, inventory, purchase,
production, adjustment, and category placeholder data.

### `GetSalesReport`

Sales-focused endpoint for dashboard/report tabs.

Fields:

- total sales count;
- commercial total/revenue;
- COGS inventory value;
- gross margin inventory value;
- gross margin percentage;
- average ticket;
- growth versus previous period;
- sales and revenue series by day/month;
- monthly revenue;
- monthly sales count;
- top products by quantity sold;
- top products by revenue;
- free sales/promotions/samples count and commercial-zero totals;
- sales by customer;
- anonymous sales.

### `GetInventoryReport`

Current stock and lot-risk endpoint.

Fields:

- total inventory value;
- low-stock item count;
- sellable zero-stock item count;
- low-stock items with current balance and reorder point;
- lots expiring in 7 days;
- lots expiring in 30 days;
- expired lots that still have remaining quantity;
- inventory value by item.

### `GetPurchaseReport`

Inbound/commercial purchasing endpoint.

Fields:

- purchase commercial total by period;
- top suppliers by commercial total;
- free-stock inbound entries using reason `FREE_STOCK`.

### `GetProductionReport`

Production and costing endpoint.

Fields:

- production quantity by recipe/product;
- direct production cost inventory value by period;
- simple yield variance: actual output versus recipe standard yield.

### `GetAdjustmentReport`

Operational quality and correction endpoint.

Fields:

- negative adjustments by reason, including waste, expiry, damage, sample, and
  documented correction;
- positive adjustments by reason, including opening balance, free stock, and
  physical count;
- exactly reversed documents/corrections by period.

### `GetCategoryMixReport`

Placeholder endpoint for the existing pie chart. V2 has no catalog category/tag
dimension yet, so this endpoint returns an explicit unavailable/empty response
until a real category dimension exists.

Fields:

- `available: false`;
- empty category rows;
- reason explaining that catalog categories/tags are not modeled yet.
