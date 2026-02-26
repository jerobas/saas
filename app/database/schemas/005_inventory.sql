CREATE TABLE IF NOT EXISTS inventory_movements (
  id INTEGER PRIMARY KEY AUTOINCREMENT,

  event_id INTEGER NOT NULL,

  -- item_type TEXT NOT NULL CHECK(item_type IN ('INGREDIENT','PRODUCT')),
  item_id   INTEGER NOT NULL,

  direction TEXT NOT NULL CHECK(direction IN ('IN','OUT')),
  quantity  REAL NOT NULL CHECK(quantity > 0),

  unit_cost INTEGER 
    CHECK(unit_cost IS NULL OR unit_cost >= 0)
    CHECK (
      (direction = 'IN' AND unit_cost IS NOT NULL)
      OR
      (direction = 'OUT')
    ),

  expires_at DATETIME DEFAULT NULL,

  origin_movement_id INTEGER DEFAULT NULL,

  occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (event_id) REFERENCES events(id),
  FOREIGN KEY (item_id) REFERENCES items(id),
  FOREIGN KEY (origin_movement_id) REFERENCES inventory_movements(id)
);

CREATE INDEX IF NOT EXISTS idx_inv_mov_event      ON inventory_movements(event_id);
CREATE INDEX IF NOT EXISTS idx_inv_mov_item       ON inventory_movements(item_id);
CREATE INDEX IF NOT EXISTS idx_inv_mov_origin     ON inventory_movements(origin_movement_id);
CREATE INDEX IF NOT EXISTS idx_inv_mov_occurred   ON inventory_movements(occurred_at);

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_no_insert_in_out_same_item
BEFORE INSERT ON inventory_movements
BEGIN
  SELECT RAISE(FAIL, 'IN and OUT inventory_movements for the same item aren''t allowed in the same event')
  WHERE EXISTS (
    SELECT 1
    FROM inventory_movements mov
    WHERE mov.event_id = NEW.event_id
      AND mov.item_id = NEW.item_id  
      AND mov.direction <> NEW.direction
  );
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_no_update_in_out_same_item
BEFORE UPDATE ON inventory_movements
BEGIN
  SELECT RAISE(FAIL, 'IN and OUT inventory_movements for the same item aren''t allowed in the same event')
  WHERE EXISTS (
    SELECT 1
    FROM inventory_movements mov
    WHERE mov.event_id = NEW.event_id
      AND mov.item_id = NEW.item_id  
      AND mov.direction <> NEW.direction
      AND mov.id <> OLD.id
  );
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_no_insert_unless_draft
BEFORE INSERT ON inventory_movements
WHEN (SELECT e.status FROM events e WHERE e.id = NEW.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'inventory_movements can only be inserted while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_no_update_unless_draft
BEFORE UPDATE ON inventory_movements
WHEN (SELECT e.status FROM events e WHERE e.id = OLD.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'inventory_movements can only be updated while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_no_delete_unless_draft
BEFORE DELETE ON inventory_movements
WHEN (SELECT e.status FROM events e WHERE e.id = OLD.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'inventory_movements can only be deleted while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_insert_origin_rules
BEFORE INSERT ON inventory_movements
WHEN NEW.origin_movement_id IS NOT NULL
BEGIN
  -- origin must exist
  SELECT RAISE(FAIL, 'origin_movement_id does not exist')
  WHERE (SELECT 1 FROM inventory_movements om WHERE om.id = NEW.origin_movement_id) IS NULL;

  -- origin must be IN
  SELECT RAISE(FAIL, 'origin movement must have direction=IN')
  WHERE (SELECT om.direction FROM inventory_movements om WHERE om.id = NEW.origin_movement_id) <> 'IN';

  -- same item_id
  SELECT RAISE(FAIL, 'origin movement must have same item_id')
  WHERE (SELECT om.item_id FROM inventory_movements om WHERE om.id = NEW.origin_movement_id) <> NEW.item_id;

  -- new row must be OUT
  SELECT RAISE(FAIL, 'movement with origin_movement_id must be direction=OUT')
  WHERE NEW.direction <> 'OUT';
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_update_origin_rules
BEFORE UPDATE ON inventory_movements
WHEN NEW.origin_movement_id IS NOT NULL
BEGIN
  -- origin must exist
  SELECT RAISE(FAIL, 'origin_movement_id does not exist')
  WHERE (SELECT 1 FROM inventory_movements om WHERE om.id = NEW.origin_movement_id) IS NULL;

  -- origin must be IN
  SELECT RAISE(FAIL, 'origin movement must have direction=IN')
  WHERE (SELECT om.direction FROM inventory_movements om WHERE om.id = NEW.origin_movement_id) <> 'IN';

  -- same item_id
  SELECT RAISE(FAIL, 'origin movement must have same item_id')
  WHERE (SELECT om.item_id FROM inventory_movements om WHERE om.id = NEW.origin_movement_id) <> NEW.item_id;

  -- new row must be OUT
  SELECT RAISE(FAIL, 'movement with origin_movement_id must be direction=OUT')
  WHERE NEW.direction <> 'OUT';
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_insert_direction_by_event_type
BEFORE INSERT ON inventory_movements
BEGIN
  -- PURCHASE => must be IN
  SELECT RAISE(FAIL, 'PURCHASE movements must be IN')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) = 'PURCHASE'
    AND NEW.direction <> 'IN';

  -- SALE => must be OUT
  SELECT RAISE(FAIL, 'SALE movements must be OUT')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) = 'SALE'
    AND NEW.direction <> 'OUT';

  -- PRODUCTION => allow both (no check)

  -- CONVERSION => allow both (cannot enforce the pair without posting logic)

  -- ADJUSTMENT => allow both (no check)
END;

CREATE TRIGGER IF NOT EXISTS trg_inventory_movements_update_direction_by_event_type
BEFORE UPDATE ON inventory_movements
BEGIN
  -- PURCHASE => must be IN
  SELECT RAISE(FAIL, 'PURCHASE movements must be IN')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) = 'PURCHASE'
    AND NEW.direction <> 'IN';

  -- SALE => must be OUT
  SELECT RAISE(FAIL, 'SALE movements must be OUT')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) = 'SALE'
    AND NEW.direction <> 'OUT';

  -- PRODUCTION => allow both (no check)

  -- CONVERSION => allow both (cannot enforce the pair without posting logic)

  -- ADJUSTMENT => allow both (no check)
