-- name: GetAppSettings :one
SELECT
    id,
    business_name,
    locale_code,
    timezone_name,
    currency_code,
    currency_minor_digits,
    hourly_labor_cost_minor,
    default_gross_margin_basis_points,
    created_at_ms,
    updated_at_ms
FROM app_settings
WHERE id = 1;

-- name: UpdateAppSettings :one
UPDATE app_settings
SET
    business_name = sqlc.arg(business_name),
    locale_code = sqlc.arg(locale_code),
    timezone_name = sqlc.arg(timezone_name),
    currency_code = sqlc.arg(currency_code),
    currency_minor_digits = sqlc.arg(currency_minor_digits),
    hourly_labor_cost_minor = sqlc.narg(hourly_labor_cost_minor),
    default_gross_margin_basis_points = sqlc.narg(default_gross_margin_basis_points),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = 1
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms)
RETURNING
    id,
    business_name,
    locale_code,
    timezone_name,
    currency_code,
    currency_minor_digits,
    hourly_labor_cost_minor,
    default_gross_margin_basis_points,
    created_at_ms,
    updated_at_ms;

-- name: GetMeasurementUnit :one
SELECT
    code,
    name,
    symbol,
    dimension,
    atomic_numerator,
    atomic_denominator,
    is_item_base,
    is_seeded
FROM measurement_units
WHERE code = sqlc.arg(code);

-- name: ListMeasurementUnits :many
SELECT
    code,
    name,
    symbol,
    dimension,
    atomic_numerator,
    atomic_denominator,
    is_item_base,
    is_seeded
FROM measurement_units
ORDER BY
    CASE dimension
        WHEN 'MASS' THEN 1
        WHEN 'VOLUME' THEN 2
        WHEN 'COUNT' THEN 3
    END,
    atomic_numerator,
    atomic_denominator,
    code;
