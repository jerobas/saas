package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	recipedomain "github.com/jerobas/saas/internal/domain/recipe"
)

type RecipeStore interface {
	GetRecipe(ctx context.Context, id domain.RecipeID) (recipedomain.Recipe, error)
	GetRecipeRevision(ctx context.Context, id domain.RecipeRevisionID) (recipedomain.Revision, error)
	ListRecipeRevisions(ctx context.Context, recipeID domain.RecipeID) ([]recipedomain.Revision, error)
	ListRecipes(ctx context.Context, input RecipeListInput) (RecipePage, error)
	CreateRecipe(ctx context.Context, input recipeCreateStoreInput) (recipedomain.Recipe, error)
	PublishRecipeRevision(ctx context.Context, input recipePublishStoreInput) (recipedomain.Revision, error)
	RenameRecipe(ctx context.Context, input recipeRenameStoreInput) (recipedomain.Recipe, error)
	ArchiveRecipe(ctx context.Context, input recipeArchiveStoreInput) (recipedomain.Recipe, error)
	RestoreRecipe(ctx context.Context, input recipeRestoreStoreInput) (recipedomain.Recipe, error)
}

type RecipeCursor struct {
	Name domain.UniqueName
	ID   domain.RecipeID
}

type RecipeListInput struct {
	Archive  domain.ArchiveFilter
	Search   domain.Option[domain.NonEmptyText]
	After    domain.Option[RecipeCursor]
	PageSize int
}

type RecipePage struct {
	items []recipedomain.RecipeSummary
	next  domain.Option[RecipeCursor]
}

func NewRecipePage(items []recipedomain.RecipeSummary, next domain.Option[RecipeCursor]) RecipePage {
	cloned := make([]recipedomain.RecipeSummary, len(items))
	copy(cloned, items)
	return RecipePage{items: cloned, next: next}
}

func (p RecipePage) Items() []recipedomain.RecipeSummary {
	items := make([]recipedomain.RecipeSummary, len(p.items))
	copy(items, p.items)
	return items
}

func (p RecipePage) Next() domain.Option[RecipeCursor] { return p.next }

type RecipeComponentSourceKind string

const (
	RecipeComponentSourceUnit      RecipeComponentSourceKind = "UNIT"
	RecipeComponentSourcePackaging RecipeComponentSourceKind = "PACKAGING"
)

type RecipeComponentSource struct {
	Kind        RecipeComponentSourceKind
	Unit        domain.Option[domain.UnitCode]
	PackagingID domain.Option[domain.PackagingID]
}

func NewRecipeUnitComponentSource(unit domain.UnitCode) RecipeComponentSource {
	return RecipeComponentSource{Kind: RecipeComponentSourceUnit, Unit: domain.Some(unit)}
}

func NewRecipePackagingComponentSource(packagingID domain.PackagingID) RecipeComponentSource {
	return RecipeComponentSource{Kind: RecipeComponentSourcePackaging, PackagingID: domain.Some(packagingID)}
}

type RecipeComponentInput struct {
	Order    domain.ComponentOrder
	ItemID   domain.ItemID
	Quantity domain.AtomicQuantity
	Source   RecipeComponentSource
}

type RecipeRevisionWriteInput struct {
	StandardYield       domain.AtomicQuantity
	Instructions        string
	PreparationTime     domain.PreparationMinutes
	EstimatedDirectCost domain.Option[domain.InventoryValue]
	Components          []RecipeComponentInput
}

type RecipeCreateInput struct {
	Name         domain.UniqueName
	OutputItemID domain.ItemID
	Revision     RecipeRevisionWriteInput
}

type RecipePublishRevisionInput struct {
	RecipeID               domain.RecipeID
	ExpectedLatestRevision domain.RevisionNumber
	ExpectedUpdatedAt      domain.UTCInstant
	Revision               RecipeRevisionWriteInput
}

type RecipeRenameInput struct {
	ID                domain.RecipeID
	Name              domain.UniqueName
	ExpectedUpdatedAt domain.UTCInstant
}

type RecipeArchiveInput struct {
	ID                domain.RecipeID
	ExpectedUpdatedAt domain.UTCInstant
}

type RecipeRestoreInput struct {
	ID                domain.RecipeID
	ExpectedUpdatedAt domain.UTCInstant
}

type recipeCreateStoreInput struct {
	RecipeCreateInput
	CreatedAt domain.UTCInstant
}

type recipePublishStoreInput struct {
	RecipePublishRevisionInput
	RevisionCreatedAt domain.UTCInstant
}

type recipeRenameStoreInput struct {
	RecipeRenameInput
	UpdatedAt domain.UTCInstant
}

type recipeArchiveStoreInput struct {
	RecipeArchiveInput
	ArchivedAt domain.UTCInstant
}

