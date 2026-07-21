-- name: GetItem :one
SELECT
    id,
    name,
    normalized_name,
    sku,
    normalized_sku,
    description,
    base_unit_code,
    is_purchasable,
    is_producible,
    is_sellable,
    default_sale_price_minor,
    reorder_quantity_atomic,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
FROM items
WHERE id = sqlc.arg(id);

-- name: ListItems :many
SELECT
    id,
    name,
    normalized_name,
    sku,
    normalized_sku,
    description,
    base_unit_code,
    is_purchasable,
    is_producible,
    is_sellable,
    default_sale_price_minor,
    reorder_quantity_atomic,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
FROM items
WHERE
    (
        CAST(sqlc.arg(archive_filter) AS INTEGER) = 2
        OR (CAST(sqlc.arg(archive_filter) AS INTEGER) = 0 AND archived_at_ms IS NULL)
        OR (CAST(sqlc.arg(archive_filter) AS INTEGER) = 1 AND archived_at_ms IS NOT NULL)
    )
    AND (CAST(sqlc.arg(require_purchasable) AS INTEGER) = 0 OR is_purchasable = 1)
    AND (CAST(sqlc.arg(require_producible) AS INTEGER) = 0 OR is_producible = 1)
    AND (CAST(sqlc.arg(require_sellable) AS INTEGER) = 0 OR is_sellable = 1)
    AND (
        CAST(sqlc.arg(search_key) AS TEXT) = ''
        OR instr(normalized_name, CAST(sqlc.arg(search_key) AS TEXT)) > 0
        OR instr(COALESCE(normalized_sku, ''), CAST(sqlc.arg(search_key) AS TEXT)) > 0
    )
    AND (
        CAST(sqlc.arg(after_normalized_name) AS TEXT) = ''
        OR normalized_name > CAST(sqlc.arg(after_normalized_name) AS TEXT)
        OR (
            normalized_name = CAST(sqlc.arg(after_normalized_name) AS TEXT)
            AND id > sqlc.arg(after_id)
        )
    )
ORDER BY normalized_name, id
LIMIT sqlc.arg(limit_count);

-- name: InsertItem :one
INSERT INTO items (
    name,
    normalized_name,
    sku,
    normalized_sku,
    description,
    base_unit_code,
    is_purchasable,
    is_producible,
    is_sellable,
    default_sale_price_minor,
    reorder_quantity_atomic,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
) VALUES (
    sqlc.arg(name),
    sqlc.arg(normalized_name),
    sqlc.narg(sku),
    sqlc.narg(normalized_sku),
    sqlc.narg(description),
    sqlc.arg(base_unit_code),
    sqlc.arg(is_purchasable),
    sqlc.arg(is_producible),
    sqlc.arg(is_sellable),
    sqlc.narg(default_sale_price_minor),
    sqlc.narg(reorder_quantity_atomic),
    sqlc.arg(created_at_ms),
    sqlc.arg(updated_at_ms),
    NULL
)
RETURNING id;

-- name: UpdateItem :execrows
UPDATE items
SET
    name = sqlc.arg(name),
    normalized_name = sqlc.arg(normalized_name),
    sku = sqlc.narg(sku),
    normalized_sku = sqlc.narg(normalized_sku),
    description = sqlc.narg(description),
    base_unit_code = sqlc.arg(base_unit_code),
    is_purchasable = sqlc.arg(is_purchasable),
    is_producible = sqlc.arg(is_producible),
    is_sellable = sqlc.arg(is_sellable),
    default_sale_price_minor = sqlc.narg(default_sale_price_minor),
    reorder_quantity_atomic = sqlc.narg(reorder_quantity_atomic),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: ArchiveItem :execrows
UPDATE items
SET
    archived_at_ms = CAST(sqlc.arg(archived_at_ms) AS INTEGER),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: RestoreItem :execrows
UPDATE items
SET
    archived_at_ms = NULL,
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NOT NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: GetItemPackaging :one
SELECT
    id,
    item_id,
    name,
    normalized_name,
    entered_unit_code,
    conversion_numerator_atomic,
    conversion_denominator,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
FROM item_packagings
WHERE id = sqlc.arg(id);

-- name: ListItemPackagings :many
SELECT
    id,
    item_id,
    name,
    normalized_name,
    entered_unit_code,
    conversion_numerator_atomic,
    conversion_denominator,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
FROM item_packagings
WHERE item_id = sqlc.arg(item_id)
  AND (CAST(sqlc.arg(include_archived) AS INTEGER) = 1 OR archived_at_ms IS NULL)
ORDER BY normalized_name, id;

-- name: InsertItemPackaging :one
INSERT INTO item_packagings (
    item_id,
    name,
    normalized_name,
    entered_unit_code,
    conversion_numerator_atomic,
    conversion_denominator,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
) VALUES (
    sqlc.arg(item_id),
    sqlc.arg(name),
    sqlc.arg(normalized_name),
    sqlc.arg(entered_unit_code),
    sqlc.arg(conversion_numerator_atomic),
    sqlc.arg(conversion_denominator),
    sqlc.arg(created_at_ms),
    sqlc.arg(updated_at_ms),
    NULL
)
RETURNING id;

-- name: UpdateItemPackaging :execrows
UPDATE item_packagings
SET
    name = sqlc.arg(name),
    normalized_name = sqlc.arg(normalized_name),
    entered_unit_code = sqlc.arg(entered_unit_code),
    conversion_numerator_atomic = sqlc.arg(conversion_numerator_atomic),
    conversion_denominator = sqlc.arg(conversion_denominator),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: ArchiveItemPackaging :execrows
UPDATE item_packagings
SET
    archived_at_ms = CAST(sqlc.arg(archived_at_ms) AS INTEGER),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: ReconfigureArchivedItemPackaging :execrows
UPDATE item_packagings
SET
    name = sqlc.arg(name),
    normalized_name = sqlc.arg(normalized_name),
    entered_unit_code = sqlc.arg(entered_unit_code),
    conversion_numerator_atomic = sqlc.arg(conversion_numerator_atomic),
    conversion_denominator = sqlc.arg(conversion_denominator),
    updated_at_ms = sqlc.arg(updated_at_ms),
    archived_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NOT NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: RestoreItemPackaging :execrows
UPDATE item_packagings
SET
    archived_at_ms = NULL,
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NOT NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);
