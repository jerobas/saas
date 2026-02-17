package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type ItemRepository struct {
	db *Database
}

func NewItemRepository(db *Database) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(item *model.Item) error {
	query := `INSERT INTO items (id, name, unit, min_stock_alert) VALUES (?, ?, ?, ?)`
	_, err := r.db.Conn.Exec(query, item.ID, item.Name, item.Unit, item.MinStockAlert)
	return err
}

func (r *ItemRepository) GetByID(id string) (*model.Item, error) {
	query := `SELECT id, name, unit, min_stock_alert, created_at FROM items WHERE id = ?`
	item := &model.Item{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&item.ID, &item.Name, &item.Unit, &item.MinStockAlert, &item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

func (r *ItemRepository) GetAll() ([]*model.Item, error) {
	query := `SELECT id, name, unit, min_stock_alert, created_at FROM items ORDER BY name`
	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		item := &model.Item{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Unit, &item.MinStockAlert, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ItemRepository) Update(item *model.Item) error {
	query := `UPDATE items SET name = ?, unit = ?, min_stock_alert = ? WHERE id = ?`
	_, err := r.db.Conn.Exec(query, item.Name, item.Unit, item.MinStockAlert, item.ID)
	return err
}

func (r *ItemRepository) Delete(id string) error {
	_, err := r.db.Conn.Exec(`DELETE FROM items WHERE id = ?`, id)
	return err
}