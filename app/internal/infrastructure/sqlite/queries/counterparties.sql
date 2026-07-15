-- name: GetCounterparty :one
SELECT
    counterparty.id,
    counterparty.name,
    counterparty.phone,
    counterparty.email,
    counterparty.notes,
    counterparty.created_at_ms,
    counterparty.updated_at_ms,
    counterparty.archived_at_ms,
    CAST(COALESCE((
        SELECT role.created_at_ms
        FROM counterparty_roles role
        WHERE role.counterparty_id = counterparty.id
          AND role.role = 'SUPPLIER'
    ), -1) AS INTEGER) AS supplier_role_created_at_ms,
    CAST(COALESCE((
        SELECT role.created_at_ms
        FROM counterparty_roles role
        WHERE role.counterparty_id = counterparty.id
          AND role.role = 'CUSTOMER'
    ), -1) AS INTEGER) AS customer_role_created_at_ms
FROM counterparties counterparty
WHERE counterparty.id = sqlc.arg(id);

-- name: ListCounterparties :many
SELECT
    counterparty.id,
    counterparty.name,
    counterparty.phone,
    counterparty.email,
    counterparty.notes,
    counterparty.created_at_ms,
    counterparty.updated_at_ms,
    counterparty.archived_at_ms,
    CAST(COALESCE((
        SELECT role.created_at_ms
        FROM counterparty_roles role
        WHERE role.counterparty_id = counterparty.id
          AND role.role = 'SUPPLIER'
    ), -1) AS INTEGER) AS supplier_role_created_at_ms,
    CAST(COALESCE((
        SELECT role.created_at_ms
        FROM counterparty_roles role
        WHERE role.counterparty_id = counterparty.id
          AND role.role = 'CUSTOMER'
    ), -1) AS INTEGER) AS customer_role_created_at_ms
FROM counterparties counterparty
WHERE
    (
        CAST(sqlc.arg(archive_filter) AS INTEGER) = 2
        OR (CAST(sqlc.arg(archive_filter) AS INTEGER) = 0 AND counterparty.archived_at_ms IS NULL)
        OR (CAST(sqlc.arg(archive_filter) AS INTEGER) = 1 AND counterparty.archived_at_ms IS NOT NULL)
    )
    AND (
        CAST(sqlc.arg(role_filter) AS TEXT) = ''
        OR EXISTS (
            SELECT 1
            FROM counterparty_roles role
            WHERE role.counterparty_id = counterparty.id
              AND role.role = CAST(sqlc.arg(role_filter) AS TEXT)
        )
    )
    -- Counterparty names intentionally have display-text, case-sensitive search.
    AND (
        CAST(sqlc.arg(search_text) AS TEXT) = ''
        OR instr(counterparty.name, CAST(sqlc.arg(search_text) AS TEXT)) > 0
    )
    AND (
        CAST(sqlc.arg(after_name) AS TEXT) = ''
        OR counterparty.name > CAST(sqlc.arg(after_name) AS TEXT)
        OR (
            counterparty.name = CAST(sqlc.arg(after_name) AS TEXT)
            AND counterparty.id > sqlc.arg(after_id)
        )
    )
ORDER BY counterparty.name, counterparty.id
LIMIT sqlc.arg(limit_count);

-- name: InsertCounterparty :one
INSERT INTO counterparties (
    name,
    phone,
    email,
    notes,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
) VALUES (
    sqlc.arg(name),
    sqlc.narg(phone),
    sqlc.narg(email),
    sqlc.narg(notes),
    sqlc.arg(created_at_ms),
    sqlc.arg(updated_at_ms),
    NULL
)
RETURNING id;

-- name: UpdateCounterparty :execrows
UPDATE counterparties
SET
    name = sqlc.arg(name),
    phone = sqlc.narg(phone),
    email = sqlc.narg(email),
    notes = sqlc.narg(notes),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: ArchiveCounterparty :execrows
UPDATE counterparties
SET
    archived_at_ms = CAST(sqlc.arg(archived_at_ms) AS INTEGER),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: RestoreCounterparty :execrows
UPDATE counterparties
SET
    archived_at_ms = NULL,
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NOT NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: ListCounterpartyRoles :many
SELECT
    counterparty_id,
    role,
    created_at_ms
FROM counterparty_roles
WHERE counterparty_id = sqlc.arg(counterparty_id)
ORDER BY role;

-- name: DeleteCounterpartyRoles :execrows
DELETE FROM counterparty_roles
WHERE counterparty_id = sqlc.arg(counterparty_id);

-- name: InsertCounterpartyRole :exec
INSERT INTO counterparty_roles (
    counterparty_id,
    role,
    created_at_ms
) VALUES (
    sqlc.arg(counterparty_id),
    sqlc.arg(role),
    sqlc.arg(created_at_ms)
);
