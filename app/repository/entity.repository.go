package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type EntityRepository struct {
	db *Database
}

func NewEntityRepository(db *Database) *EntityRepository {
	return &EntityRepository{db: db}
}

func (r *EntityRepository) Create(ett *model.EntityInsertDTO) (int64, error) {
	query := `
		INSERT INTO entities
			(name, phone, email)
		VALUES
			(?, ?, ?)
	`

	res, err := r.db.Conn.Exec(
		query,
		ett.Name,
		ett.Phone,
		ett.Email,
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

func (r *EntityRepository) GetByID(id int64) (*model.Entity, error) {
	query := `
		SELECT 
			id,
			name,
			phone,
			email,
			created_at 
		FROM entities
		WHERE id = ?
	`
	
	ett := &model.Entity{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&ett.ID,
		&ett.Name,
		&ett.Phone,
		&ett.Email,
		&ett.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return ett, err
}

func (r *EntityRepository) GetAll() ([]*model.Entity, error) {
	query := `
		SELECT 
			id,
			name,
			phone,
			email,
			created_at
		FROM entities
		ORDER BY name ASC
	`

	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	etts := []*model.Entity{}
	for rows.Next() {
		ett := &model.Entity{}
		if err := rows.Scan(
			&ett.ID,
			&ett.Name,
			&ett.Phone,
			&ett.Email,
			&ett.CreatedAt,
		); err != nil {
			return nil, err
		}
		etts = append(etts, ett)
	}
	return etts, rows.Err()
}

// func (r *EntityRepository) Delete(id int64) error {
// 	query := `DELETE FROM entities WHERE id = ?`
// 	_, err := r.db.Conn.Exec(query, id)
// 	return err
// }