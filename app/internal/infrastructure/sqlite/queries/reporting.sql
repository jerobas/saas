-- name: GetReportingCurrency :one
SELECT currency_code, currency_minor_digits
FROM app_settings
WHERE id = 1;

-- name: GetSalesReportTotals :one
WITH active_sale_lines AS (
    SELECT
        document.id AS document_id,
        line.quantity_atomic,
        line.commercial_total_minor,
        line.inventory_value_micro
    FROM stock_documents document
    JOIN stock_document_lines line ON line.document_id = document.id
    WHERE document.kind = 'SALE'
      AND document.occurred_on >= CAST(sqlc.arg(from_occurred_on) AS TEXT)
      AND document.occurred_on <= CAST(sqlc.arg(to_occurred_on) AS TEXT)
      AND NOT EXISTS (
          SELECT 1
          FROM stock_documents reversal
          WHERE reversal.kind = 'REVERSAL'
            AND reversal.reverses_document_id = document.id
      )
)
SELECT
    CAST(COUNT(DISTINCT document_id) AS INTEGER) AS sales_count,
    CAST(COALESCE(SUM(quantity_atomic), 0) AS INTEGER) AS quantity_atomic,
    CAST(COALESCE(SUM(commercial_total_minor), 0) AS INTEGER) AS revenue_minor,
    CAST(COALESCE(SUM(inventory_value_micro), 0) AS INTEGER) AS cogs_micro
FROM active_sale_lines;

-- name: ListSalesRevenueSeries :many
WITH active_sale_lines AS (
    SELECT
        document.id AS document_id,
        CAST(
            CASE
                WHEN CAST(sqlc.arg(granularity) AS TEXT) = 'DAY'
                    THEN document.occurred_on
                ELSE substr(document.occurred_on, 1, 7)
            END AS TEXT
        ) AS bucket,
        line.quantity_atomic,
        line.commercial_total_minor,
        line.inventory_value_micro
    FROM stock_documents document
    JOIN stock_document_lines line ON line.document_id = document.id
    WHERE document.kind = 'SALE'
      AND document.occurred_on >= CAST(sqlc.arg(from_occurred_on) AS TEXT)
      AND document.occurred_on <= CAST(sqlc.arg(to_occurred_on) AS TEXT)
      AND NOT EXISTS (
          SELECT 1
          FROM stock_documents reversal
          WHERE reversal.kind = 'REVERSAL'
            AND reversal.reverses_document_id = document.id
      )
)
SELECT
    CAST(bucket AS TEXT) AS bucket,
    CAST(bucket AS TEXT) AS label,
    CAST(COUNT(DISTINCT document_id) AS INTEGER) AS sales_count,
    CAST(COALESCE(SUM(quantity_atomic), 0) AS INTEGER) AS quantity_atomic,
    CAST(COALESCE(SUM(commercial_total_minor), 0) AS INTEGER) AS revenue_minor,
    CAST(COALESCE(SUM(inventory_value_micro), 0) AS INTEGER) AS cogs_micro
FROM active_sale_lines
GROUP BY bucket
ORDER BY bucket;

-- name: ListTopSalesProductsByQuantity :many
WITH active_sale_lines AS (
    SELECT
        line.item_id,
        item.name AS item_name,
        item.base_unit_code,
        line.quantity_atomic,
        line.commercial_total_minor,
        line.inventory_value_micro
    FROM stock_documents document
    JOIN stock_document_lines line ON line.document_id = document.id
    JOIN items item ON item.id = line.item_id
    WHERE document.kind = 'SALE'
      AND document.occurred_on >= CAST(sqlc.arg(from_occurred_on) AS TEXT)
      AND document.occurred_on <= CAST(sqlc.arg(to_occurred_on) AS TEXT)
      AND NOT EXISTS (
          SELECT 1
          FROM stock_documents reversal
          WHERE reversal.kind = 'REVERSAL'
            AND reversal.reverses_document_id = document.id
      )
)
SELECT
    item_id,
    item_name,
    base_unit_code,
    CAST(COALESCE(SUM(quantity_atomic), 0) AS INTEGER) AS quantity_atomic,
    CAST(COALESCE(SUM(commercial_total_minor), 0) AS INTEGER) AS revenue_minor,
    CAST(COALESCE(SUM(inventory_value_micro), 0) AS INTEGER) AS cogs_micro
FROM active_sale_lines
GROUP BY item_id, item_name, base_unit_code
ORDER BY quantity_atomic DESC, revenue_minor DESC, item_name, item_id
LIMIT sqlc.arg(limit_count);

