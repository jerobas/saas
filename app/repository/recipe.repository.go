package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type RecipeRepository struct {
	db Executor
}

func NewRecipeRepository(db Executor) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (r *RecipeRepository) Create(rcp *model.RecipeInsertDTO) (int64, error) {
	query := `
		INSERT INTO recipes
			(name, output_item_id, preparation_time_minutes, instructions, standard_yield_quantity)
		VALUES
			(?, ?, ?, ?, ?)
	`

	res, err := r.db.Exec(
		query,
		rcp.Name,
		rcp.OutputItemID,
		rcp.PreparationTimeMinutes,
		rcp.Instructions,
		rcp.StandardYieldQuantity,
	)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *RecipeRepository) GetByID(id int64) (*model.Recipe, error) {
	query := `
		SELECT
			id,
			name,
			output_item_id,
			preparation_time_minutes,
			instructions,
			standard_yield_quantity,
			created_at
		FROM recipes
		WHERE id = ?
	`

	rcp := &model.Recipe{}
	err := r.db.QueryRow(query, id).Scan(
		&rcp.ID,
		&rcp.Name,
		&rcp.OutputItemID,
		&rcp.PreparationTimeMinutes,
		&rcp.Instructions,
		&rcp.StandardYieldQuantity,
		&rcp.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return rcp, nil
}

func (r *RecipeRepository) GetAll() ([]*model.Recipe, error) {
	query := `
		SELECT
			id,
			name,
			output_item_id,
			preparation_time_minutes,
			instructions,
			standard_yield_quantity,
			created_at
		FROM recipes
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rcps := []*model.Recipe{}
	for rows.Next() {
		rcp := &model.Recipe{}
		if err := rows.Scan(
			&rcp.ID,
			&rcp.Name,
			&rcp.OutputItemID,
			&rcp.PreparationTimeMinutes,
			&rcp.Instructions,
			&rcp.StandardYieldQuantity,
			&rcp.CreatedAt,
		); err != nil {
			return nil, err
		}
		rcps = append(rcps, rcp)
	}

	return rcps, rows.Err()
}

func (r *RecipeRepository) GetAllByOutputID(outputID int64) ([]*model.Recipe, error) {
	query := `
		SELECT
			id,
			name,
			output_item_id,
			preparation_time_minutes,
			instructions,
			standard_yield_quantity,
			created_at
		FROM recipes
		WHERE output_item_id = ?
		ORDER BY name ASC
	`

	rows, err := r.db.Query(query, outputID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rcps := []*model.Recipe{}
	for rows.Next() {
		rcp := &model.Recipe{}
		if err := rows.Scan(
			&rcp.ID,
			&rcp.Name,
			&rcp.OutputItemID,
			&rcp.PreparationTimeMinutes,
			&rcp.Instructions,
			&rcp.StandardYieldQuantity,
			&rcp.CreatedAt,
		); err != nil {
			return nil, err
		}
		rcps = append(rcps, rcp)
	}

	return rcps, rows.Err()
}

func (r *RecipeRepository) Update(recipeID int64, recipe *model.RecipeInsertDTO) error {
	_, err := r.db.Exec(`
		UPDATE recipes
		SET name = ?, output_item_id = ?, preparation_time_minutes = ?,
			instructions = ?, standard_yield_quantity = ?
		WHERE id = ?
	`, recipe.Name, recipe.OutputItemID, recipe.PreparationTimeMinutes,
		recipe.Instructions, recipe.StandardYieldQuantity, recipeID)
	return err
}

func (r *RecipeRepository) Delete(id int64) error {
	query := `DELETE FROM recipes WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
