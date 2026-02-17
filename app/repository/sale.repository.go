package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
	"github.com/jerobas/saas/database"
)

type Database = database.Database

type SaleRepository struct {
	db *Database
}

func NewSaleRepository(db *Database) *SaleRepository {
	return &SaleRepository{db: db}
}

func (r *SaleRepository) Create(sale *model.Sale) error {
	query := `INSERT INTO sales (id, product_id, quantity_sold, unit_price, total_price) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Conn.Exec(query, sale.ID, sale.ProductID, sale.QuantitySold, sale.UnitPrice, sale.TotalPrice)
	return err
}

func (r *SaleRepository) GetByID(id string) (*model.Sale, error) {
	query := `SELECT id, product_id, quantity_sold, unit_price, total_price, sold_at FROM sales WHERE id = ?`
	sale := &model.Sale{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&sale.ID, &sale.ProductID, &sale.QuantitySold, &sale.UnitPrice, &sale.TotalPrice, &sale.SoldAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return sale, err
}

func (r *SaleRepository) GetAll() ([]*model.Sale, error) {
	query := `SELECT id, product_id, quantity_sold, unit_price, total_price, sold_at FROM sales ORDER BY sold_at DESC`
	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := []*model.Sale{}
	for rows.Next() {
		sale := &model.Sale{}
		if err := rows.Scan(&sale.ID, &sale.ProductID, &sale.QuantitySold,
			&sale.UnitPrice, &sale.TotalPrice, &sale.SoldAt); err != nil {
			return nil, err
		}
		sales = append(sales, sale)
	}
	return sales, rows.Err()
}

func (r *SaleRepository) Delete(id string) error {
	_, err := r.db.Conn.Exec(`DELETE FROM sales WHERE id = ?`, id)
	return err
}