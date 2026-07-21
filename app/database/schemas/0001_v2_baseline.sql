-- Sweeters V2 persistence baseline.
-- Domain timestamps are UTC Unix milliseconds and business dates are ISO 8601
-- calendar dates. Quantities and money are integers in their documented scales.

CREATE TABLE app_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    business_name TEXT NOT NULL CHECK (length(trim(business_name)) > 0),
    locale_code TEXT NOT NULL CHECK (length(trim(locale_code)) > 0),
    timezone_name TEXT NOT NULL CHECK (length(trim(timezone_name)) > 0),
    currency_code TEXT NOT NULL CHECK (
        length(currency_code) = 3
        AND currency_code = upper(currency_code)
        AND currency_code GLOB '[A-Z][A-Z][A-Z]'
    ),
    currency_minor_digits INTEGER NOT NULL CHECK (currency_minor_digits BETWEEN 0 AND 6),
    hourly_labor_cost_minor INTEGER CHECK (hourly_labor_cost_minor >= 0),
    default_gross_margin_basis_points INTEGER CHECK (
        default_gross_margin_basis_points BETWEEN 0 AND 9999
    ),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    updated_at_ms INTEGER NOT NULL CHECK (updated_at_ms >= created_at_ms)
) STRICT;

INSERT INTO app_settings (
    id,
    business_name,
    locale_code,
    timezone_name,
    currency_code,
    currency_minor_digits,
    created_at_ms,
    updated_at_ms
) VALUES (1, 'Sweeters', 'pt-BR', 'America/Sao_Paulo', 'BRL', 2, 0, 0);

CREATE TABLE measurement_units (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    symbol TEXT NOT NULL CHECK (length(trim(symbol)) > 0),
    dimension TEXT NOT NULL CHECK (dimension IN ('MASS', 'VOLUME', 'COUNT')),
    atomic_numerator INTEGER NOT NULL CHECK (atomic_numerator > 0),
    atomic_denominator INTEGER NOT NULL CHECK (atomic_denominator > 0),
    is_item_base INTEGER NOT NULL CHECK (is_item_base IN (0, 1)),
    is_seeded INTEGER NOT NULL CHECK (is_seeded IN (0, 1))
) STRICT;

CREATE UNIQUE INDEX measurement_units_one_item_base_per_dimension
    ON measurement_units (dimension)
    WHERE is_item_base = 1;

INSERT INTO measurement_units (
    code, name, symbol, dimension, atomic_numerator, atomic_denominator, is_item_base, is_seeded
) VALUES
    ('mg', 'milligram', 'mg', 'MASS', 1, 1, 0, 1),
    ('g', 'gram', 'g', 'MASS', 1000, 1, 1, 1),
    ('kg', 'kilogram', 'kg', 'MASS', 1000000, 1, 0, 1),
    ('ul', 'microlitre', 'uL', 'VOLUME', 1, 1, 0, 1),
    ('ml', 'millilitre', 'mL', 'VOLUME', 1000, 1, 1, 1),
    ('l', 'litre', 'L', 'VOLUME', 1000000, 1, 0, 1),
    ('milli_each', 'thousandth of an item', 'milli-each', 'COUNT', 1, 1, 0, 1),
    ('each', 'item', 'each', 'COUNT', 1000, 1, 1, 1),
    ('dozen', 'dozen', 'dozen', 'COUNT', 12000, 1, 0, 1);

CREATE TABLE items (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    normalized_name TEXT NOT NULL UNIQUE CHECK (length(trim(normalized_name)) > 0),
    sku TEXT CHECK (sku IS NULL OR length(trim(sku)) > 0),
    normalized_sku TEXT UNIQUE CHECK (
        normalized_sku IS NULL OR length(trim(normalized_sku)) > 0
    ),
    description TEXT CHECK (description IS NULL OR length(trim(description)) > 0),
    base_unit_code TEXT NOT NULL REFERENCES measurement_units(code)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    is_purchasable INTEGER NOT NULL CHECK (is_purchasable IN (0, 1)),
    is_producible INTEGER NOT NULL CHECK (is_producible IN (0, 1)),
    is_sellable INTEGER NOT NULL CHECK (is_sellable IN (0, 1)),
    default_sale_price_minor INTEGER CHECK (default_sale_price_minor >= 0),
    reorder_quantity_atomic INTEGER CHECK (reorder_quantity_atomic >= 0),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    updated_at_ms INTEGER NOT NULL CHECK (updated_at_ms >= created_at_ms),
    archived_at_ms INTEGER CHECK (archived_at_ms >= updated_at_ms),
    CHECK ((sku IS NULL) = (normalized_sku IS NULL)),
    CHECK (
        archived_at_ms IS NOT NULL
        OR is_purchasable = 1
        OR is_producible = 1
        OR is_sellable = 1
    ),
    CHECK (default_sale_price_minor IS NULL OR is_sellable = 1)
) STRICT;

