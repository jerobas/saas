PRAGMA foreign_keys = ON;

-- =========================
-- PROFILE (Enterprise Settings)
-- =========================
CREATE TABLE IF NOT EXISTS profile (
  id INTEGER PRIMARY KEY CHECK (id = 1),

  hourly_cost REAL NOT NULL,
  default_profit_margin REAL NOT NULL,

  expected_monthly_profit REAL,
  fixed_monthly_expenses REAL,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


-- =========================
-- ITEMS
-- =========================
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    unit TEXT NOT NULL,
    min_stock REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- INVENTORY BATCHES
-- =========================
CREATE TABLE IF NOT EXISTS inventory_batches (
    id TEXT PRIMARY KEY,
    item_id TEXT NOT NULL,
    quantity_total REAL NOT NULL,
    quantity_remaining REAL NOT NULL,
    purchase_price_total REAL NOT NULL,
    unit_price REAL NOT NULL,
    purchased_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id)
);

CREATE INDEX IF NOT EXISTS idx_inventory_batches_item ON inventory_batches(item_id);

-- =========================
-- RECIPES
-- =========================
CREATE TABLE IF NOT EXISTS recipes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    profit_margin_percent REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- RECIPE INGREDIENTS
-- =========================
CREATE TABLE IF NOT EXISTS recipe_ingredients (
    recipe_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    quantity_needed REAL NOT NULL,
    PRIMARY KEY (recipe_id, item_id),
    FOREIGN KEY (recipe_id) REFERENCES recipes(id),
    FOREIGN KEY (item_id) REFERENCES items(id)
);

-- =========================
-- PRODUCTS
-- =========================
CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    recipe_id TEXT NOT NULL,
    quantity_produced INTEGER NOT NULL,
    unit_cost REAL NOT NULL,
    sale_price REAL NOT NULL,
    produced_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

-- =========================
-- PRODUCT INVENTORY
-- =========================
CREATE TABLE IF NOT EXISTS product_inventory (
    product_id TEXT PRIMARY KEY,
    quantity_available INTEGER NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id)
);

-- =========================
-- SALES
-- =========================
CREATE TABLE IF NOT EXISTS sales (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    quantity_sold INTEGER NOT NULL,
    unit_price REAL NOT NULL,
    total_price REAL NOT NULL,
    sold_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE INDEX IF NOT EXISTS idx_sales_product ON sales(product_id);
CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(sold_at);

-- =========================
-- CONSUMPTION LOG
-- =========================
CREATE TABLE IF NOT EXISTS batch_consumption (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    batch_id TEXT NOT NULL,
    quantity_used REAL NOT NULL,
    consumed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (batch_id) REFERENCES inventory_batches(id)
);

CREATE INDEX IF NOT EXISTS idx_batch_consumption_product ON batch_consumption(product_id);

-- =========================
-- VIEWS
-- Para Views, o SQLite não suporta IF NOT EXISTS diretamente no CREATE VIEW.
-- O ideal é dar um DROP antes ou ignorar o erro.
-- =========================
DROP VIEW IF EXISTS current_stock;
CREATE VIEW current_stock AS
SELECT
    i.id AS item_id,
    i.name,
    i.unit,
    SUM(b.quantity_remaining) AS total_quantity,
    i.min_stock_alert
FROM items i
LEFT JOIN inventory_batches b ON b.item_id = i.id
GROUP BY i.id;

DROP VIEW IF EXISTS sales_profit;
CREATE VIEW sales_profit AS
SELECT
    s.id AS sale_id,
    s.sold_at,
    p.id AS product_id,
    s.quantity_sold,
    p.unit_cost,
    s.unit_price,
    (s.unit_price - p.unit_cost) * s.quantity_sold AS profit
FROM sales s
JOIN products p ON p.id = s.product_id;