type recipeRestoreStoreInput struct {
	RecipeRestoreInput
	UpdatedAt domain.UTCInstant
}

type RecipeService struct {
	store RecipeStore
	clock Clock
}

func NewRecipeService(store RecipeStore, clock Clock) *RecipeService {
	if store == nil {
		panic("recipe service requires a store")
	}
	if clock == nil {
		panic("recipe service requires a clock")
	}
	return &RecipeService{store: store, clock: clock}
}

func (s *RecipeService) GetRecipe(ctx context.Context, id domain.RecipeID) (recipedomain.Recipe, error) {
	value, err := s.store.GetRecipe(ctx, id)
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("get recipe: %w", err)
	}
	return value, nil
}

func (s *RecipeService) GetRecipeRevision(ctx context.Context, id domain.RecipeRevisionID) (recipedomain.Revision, error) {
	value, err := s.store.GetRecipeRevision(ctx, id)
	if err != nil {
		return recipedomain.Revision{}, fmt.Errorf("get recipe revision: %w", err)
	}
	return value, nil
}

func (s *RecipeService) ListRecipeRevisions(ctx context.Context, recipeID domain.RecipeID) ([]recipedomain.Revision, error) {
	values, err := s.store.ListRecipeRevisions(ctx, recipeID)
	if err != nil {
		return nil, fmt.Errorf("list recipe revisions: %w", err)
	}
	return values, nil
}

func (s *RecipeService) ListRecipes(ctx context.Context, input RecipeListInput) (RecipePage, error) {
	page, err := s.store.ListRecipes(ctx, input)
	if err != nil {
		return RecipePage{}, fmt.Errorf("list recipes: %w", err)
	}
	return page, nil
}

func (s *RecipeService) CreateRecipe(ctx context.Context, input RecipeCreateInput) (recipedomain.Recipe, error) {
	now, err := s.clock.Now()
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("read clock: %w", err)
	}
	value, err := s.store.CreateRecipe(ctx, recipeCreateStoreInput{
		RecipeCreateInput: input,
		CreatedAt:         now,
	})
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("create recipe: %w", err)
	}
	if !value.CreatedAt().Equal(now) || !value.UpdatedAt().Equal(now) || !value.CurrentRevision().CreatedAt().Equal(now) {
		return recipedomain.Recipe{}, domain.ErrInvariant
	}
	return value, nil
}

func (s *RecipeService) PublishRecipeRevision(ctx context.Context, input RecipePublishRevisionInput) (recipedomain.Revision, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return recipedomain.Revision{}, fmt.Errorf("read clock: %w", err)
	}
	value, err := s.store.PublishRecipeRevision(ctx, recipePublishStoreInput{
		RecipePublishRevisionInput: input,
		RevisionCreatedAt:          now,
	})
	if err != nil {
		return recipedomain.Revision{}, fmt.Errorf("publish recipe revision: %w", err)
	}
	if !value.CreatedAt().Equal(now) {
		return recipedomain.Revision{}, domain.ErrInvariant
	}
	return value, nil
}

func (s *RecipeService) RenameRecipe(ctx context.Context, input RecipeRenameInput) (recipedomain.Recipe, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("read clock: %w", err)
	}
	value, err := s.store.RenameRecipe(ctx, recipeRenameStoreInput{RecipeRenameInput: input, UpdatedAt: now})
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("rename recipe: %w", err)
	}
	if !value.UpdatedAt().Equal(now) {
		return recipedomain.Recipe{}, domain.ErrInvariant
	}
	return value, nil
}

func (s *RecipeService) ArchiveRecipe(ctx context.Context, input RecipeArchiveInput) (recipedomain.Recipe, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("read clock: %w", err)
	}
	value, err := s.store.ArchiveRecipe(ctx, recipeArchiveStoreInput{RecipeArchiveInput: input, ArchivedAt: now})
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("archive recipe: %w", err)
	}
	archivedAt, ok := value.ArchivedAt().Get()
	if !ok || !archivedAt.Equal(now) {
		return recipedomain.Recipe{}, domain.ErrInvariant
	}
	return value, nil
}

func (s *RecipeService) RestoreRecipe(ctx context.Context, input RecipeRestoreInput) (recipedomain.Recipe, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("read clock: %w", err)
	}
	value, err := s.store.RestoreRecipe(ctx, recipeRestoreStoreInput{RecipeRestoreInput: input, UpdatedAt: now})
	if err != nil {
		return recipedomain.Recipe{}, fmt.Errorf("restore recipe: %w", err)
	}
	if value.IsArchived() || !value.UpdatedAt().Equal(now) {
		return recipedomain.Recipe{}, domain.ErrInvariant
	}
	return value, nil
}
