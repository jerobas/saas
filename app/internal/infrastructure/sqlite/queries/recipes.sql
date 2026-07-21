-- name: GetRecipe :one
SELECT
    id,
    name,
    normalized_name,
    output_item_id,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
FROM recipes
WHERE id = sqlc.arg(id);

-- name: GetCurrentRecipe :one
SELECT
    recipe.id AS recipe_id,
    recipe.name AS recipe_name,
    recipe.normalized_name AS recipe_normalized_name,
    recipe.output_item_id,
    recipe.created_at_ms AS recipe_created_at_ms,
    recipe.updated_at_ms AS recipe_updated_at_ms,
    recipe.archived_at_ms AS recipe_archived_at_ms,
    revision.id AS revision_id,
    revision.revision_number,
    revision.standard_yield_quantity_atomic,
    revision.instructions,
    revision.preparation_time_minutes,
    revision.estimated_direct_cost_micro,
    revision.created_at_ms AS revision_created_at_ms,
    CAST(COALESCE(revision_chain.revision_count, 0) AS INTEGER) AS revision_count,
    CAST(COALESCE(revision_chain.minimum_revision_number, 0) AS INTEGER) AS minimum_revision_number
FROM recipes recipe
JOIN (
    SELECT
        historical.recipe_id,
        CAST(COUNT(*) AS INTEGER) AS revision_count,
        CAST(MIN(historical.revision_number) AS INTEGER) AS minimum_revision_number,
        CAST(MAX(historical.revision_number) AS INTEGER) AS latest_revision_number
    FROM recipe_revisions historical
    WHERE historical.recipe_id = sqlc.arg(target_recipe_id)
    GROUP BY historical.recipe_id
) revision_chain ON revision_chain.recipe_id = recipe.id
JOIN recipe_revisions revision
  ON revision.recipe_id = revision_chain.recipe_id
 AND revision.revision_number = revision_chain.latest_revision_number
WHERE recipe.id = sqlc.arg(target_recipe_id);

-- name: ListRecipes :many
SELECT
    recipe.id,
    recipe.name,
    recipe.normalized_name,
    recipe.output_item_id,
    output.name AS output_item_name,
    recipe.created_at_ms,
    recipe.updated_at_ms,
    recipe.archived_at_ms,
    current_revision.id AS current_revision_id,
    current_revision.revision_number AS current_revision_number,
    current_revision.standard_yield_quantity_atomic AS current_standard_yield_quantity_atomic,
    CAST(COALESCE(revision_chain.revision_count, 0) AS INTEGER) AS revision_count,
    CAST(COALESCE(revision_chain.minimum_revision_number, 0) AS INTEGER) AS minimum_revision_number
FROM recipes recipe
JOIN items output ON output.id = recipe.output_item_id
LEFT JOIN (
    SELECT
        historical.recipe_id,
        CAST(COUNT(*) AS INTEGER) AS revision_count,
        CAST(MIN(historical.revision_number) AS INTEGER) AS minimum_revision_number,
        CAST(MAX(historical.revision_number) AS INTEGER) AS latest_revision_number
    FROM recipe_revisions historical
    GROUP BY historical.recipe_id
) revision_chain ON revision_chain.recipe_id = recipe.id
LEFT JOIN recipe_revisions current_revision
  ON current_revision.recipe_id = revision_chain.recipe_id
 AND current_revision.revision_number = revision_chain.latest_revision_number
WHERE
    (
        CAST(sqlc.arg(archive_filter) AS INTEGER) = 2
        OR (CAST(sqlc.arg(archive_filter) AS INTEGER) = 0 AND recipe.archived_at_ms IS NULL)
        OR (CAST(sqlc.arg(archive_filter) AS INTEGER) = 1 AND recipe.archived_at_ms IS NOT NULL)
    )
    AND (
        CAST(sqlc.arg(search_key) AS TEXT) = ''
        OR instr(recipe.normalized_name, CAST(sqlc.arg(search_key) AS TEXT)) > 0
    )
    AND (
        CAST(sqlc.arg(after_normalized_name) AS TEXT) = ''
        OR recipe.normalized_name > CAST(sqlc.arg(after_normalized_name) AS TEXT)
        OR (
            recipe.normalized_name = CAST(sqlc.arg(after_normalized_name) AS TEXT)
            AND recipe.id > sqlc.arg(after_id)
        )
    )
ORDER BY recipe.normalized_name, recipe.id
LIMIT sqlc.arg(limit_count);

-- name: InsertRecipe :one
INSERT INTO recipes (
    name,
    normalized_name,
    output_item_id,
    created_at_ms,
    updated_at_ms,
    archived_at_ms
) VALUES (
    sqlc.arg(name),
    sqlc.arg(normalized_name),
    sqlc.arg(output_item_id),
    sqlc.arg(created_at_ms),
    sqlc.arg(updated_at_ms),
    NULL
)
RETURNING id;

