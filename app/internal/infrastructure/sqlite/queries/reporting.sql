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

-- name: GetInventoryReportTotals :one
SELECT
    CAST(COALESCE(SUM(balance.inventory_value_micro), 0) AS INTEGER) AS total_inventory_value_micro,
    CAST(COALESCE(SUM(
        CASE
            WHEN item.archived_at_ms IS NULL
             AND item.reorder_quantity_atomic IS NOT NULL
             AND balance.quantity_atomic <= item.reorder_quantity_atomic
                THEN 1 ELSE 0
        END
    ), 0) AS INTEGER) AS low_stock_item_count,
    CAST(COALESCE(SUM(
        CASE
            WHEN item.archived_at_ms IS NULL
             AND item.is_sellable = 1
             AND balance.quantity_atomic = 0
                THEN 1 ELSE 0
        END
    ), 0) AS INTEGER) AS zero_stock_sellable_count
FROM inventory_balances balance
JOIN items item ON item.id = balance.item_id;

-- name: ListLowStockItems :many
SELECT
    item.id AS item_id,
    item.name AS item_name,
    item.base_unit_code,
    balance.quantity_atomic,
    balance.inventory_value_micro,
    item.reorder_quantity_atomic
FROM inventory_balances balance
JOIN items item ON item.id = balance.item_id
WHERE item.archived_at_ms IS NULL
  AND item.reorder_quantity_atomic IS NOT NULL
  AND balance.quantity_atomic <= item.reorder_quantity_atomic
ORDER BY (item.reorder_quantity_atomic - balance.quantity_atomic) DESC, item.name, item.id
LIMIT sqlc.arg(limit_count);

-- name: ListInventoryValueByItem :many
SELECT
    item.id AS item_id,
    item.name AS item_name,
    item.base_unit_code,
    balance.quantity_atomic,
    balance.inventory_value_micro
FROM inventory_balances balance
JOIN items item ON item.id = balance.item_id
WHERE item.archived_at_ms IS NULL
  AND balance.inventory_value_micro > 0
ORDER BY balance.inventory_value_micro DESC, item.name, item.id
LIMIT sqlc.arg(limit_count);

-- name: ListExpiringLots :many
WITH lot_facts AS (
    SELECT
        lot.id,
        lot.item_id,
        item.name AS item_name,
        lot.source_line_id,
        lot.initial_quantity_atomic,
        lot.lot_code,
        lot.expires_on,
        source_line.inventory_value_micro AS source_inventory_value_micro,
        source_line.quantity_atomic AS source_quantity_atomic,
        lot.initial_quantity_atomic
            - COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0)
            + COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NOT NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0) AS available_quantity_atomic
    FROM inventory_lots lot
    JOIN items item ON item.id = lot.item_id
    JOIN stock_document_lines source_line ON source_line.id = lot.source_line_id
    LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
    WHERE item.archived_at_ms IS NULL
      AND lot.expires_on IS NOT NULL
    GROUP BY lot.id
)
SELECT
    id AS lot_id,
    item_id,
    item_name,
    lot_code,
    expires_on,
    CAST(available_quantity_atomic AS INTEGER) AS available_quantity_atomic,
    CAST(
        (source_inventory_value_micro * available_quantity_atomic) / source_quantity_atomic
        AS INTEGER
    ) AS inventory_value_micro
FROM lot_facts
WHERE available_quantity_atomic > 0
  AND expires_on > CAST(sqlc.arg(reference_date) AS TEXT)
  AND expires_on <= date(
      CAST(sqlc.arg(reference_date) AS TEXT),
      '+' || CAST(sqlc.arg(days_ahead) AS INTEGER) || ' day'
  )
ORDER BY expires_on, item_name, id
LIMIT sqlc.arg(limit_count);

-- name: ListExpiredLotsWithStock :many
WITH lot_facts AS (
    SELECT
        lot.id,
        lot.item_id,
        item.name AS item_name,
        lot.source_line_id,
        lot.initial_quantity_atomic,
        lot.lot_code,
        lot.expires_on,
        source_line.inventory_value_micro AS source_inventory_value_micro,
        source_line.quantity_atomic AS source_quantity_atomic,
        lot.initial_quantity_atomic
            - COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0)
            + COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NOT NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0) AS available_quantity_atomic
    FROM inventory_lots lot
    JOIN items item ON item.id = lot.item_id
    JOIN stock_document_lines source_line ON source_line.id = lot.source_line_id
    LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
    WHERE item.archived_at_ms IS NULL
      AND lot.expires_on IS NOT NULL
    GROUP BY lot.id
)
SELECT
    id AS lot_id,
    item_id,
    item_name,
    lot_code,
    expires_on,
    CAST(available_quantity_atomic AS INTEGER) AS available_quantity_atomic,
    CAST(
        (source_inventory_value_micro * available_quantity_atomic) / source_quantity_atomic
        AS INTEGER
    ) AS inventory_value_micro
FROM lot_facts
WHERE available_quantity_atomic > 0
  AND expires_on <= CAST(sqlc.arg(reference_date) AS TEXT)
ORDER BY expires_on, item_name, id
LIMIT sqlc.arg(limit_count);
