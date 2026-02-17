package service

import (
	"github.com/google/uuid"
	"github.com/jerobas/saas/repository"
	"github.com/jerobas/saas/model"
)

type RecipeService struct {
	repo *repository.RecipeRepository
}

func NewRecipeService(db *Database) *RecipeService {
	return &RecipeService{repo: repository.NewRecipeRepository(db)}
}

func (s *RecipeService) CreateRecipe(name string, profitMargin float64) (*model.RecipeDTO, error) {
	recipe := &model.Recipe{
		ID:                  uuid.New().String(),
		Name:                name,
		ProfitMarginPercent: profitMargin,
	}
	if err := s.repo.Create(recipe); err != nil {
		return nil, err
	}
	
	created, err := s.repo.GetByID(recipe.ID)
	if err != nil {
		return nil, err
	}
	return model.ToRecipeDTO(created), nil
}

func (s *RecipeService) GetRecipe(id string) (*model.RecipeDTO, error) {
	recipe, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return model.ToRecipeDTO(recipe), nil
}

func (s *RecipeService) GetAllRecipes() ([]*model.RecipeDTO, error) {
	recipes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	return model.ToRecipeDTOList(recipes), nil
}

func (s *RecipeService) UpdateRecipe(id, name string, profitMargin float64) error {
	recipe := &model.Recipe{
		ID:                  id,
		Name:                name,
		ProfitMarginPercent: profitMargin,
	}
	return s.repo.Update(recipe)
}

func (s *RecipeService) DeleteRecipe(id string) error {
	return s.repo.Delete(id)
}

func (s *RecipeService) AddIngredient(recipeID, itemID string, quantity float64) error {
	ingredient := &model.RecipeIngredient{
		RecipeID:       recipeID,
		ItemID:         itemID,
		QuantityNeeded: quantity,
	}
	return s.repo.AddIngredient(ingredient)
}

func (s *RecipeService) GetIngredients(recipeID string) ([]*model.RecipeIngredient, error) {
	return s.repo.GetIngredients(recipeID)
}

func (s *RecipeService) RemoveIngredient(recipeID, itemID string) error {
	return s.repo.RemoveIngredient(recipeID, itemID)
}