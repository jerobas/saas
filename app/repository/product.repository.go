package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type ProductRepository struct {
	db *Database
}

func NewProductRepository(db *Database) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *model.Product) error {
	query := `INSERT INTO products (id, recipe_id, quantity_produced, unit_cost, sale_price) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Conn.Exec(query, product.ID, product.RecipeID, product.QuantityProduced, product.UnitCost, product.SalePrice)
	return err
}

func (r *ProductRepository) GetByID(id string) (*model.Product, error) {
	query := `SELECT id, recipe_id, quantity_produced, unit_cost, sale_price, produced_at FROM products WHERE id = ?`
	product := &model.Product{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&product.ID, &product.RecipeID, &product.QuantityProduced,
		&product.UnitCost, &product.SalePrice, &product.ProducedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return product, err
}

func (r *ProductRepository) GetAll() ([]*model.Product, error) {
	query := `SELECT id, recipe_id, quantity_produced, unit_cost, sale_price, produced_at FROM products ORDER BY produced_at DESC`
	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []*model.Product{}
	for rows.Next() {
		product := &model.Product{}
		if err := rows.Scan(&product.ID, &product.RecipeID, &product.QuantityProduced,
			&product.UnitCost, &product.SalePrice, &product.ProducedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (r *ProductRepository) Delete(id string) error {
	_, err := r.db.Conn.Exec(`DELETE FROM products WHERE id = ?`, id)
	return err
}