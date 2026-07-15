CREATE TRIGGER items_preserve_active_recipe_output
BEFORE UPDATE OF archived_at_ms, is_producible ON items
WHEN (
    NEW.archived_at_ms IS NOT NULL
    OR NEW.is_producible <> 1
)
AND EXISTS (
    SELECT 1
    FROM recipes recipe
    WHERE recipe.output_item_id = OLD.id
      AND recipe.archived_at_ms IS NULL
)
BEGIN
    SELECT RAISE(ABORT, 'active recipe output must remain active and producible');
END;

CREATE TRIGGER recipes_validate_output_restore
BEFORE UPDATE OF archived_at_ms ON recipes
WHEN OLD.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms IS NULL
 AND NOT EXISTS (
    SELECT 1
    FROM items output
    WHERE output.id = NEW.output_item_id
      AND output.archived_at_ms IS NULL
      AND output.is_producible = 1
 )
BEGIN
    SELECT RAISE(ABORT, 'restored recipe output must be active and producible');
END;

CREATE TRIGGER recipe_revisions_require_next_number
BEFORE INSERT ON recipe_revisions
WHEN NEW.revision_number <> (
    SELECT COALESCE(MAX(existing.revision_number), 0) + 1
    FROM recipe_revisions existing
    WHERE existing.recipe_id = NEW.recipe_id
)
BEGIN
    SELECT RAISE(ABORT, 'recipe revision number must be the next contiguous value');
END;

CREATE TRIGGER items_archive_version_insert
BEFORE INSERT ON items
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'item archive timestamp must equal its optimistic version');
END;

CREATE TRIGGER items_archive_version_update
BEFORE UPDATE OF archived_at_ms, updated_at_ms ON items
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'item archive timestamp must equal its optimistic version');
END;

CREATE TRIGGER item_packagings_archive_version_insert
BEFORE INSERT ON item_packagings
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'item packaging archive timestamp must equal its optimistic version');
END;

CREATE TRIGGER item_packagings_archive_version_update
BEFORE UPDATE OF archived_at_ms, updated_at_ms ON item_packagings
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'item packaging archive timestamp must equal its optimistic version');
END;

CREATE TRIGGER counterparties_archive_version_insert
BEFORE INSERT ON counterparties
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'counterparty archive timestamp must equal its optimistic version');
END;

CREATE TRIGGER counterparties_archive_version_update
BEFORE UPDATE OF archived_at_ms, updated_at_ms ON counterparties
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'counterparty archive timestamp must equal its optimistic version');
END;

CREATE TRIGGER recipes_archive_version_insert
BEFORE INSERT ON recipes
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'recipe archive timestamp must equal its optimistic version');
END;

CREATE TRIGGER recipes_archive_version_update
BEFORE UPDATE OF archived_at_ms, updated_at_ms ON recipes
WHEN NEW.archived_at_ms IS NOT NULL
 AND NEW.archived_at_ms <> NEW.updated_at_ms
BEGIN
    SELECT RAISE(ABORT, 'recipe archive timestamp must equal its optimistic version');
END;

-- Re-run the new guards over version-one rows. A database containing a state
-- that the new contract cannot represent fails this migration atomically
-- instead of recording version two over invalid data.
CREATE TABLE migration_0002_recipe_chain_guard (
    singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
    valid INTEGER NOT NULL CHECK (valid = 1)
) STRICT;

INSERT INTO migration_0002_recipe_chain_guard (singleton, valid)
SELECT 1, CASE WHEN EXISTS (
    SELECT recipe.id
    FROM recipes recipe
    LEFT JOIN recipe_revisions revision ON revision.recipe_id = recipe.id
    GROUP BY recipe.id
    HAVING COUNT(revision.id) = 0
        OR MIN(revision.revision_number) <> 1
        OR COUNT(revision.id) <> MAX(revision.revision_number)
) THEN 0 ELSE 1 END;

DROP TABLE migration_0002_recipe_chain_guard;

UPDATE items
SET is_producible = is_producible,
    updated_at_ms = updated_at_ms;

UPDATE item_packagings
SET updated_at_ms = updated_at_ms;

UPDATE counterparties
SET updated_at_ms = updated_at_ms;

UPDATE recipes
SET updated_at_ms = updated_at_ms;
