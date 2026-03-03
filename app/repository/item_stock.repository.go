package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type ItemStockRepository struct {
	db *Database
}

func NewItemStockRepository(db *Database) *ItemStockRepository {
	return &ItemStockRepository{db: db}
}

func (r *ItemStockRepository) GetByID(itemID int64) (*model.ItemStock, error) {
	query := `
		SELECT 
			item_id,
			quantity,
			average_unit_cost,
			updated_at 
		FROM item_stock
		WHERE item_id = ?
	`
	
	sto := &model.ItemStock{}
	err := r.db.Conn.QueryRow(query, itemID).Scan(
		&sto.ItemID,
		&sto.Quantity,
		&sto.AverageUnitCost,
		&sto.UpdatedAt
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return sto, err
}

func (r *ItemStockRepository) GetAll() ([]*model.ItemStock, error) {
	query := `
		SELECT 
			item_id,
			quantity,
			average_unit_cost,
			updated_at 
		FROM item_stock
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stos := []*model.ItemStock{}
	for rows.Next() {
		sto := &model.ItemStock{}
		if err := rows.Scan(
			&sto.ItemID,
			&sto.Quantity,
			&sto.AverageUnitCost,
			&sto.UpdatedAt
		); err != nil {
			return nil, err
		}
		stos = append(stos, sto)
	}
	return stos, rows.Err()
}