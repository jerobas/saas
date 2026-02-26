CREATE TABLE IF NOT EXISTS sale_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT,

  event_id INTEGER NOT NULL,      -- must be an event_type='SALE' (enforced later via triggers/app)
  item_id  INTEGER NOT NULL,      -- unified catalog item (sellable)

  quantity REAL NOT NULL CHECK(quantity > 0),

  unit_price INTEGER NOT NULL CHECK(unit_price >= 0),

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (event_id) REFERENCES events(id),
  FOREIGN KEY (item_id) REFERENCES items(id)
);

CREATE INDEX IF NOT EXISTS idx_sale_lines_event ON sale_lines(event_id);
CREATE INDEX IF NOT EXISTS idx_sale_lines_item  ON sale_lines(item_id);

CREATE TRIGGER IF NOT EXISTS trg_sale_lines_no_insert_unless_draft
BEFORE INSERT ON sale_lines
WHEN (SELECT e.status FROM events e WHERE e.id = NEW.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'sale_lines can only be inserted while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_sale_lines_no_update_unless_draft
BEFORE UPDATE ON sale_lines
WHEN (SELECT e.status FROM events e WHERE e.id = OLD.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'sale_lines can only be updated while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_sale_lines_no_delete_unless_draft
BEFORE DELETE ON sale_lines
WHEN (SELECT e.status FROM events e WHERE e.id = OLD.event_id) <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'sale_lines can only be deleted while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_sale_lines_insert_require_event_is_sale
BEFORE INSERT ON sale_lines
BEGIN
  SELECT RAISE(FAIL, 'sale_lines.event_id not found')
  WHERE (SELECT 1 FROM events e WHERE e.id = NEW.event_id) IS NULL;

  SELECT RAISE(FAIL, 'sale_lines require event_type=SALE')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) <> 'SALE';
END;

CREATE TRIGGER IF NOT EXISTS trg_sale_lines_update_require_event_is_sale
BEFORE UPDATE ON sale_lines
BEGIN
  SELECT RAISE(FAIL, 'sale_lines.event_id not found')
  WHERE (SELECT 1 FROM events e WHERE e.id = NEW.event_id) IS NULL;

  SELECT RAISE(FAIL, 'sale_lines require event_type=SALE')
  WHERE (SELECT e.event_type FROM events e WHERE e.id = NEW.event_id) <> 'SALE';
END;

CREATE TRIGGER IF NOT EXISTS trg_sale_lines_insert_item_must_be_sellable
BEFORE INSERT ON sale_lines
BEGIN
  SELECT RAISE(FAIL, 'sale_lines require sellable item')
  WHERE (SELECT i.sellable FROM items i WHERE i.id = NEW.item_id) <> 1;
END;

CREATE TRIGGER IF NOT EXISTS trg_sale_lines_update_item_must_be_sellable
BEFORE UPDATE ON sale_lines
BEGIN
  SELECT RAISE(FAIL, 'sale_lines require sellable item')
  WHERE (SELECT i.sellable FROM items i WHERE i.id = NEW.item_id) <> 1;
END;