-- name: ListTopSalesProductsByRevenue :many
WITH active_sale_lines AS (
    SELECT
        line.item_id,
        item.name AS item_name,
        item.base_unit_code,
        line.quantity_atomic,
        line.commercial_total_minor,
        line.inventory_value_micro
    FROM stock_documents document
    JOIN stock_document_lines line ON line.document_id = document.id
    JOIN items item ON item.id = line.item_id
    WHERE document.kind = 'SALE'
      AND document.occurred_on >= CAST(sqlc.arg(from_occurred_on) AS TEXT)
      AND document.occurred_on <= CAST(sqlc.arg(to_occurred_on) AS TEXT)
      AND NOT EXISTS (
          SELECT 1
          FROM stock_documents reversal
          WHERE reversal.kind = 'REVERSAL'
            AND reversal.reverses_document_id = document.id
      )
)
SELECT
    item_id,
    item_name,
    base_unit_code,
    CAST(COALESCE(SUM(quantity_atomic), 0) AS INTEGER) AS quantity_atomic,
    CAST(COALESCE(SUM(commercial_total_minor), 0) AS INTEGER) AS revenue_minor,
    CAST(COALESCE(SUM(inventory_value_micro), 0) AS INTEGER) AS cogs_micro
FROM active_sale_lines
GROUP BY item_id, item_name, base_unit_code
ORDER BY revenue_minor DESC, quantity_atomic DESC, item_name, item_id
LIMIT sqlc.arg(limit_count);

-- name: GetFreeSalesTotals :one
WITH free_sale_lines AS (
    SELECT
        document.id AS document_id,
        line.quantity_atomic,
        line.commercial_total_minor,
        line.inventory_value_micro
    FROM stock_documents document
    JOIN stock_document_lines line ON line.document_id = document.id
    WHERE document.kind = 'SALE'
      AND document.reason_code IN ('PROMOTION', 'SAMPLE')
      AND line.commercial_total_minor = 0
      AND document.occurred_on >= CAST(sqlc.arg(from_occurred_on) AS TEXT)
      AND document.occurred_on <= CAST(sqlc.arg(to_occurred_on) AS TEXT)
      AND NOT EXISTS (
          SELECT 1
          FROM stock_documents reversal
          WHERE reversal.kind = 'REVERSAL'
            AND reversal.reverses_document_id = document.id
      )
)
SELECT
    CAST(COUNT(DISTINCT document_id) AS INTEGER) AS document_count,
    CAST(COALESCE(SUM(quantity_atomic), 0) AS INTEGER) AS quantity_atomic,
    CAST(COALESCE(SUM(commercial_total_minor), 0) AS INTEGER) AS revenue_minor,
    CAST(COALESCE(SUM(inventory_value_micro), 0) AS INTEGER) AS cogs_micro
FROM free_sale_lines;

-- name: ListSalesByCustomer :many
WITH active_sale_lines AS (
    SELECT
        document.id AS document_id,
        document.counterparty_id,
        counterparty.name AS counterparty_name,
        line.commercial_total_minor
    FROM stock_documents document
    JOIN stock_document_lines line ON line.document_id = document.id
    JOIN counterparties counterparty ON counterparty.id = document.counterparty_id
    WHERE document.kind = 'SALE'
      AND document.counterparty_id IS NOT NULL
      AND document.occurred_on >= CAST(sqlc.arg(from_occurred_on) AS TEXT)
      AND document.occurred_on <= CAST(sqlc.arg(to_occurred_on) AS TEXT)
      AND NOT EXISTS (
          SELECT 1
          FROM stock_documents reversal
          WHERE reversal.kind = 'REVERSAL'
            AND reversal.reverses_document_id = document.id
      )
)
SELECT
    counterparty_id,
    counterparty_name,
    CAST(COUNT(DISTINCT document_id) AS INTEGER) AS document_count,
    CAST(COALESCE(SUM(commercial_total_minor), 0) AS INTEGER) AS revenue_minor
FROM active_sale_lines
GROUP BY counterparty_id, counterparty_name
ORDER BY revenue_minor DESC, document_count DESC, counterparty_name, counterparty_id
LIMIT sqlc.arg(limit_count);

-- name: GetAnonymousSalesTotals :one
WITH anonymous_sale_lines AS (
    SELECT
        document.id AS document_id,
        line.commercial_total_minor
    FROM stock_documents document
    JOIN stock_document_lines line ON line.document_id = document.id
    WHERE document.kind = 'SALE'
      AND document.counterparty_id IS NULL
      AND document.occurred_on >= CAST(sqlc.arg(from_occurred_on) AS TEXT)
      AND document.occurred_on <= CAST(sqlc.arg(to_occurred_on) AS TEXT)
      AND NOT EXISTS (
          SELECT 1
          FROM stock_documents reversal
          WHERE reversal.kind = 'REVERSAL'
            AND reversal.reverses_document_id = document.id
      )
)
SELECT
    CAST(COUNT(DISTINCT document_id) AS INTEGER) AS document_count,
    CAST(COALESCE(SUM(commercial_total_minor), 0) AS INTEGER) AS revenue_minor
FROM anonymous_sale_lines;
