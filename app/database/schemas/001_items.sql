CREATE TABLE IF NOT EXISTS items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,

  name TEXT NOT NULL UNIQUE,
  unit TEXT NOT NULL, -- base unit: 'g', 'ml', 'un', etc.

  sellable    INTEGER NOT NULL DEFAULT 0 CHECK (sellable IN (0,1)),
  purchasable INTEGER NOT NULL DEFAULT 0 CHECK (purchasable IN (0,1)),
  producible  INTEGER NOT NULL DEFAULT 0 CHECK (producible IN (0,1)),

  default_sale_price INTEGER CHECK (default_sale_price IS NULL OR default_sale_price >= 0)
    CHECK (default_sale_price IS NULL OR default_sale_price >= 0)
    CHECK (
      (sellable = 1 AND default_sale_price IS NOT NULL)
      OR
      (sellable = 0 AND default_sale_price IS NULL)
    ),

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);

CREATE TABLE IF NOT EXISTS item_stock (
  -- item_type TEXT NOT NULL CHECK(item_type IN ('INGREDIENT','PRODUCT')),
  item_id INTEGER PRIMARY KEY NOT NULL,

  quantity REAL NOT NULL DEFAULT 0 CHECK(quantity >= 0),

  average_unit_cost INTEGER CHECK(average_unit_cost IS NULL OR average_unit_cost >= 0), --NOT NULL later

  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (item_id) REFERENCES items(id)
);

CREATE TABLE IF NOT EXISTS item_conversions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_item_id INTEGER NOT NULL,
    to_item_id INTEGER NOT NULL,
    factor REAL NOT NULL CHECK(factor > 0),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (from_item_id) REFERENCES items(id),
    FOREIGN KEY (to_item_id) REFERENCES items(id),

    UNIQUE(from_item_id, to_item_id),
    CHECK(from_item_id <> to_item_id)
);

CREATE INDEX IF NOT EXISTS idx_item_conversions_from ON item_conversions(from_item_id);
CREATE INDEX IF NOT EXISTS idx_item_conversions_to   ON item_conversions(to_item_id);

CREATE TRIGGER IF NOT EXISTS trg_items_no_delete
BEFORE DELETE ON items
BEGIN
  SELECT RAISE(FAIL, 'items cannot be deleted (use soft delete/status instead)');
END;

CREATE TRIGGER IF NOT EXISTS trg_item_stock_insert
AFTER INSERT ON items
BEGIN
  INSERT INTO item_stock (item_id, quantity, average_unit_cost, updated_at) VALUES (NEW.id, 0, NULL, CURRENT_TIMESTAMP);
END;

CREATE TRIGGER IF NOT EXISTS trg_item_stock_no_delete
BEFORE DELETE ON item_stock
BEGIN
  SELECT RAISE(FAIL, 'item_stock cannot be deleted, soft delete the item to deactivate stock');
END;