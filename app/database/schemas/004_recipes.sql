CREATE TABLE IF NOT EXISTS recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,

    output_item_id INTEGER NOT NULL, -- what this recipe produces into stock

    preparation_time_minutes INTEGER NOT NULL CHECK(preparation_time_minutes >= 0),
    instructions TEXT NOT NULL,
    standard_yield_quantity REAL NOT NULL CHECK(standard_yield_quantity > 0),

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (output_item_id) REFERENCES items(id)
);

CREATE INDEX IF NOT EXISTS idx_recipes_output_item ON recipes(output_item_id);

CREATE TABLE IF NOT EXISTS recipe_components (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recipe_id INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    quantity REAL NOT NULL CHECK(quantity > 0),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (recipe_id) REFERENCES recipes(id),
    FOREIGN KEY (item_id) REFERENCES items(id),

    UNIQUE(recipe_id, item_id)
);

CREATE INDEX IF NOT EXISTS idx_recipe_components_recipe ON recipe_components(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_components_item   ON recipe_components(item_id);

CREATE TRIGGER IF NOT EXISTS trg_recipes_output_must_be_producible
BEFORE INSERT ON recipes
BEGIN
  SELECT RAISE(FAIL, 'recipes.output_item_id must be producible')
  WHERE (SELECT i.producible FROM items i WHERE i.id = NEW.output_item_id) <> 1;
END;

CREATE TRIGGER IF NOT EXISTS trg_recipes_output_must_be_producible_update
BEFORE UPDATE ON recipes
BEGIN
  SELECT RAISE(FAIL, 'recipes.output_item_id must be producible')
  WHERE (SELECT i.producible FROM items i WHERE i.id = NEW.output_item_id) <> 1;
END;