package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type PurchaseLineRepository struct {
	db *Database
}

func NewPurchaseLineRepository(db *Database) *PurchaseLineRepository {
	return &PurchaseLineRepository{db: db}
}

func (r *PurchaseLineRepository) Create(pll *model.PurchaseLineInsertDTO) (int64, error) {
	query := `
		INSERT INTO purchase_lines
			(event_id, item_id, quantity, unit_price)
		VALUES 
			(?, ?, ?, ?)
	`
	
	res, err := r.db.Conn.Exec(
		query,
		pll.EventID,
		pll.ItemID,
		pll.Quantity,
		pll.UnitPrice
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

func (r *PurchaseLineRepository) GetByID(id int64) (*model.PurchaseLine, error) {
	query := `
		SELECT 
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at 
		FROM purchase_lines
		WHERE id = ?
	`
	
	pll := &model.PurchaseLine{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&pll.ID,
		&pll.EventID,
		&pll.ItemID,
		&pll.Quantity,
		&pll.UnitPrice,
		&pll.CreatedAt
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return pll, err
}

func (r *PurchaseLineRepository) GetAll() ([]*model.PurchaseLine, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at
		FROM purchase_lines
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plls := []*model.PurchaseLine{}
	for rows.Next() {
		pll := &model.PurchaseLine{}
		if err := rows.Scan(
			&pll.ID,
			&pll.EventID,
			&pll.ItemID,
			&pll.Quantity,
			&pll.UnitPrice,
			&pll.CreatedAt
		); err != nil {
			return nil, err
		}
		plls = append(plls, pll)
	}
	return plls, rows.Err()
}

func (r *PurchaseLineRepository) GetAllByEventID(eventID int64) ([]*model.PurchaseLine, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at
		FROM purchase_lines
		WHERE event_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Conn.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plls := []*model.PurchaseLine{}
	for rows.Next() {
		pll := &model.PurchaseLine{}
		if err := rows.Scan(
			&pll.ID,
			&pll.EventID,
			&pll.ItemID,
			&pll.Quantity,
			&pll.UnitPrice,
			&pll.CreatedAt
		); err != nil {
			return nil, err
		}
		plls = append(plls, pll)
	}
	return plls, rows.Err()
}

func (r *PurchaseLineRepository) GetAllByItemID(itemId int64) ([]*model.PurchaseLine, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at
		FROM purchase_lines
		WHERE item_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Conn.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plls := []*model.PurchaseLine{}
	for rows.Next() {
		pll := &model.PurchaseLine{}
		if err := rows.Scan(
			&pll.ID,
			&pll.EventID,
			&pll.ItemID,
			&pll.Quantity,
			&pll.UnitPrice,
			&pll.CreatedAt
		); err != nil {
			return nil, err
		}
		plls = append(plls, pll)
	}
	return plls, rows.Err()
}

func (r *PurchaseLineRepository) Delete(id int64) error {
	query := `DELETE FROM purchase_lines WHERE id = ?`
	_, err := r.db.Conn.Exec(query, id)
	return err
}