package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type RecipeComponentRepository struct {
	db *Database
}

func NewRecipeComponentRepository(db *Database) *RecipeComponentRepository {
	return &RecipeComponentRepository{db: db}
}

func (r *RecipeComponentRepository) Create(cpn *model.RecipeComponentInsertDTO) (int64, error) {
	query := `
		INSERT INTO recipe_components
			(recipe_id, item_id, quantity)
		VALUES
			(?, ?, ?)
	`

	res, err := r.db.Conn.Exec(
		query,
		cpn.RecipeID,
		cpn.ItemID,
		cpn.Quantity
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

func (r *RecipeComponentRepository) GetByID(id int64) (*model.RecipeComponent, error) {
	query := `
		SELECT
			id,
			recipe_id,
			item_id,
			quantity,
			created_at
		FROM recipe_components
		WHERE id = ?
	`

	cpn := &model.RecipeComponent{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&cpn.ID,
		&cpn.RecipeID,
		&cpn.ItemID,
		&cpn.Quantity,
		&cpn.CreatedAt
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return cpn, nil
}

func (r *RecipeComponentRepository) GetAll() ([]*model.RecipeComponent, error) {
	query := `
		SELECT
			id,
			recipe_id,
			item_id,
			quantity,
			created_at
		FROM recipe_components
		ORDER BY recipe_id ASC
	`

	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cpns := []*model.RecipeComponent{}
	for rows.Next() {
		cpn := &model.RecipeComponent{}
		if err := rows.Scan(
			&cpn.ID,
			&cpn.RecipeID,
			&cpn.ItemID,
			&cpn.Quantity,
			&cpn.CreatedAt
		); err != nil {
			return nil, err
		}
		cpns = append(cpns, cpn)
	}

	return cpns, rows.Err()
}

func (r *RecipeComponentRepository) GetAllByRecipeID(recipeID int64) ([]*model.RecipeComponent, error) {
	query := `
		SELECT
			id,
			recipe_id,
			item_id,
			quantity,
			created_at
		FROM recipe_components
		WHERE recipe_id = ?
		ORDER BY item_id ASC
	`

	rows, err := r.db.Conn.Query(query, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cpns := []*model.RecipeComponent{}
	for rows.Next() {
		cpn := &model.RecipeComponent{}
		if err := rows.Scan(
			&cpn.ID,
			&cpn.RecipeID,
			&cpn.ItemID,
			&cpn.Quantity,
			&cpn.CreatedAt
		); err != nil {
			return nil, err
		}
		cpns = append(cpns, cpn)
	}

	return cpns, rows.Err()
}

func (r *RecipeComponentRepository) GetAllByItemID(itemID int64) ([]*model.RecipeComponent, error) {
	query := `
		SELECT
			id,
			recipe_id,
			item_id,
			quantity,
			created_at
		FROM recipe_components
		WHERE item_id = ?
		ORDER BY recipe_id ASC
	`

	rows, err := r.db.Conn.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cpns := []*model.RecipeComponent{}
	for rows.Next() {
		cpn := &model.RecipeComponent{}
		if err := rows.Scan(
			&cpn.ID,
			&cpn.RecipeID,
			&cpn.ItemID,
			&cpn.Quantity,
			&cpn.CreatedAt
		); err != nil {
			return nil, err
		}
		cpns = append(cpns, cpn)
	}

	return cpns, rows.Err()
}

func (r *RecipeComponentRepository) Delete(id int64) error {
	query := `DELETE FROM recipe_components WHERE id = ?`
	_, err := r.db.Conn.Exec(query, id)
	return err
}