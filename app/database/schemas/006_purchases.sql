CREATE TABLE IF NOT EXISTS purchase_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT,

  event_id INTEGER NOT NULL,      -- must be an event_type='PURCHASE' (enforced later via triggers/app)
  item_id  INTEGER NOT NULL,      -- unified catalog item (purchasable)

  quantity REAL NOT NULL CHECK(quantity > 0),

  unit_cost INTEGER NOT NULL CHECK(unit_cost >= 0),

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (event_id) REFERENCES events(id),
  FOREIGN KEY (item_id) REFERENCES items(id)
);

CREATE INDEX IF NOT EXISTS idx_purchase_lines_event ON purchase_lines(event_id);
CREATE INDEX IF NOT EXISTS idx_purchase_lines_item  ON purchase_lines(item_id);

CREATE TRIGGER IF NOT EXISTS trg_purchase_lines_no_insert_unless_draft
BEFORE INSERT ON purchase_lines
WHEN (SELECT e.status FROM events e WHERE e.id = NEW.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'purchase_lines can only be inserted while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_purchase_lines_no_update_unless_draft
BEFORE UPDATE ON purchase_lines
WHEN (SELECT e.status FROM events e WHERE e.id = OLD.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'purchase_lines can only be updated while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_purchase_lines_no_delete_unless_draft
BEFORE DELETE ON purchase_lines
WHEN (SELECT e.status FROM events e WHERE e.id = OLD.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'purchase_lines can only be deleted while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_purchase_lines_insert_require_event_is_purchase
BEFORE INSERT ON purchase_lines
BEGIN
  -- must exist
  SELECT RAISE(FAIL, 'purchase_lines.event_id not found')
  WHERE (SELECT 1 FROM events e WHERE e.id = NEW.event_id) IS NULL;

  -- must be PURCHASE
  SELECT RAISE(FAIL, 'purchase_lines require event_type=PURCHASE')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) <> 'PURCHASE';
END;

CREATE TRIGGER IF NOT EXISTS trg_purchase_lines_update_require_event_is_purchase
BEFORE UPDATE ON purchase_lines
BEGIN
  -- must exist
  SELECT RAISE(FAIL, 'purchase_lines.event_id not found')
  WHERE (SELECT 1 FROM events e WHERE e.id = NEW.event_id) IS NULL;

  -- must be PURCHASE
  SELECT RAISE(FAIL, 'purchase_lines require event_type=PURCHASE')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) <> 'PURCHASE';
END;

CREATE TRIGGER IF NOT EXISTS trg_purchase_lines_insert_item_must_be_purchasable
BEFORE INSERT ON purchase_lines
BEGIN
  SELECT RAISE(FAIL, 'purchase_lines require purchasable item')
  WHERE (SELECT i.purchasable FROM items i WHERE i.id = NEW.item_id) <> 1;
END;

CREATE TRIGGER IF NOT EXISTS trg_purchase_lines_update_item_must_be_purchasable
BEFORE UPDATE ON purchase_lines
BEGIN
  SELECT RAISE(FAIL, 'purchase_lines require purchasable item')
  WHERE (SELECT i.purchasable FROM items i WHERE i.id = NEW.item_id) <> 1;
END;