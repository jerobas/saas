package service

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/jerobas/saas/model"
	"github.com/jerobas/saas/repository"
)

type RecipeService struct {
	db *Database
}

func NewRecipeService(db *Database) *RecipeService {
	return &RecipeService{db: db}
}

func (s *RecipeService) CreateRecipe(name string, ingredients []RecipeIngredientInput) (*RecipeDTO, error) {
	if len(ingredients) == 0 {
		return nil, errors.New("recipe requires at least one ingredient")
	}
	tx, err := s.db.Conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	itemRepo := repository.NewItemRepository(tx)
	outputID, err := itemRepo.Create(&model.ItemInsertDTO{
		Name:             "Produto: " + name,
		Unit:             "un",
		Producible:       1,
		DefaultSalePrice: sql.NullInt64{},
	})
	if err != nil {
		return nil, err
	}
	recipeRepo := repository.NewRecipeRepository(tx)
	recipeID, err := recipeRepo.Create(&model.RecipeInsertDTO{
		Name:                   name,
		OutputItemID:           outputID,
		PreparationTimeMinutes: 0,
		Instructions:           "",
		StandardYieldQuantity:  1,
	})
	if err != nil {
		return nil, err
	}
	if err := createRecipeComponents(repository.NewRecipeComponentRepository(tx), recipeID, ingredients); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetRecipe(strconv.FormatInt(recipeID, 10))
}

func (s *RecipeService) GetRecipe(id string) (*RecipeDTO, error) {
	recipeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}
	recipe, err := repository.NewRecipeRepository(s.db.Conn).GetByID(recipeID)
	if err != nil {
		return nil, err
	}
	components, err := repository.NewRecipeComponentRepository(s.db.Conn).GetAllByRecipeID(recipeID)
	if err != nil {
		return nil, err
	}
	return toRecipeDTO(recipe, components), nil
}

func (s *RecipeService) GetAllRecipes() ([]*RecipeDTO, error) {
	recipes, err := repository.NewRecipeRepository(s.db.Conn).GetAll()
	if err != nil {
		return nil, err
	}
	componentRepo := repository.NewRecipeComponentRepository(s.db.Conn)
	result := make([]*RecipeDTO, 0, len(recipes))
	for _, recipe := range recipes {
		components, err := componentRepo.GetAllByRecipeID(recipe.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, toRecipeDTO(recipe, components))
	}
	return result, nil
}

func (s *RecipeService) UpdateRecipe(id, name string, ingredients []RecipeIngredientInput) error {
	recipeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	tx, err := s.db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	recipeRepo := repository.NewRecipeRepository(tx)
	current, err := recipeRepo.GetByID(recipeID)
	if err != nil {
		return err
	}
	if err := recipeRepo.Update(recipeID, &model.RecipeInsertDTO{
		Name:                   name,
		OutputItemID:           current.OutputItemID,
		PreparationTimeMinutes: current.PreparationTimeMinutes,
		Instructions:           current.Instructions,
		StandardYieldQuantity:  current.StandardYieldQuantity,
	}); err != nil {
		return err
	}
	componentRepo := repository.NewRecipeComponentRepository(tx)
	if err := componentRepo.DeleteByRecipeID(recipeID); err != nil {
		return err
	}
	if err := createRecipeComponents(componentRepo, recipeID, ingredients); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *RecipeService) DeleteRecipe(id string) error {
	recipeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	tx, err := s.db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := repository.NewRecipeComponentRepository(tx).DeleteByRecipeID(recipeID); err != nil {
		return err
	}
	if err := repository.NewRecipeRepository(tx).Delete(recipeID); err != nil {
		return err
	}
	return tx.Commit()
}

func createRecipeComponents(repo *repository.RecipeComponentRepository, recipeID int64, ingredients []RecipeIngredientInput) error {
	for _, ingredient := range ingredients {
		itemID, err := strconv.ParseInt(ingredient.ItemID, 10, 64)
		if err != nil {
			return err
		}
		if ingredient.Quantity <= 0 {
			return errors.New("ingredient quantity must be greater than zero")
		}
		if _, err := repo.Create(&model.RecipeComponentInsertDTO{
			RecipeID: recipeID,
			ItemID:   itemID,
			Quantity: ingredient.Quantity,
		}); err != nil {
			return err
		}
	}
	return nil
}

func toRecipeDTO(recipe *model.Recipe, components []*model.RecipeComponent) *RecipeDTO {
	result := &RecipeDTO{
		ID:          strconv.FormatInt(recipe.ID, 10),
		Name:        recipe.Name,
		Ingredients: make([]RecipeIngredientInput, 0, len(components)),
		CreatedAt:   recipe.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	for _, component := range components {
		result.Ingredients = append(result.Ingredients, RecipeIngredientInput{
			ItemID:   strconv.FormatInt(component.ItemID, 10),
			Quantity: component.Quantity,
		})
	}
	return result
}
