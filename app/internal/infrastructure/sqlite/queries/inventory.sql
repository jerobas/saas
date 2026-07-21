-- name: GetInventoryBalance :one
SELECT
    balance.item_id,
    item.name AS item_name,
    item.normalized_name AS item_normalized_name,
    item.base_unit_code,
    item.archived_at_ms AS item_archived_at_ms,
    balance.quantity_atomic,
    balance.inventory_value_micro,
    balance.last_document_id,
    balance.updated_at_ms
FROM inventory_balances balance
JOIN items item ON item.id = balance.item_id
WHERE balance.item_id = sqlc.arg(item_id);

-- name: ListInventoryBalances :many
SELECT
    balance.item_id,
    item.name AS item_name,
    item.normalized_name AS item_normalized_name,
    item.base_unit_code,
    item.is_purchasable,
    item.is_producible,
    item.is_sellable,
    item.reorder_quantity_atomic,
    item.archived_at_ms AS item_archived_at_ms,
    balance.quantity_atomic,
    balance.inventory_value_micro,
    balance.last_document_id,
    balance.updated_at_ms
FROM inventory_balances balance
JOIN items item ON item.id = balance.item_id
WHERE
    (CAST(sqlc.arg(include_archived) AS INTEGER) = 1 OR item.archived_at_ms IS NULL)
    AND (
        CAST(sqlc.arg(search_key) AS TEXT) = ''
        OR instr(item.normalized_name, CAST(sqlc.arg(search_key) AS TEXT)) > 0
    )
    AND (
        CAST(sqlc.arg(after_normalized_name) AS TEXT) = ''
        OR item.normalized_name > CAST(sqlc.arg(after_normalized_name) AS TEXT)
        OR (
            item.normalized_name = CAST(sqlc.arg(after_normalized_name) AS TEXT)
            AND item.id > sqlc.arg(after_item_id)
        )
    )
ORDER BY item.normalized_name, item.id
LIMIT sqlc.arg(limit_count);

-- name: ListItemLotFacts :many
WITH lot_facts AS (
    SELECT
        lot.id,
        lot.item_id,
        lot.source_line_id,
        lot.initial_quantity_atomic,
        lot.lot_code,
        lot.originated_on,
        lot.expires_on,
        lot.created_at_ms,
        source_document.id AS source_document_id,
        source_document.kind AS source_document_kind,
        source_document.posting_sequence AS source_posting_sequence,
        source_document.occurred_on AS source_occurred_on,
        CAST(COALESCE(SUM(
            CASE WHEN allocation.restores_allocation_id IS NULL
                THEN allocation.quantity_atomic ELSE 0 END
        ), 0) AS INTEGER) AS consumed_quantity_atomic,
        CAST(COALESCE(SUM(
            CASE WHEN allocation.restores_allocation_id IS NOT NULL
                THEN allocation.quantity_atomic ELSE 0 END
        ), 0) AS INTEGER) AS restored_quantity_atomic,
        lot.initial_quantity_atomic
            - COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0)
            + COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NOT NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0) AS remaining_quantity_atomic
    FROM inventory_lots lot
    JOIN stock_document_lines source_line ON source_line.id = lot.source_line_id
    JOIN stock_documents source_document ON source_document.id = source_line.document_id
    LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
    WHERE lot.item_id = sqlc.arg(item_id)
    GROUP BY
        lot.id,
        lot.item_id,
        lot.source_line_id,
        lot.initial_quantity_atomic,
        lot.lot_code,
        lot.originated_on,
        lot.expires_on,
        lot.created_at_ms,
        source_document.id,
        source_document.kind,
        source_document.posting_sequence,
        source_document.occurred_on
)
SELECT
    id,
    item_id,
    source_line_id,
    initial_quantity_atomic,
    consumed_quantity_atomic,
    restored_quantity_atomic,
    remaining_quantity_atomic,
    lot_code,
    originated_on,
    expires_on,
    created_at_ms,
    source_document_id,
    source_document_kind,
    source_posting_sequence,
    source_occurred_on
FROM lot_facts
ORDER BY
    expires_on IS NULL,
    expires_on,
    source_posting_sequence,
    id;