-- name: RenameRecipe :execrows
UPDATE recipes
SET
    name = sqlc.arg(name),
    normalized_name = sqlc.arg(normalized_name),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: AdvanceRecipeVersion :execrows
UPDATE recipes
SET updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: ArchiveRecipe :execrows
UPDATE recipes
SET
    archived_at_ms = CAST(sqlc.arg(archived_at_ms) AS INTEGER),
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: RestoreRecipe :execrows
UPDATE recipes
SET
    archived_at_ms = NULL,
    updated_at_ms = sqlc.arg(updated_at_ms)
WHERE id = sqlc.arg(id)
  AND archived_at_ms IS NOT NULL
  AND updated_at_ms = sqlc.arg(expected_updated_at_ms);

-- name: GetLatestRecipeRevisionNumber :one
SELECT CAST(COALESCE(MAX(revision_number), 0) AS INTEGER) AS latest_revision_number
FROM recipe_revisions
WHERE recipe_id = sqlc.arg(recipe_id);

-- name: GetRecipeRevision :one
SELECT
    id,
    recipe_id,
    revision_number,
    standard_yield_quantity_atomic,
    instructions,
    preparation_time_minutes,
    estimated_direct_cost_micro,
    created_at_ms,
    CAST((
        SELECT COUNT(*)
        FROM recipe_revisions historical
        WHERE historical.recipe_id = revision.recipe_id
    ) AS INTEGER) AS revision_count,
    CAST((
        SELECT MIN(historical.revision_number)
        FROM recipe_revisions historical
        WHERE historical.recipe_id = revision.recipe_id
    ) AS INTEGER) AS minimum_revision_number,
    CAST((
        SELECT MAX(historical.revision_number)
        FROM recipe_revisions historical
        WHERE historical.recipe_id = revision.recipe_id
    ) AS INTEGER) AS latest_revision_number
FROM recipe_revisions revision
WHERE revision.id = sqlc.arg(id);

-- name: ListRecipeRevisions :many
SELECT
    id,
    recipe_id,
    revision_number,
    standard_yield_quantity_atomic,
    instructions,
    preparation_time_minutes,
    estimated_direct_cost_micro,
    created_at_ms
FROM recipe_revisions
WHERE recipe_id = sqlc.arg(recipe_id)
ORDER BY revision_number DESC;

-- name: InsertRecipeRevision :one
INSERT INTO recipe_revisions (
    recipe_id,
    revision_number,
    standard_yield_quantity_atomic,
    instructions,
    preparation_time_minutes,
    estimated_direct_cost_micro,
    created_at_ms
) SELECT
    sqlc.arg(recipe_id),
    sqlc.arg(revision_number),
    sqlc.arg(standard_yield_quantity_atomic),
    sqlc.arg(instructions),
    sqlc.arg(preparation_time_minutes),
    sqlc.narg(estimated_direct_cost_micro),
    sqlc.arg(created_at_ms)
WHERE CAST(sqlc.arg(expected_latest_revision_number) AS INTEGER) = (
    SELECT CAST(COALESCE(MAX(existing.revision_number), 0) AS INTEGER)
    FROM recipe_revisions existing
    WHERE existing.recipe_id = sqlc.arg(recipe_id)
)
  AND sqlc.arg(revision_number) = CAST(sqlc.arg(expected_latest_revision_number) AS INTEGER) + 1
RETURNING id;

-- name: ListRecipeRevisionComponents :many
SELECT
    id,
    recipe_revision_id,
    component_order,
    item_id,
    quantity_atomic,
    entered_unit_code,
    entered_packaging_name,
    conversion_numerator_atomic,
    conversion_denominator,
    created_at_ms
FROM recipe_revision_components
WHERE recipe_revision_id = sqlc.arg(recipe_revision_id)
ORDER BY component_order, id;

-- name: InsertRecipeRevisionComponent :one
INSERT INTO recipe_revision_components (
    recipe_revision_id,
    component_order,
    item_id,
    quantity_atomic,
    entered_unit_code,
    entered_packaging_name,
    conversion_numerator_atomic,
    conversion_denominator,
    created_at_ms
) VALUES (
    sqlc.arg(recipe_revision_id),
    sqlc.arg(component_order),
    sqlc.arg(item_id),
    sqlc.arg(quantity_atomic),
    sqlc.arg(entered_unit_code),
    sqlc.narg(entered_packaging_name),
    sqlc.arg(conversion_numerator_atomic),
    sqlc.arg(conversion_denominator),
    sqlc.arg(created_at_ms)
)
RETURNING id;