CREATE TABLE item_packagings (
    id INTEGER PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    normalized_name TEXT NOT NULL CHECK (length(trim(normalized_name)) > 0),
    entered_unit_code TEXT NOT NULL REFERENCES measurement_units(code)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    conversion_numerator_atomic INTEGER NOT NULL CHECK (conversion_numerator_atomic > 0),
    conversion_denominator INTEGER NOT NULL CHECK (conversion_denominator > 0),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    updated_at_ms INTEGER NOT NULL CHECK (updated_at_ms >= created_at_ms),
    archived_at_ms INTEGER CHECK (archived_at_ms >= updated_at_ms),
    UNIQUE (item_id, normalized_name)
) STRICT;

CREATE TABLE counterparties (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    phone TEXT CHECK (phone IS NULL OR length(trim(phone)) > 0),
    email TEXT CHECK (email IS NULL OR length(trim(email)) > 0),
    notes TEXT CHECK (notes IS NULL OR length(trim(notes)) > 0),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    updated_at_ms INTEGER NOT NULL CHECK (updated_at_ms >= created_at_ms),
    archived_at_ms INTEGER CHECK (archived_at_ms >= updated_at_ms)
) STRICT;

CREATE TABLE counterparty_roles (
    counterparty_id INTEGER NOT NULL REFERENCES counterparties(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    role TEXT NOT NULL CHECK (role IN ('SUPPLIER', 'CUSTOMER')),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    PRIMARY KEY (counterparty_id, role)
) STRICT;

CREATE TABLE recipes (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    normalized_name TEXT NOT NULL UNIQUE CHECK (length(trim(normalized_name)) > 0),
    output_item_id INTEGER NOT NULL REFERENCES items(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    updated_at_ms INTEGER NOT NULL CHECK (updated_at_ms >= created_at_ms),
    archived_at_ms INTEGER CHECK (archived_at_ms >= updated_at_ms)
) STRICT;

CREATE TABLE recipe_revisions (
    id INTEGER PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipes(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    revision_number INTEGER NOT NULL CHECK (revision_number > 0),
    standard_yield_quantity_atomic INTEGER NOT NULL CHECK (
        standard_yield_quantity_atomic > 0
    ),
    instructions TEXT NOT NULL,
    preparation_time_minutes INTEGER NOT NULL CHECK (preparation_time_minutes >= 0),
    estimated_direct_cost_micro INTEGER CHECK (estimated_direct_cost_micro >= 0),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    UNIQUE (recipe_id, revision_number)
) STRICT;

CREATE TABLE recipe_revision_components (
    id INTEGER PRIMARY KEY,
    recipe_revision_id INTEGER NOT NULL REFERENCES recipe_revisions(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    component_order INTEGER NOT NULL CHECK (component_order > 0),
    item_id INTEGER NOT NULL REFERENCES items(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    quantity_atomic INTEGER NOT NULL CHECK (quantity_atomic > 0),
    entered_unit_code TEXT NOT NULL REFERENCES measurement_units(code)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    entered_packaging_name TEXT CHECK (
        entered_packaging_name IS NULL OR length(trim(entered_packaging_name)) > 0
    ),
    conversion_numerator_atomic INTEGER NOT NULL CHECK (conversion_numerator_atomic > 0),
    conversion_denominator INTEGER NOT NULL CHECK (conversion_denominator > 0),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    UNIQUE (recipe_revision_id, component_order),
    UNIQUE (recipe_revision_id, item_id)
) STRICT;

CREATE TABLE stock_documents (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (
        kind IN ('PURCHASE', 'SALE', 'PRODUCTION', 'ADJUSTMENT', 'REVERSAL')
    ),
    idempotency_key TEXT NOT NULL UNIQUE CHECK (length(trim(idempotency_key)) > 0),
    posting_sequence INTEGER NOT NULL UNIQUE CHECK (posting_sequence > 0),
    counterparty_id INTEGER REFERENCES counterparties(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    occurred_on TEXT NOT NULL CHECK (
        length(occurred_on) = 10
        AND occurred_on GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'
    ),
    posted_at_ms INTEGER NOT NULL CHECK (posted_at_ms >= 0),
    currency_code TEXT NOT NULL CHECK (
        length(currency_code) = 3
        AND currency_code = upper(currency_code)
        AND currency_code GLOB '[A-Z][A-Z][A-Z]'
    ),
    currency_minor_digits INTEGER NOT NULL CHECK (currency_minor_digits BETWEEN 0 AND 6),
    reason_code TEXT CHECK (
        reason_code IS NULL OR reason_code IN (
            'FREE_STOCK',
            'PROMOTION',
            'SAMPLE',
            'OPENING_BALANCE',
            'PHYSICAL_COUNT',
            'WASTE',
            'EXPIRY',
            'DAMAGE',
            'DOCUMENTED_CORRECTION',
            'EXACT_REVERSAL'
        )
    ),
    notes TEXT CHECK (notes IS NULL OR length(trim(notes)) > 0),
    reverses_document_id INTEGER UNIQUE REFERENCES stock_documents(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    CHECK (counterparty_id IS NULL OR kind IN ('PURCHASE', 'SALE')),
    CHECK ((kind = 'REVERSAL') = (reverses_document_id IS NOT NULL)),
    CHECK (reverses_document_id IS NULL OR reverses_document_id <> id),
    CHECK (
        (kind = 'PURCHASE' AND (reason_code IS NULL OR reason_code = 'FREE_STOCK'))
        OR (kind = 'SALE' AND (reason_code IS NULL OR reason_code IN ('PROMOTION', 'SAMPLE')))
        OR (kind = 'PRODUCTION' AND reason_code IS NULL)
        OR (kind = 'ADJUSTMENT' AND reason_code IN (
            'OPENING_BALANCE',
            'FREE_STOCK',
            'PHYSICAL_COUNT',
            'WASTE',
            'EXPIRY',
            'DAMAGE',
            'SAMPLE',
            'DOCUMENTED_CORRECTION'
        ))
        OR (kind = 'REVERSAL' AND reason_code = 'EXACT_REVERSAL')
    )
) STRICT;

CREATE TABLE stock_document_lines (
    id INTEGER PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES stock_documents(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    line_order INTEGER NOT NULL CHECK (line_order > 0),
    item_id INTEGER NOT NULL REFERENCES items(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    direction TEXT NOT NULL CHECK (direction IN ('IN', 'OUT')),
    quantity_atomic INTEGER NOT NULL CHECK (quantity_atomic > 0),
    entered_unit_code TEXT NOT NULL REFERENCES measurement_units(code)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    entered_packaging_name TEXT CHECK (
        entered_packaging_name IS NULL OR length(trim(entered_packaging_name)) > 0
    ),
    conversion_numerator_atomic INTEGER NOT NULL CHECK (conversion_numerator_atomic > 0),
    conversion_denominator INTEGER NOT NULL CHECK (conversion_denominator > 0),
    inventory_value_micro INTEGER NOT NULL CHECK (inventory_value_micro >= 0),
    commercial_total_minor INTEGER CHECK (commercial_total_minor >= 0),
    reverses_line_id INTEGER UNIQUE REFERENCES stock_document_lines(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    UNIQUE (document_id, line_order)
) STRICT;

CREATE TABLE adjustment_line_details (
    line_id INTEGER PRIMARY KEY REFERENCES stock_document_lines(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    expected_quantity_atomic INTEGER NOT NULL CHECK (expected_quantity_atomic >= 0),
    observed_quantity_atomic INTEGER NOT NULL CHECK (observed_quantity_atomic >= 0),
    CHECK (expected_quantity_atomic <> observed_quantity_atomic)
) STRICT;

CREATE TABLE production_runs (
    document_id INTEGER PRIMARY KEY REFERENCES stock_documents(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    recipe_revision_id INTEGER NOT NULL REFERENCES recipe_revisions(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    output_line_id INTEGER NOT NULL UNIQUE REFERENCES stock_document_lines(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    direct_production_cost_micro INTEGER NOT NULL DEFAULT 0 CHECK (
        direct_production_cost_micro >= 0
    )
) STRICT;

CREATE TABLE inventory_lots (
    id INTEGER PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES items(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    source_line_id INTEGER NOT NULL UNIQUE REFERENCES stock_document_lines(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    initial_quantity_atomic INTEGER NOT NULL CHECK (initial_quantity_atomic > 0),
    lot_code TEXT CHECK (lot_code IS NULL OR length(trim(lot_code)) > 0),
    originated_on TEXT NOT NULL CHECK (
        length(originated_on) = 10
        AND originated_on GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'
    ),
    expires_on TEXT CHECK (
        expires_on IS NULL OR (
            length(expires_on) = 10
            AND expires_on GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'
        )
    ),
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0)
) STRICT;

CREATE TABLE lot_allocations (
    id INTEGER PRIMARY KEY,
    line_id INTEGER NOT NULL REFERENCES stock_document_lines(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    lot_id INTEGER NOT NULL REFERENCES inventory_lots(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    quantity_atomic INTEGER NOT NULL CHECK (quantity_atomic > 0),
    restores_allocation_id INTEGER UNIQUE REFERENCES lot_allocations(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    created_at_ms INTEGER NOT NULL CHECK (created_at_ms >= 0),
    CHECK (restores_allocation_id IS NULL OR restores_allocation_id <> id),
    UNIQUE (line_id, lot_id)
) STRICT;

CREATE TABLE inventory_balances (
    item_id INTEGER PRIMARY KEY REFERENCES items(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    quantity_atomic INTEGER NOT NULL CHECK (quantity_atomic >= 0),
    inventory_value_micro INTEGER NOT NULL CHECK (inventory_value_micro >= 0),
    last_document_id INTEGER REFERENCES stock_documents(id)
        ON UPDATE RESTRICT ON DELETE RESTRICT,
    updated_at_ms INTEGER NOT NULL CHECK (updated_at_ms >= 0),
    CHECK (quantity_atomic <> 0 OR inventory_value_micro = 0)
) STRICT;

CREATE INDEX items_active_name
    ON items (archived_at_ms, normalized_name);
CREATE INDEX items_active_purchasable
    ON items (normalized_name) WHERE archived_at_ms IS NULL AND is_purchasable = 1;
CREATE INDEX items_active_producible
    ON items (normalized_name) WHERE archived_at_ms IS NULL AND is_producible = 1;
CREATE INDEX items_active_sellable
    ON items (normalized_name) WHERE archived_at_ms IS NULL AND is_sellable = 1;
CREATE INDEX item_packagings_item_active_name
    ON item_packagings (item_id, archived_at_ms, normalized_name);
CREATE INDEX counterparties_active_name
    ON counterparties (archived_at_ms, name);
CREATE INDEX counterparty_roles_role
    ON counterparty_roles (role, counterparty_id);
CREATE INDEX recipes_output_item
    ON recipes (output_item_id);
CREATE INDEX recipes_active_name
    ON recipes (archived_at_ms, normalized_name);
CREATE INDEX recipe_revision_components_item
    ON recipe_revision_components (item_id);
CREATE INDEX stock_documents_kind_date_sequence
    ON stock_documents (kind, occurred_on, posting_sequence);
CREATE INDEX stock_documents_counterparty_sequence
    ON stock_documents (counterparty_id, posting_sequence);
CREATE INDEX stock_document_lines_item_document
    ON stock_document_lines (item_id, document_id);
CREATE INDEX production_runs_revision
    ON production_runs (recipe_revision_id);
CREATE INDEX inventory_lots_fefo
    ON inventory_lots (item_id, expires_on, id);
CREATE INDEX lot_allocations_lot
    ON lot_allocations (lot_id);
CREATE INDEX lot_allocations_line
    ON lot_allocations (line_id);

CREATE TRIGGER schema_migrations_no_update
BEFORE UPDATE ON schema_migrations
BEGIN
    SELECT RAISE(ABORT, 'applied migrations are immutable');
END;

CREATE TRIGGER schema_migrations_no_delete
BEFORE DELETE ON schema_migrations
BEGIN
    SELECT RAISE(ABORT, 'applied migrations are immutable');
END;

CREATE TRIGGER app_settings_no_delete
BEFORE DELETE ON app_settings
BEGIN
    SELECT RAISE(ABORT, 'the settings row cannot be deleted');
END;

CREATE TRIGGER app_settings_currency_lock
BEFORE UPDATE OF currency_code, currency_minor_digits ON app_settings
WHEN EXISTS (SELECT 1 FROM stock_documents)
 AND (
    NEW.currency_code <> OLD.currency_code
    OR NEW.currency_minor_digits <> OLD.currency_minor_digits
 )
BEGIN
    SELECT RAISE(ABORT, 'currency is locked after the first stock document');
END;

CREATE TRIGGER measurement_units_seeded_no_update
BEFORE UPDATE ON measurement_units
WHEN OLD.is_seeded = 1
BEGIN
    SELECT RAISE(ABORT, 'seeded measurement units are immutable');
END;

CREATE TRIGGER measurement_units_seeded_no_delete
BEFORE DELETE ON measurement_units
WHEN OLD.is_seeded = 1
BEGIN
    SELECT RAISE(ABORT, 'seeded measurement units are immutable');
END;

CREATE TRIGGER items_validate_base_unit_insert
BEFORE INSERT ON items
WHEN NOT EXISTS (
    SELECT 1 FROM measurement_units
    WHERE code = NEW.base_unit_code AND is_item_base = 1
)
BEGIN
    SELECT RAISE(ABORT, 'item base unit must be a canonical item unit');
END;

CREATE TRIGGER items_validate_base_unit_update
BEFORE UPDATE OF base_unit_code ON items
WHEN NOT EXISTS (
    SELECT 1 FROM measurement_units
    WHERE code = NEW.base_unit_code AND is_item_base = 1
)
BEGIN
    SELECT RAISE(ABORT, 'item base unit must be a canonical item unit');
END;

CREATE TRIGGER items_lock_used_base_unit
BEFORE UPDATE OF base_unit_code ON items
WHEN NEW.base_unit_code <> OLD.base_unit_code
 AND (
    EXISTS (
        SELECT 1 FROM item_packagings
        WHERE item_id = OLD.id AND archived_at_ms IS NULL
    )
    OR EXISTS (
        SELECT 1
        FROM recipes recipe
        JOIN recipe_revisions revision ON revision.recipe_id = recipe.id
        WHERE recipe.output_item_id = OLD.id
    )
    OR EXISTS (SELECT 1 FROM recipe_revision_components WHERE item_id = OLD.id)
    OR EXISTS (SELECT 1 FROM stock_document_lines WHERE item_id = OLD.id)
 )
BEGIN
    SELECT RAISE(ABORT, 'item base unit is immutable after use');
END;

CREATE TRIGGER items_no_delete
BEFORE DELETE ON items
BEGIN
    SELECT RAISE(ABORT, 'items must be archived, not deleted');
END;

CREATE TRIGGER item_packagings_validate_dimension_insert
BEFORE INSERT ON item_packagings
WHEN NOT EXISTS (
    SELECT 1
    FROM items item
    JOIN measurement_units base_unit ON base_unit.code = item.base_unit_code
    JOIN measurement_units entered_unit ON entered_unit.code = NEW.entered_unit_code
    WHERE item.id = NEW.item_id
      AND base_unit.dimension = entered_unit.dimension
)
BEGIN
    SELECT RAISE(ABORT, 'packaging unit dimension must match the item');
END;

CREATE TRIGGER item_packagings_validate_dimension_update
BEFORE UPDATE OF item_id, entered_unit_code, archived_at_ms ON item_packagings
WHEN NOT EXISTS (
    SELECT 1
    FROM items item
    JOIN measurement_units base_unit ON base_unit.code = item.base_unit_code
    JOIN measurement_units entered_unit ON entered_unit.code = NEW.entered_unit_code
    WHERE item.id = NEW.item_id
      AND base_unit.dimension = entered_unit.dimension
)
BEGIN
    SELECT RAISE(ABORT, 'packaging unit dimension must match the item');
END;

CREATE TRIGGER item_packagings_no_delete
BEFORE DELETE ON item_packagings
BEGIN
    SELECT RAISE(ABORT, 'item packagings must be archived, not deleted');
END;

CREATE TRIGGER counterparties_no_delete
BEFORE DELETE ON counterparties
BEGIN
    SELECT RAISE(ABORT, 'counterparties must be archived, not deleted');
END;

CREATE TRIGGER recipes_validate_output_insert
BEFORE INSERT ON recipes
WHEN NOT EXISTS (
    SELECT 1 FROM items
    WHERE id = NEW.output_item_id
      AND archived_at_ms IS NULL
      AND is_producible = 1
)
BEGIN
    SELECT RAISE(ABORT, 'recipe output must be an active producible item');
END;

CREATE TRIGGER recipes_validate_output_update
BEFORE UPDATE OF output_item_id ON recipes
WHEN NEW.output_item_id <> OLD.output_item_id
BEGIN
    SELECT RAISE(ABORT, 'recipe output item is immutable');
END;

CREATE TRIGGER recipes_no_delete
BEFORE DELETE ON recipes
BEGIN
    SELECT RAISE(ABORT, 'recipes must be archived, not deleted');
END;

CREATE TRIGGER recipe_revisions_no_update
BEFORE UPDATE ON recipe_revisions
BEGIN
    SELECT RAISE(ABORT, 'recipe revisions are immutable');
END;

CREATE TRIGGER recipe_revisions_no_delete
BEFORE DELETE ON recipe_revisions
BEGIN
    SELECT RAISE(ABORT, 'recipe revisions are immutable');
END;

CREATE TRIGGER recipe_components_validate_insert
BEFORE INSERT ON recipe_revision_components
WHEN NOT EXISTS (
    SELECT 1
    FROM recipe_revisions revision
    JOIN recipes recipe ON recipe.id = revision.recipe_id
    JOIN items component ON component.id = NEW.item_id
    JOIN measurement_units component_base ON component_base.code = component.base_unit_code
    JOIN measurement_units entered_unit ON entered_unit.code = NEW.entered_unit_code
    WHERE revision.id = NEW.recipe_revision_id
      AND recipe.output_item_id <> NEW.item_id
      AND component.archived_at_ms IS NULL
      AND component_base.dimension = entered_unit.dimension
)
BEGIN
    SELECT RAISE(ABORT, 'invalid recipe component or entered unit');
END;

CREATE TRIGGER recipe_components_no_update
BEFORE UPDATE ON recipe_revision_components
BEGIN
    SELECT RAISE(ABORT, 'recipe revision components are immutable');
END;

CREATE TRIGGER recipe_components_no_delete
BEFORE DELETE ON recipe_revision_components
BEGIN
    SELECT RAISE(ABORT, 'recipe revision components are immutable');
END;

CREATE TRIGGER stock_documents_validate_insert
BEFORE INSERT ON stock_documents
BEGIN
    SELECT CASE
        WHEN NEW.posting_sequence <= COALESCE((SELECT MAX(posting_sequence) FROM stock_documents), 0)
        THEN RAISE(ABORT, 'posting sequence must increase monotonically')
    END;
    SELECT CASE
        WHEN NEW.currency_code <> (SELECT currency_code FROM app_settings WHERE id = 1)
          OR NEW.currency_minor_digits <> (
              SELECT currency_minor_digits FROM app_settings WHERE id = 1
          )
        THEN RAISE(ABORT, 'document currency must match application settings')
    END;
    SELECT CASE
        WHEN NEW.counterparty_id IS NOT NULL
         AND NOT EXISTS (
             SELECT 1
             FROM counterparties counterparty
             JOIN counterparty_roles role ON role.counterparty_id = counterparty.id
             WHERE counterparty.id = NEW.counterparty_id
               AND counterparty.archived_at_ms IS NULL
               AND role.role = CASE NEW.kind
                   WHEN 'PURCHASE' THEN 'SUPPLIER'
                   WHEN 'SALE' THEN 'CUSTOMER'
               END
         )
        THEN RAISE(ABORT, 'counterparty is not eligible for this document kind')
    END;
    SELECT CASE
        WHEN NEW.kind = 'REVERSAL'
         AND NOT EXISTS (
             SELECT 1 FROM stock_documents target
             WHERE target.id = NEW.reverses_document_id
               AND target.kind <> 'REVERSAL'
         )
        THEN RAISE(ABORT, 'a reversal must target a non-reversal document')
    END;
END;

CREATE TRIGGER stock_documents_no_update
BEFORE UPDATE ON stock_documents
BEGIN
    SELECT RAISE(ABORT, 'stock documents are immutable');
END;

CREATE TRIGGER stock_documents_no_delete
BEFORE DELETE ON stock_documents
BEGIN
    SELECT RAISE(ABORT, 'stock documents are immutable');
END;

CREATE TRIGGER stock_document_lines_validate_insert
BEFORE INSERT ON stock_document_lines
BEGIN
    SELECT CASE
        WHEN NOT EXISTS (
            SELECT 1
            FROM items item
            JOIN measurement_units base_unit ON base_unit.code = item.base_unit_code
            JOIN measurement_units entered_unit ON entered_unit.code = NEW.entered_unit_code
            JOIN stock_documents document ON document.id = NEW.document_id
            WHERE item.id = NEW.item_id
              AND base_unit.dimension = entered_unit.dimension
              AND (
                  document.kind = 'REVERSAL'
                  OR (
                      item.archived_at_ms IS NULL
                      AND (
                          (document.kind = 'PURCHASE' AND item.is_purchasable = 1)
                          OR (document.kind = 'SALE' AND item.is_sellable = 1)
                          OR (document.kind = 'PRODUCTION' AND (
                              (NEW.direction = 'IN' AND item.is_producible = 1)
                              OR NEW.direction = 'OUT'
                          ))
                          OR document.kind = 'ADJUSTMENT'
                      )
                  )
              )
        )
        THEN RAISE(ABORT, 'item or entered unit is invalid for this document line')
    END;
    SELECT CASE
        WHEN EXISTS (
            SELECT 1
            FROM stock_document_lines line
            WHERE line.document_id = NEW.document_id
              AND line.item_id = NEW.item_id
              AND line.direction <> NEW.direction
        )
        THEN RAISE(ABORT, 'a document cannot move one item in both directions')
    END;
    SELECT CASE
        WHEN EXISTS (
            SELECT 1 FROM stock_documents document
            WHERE document.id = NEW.document_id
              AND (
                  (document.kind = 'PURCHASE' AND (
                      NEW.direction <> 'IN' OR NEW.commercial_total_minor IS NULL
                  ))
                  OR (document.kind = 'SALE' AND (
                      NEW.direction <> 'OUT' OR NEW.commercial_total_minor IS NULL
                  ))
                  OR (document.kind IN ('PRODUCTION', 'ADJUSTMENT')
                      AND NEW.commercial_total_minor IS NOT NULL)
                  OR (document.kind <> 'REVERSAL' AND NEW.reverses_line_id IS NOT NULL)
                  OR (document.kind = 'REVERSAL' AND NEW.reverses_line_id IS NULL)
              )
        )
        THEN RAISE(ABORT, 'line shape does not match its document kind')
    END;
    SELECT CASE
        WHEN EXISTS (
            SELECT 1 FROM stock_documents document
            WHERE document.id = NEW.document_id
              AND document.kind = 'PURCHASE'
              AND NEW.commercial_total_minor = 0
              AND document.reason_code IS NOT 'FREE_STOCK'
        )
        THEN RAISE(ABORT, 'zero-cost purchase requires FREE_STOCK')
    END;
    SELECT CASE
        WHEN EXISTS (
            SELECT 1 FROM stock_documents document
            WHERE document.id = NEW.document_id
              AND document.kind = 'SALE'
              AND NEW.commercial_total_minor = 0
              AND NOT (
                  document.reason_code IS 'PROMOTION'
                  OR document.reason_code IS 'SAMPLE'
              )
        )
        THEN RAISE(ABORT, 'zero-price sale requires PROMOTION or SAMPLE')
    END;
    SELECT CASE
        WHEN EXISTS (
            SELECT 1 FROM stock_documents document
            WHERE document.id = NEW.document_id
              AND document.kind = 'ADJUSTMENT'
              AND (
                  (document.reason_code IN ('OPENING_BALANCE', 'FREE_STOCK')
                      AND NEW.direction <> 'IN')
                  OR (document.reason_code IN ('WASTE', 'EXPIRY', 'DAMAGE', 'SAMPLE')
                      AND NEW.direction <> 'OUT')
              )
        )
        THEN RAISE(ABORT, 'adjustment direction does not match its reason')
    END;
    SELECT CASE
        WHEN EXISTS (
            SELECT 1 FROM stock_documents document
            WHERE document.id = NEW.document_id
              AND document.kind = 'PRODUCTION'
              AND NEW.direction = 'IN'
        )
         AND EXISTS (
             SELECT 1
             FROM stock_document_lines other
             WHERE other.document_id = NEW.document_id
               AND other.direction = 'IN'
         )
        THEN RAISE(ABORT, 'production can have only one output line')
    END;
    SELECT CASE
        WHEN EXISTS (
            SELECT 1
            FROM stock_documents document
            JOIN stock_document_lines target
              ON target.id = NEW.reverses_line_id
             AND target.document_id = document.reverses_document_id
            WHERE document.id = NEW.document_id
              AND document.kind = 'REVERSAL'
              AND NEW.item_id = target.item_id
              AND NEW.direction <> target.direction
              AND NEW.quantity_atomic = target.quantity_atomic
              AND NEW.entered_unit_code = target.entered_unit_code
              AND NEW.entered_packaging_name IS target.entered_packaging_name
              AND NEW.conversion_numerator_atomic = target.conversion_numerator_atomic
              AND NEW.conversion_denominator = target.conversion_denominator
              AND NEW.inventory_value_micro = target.inventory_value_micro
              AND NEW.commercial_total_minor IS target.commercial_total_minor
        ) = 0
         AND EXISTS (
             SELECT 1 FROM stock_documents
             WHERE id = NEW.document_id AND kind = 'REVERSAL'
         )
        THEN RAISE(ABORT, 'reversal line must exactly invert a target line')
    END;
END;

CREATE TRIGGER stock_document_lines_no_update
BEFORE UPDATE ON stock_document_lines
BEGIN
    SELECT RAISE(ABORT, 'stock document lines are immutable');
END;

CREATE TRIGGER stock_document_lines_no_delete
BEFORE DELETE ON stock_document_lines
BEGIN
    SELECT RAISE(ABORT, 'stock document lines are immutable');
END;

CREATE TRIGGER adjustment_line_details_validate_insert
BEFORE INSERT ON adjustment_line_details
WHEN NOT EXISTS (
    SELECT 1
    FROM stock_document_lines line
    JOIN stock_documents document ON document.id = line.document_id
    WHERE line.id = NEW.line_id
      AND document.kind = 'ADJUSTMENT'
      AND document.reason_code = 'PHYSICAL_COUNT'
      AND line.quantity_atomic = abs(
          NEW.observed_quantity_atomic - NEW.expected_quantity_atomic
      )
      AND line.direction = CASE
          WHEN NEW.observed_quantity_atomic > NEW.expected_quantity_atomic THEN 'IN'
          ELSE 'OUT'
      END
)
BEGIN
    SELECT RAISE(ABORT, 'physical-count detail does not match its adjustment line');
END;

CREATE TRIGGER adjustment_line_details_no_update
BEFORE UPDATE ON adjustment_line_details
BEGIN
    SELECT RAISE(ABORT, 'adjustment details are immutable');
END;

CREATE TRIGGER adjustment_line_details_no_delete
BEFORE DELETE ON adjustment_line_details
BEGIN
    SELECT RAISE(ABORT, 'adjustment details are immutable');
END;

CREATE TRIGGER production_runs_validate_insert
BEFORE INSERT ON production_runs
WHEN NOT EXISTS (
    SELECT 1
    FROM stock_documents document
    JOIN stock_document_lines output_line
      ON output_line.id = NEW.output_line_id
     AND output_line.document_id = document.id
    JOIN recipe_revisions revision ON revision.id = NEW.recipe_revision_id
    JOIN recipes recipe ON recipe.id = revision.recipe_id
    WHERE document.id = NEW.document_id
      AND document.kind = 'PRODUCTION'
      AND output_line.direction = 'IN'
      AND output_line.item_id = recipe.output_item_id
)
BEGIN
    SELECT RAISE(ABORT, 'production run output does not match its recipe revision');
END;

CREATE TRIGGER production_runs_no_update
BEFORE UPDATE ON production_runs
BEGIN
    SELECT RAISE(ABORT, 'production runs are immutable');
END;

CREATE TRIGGER production_runs_no_delete
BEFORE DELETE ON production_runs
BEGIN
    SELECT RAISE(ABORT, 'production runs are immutable');
END;

CREATE TRIGGER inventory_lots_validate_insert
BEFORE INSERT ON inventory_lots
WHEN NOT EXISTS (
    SELECT 1
    FROM stock_document_lines line
    JOIN stock_documents document ON document.id = line.document_id
    WHERE line.id = NEW.source_line_id
      AND document.kind <> 'REVERSAL'
      AND line.direction = 'IN'
      AND line.item_id = NEW.item_id
      AND line.quantity_atomic = NEW.initial_quantity_atomic
)
BEGIN
    SELECT RAISE(ABORT, 'lot must exactly represent a normal inbound line');
END;

CREATE TRIGGER inventory_lots_no_update
BEFORE UPDATE ON inventory_lots
BEGIN
    SELECT RAISE(ABORT, 'inventory lots are immutable');
END;

CREATE TRIGGER inventory_lots_no_delete
BEFORE DELETE ON inventory_lots
BEGIN
    SELECT RAISE(ABORT, 'inventory lots are immutable');
END;

CREATE TRIGGER lot_allocations_validate_insert
BEFORE INSERT ON lot_allocations
BEGIN
    SELECT CASE
        WHEN NOT EXISTS (
            SELECT 1
            FROM stock_document_lines line
            JOIN inventory_lots lot ON lot.id = NEW.lot_id
            WHERE line.id = NEW.line_id
              AND line.item_id = lot.item_id
        )
        THEN RAISE(ABORT, 'allocation item must match its lot')
    END;
    SELECT CASE
        WHEN NEW.restores_allocation_id IS NULL
         AND NOT EXISTS (
             SELECT 1
             FROM stock_document_lines line
             JOIN inventory_lots lot ON lot.id = NEW.lot_id
             JOIN stock_document_lines source ON source.id = lot.source_line_id
             JOIN stock_documents consuming_document ON consuming_document.id = line.document_id
             JOIN stock_documents source_document ON source_document.id = source.document_id
             WHERE line.id = NEW.line_id
               AND line.direction = 'OUT'
               AND consuming_document.posting_sequence > source_document.posting_sequence
               AND (
                   consuming_document.kind <> 'REVERSAL'
                   OR source.id = line.reverses_line_id
               )
         )
        THEN RAISE(ABORT, 'normal allocation must consume an earlier-document lot')
    END;
    SELECT CASE
        WHEN NEW.restores_allocation_id IS NOT NULL
         AND NOT EXISTS (
             SELECT 1
             FROM lot_allocations original
             JOIN stock_document_lines original_line ON original_line.id = original.line_id
             JOIN stock_document_lines restoring_line ON restoring_line.id = NEW.line_id
             JOIN stock_documents restoring_document
               ON restoring_document.id = restoring_line.document_id
             WHERE original.id = NEW.restores_allocation_id
               AND original.restores_allocation_id IS NULL
               AND original.lot_id = NEW.lot_id
               AND original.quantity_atomic = NEW.quantity_atomic
               AND restoring_line.direction = 'IN'
               AND restoring_document.kind = 'REVERSAL'
               AND restoring_line.reverses_line_id = original_line.id
         )
        THEN RAISE(ABORT, 'restoration must exactly reverse an original allocation')
    END;
    SELECT CASE
        WHEN (
            SELECT COALESCE(SUM(
                CASE WHEN restores_allocation_id IS NULL
                    THEN quantity_atomic ELSE -quantity_atomic END
            ), 0)
            FROM lot_allocations
            WHERE lot_id = NEW.lot_id
        ) + CASE WHEN NEW.restores_allocation_id IS NULL
            THEN NEW.quantity_atomic ELSE -NEW.quantity_atomic END
        NOT BETWEEN 0 AND (
            SELECT initial_quantity_atomic FROM inventory_lots WHERE id = NEW.lot_id
        )
        THEN RAISE(ABORT, 'allocation would make lot consumption invalid')
    END;
END;

CREATE TRIGGER lot_allocations_no_update
BEFORE UPDATE ON lot_allocations
BEGIN
    SELECT RAISE(ABORT, 'lot allocations are immutable');
END;

CREATE TRIGGER lot_allocations_no_delete
BEFORE DELETE ON lot_allocations
BEGIN
    SELECT RAISE(ABORT, 'lot allocations are immutable');
END;

CREATE TRIGGER items_create_zero_balance
AFTER INSERT ON items
BEGIN
    INSERT INTO inventory_balances (
        item_id, quantity_atomic, inventory_value_micro, last_document_id, updated_at_ms
    ) VALUES (NEW.id, 0, 0, NULL, NEW.created_at_ms);
END;

CREATE TRIGGER inventory_balances_no_delete
BEFORE DELETE ON inventory_balances
BEGIN
    SELECT RAISE(ABORT, 'inventory balance rows cannot be deleted');
END;
