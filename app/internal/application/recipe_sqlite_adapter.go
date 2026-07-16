package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/recipe"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteRecipeStore struct {
	store *sqlite.Store
}

func NewSQLiteRecipeStore(store *sqlite.Store) RecipeStore {
	if store == nil {
		panic("sqlite recipe store requires a store")
	}
	return &sqliteRecipeStore{store: store}
}

func (s *sqliteRecipeStore) GetRecipe(ctx context.Context, id domain.RecipeID) (recipe.Recipe, error) {
	return s.store.GetRecipe(ctx, id)
}

func (s *sqliteRecipeStore) GetRecipeRevision(ctx context.Context, id domain.RecipeRevisionID) (recipe.Revision, error) {
	return s.store.GetRecipeRevision(ctx, id)
}

func (s *sqliteRecipeStore) ListRecipeRevisions(ctx context.Context, recipeID domain.RecipeID) ([]recipe.Revision, error) {
	return s.store.ListRecipeRevisions(ctx, recipeID)
}

func (s *sqliteRecipeStore) ListRecipes(ctx context.Context, input RecipeListInput) (RecipePage, error) {
	pageSize, err := sqlite.NewRecipePageSize(input.PageSize)
	if err != nil {
		return RecipePage{}, err
	}
	after := domain.None[sqlite.RecipeCursor]()
	if cursor, ok := input.After.Get(); ok {
		after = domain.Some(sqlite.RecipeCursor{Name: cursor.Name, ID: cursor.ID})
	}
	page, err := s.store.ListRecipes(ctx, sqlite.RecipeListFilter{
		Archive:  input.Archive,
		Search:   input.Search,
		After:    after,
		PageSize: pageSize,
	})
	if err != nil {
		return RecipePage{}, err
	}
	next := domain.None[RecipeCursor]()
	if cursor, ok := page.Next().Get(); ok {
		next = domain.Some(RecipeCursor{Name: cursor.Name, ID: cursor.ID})
	}
	return NewRecipePage(page.Items(), next), nil
}

func (s *sqliteRecipeStore) CreateRecipe(ctx context.Context, input recipeCreateStoreInput) (recipe.Recipe, error) {
	revision, err := mapRecipeRevisionInput(input.Revision, input.CreatedAt)
	if err != nil {
		return recipe.Recipe{}, err
	}
	return s.store.CreateRecipe(ctx, sqlite.CreateRecipeInput{
		Name:         input.Name,
		OutputItemID: input.OutputItemID,
		Revision:     revision,
		CreatedAt:    input.CreatedAt,
	})
}

func (s *sqliteRecipeStore) PublishRecipeRevision(ctx context.Context, input recipePublishStoreInput) (recipe.Revision, error) {
	revision, err := mapRecipeRevisionInput(input.Revision, input.RevisionCreatedAt)
	if err != nil {
		return recipe.Revision{}, err
	}
	return s.store.PublishRecipeRevision(ctx, sqlite.PublishRecipeRevisionInput{
		RecipeID:               input.RecipeID,
		ExpectedLatestRevision: input.ExpectedLatestRevision,
		ExpectedUpdatedAt:      input.ExpectedUpdatedAt,
		Revision:               revision,
	})
}

func (s *sqliteRecipeStore) RenameRecipe(ctx context.Context, input recipeRenameStoreInput) (recipe.Recipe, error) {
	return s.store.RenameRecipe(ctx, sqlite.RenameRecipeInput{
		ID: input.ID, Name: input.Name, ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		UpdatedAt: input.UpdatedAt,
	})
}

func (s *sqliteRecipeStore) ArchiveRecipe(ctx context.Context, input recipeArchiveStoreInput) (recipe.Recipe, error) {
	return s.store.ArchiveRecipe(ctx, sqlite.ArchiveRecipeInput{
		ID: input.ID, ExpectedUpdatedAt: input.ExpectedUpdatedAt, ArchivedAt: input.ArchivedAt,
	})
}

func (s *sqliteRecipeStore) RestoreRecipe(ctx context.Context, input recipeRestoreStoreInput) (recipe.Recipe, error) {
	return s.store.RestoreRecipe(ctx, sqlite.RestoreRecipeInput{
		ID: input.ID, ExpectedUpdatedAt: input.ExpectedUpdatedAt, UpdatedAt: input.UpdatedAt,
	})
}

func mapRecipeRevisionInput(input RecipeRevisionWriteInput, createdAt domain.UTCInstant) (sqlite.RecipeRevisionInput, error) {
	components := make([]sqlite.RecipeComponentInput, 0, len(input.Components))
	for _, component := range input.Components {
		source, err := mapRecipeComponentSource(component.Source)
		if err != nil {
			return sqlite.RecipeRevisionInput{}, err
		}
		components = append(components, sqlite.RecipeComponentInput{
			Order: component.Order, ItemID: component.ItemID, Quantity: component.Quantity, Source: source,
		})
	}
	return sqlite.RecipeRevisionInput{
		StandardYield:       input.StandardYield,
		Instructions:        input.Instructions,
		PreparationTime:     input.PreparationTime,
		EstimatedDirectCost: input.EstimatedDirectCost,
		Components:          components,
		CreatedAt:           createdAt,
	}, nil
}

func mapRecipeComponentSource(source RecipeComponentSource) (sqlite.RecipeComponentSource, error) {
	switch source.Kind {
	case RecipeComponentSourceUnit:
		unit, ok := source.Unit.Get()
		if !ok {
			return sqlite.RecipeComponentSource{}, domain.Invalid("component_unit", domain.ViolationRequired, "UNIT-005")
		}
		return sqlite.NewRecipeUnitSource(unit)
	case RecipeComponentSourcePackaging:
		packagingID, ok := source.PackagingID.Get()
		if !ok {
			return sqlite.RecipeComponentSource{}, domain.Invalid("component_packaging", domain.ViolationRequired, "UNIT-005")
		}
		return sqlite.NewRecipePackagingSource(packagingID)
	default:
		return sqlite.RecipeComponentSource{}, fmt.Errorf("%w: component source kind %q", domain.ErrValidation, source.Kind)
	}
}
