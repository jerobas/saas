package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type RecipeRepository struct {
	db *Database
}

func NewRecipeRepository(db *Database) *RecipeRepository {
	return &RecipeRepository{db: db}
}

func (r *RecipeRepository) Create(recipe *model.Recipe) error {
	query := `INSERT INTO recipes (id, name, profit_margin_percent) VALUES (?, ?, ?)`
	_, err := r.db.Conn.Exec(query, recipe.ID, recipe.Name, recipe.ProfitMarginPercent)
	return err
}

func (r *RecipeRepository) GetByID(id string) (*model.Recipe, error) {
	query := `SELECT id, name, profit_margin_percent, created_at FROM recipes WHERE id = ?`
	recipe := &model.Recipe{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&recipe.ID, &recipe.Name, &recipe.ProfitMarginPercent, &recipe.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return recipe, err
}

func (r *RecipeRepository) GetAll() ([]*model.Recipe, error) {
	query := `SELECT id, name, profit_margin_percent, created_at FROM recipes ORDER BY name`
	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := []*model.Recipe{}
	for rows.Next() {
		recipe := &model.Recipe{}
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.ProfitMarginPercent, &recipe.CreatedAt); err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	return recipes, rows.Err()
}

func (r *RecipeRepository) Update(recipe *model.Recipe) error {
	query := `UPDATE recipes SET name = ?, profit_margin_percent = ? WHERE id = ?`
	_, err := r.db.Conn.Exec(query, recipe.Name, recipe.ProfitMarginPercent, recipe.ID)
	return err
}

func (r *RecipeRepository) Delete(id string) error {
	_, err := r.db.Conn.Exec(`DELETE FROM recipes WHERE id = ?`, id)
	return err
}

func (r *RecipeRepository) AddIngredient(ingredient *model.RecipeIngredient) error {
	query := `INSERT INTO recipe_ingredients (recipe_id, item_id, quantity_needed) VALUES (?, ?, ?)`
	_, err := r.db.Conn.Exec(query, ingredient.RecipeID, ingredient.ItemID, ingredient.QuantityNeeded)
	return err
}

func (r *RecipeRepository) GetIngredients(recipeID string) ([]*model.RecipeIngredient, error) {
	query := `SELECT recipe_id, item_id, quantity_needed FROM recipe_ingredients WHERE recipe_id = ?`
	rows, err := r.db.Conn.Query(query, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ingredients := []*model.RecipeIngredient{}
	for rows.Next() {
		ing := &model.RecipeIngredient{}
		if err := rows.Scan(&ing.RecipeID, &ing.ItemID, &ing.QuantityNeeded); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ing)
	}
	return ingredients, rows.Err()
}

func (r *RecipeRepository) RemoveIngredient(recipeID, itemID string) error {
	_, err := r.db.Conn.Exec(`DELETE FROM recipe_ingredients WHERE recipe_id = ? AND item_id = ?`, recipeID, itemID)
	return err
}