END;

-- ====================
-- SPECIAL CASE
-- ====================

CREATE TRIGGER IF NOT EXISTS trg_post_event_processing
BEFORE UPDATE ON events
WHEN OLD.status = 'DRAFT' AND NEW.status = 'POSTED'
BEGIN
  SELECT RAISE(FAIL, 'Data corruption: found item without equivalent item_stock')
    WHERE EXISTS (
    SELECT 1
    FROM inventory_movements mov
    LEFT JOIN item_stock sto ON sto.item_id = mov.item_id
    WHERE mov.event_id = NEW.id
      AND sto.item_id IS NULL
    LIMIT 1
  );

  SELECT RAISE(FAIL, 'Data corruption: OUT movement without item_stock.average_unit_cost')
  WHERE EXISTS (
    SELECT 1
    FROM inventory_movements mov, item_stock sto
    WHERE sto.item_id = mov.item_id
      AND mov.event_id = NEW.id
      AND mov.direction = 'OUT'
      AND sto.average_unit_cost IS NULL
    LIMIT 1
  );

  SELECT RAISE(FAIL, 'Not enough stock to post this event')
  WHERE EXISTS (
    SELECT 1
    FROM (
        SELECT
          mov.item_id,
          SUM(CASE WHEN mov.direction = 'IN' THEN mov.quantity ELSE -mov.quantity END) AS quantity
        FROM inventory_movements mov
        WHERE mov.event_id = NEW.id
        GROUP BY mov.item_id
      ) delta,
      item_stock sto
    WHERE sto.item_id = delta.item_id AND sto.quantity + delta.quantity < 0
    LIMIT 1
  );

  UPDATE inventory_movements
  SET unit_cost = (
    SELECT sto.average_unit_cost
    FROM item_stock sto
    WHERE sto.item_id = inventory_movements.item_id
  )
  WHERE event_id = NEW.id
    AND direction = 'OUT'; 
END;

CREATE TRIGGER IF NOT EXISTS trg_item_stock_update_when_posting
AFTER UPDATE ON events
WHEN OLD.status = 'DRAFT' AND NEW.status = 'POSTED'
BEGIN
  UPDATE item_stock SET 
    quantity = quantity + (
      SELECT
        SUM(CASE WHEN mov.direction = 'IN' THEN mov.quantity ELSE -mov.quantity END)
      FROM inventory_movements mov
      WHERE mov.event_id = NEW.id
        AND mov.item_id = item_stock.item_id
      GROUP BY mov.item_id
    ),
    average_unit_cost = (
      CASE WHEN quantity + (
        SELECT
          SUM(CASE WHEN mov.direction = 'IN' THEN mov.quantity ELSE -mov.quantity END)
        FROM inventory_movements mov
        WHERE mov.event_id = NEW.id
          AND mov.item_id = item_stock.item_id
        GROUP BY mov.item_id
      ) = 0
      THEN NULL
      ELSE ROUND(
        (COALESCE(average_unit_cost, 0) * quantity + (
          SELECT
            SUM(COALESCE(mov.unit_cost, 0) * (CASE WHEN mov.direction = 'IN' THEN mov.quantity ELSE -mov.quantity END))
          FROM inventory_movements mov
          WHERE mov.event_id = NEW.id
            AND mov.item_id = item_stock.item_id
          GROUP BY mov.item_id
        )) /
        (quantity + (
          SELECT
            SUM(CASE WHEN mov.direction = 'IN' THEN mov.quantity ELSE -mov.quantity END)
          FROM inventory_movements mov
          WHERE mov.event_id = NEW.id
            AND mov.item_id = item_stock.item_id
          GROUP BY mov.item_id
      ))) END
    ),
    updated_at = CURRENT_TIMESTAMP
  WHERE item_stock.item_id IN (
    SELECT
      mov.item_id AS item_id
    FROM inventory_movements mov
    WHERE mov.event_id = NEW.id
    GROUP BY mov.item_id
  );
END;