package main

import (
	"database/sql"
	"time"
)

// ============================================
// ITEMS
// ============================================

type Item struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Unit          string    `json:"unit"`
	MinStockAlert float64   `json:"min_stock_alert"`
	CreatedAt     time.Time `json:"created_at"`
}

type ItemRepository struct {
	db *Database
}

func NewItemRepository(db *Database) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(item *Item) error {
	query := `INSERT INTO items (id, name, unit, min_stock_alert) VALUES (?, ?, ?, ?)`
	_, err := r.db.conn.Exec(query, item.ID, item.Name, item.Unit, item.MinStockAlert)
	return err
}

func (r *ItemRepository) GetByID(id string) (*Item, error) {
	query := `SELECT id, name, unit, min_stock_alert, created_at FROM items WHERE id = ?`
	item := &Item{}
	err := r.db.conn.QueryRow(query, id).Scan(
		&item.ID, &item.Name, &item.Unit, &item.MinStockAlert, &item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

func (r *ItemRepository) GetAll() ([]*Item, error) {
	query := `SELECT id, name, unit, min_stock_alert, created_at FROM items ORDER BY name`
	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*Item{}
	for rows.Next() {
		item := &Item{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Unit, &item.MinStockAlert, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ItemRepository) Update(item *Item) error {
	query := `UPDATE items SET name = ?, unit = ?, min_stock_alert = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, item.Name, item.Unit, item.MinStockAlert, item.ID)
	return err
}

func (r *ItemRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM items WHERE id = ?`, id)
	return err
}

// ============================================
// INVENTORY BATCHES
// ============================================

type InventoryBatch struct {
	ID                 string    `json:"id"`
	ItemID             string    `json:"item_id"`
	QuantityTotal      float64   `json:"quantity_total"`
	QuantityRemaining  float64   `json:"quantity_remaining"`
	PurchasePriceTotal float64   `json:"purchase_price_total"`
	UnitPrice          float64   `json:"unit_price"`
	PurchasedAt        time.Time `json:"purchased_at"`
}

type InventoryBatchRepository struct {
	db *Database
}

func NewInventoryBatchRepository(db *Database) *InventoryBatchRepository {
	return &InventoryBatchRepository{db: db}
}

func (r *InventoryBatchRepository) Create(batch *InventoryBatch) error {
	query := `
		INSERT INTO inventory_batches 
		(id, item_id, quantity_total, quantity_remaining, purchase_price_total, unit_price)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.conn.Exec(query, batch.ID, batch.ItemID, batch.QuantityTotal,
		batch.QuantityRemaining, batch.PurchasePriceTotal, batch.UnitPrice)
	return err
}

func (r *InventoryBatchRepository) GetByID(id string) (*InventoryBatch, error) {
	query := `
		SELECT id, item_id, quantity_total, quantity_remaining, 
		       purchase_price_total, unit_price, purchased_at
		FROM inventory_batches WHERE id = ?
	`
	batch := &InventoryBatch{}
	err := r.db.conn.QueryRow(query, id).Scan(
		&batch.ID, &batch.ItemID, &batch.QuantityTotal, &batch.QuantityRemaining,
		&batch.PurchasePriceTotal, &batch.UnitPrice, &batch.PurchasedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return batch, err
}

func (r *InventoryBatchRepository) GetByItemID(itemID string) ([]*InventoryBatch, error) {
	query := `
		SELECT id, item_id, quantity_total, quantity_remaining,
		       purchase_price_total, unit_price, purchased_at
		FROM inventory_batches 
		WHERE item_id = ? AND quantity_remaining > 0
		ORDER BY purchased_at
	`
	rows, err := r.db.conn.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := []*InventoryBatch{}
	for rows.Next() {
		batch := &InventoryBatch{}
		if err := rows.Scan(&batch.ID, &batch.ItemID, &batch.QuantityTotal,
			&batch.QuantityRemaining, &batch.PurchasePriceTotal, &batch.UnitPrice,
			&batch.PurchasedAt); err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}
	return batches, rows.Err()
}

func (r *InventoryBatchRepository) UpdateQuantity(id string, quantity float64) error {
	_, err := r.db.conn.Exec(`UPDATE inventory_batches SET quantity_remaining = ? WHERE id = ?`, quantity, id)
	return err
}

func (r *InventoryBatchRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM inventory_batches WHERE id = ?`, id)
	return err
}

// ============================================
// RECIPES
// ============================================

type Recipe struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	ProfitMarginPercent float64   `json:"profit_margin_percent"`
	CreatedAt           time.Time `json:"created_at"`
}

type RecipeIngredient struct {
	RecipeID       string  `json:"recipe_id"`
	ItemID         string  `json:"item_id"`
	QuantityNeeded float64 `json:"quantity_needed"`
}

type RecipeRepository struct {
	db *Database
}

func NewRecipeRepository(db *Database) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (r *RecipeRepository) Create(recipe *Recipe) error {
	query := `INSERT INTO recipes (id, name, profit_margin_percent) VALUES (?, ?, ?)`
	_, err := r.db.conn.Exec(query, recipe.ID, recipe.Name, recipe.ProfitMarginPercent)
	return err
}

func (r *RecipeRepository) GetByID(id string) (*Recipe, error) {
	query := `SELECT id, name, profit_margin_percent, created_at FROM recipes WHERE id = ?`
	recipe := &Recipe{}
	err := r.db.conn.QueryRow(query, id).Scan(
		&recipe.ID, &recipe.Name, &recipe.ProfitMarginPercent, &recipe.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return recipe, err
}

func (r *RecipeRepository) GetAll() ([]*Recipe, error) {
	query := `SELECT id, name, profit_margin_percent, created_at FROM recipes ORDER BY name`
	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := []*Recipe{}
	for rows.Next() {
		recipe := &Recipe{}
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.ProfitMarginPercent, &recipe.CreatedAt); err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	return recipes, rows.Err()
}

func (r *RecipeRepository) Update(recipe *Recipe) error {
	query := `UPDATE recipes SET name = ?, profit_margin_percent = ? WHERE id = ?`
	_, err := r.db.conn.Exec(query, recipe.Name, recipe.ProfitMarginPercent, recipe.ID)
	return err
}

func (r *RecipeRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM recipes WHERE id = ?`, id)
	return err
}

func (r *RecipeRepository) AddIngredient(ingredient *RecipeIngredient) error {
	query := `INSERT INTO recipe_ingredients (recipe_id, item_id, quantity_needed) VALUES (?, ?, ?)`
	_, err := r.db.conn.Exec(query, ingredient.RecipeID, ingredient.ItemID, ingredient.QuantityNeeded)
	return err
}

func (r *RecipeRepository) GetIngredients(recipeID string) ([]*RecipeIngredient, error) {
	query := `SELECT recipe_id, item_id, quantity_needed FROM recipe_ingredients WHERE recipe_id = ?`
	rows, err := r.db.conn.Query(query, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ingredients := []*RecipeIngredient{}
	for rows.Next() {
		ing := &RecipeIngredient{}
		if err := rows.Scan(&ing.RecipeID, &ing.ItemID, &ing.QuantityNeeded); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ing)
	}
	return ingredients, rows.Err()
}

func (r *RecipeRepository) RemoveIngredient(recipeID, itemID string) error {
	_, err := r.db.conn.Exec(`DELETE FROM recipe_ingredients WHERE recipe_id = ? AND item_id = ?`, recipeID, itemID)
	return err
}

// ============================================
// PRODUCTS
// ============================================

type Product struct {
	ID               string    `json:"id"`
	RecipeID         string    `json:"recipe_id"`
	QuantityProduced int       `json:"quantity_produced"`
	UnitCost         float64   `json:"unit_cost"`
	SalePrice        float64   `json:"sale_price"`
	ProducedAt       time.Time `json:"produced_at"`
}

type ProductRepository struct {
	db *Database
}

func NewProductRepository(db *Database) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *Product) error {
	query := `INSERT INTO products (id, recipe_id, quantity_produced, unit_cost, sale_price) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.conn.Exec(query, product.ID, product.RecipeID, product.QuantityProduced, product.UnitCost, product.SalePrice)
	return err
}

func (r *ProductRepository) GetByID(id string) (*Product, error) {
	query := `SELECT id, recipe_id, quantity_produced, unit_cost, sale_price, produced_at FROM products WHERE id = ?`
	product := &Product{}
	err := r.db.conn.QueryRow(query, id).Scan(
		&product.ID, &product.RecipeID, &product.QuantityProduced,
		&product.UnitCost, &product.SalePrice, &product.ProducedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return product, err
}

func (r *ProductRepository) GetAll() ([]*Product, error) {
	query := `SELECT id, recipe_id, quantity_produced, unit_cost, sale_price, produced_at FROM products ORDER BY produced_at DESC`
	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []*Product{}
	for rows.Next() {
		product := &Product{}
		if err := rows.Scan(&product.ID, &product.RecipeID, &product.QuantityProduced,
			&product.UnitCost, &product.SalePrice, &product.ProducedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (r *ProductRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM products WHERE id = ?`, id)
	return err
}

// ============================================
// SALES
// ============================================

type Sale struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	QuantitySold int       `json:"quantity_sold"`
	UnitPrice    float64   `json:"unit_price"`
	TotalPrice   float64   `json:"total_price"`
	SoldAt       time.Time `json:"sold_at"`
}

type SaleRepository struct {
	db *Database
}

func NewSaleRepository(db *Database) *SaleRepository {
	return &SaleRepository{db: db}
}

func (r *SaleRepository) Create(sale *Sale) error {
	query := `INSERT INTO sales (id, product_id, quantity_sold, unit_price, total_price) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.conn.Exec(query, sale.ID, sale.ProductID, sale.QuantitySold, sale.UnitPrice, sale.TotalPrice)
	return err
}

func (r *SaleRepository) GetByID(id string) (*Sale, error) {
	query := `SELECT id, product_id, quantity_sold, unit_price, total_price, sold_at FROM sales WHERE id = ?`
	sale := &Sale{}
	err := r.db.conn.QueryRow(query, id).Scan(
		&sale.ID, &sale.ProductID, &sale.QuantitySold, &sale.UnitPrice, &sale.TotalPrice, &sale.SoldAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return sale, err
}

func (r *SaleRepository) GetAll() ([]*Sale, error) {
	query := `SELECT id, product_id, quantity_sold, unit_price, total_price, sold_at FROM sales ORDER BY sold_at DESC`
	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := []*Sale{}
	for rows.Next() {
		sale := &Sale{}
		if err := rows.Scan(&sale.ID, &sale.ProductID, &sale.QuantitySold,
			&sale.UnitPrice, &sale.TotalPrice, &sale.SoldAt); err != nil {
			return nil, err
		}
		sales = append(sales, sale)
	}
	return sales, rows.Err()
}

func (r *SaleRepository) Delete(id string) error {
	_, err := r.db.conn.Exec(`DELETE FROM sales WHERE id = ?`, id)
	return err
}