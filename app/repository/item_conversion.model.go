package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type ItemConversionRepository struct {
	db *Database
}

func NewItemConversionRepository(db *Database) *ItemConversionRepository {
	return &ItemConversionRepository{db: db}
}

func (r *ItemConversionRepository) Create(cnv *model.ItemConversionInsertDTO) (int64, error) {
	query := `
		INSERT INTO item_conversions
			(from_item_id, to_item_id, factor)
		VALUES
			(?, ?, ?)
	`

	res, err := r.db.Conn.Exec(
		query,
		cnv.FromItemID,
		cnv.ToItemID,
		cnv.Factor
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

func (r *ItemConversionRepository) GetByID(id int64) (*model.ItemConversion, error) {
	query := `
		SELECT 
			id,
			from_item_id,
			to_item_id,
			created_at 
		FROM item_conversions
		WHERE id = ?
	`
	
	cnv := &model.ItemConversion{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&cnv.ID,
		&cnv.FromItemID,
		&cnv.ToItemID,
		&cnv.Factor,
		&cnv.CreatedAt
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return cnv, err
}

func (r *ItemConversionRepository) GetAll() ([]*model.ItemConversion, error) {
	query := `
		SELECT 
			id,
			from_item_id,
			to_item_id,
			created_at 
		FROM item_conversions
		ORDER BY created_at DESC
	`

	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cnvs := []*model.ItemConversion{}
	for rows.Next() {
		cnv := &model.ItemConversion{}
		if err := rows.Scan(
			&cnv.ID,
			&cnv.FromItemID,
			&cnv.ToItemID,
			&cnv.Factor,
			&cnv.CreatedAt
		); err != nil {
			return nil, err
		}
		cnvs = append(cnvs, cnv)
	}
	return cnvs, rows.Err()
}

func (r *ItemConversionRepository) GetAllByFromID(fromID int64) ([]*model.ItemConversion, error) {
	query := `
		SELECT 
			id,
			from_item_id,
			to_item_id,
			created_at 
		FROM item_conversions
		WHERE from_item_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Conn.Query(query, fromID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cnvs := []*model.ItemConversion{}
	for rows.Next() {
		cnv := &model.ItemConversion{}
		if err := rows.Scan(
			&cnv.ID,
			&cnv.FromItemID,
			&cnv.ToItemID,
			&cnv.Factor,
			&cnv.CreatedAt
		); err != nil {
			return nil, err
		}
		cnvs = append(cnvs, cnv)
	}
	return cnvs, rows.Err()
}

func (r *ItemConversionRepository) GetAllByToID(toID int64) ([]*model.ItemConversion, error) {
	query := `
		SELECT 
			id,
			from_item_id,
			to_item_id,
			created_at 
		FROM item_conversions
		WHERE to_item_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Conn.Query(query, toID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cnvs := []*model.ItemConversion{}
	for rows.Next() {
		cnv := &model.ItemConversion{}
		if err := rows.Scan(
			&cnv.ID,
			&cnv.FromItemID,
			&cnv.ToItemID,
			&cnv.Factor,
			&cnv.CreatedAt
		); err != nil {
			return nil, err
		}
		cnvs = append(cnvs, cnv)
	}
	return cnvs, rows.Err()
}

func (r *ItemConversionRepository) Delete(id int64) error {
	query := `DELETE FROM item_conversions WHERE id = ?`
	_, err := r.db.Conn.Exec(query, id)
	return err
}