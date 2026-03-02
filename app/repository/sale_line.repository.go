package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
	"github.com/jerobas/saas/database"
)

type Database = database.Database

type SaleLineRepository struct {
	db *Database
}

func NewSaleLineRepository(db *Database) *SaleLineRepository {
	return &SaleLineRepository{db: db}
}

func (r *SaleLineRepository) Create(sll *model.SaleLineInsertDTO) (int64, error) {
	query := `
		INSERT INTO sale_lines
			(event_id, item_id, quantity, unit_price)
		VALUES 
			(?, ?, ?, ?)
	`
	
	res, err := r.db.Conn.Exec(
		query,
		sll.EventID,
		sll.ItemID,
		sll.Quantity,
		sll.UnitPrice
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

func (r *SaleLineRepository) GetByID(id int64) (*model.SaleLine, error) {
	query := `
		SELECT 
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at 
		FROM sale_lines
		WHERE id = ?
	`
	
	sll := &model.SaleLine{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&sll.ID,
		&sll.EventID,
		&sll.ItemID,
		&sll.Quantity,
		&sll.UnitPrice,
		&sll.CreatedAt
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return sll, err
}

func (r *SaleLineRepository) GetAll() ([]*model.SaleLine, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at
		FROM sale_lines
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slls := []*model.SaleLine{}
	for rows.Next() {
		sll := &model.SaleLine{}
		if err := rows.Scan(
			&sll.ID,
			&sll.EventID,
			&sll.ItemID,
			&sll.Quantity,
			&sll.UnitPrice,
			&sll.CreatedAt
		); err != nil {
			return nil, err
		}
		slls = append(slls, sll)
	}
	return slls, rows.Err()
}

func (r *SaleLineRepository) GetAllByEventID(eventID int64) ([]*model.SaleLine, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at
		FROM sale_lines
		WHERE event_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Conn.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slls := []*model.SaleLine{}
	for rows.Next() {
		sll := &model.SaleLine{}
		if err := rows.Scan(
			&sll.ID,
			&sll.EventID,
			&sll.ItemID,
			&sll.Quantity,
			&sll.UnitPrice,
			&sll.CreatedAt
		); err != nil {
			return nil, err
		}
		slls = append(slls, sll)
	}
	return slls, rows.Err()
}

func (r *SaleLineRepository) GetAllByItemID(itemID int64) ([]*model.SaleLine, error) {
	query := `
		SELECT
			id,
			event_id,
			item_id,
			quantity,
			unit_price,
			created_at
		FROM sale_lines
		WHERE item_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Conn.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slls := []*model.SaleLine{}
	for rows.Next() {
		sll := &model.SaleLine{}
		if err := rows.Scan(
			&sll.ID,
			&sll.EventID,
			&sll.ItemID,
			&sll.Quantity,
			&sll.UnitPrice,
			&sll.CreatedAt
		); err != nil {
			return nil, err
		}
		slls = append(slls, sll)
	}
	return slls, rows.Err()
}

func (r *SaleLineRepository) Delete(id int64) error {
	query := `DELETE FROM sale_lines WHERE id = ?`
	_, err := r.db.Conn.Exec(query, id)
	return err
}