-- name: ListEligibleFEFOLots :many
WITH lot_facts AS (
    SELECT
        lot.id,
        lot.item_id,
        lot.source_line_id,
        lot.initial_quantity_atomic,
        lot.lot_code,
        lot.originated_on,
        lot.expires_on,
        lot.created_at_ms,
        source_document.id AS source_document_id,
        source_document.kind AS source_document_kind,
        source_document.posting_sequence AS source_posting_sequence,
        source_document.occurred_on AS source_occurred_on,
        CAST(COALESCE(SUM(
            CASE WHEN allocation.restores_allocation_id IS NULL
                THEN allocation.quantity_atomic ELSE 0 END
        ), 0) AS INTEGER) AS consumed_quantity_atomic,
        CAST(COALESCE(SUM(
            CASE WHEN allocation.restores_allocation_id IS NOT NULL
                THEN allocation.quantity_atomic ELSE 0 END
        ), 0) AS INTEGER) AS restored_quantity_atomic,
        lot.initial_quantity_atomic
            - COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0)
            + COALESCE(SUM(
                CASE WHEN allocation.restores_allocation_id IS NOT NULL
                    THEN allocation.quantity_atomic ELSE 0 END
            ), 0) AS remaining_quantity_atomic
    FROM inventory_lots lot
    JOIN stock_document_lines source_line ON source_line.id = lot.source_line_id
    JOIN stock_documents source_document ON source_document.id = source_line.document_id
    LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
    WHERE lot.item_id = sqlc.arg(item_id)
    GROUP BY
        lot.id,
        lot.item_id,
        lot.source_line_id,
        lot.initial_quantity_atomic,
        lot.lot_code,
        lot.originated_on,
        lot.expires_on,
        lot.created_at_ms,
        source_document.id,
        source_document.kind,
        source_document.posting_sequence,
        source_document.occurred_on
)
SELECT
    id,
    item_id,
    source_line_id,
    initial_quantity_atomic,
    consumed_quantity_atomic,
    restored_quantity_atomic,
    remaining_quantity_atomic,
    lot_code,
    originated_on,
    expires_on,
    created_at_ms,
    source_document_id,
    source_document_kind,
    source_posting_sequence,
    source_occurred_on
FROM lot_facts
WHERE remaining_quantity_atomic > 0
  AND (expires_on IS NULL OR expires_on >= CAST(sqlc.arg(business_date) AS TEXT))
ORDER BY
    expires_on IS NULL,
    expires_on,
    source_posting_sequence,
    id;

-- name: ListItemLedgerPage :many
SELECT
    line.id AS line_id,
    line.document_id,
    line.line_order,
    line.item_id,
    line.direction,
    line.quantity_atomic,
    line.entered_unit_code,
    line.entered_packaging_name,
    line.conversion_numerator_atomic,
    line.conversion_denominator,
    line.inventory_value_micro,
    line.commercial_total_minor,
    line.reverses_line_id,
    document.kind AS document_kind,
    document.idempotency_key,
    document.posting_sequence,
    document.counterparty_id,
    counterparty.name AS counterparty_name,
    document.occurred_on,
    document.posted_at_ms,
    document.currency_code,
    document.currency_minor_digits,
    document.reason_code,
    document.notes,
    document.reverses_document_id
FROM stock_document_lines line
JOIN stock_documents document ON document.id = line.document_id
LEFT JOIN counterparties counterparty ON counterparty.id = document.counterparty_id
WHERE line.item_id = sqlc.arg(item_id)
  AND (
      CAST(sqlc.arg(before_posting_sequence) AS INTEGER) = 0
      OR document.posting_sequence < CAST(sqlc.arg(before_posting_sequence) AS INTEGER)
      OR (
          document.posting_sequence = CAST(sqlc.arg(before_posting_sequence) AS INTEGER)
          AND line.line_order > sqlc.arg(after_line_order)
      )
      OR (
          document.posting_sequence = CAST(sqlc.arg(before_posting_sequence) AS INTEGER)
          AND line.line_order = sqlc.arg(after_line_order)
          AND line.id > sqlc.arg(after_line_id)
      )
  )
ORDER BY document.posting_sequence DESC, line.line_order, line.id
LIMIT sqlc.arg(limit_count);

-- name: ListLineAllocations :many
SELECT
    allocation.id,
    allocation.line_id,
    allocation.lot_id,
    allocation.quantity_atomic,
    allocation.restores_allocation_id,
    allocation.created_at_ms,
    lot.source_line_id,
    lot.initial_quantity_atomic AS lot_initial_quantity_atomic,
    lot.lot_code,
    lot.originated_on,
    lot.expires_on
FROM lot_allocations allocation
JOIN inventory_lots lot ON lot.id = allocation.lot_id
WHERE allocation.line_id = sqlc.arg(line_id)
ORDER BY allocation.id;
