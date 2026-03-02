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

func (r *ItemRepository) Create(itm *model.ItemInsertDTO) (int64, error) {
	query := `
		INSERT INTO items
			(name, unit, sellable, purchasable, producible, default_sale_price)
		VALUES
			(?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.Conn.Exec(
		query,
		itm.Name,
		itm.Unit,
		itm.Sellable,
		itm.Purchasable,
		itm.Producible,
		itm.DefaultSalePrice,
	)
	
	if err != nil {
		return (-1, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return (-1, err)
	}

	return (id, nil)
}

func (r *ItemRepository) GetByID(id int64) (*model.Item, error) {
	query := `
		SELECT 
			id,
			name,
			unit,
			sellable,
			purchasable,
			producible,
			default_sale_price,
			created_at 
		FROM items
		WHERE id = ?
	`
	
	itm := &model.Item{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&itm.ID,
		&itm.Name,
		&itm.Unit,
		&itm.Sellable,
		&itm.Purchasable,
		&itm.Producible,
		&itm.DefaultSalePrice,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return itm, err
}

func (r *ItemRepository) GetAll() ([]*model.Item, error) {
	query := `
		SELECT 
			id,
			name,
			unit,
			sellable,
			purchasable,
			producible,
			default_sale_price,
			created_at 
		FROM items
		ORDER BY name ASC
	`

	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itms := []*model.Item{}
	for rows.Next() {
		itm := &model.Item{}
		if err := rows.Scan(
			&itm.ID,
			&itm.Name,
			&itm.Unit,
			&itm.Sellable,
			&itm.Purchasable,
			&itm.Producible,
			&itm.DefaultSalePrice,
		); err != nil {
			return nil, err
		}
		itms = append(itms, itm)
	}
	return itms, rows.Err()
}

// func (r *ItemRepository) Delete(id int64) error {
// 	query := `DELETE FROM items WHERE id = ?`
// 	_, err := r.db.Conn.Exec(query, id)
// 	return err
// }