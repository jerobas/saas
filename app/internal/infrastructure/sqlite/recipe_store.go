package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	recipedomain "github.com/jerobas/saas/internal/domain/recipe"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

const maximumRecipePageSize = 100

type RecipePageSize struct{ value int64 }

func NewRecipePageSize(value int) (RecipePageSize, error) {
	if value < 1 || value > maximumRecipePageSize {
		return RecipePageSize{}, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	return RecipePageSize{value: int64(value)}, nil
}

func (s RecipePageSize) Int() int { return int(s.value) }

type RecipeCursor struct {
	Name domain.UniqueName
	ID   domain.RecipeID
}

type RecipeListFilter struct {
	Archive  domain.ArchiveFilter
	Search   domain.Option[domain.NonEmptyText]
	After    domain.Option[RecipeCursor]
	PageSize RecipePageSize
}

type RecipePage struct {
	items []recipedomain.RecipeSummary
	next  domain.Option[RecipeCursor]
}

func (p RecipePage) Items() []recipedomain.RecipeSummary {
	items := make([]recipedomain.RecipeSummary, len(p.items))
	copy(items, p.items)
	return items
}

func (p RecipePage) Next() domain.Option[RecipeCursor] { return p.next }

type recipeComponentSourceKind uint8

const (
	recipeComponentSourceUnit recipeComponentSourceKind = iota + 1
	recipeComponentSourcePackaging
)

// RecipeComponentSource is a closed choice between a controlled unit and an
// active item packaging. Its private fields prevent callers from providing an
// arbitrary historical conversion snapshot.
type RecipeComponentSource struct {
	kind        recipeComponentSourceKind
	unit        domain.UnitCode
	packagingID domain.PackagingID
}

func NewRecipeUnitSource(unit domain.UnitCode) (RecipeComponentSource, error) {
	if unit.String() == "" {
		return RecipeComponentSource{}, domain.Invalid("unit_code", domain.ViolationRequired, "UNIT-005")
	}
	return RecipeComponentSource{kind: recipeComponentSourceUnit, unit: unit}, nil
}

func NewRecipePackagingSource(packagingID domain.PackagingID) (RecipeComponentSource, error) {
	if packagingID.IsZero() {
		return RecipeComponentSource{}, domain.Invalid("packaging_id", domain.ViolationRequired, "UNIT-005")
	}
	return RecipeComponentSource{kind: recipeComponentSourcePackaging, packagingID: packagingID}, nil
}

type RecipeComponentInput struct {
	Order    domain.ComponentOrder
	ItemID   domain.ItemID
	Quantity domain.AtomicQuantity
	Source   RecipeComponentSource
}

type RecipeRevisionInput struct {
	StandardYield       domain.AtomicQuantity
	Instructions        string
	PreparationTime     domain.PreparationMinutes
	EstimatedDirectCost domain.Option[domain.InventoryValue]
	Components          []RecipeComponentInput
	CreatedAt           domain.UTCInstant
}

type CreateRecipeInput struct {
	Name         domain.UniqueName
	OutputItemID domain.ItemID
	Revision     RecipeRevisionInput
	CreatedAt    domain.UTCInstant
}

type PublishRecipeRevisionInput struct {
	RecipeID               domain.RecipeID
	ExpectedLatestRevision domain.RevisionNumber
	ExpectedUpdatedAt      domain.UTCInstant
	Revision               RecipeRevisionInput
}

type RenameRecipeInput struct {
	ID                domain.RecipeID
	Name              domain.UniqueName
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

type ArchiveRecipeInput struct {
	ID                domain.RecipeID
	ExpectedUpdatedAt domain.UTCInstant
	ArchivedAt        domain.UTCInstant
}

type RestoreRecipeInput struct {
	ID                domain.RecipeID
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

type recipeReadStage uint8

const recipeRevisionRowsLoaded recipeReadStage = 1

func (s *Store) GetRecipe(ctx context.Context, id domain.RecipeID) (recipedomain.Recipe, error) {
	if id.IsZero() {
		return recipedomain.Recipe{}, domain.Invalid("recipe_id", domain.ViolationRequired, "")
	}
	var value recipedomain.Recipe
	err := s.withReadQueries(ctx, "get recipe", func(queries *sqlcgen.Queries) error {
		var err error
		value, err = loadRecipeAggregate(ctx, queries, id)
		return err
	})
	return value, err
}

func (s *Store) GetRecipeRevision(ctx context.Context, id domain.RecipeRevisionID) (recipedomain.Revision, error) {
	if id.IsZero() {
		return recipedomain.Revision{}, domain.Invalid("recipe_revision_id", domain.ViolationRequired, "")
	}
	var value recipedomain.Revision
	err := s.withReadQueries(ctx, "get recipe revision", func(queries *sqlcgen.Queries) error {
		var err error
		value, err = loadRecipeRevision(ctx, queries, id)
		return err
	})
	return value, err
}

func (s *Store) ListRecipeRevisions(ctx context.Context, recipeID domain.RecipeID) ([]recipedomain.Revision, error) {
	if recipeID.IsZero() {
		return nil, domain.Invalid("recipe_id", domain.ViolationRequired, "")
	}
	var result []recipedomain.Revision
	err := s.withReadQueries(ctx, "list recipe revisions", func(queries *sqlcgen.Queries) error {
		var err error
		result, err = s.listRecipeRevisions(ctx, queries, recipeID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) listRecipeRevisions(
	ctx context.Context,
	queries *sqlcgen.Queries,
	recipeID domain.RecipeID,
) ([]recipedomain.Revision, error) {
	rows, err := queries.ListRecipeRevisions(ctx, recipeID.Int64())
	if err != nil {
		return nil, err
	}
	if s.recipeReadHook != nil {
		if err := s.recipeReadHook(recipeRevisionRowsLoaded); err != nil {
			return nil, err
		}
	}
	if len(rows) == 0 {
		if _, headerErr := queries.GetRecipe(ctx, recipeID.Int64()); headerErr != nil {
			return nil, headerErr
		}
		return nil, domain.Corrupt(domain.ErrInvariant)
	}
	result := make([]recipedomain.Revision, 0, len(rows))
	for index, row := range rows {
		expectedNumber := int64(len(rows) - index)
		if row.RevisionNumber != expectedNumber || row.RecipeID != recipeID.Int64() {
			return nil, domain.Corrupt(domain.Invalid("revision_number", domain.ViolationInvariant, "REC-005"))
		}
		value, mapErr := mapRecipeRevisionWithComponents(ctx, queries, row)
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, value)
	}
	return result, nil
}

func (s *Store) ListRecipes(ctx context.Context, filter RecipeListFilter) (RecipePage, error) {
	query, err := recipeListQuery(filter)
	if err != nil {
		return RecipePage{}, err
	}
	rows, err := s.queries.ListRecipes(ctx, query)
	if err != nil {
		return RecipePage{}, classifyError("list recipes", err)
	}
	hasMore := int64(len(rows)) > filter.PageSize.value
	if hasMore {
		rows = rows[:filter.PageSize.value]
	}
	items := make([]recipedomain.RecipeSummary, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapRecipeSummary(row)
		if mapErr != nil {
			return RecipePage{}, corruptDataError("map recipe summary", mapErr)
		}
		items = append(items, item)
	}
	next := domain.None[RecipeCursor]()
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		next = domain.Some(RecipeCursor{Name: last.Name(), ID: last.ID()})
	}
	return RecipePage{items: items, next: next}, nil
}

func (s *Store) CreateRecipe(ctx context.Context, input CreateRecipeInput) (recipedomain.Recipe, error) {
	if err := validateCreateRecipeInput(input); err != nil {
		return recipedomain.Recipe{}, err
	}
	var created recipedomain.Recipe
	err := s.withWriteQueries(ctx, "create recipe", func(queries *sqlcgen.Queries) error {
		if _, err := loadActiveProducibleRecipeItem(ctx, queries, input.OutputItemID); err != nil {
			return err
		}
		components, err := resolveRecipeComponents(ctx, queries, input.OutputItemID, input.Revision.Components)
		if err != nil {
			return err
		}
		recipeIDValue, err := queries.InsertRecipe(ctx, sqlcgen.InsertRecipeParams{
			Name: input.Name.Display(), NormalizedName: input.Name.Key(),
			OutputItemID: input.OutputItemID.Int64(), CreatedAtMs: input.CreatedAt.UnixMilli(),
			UpdatedAtMs: input.Revision.CreatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		recipeID, err := domain.NewRecipeID(recipeIDValue)
		if err != nil {
			return corruptDataError("map created recipe id", err)
		}
		revisionID, err := insertRecipeRevision(ctx, queries, recipeID, 1, 0, input.Revision, components)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: revision one insert matched no version", domain.ErrConflict)
		}
		if err != nil {
			return err
		}
		if revisionID.IsZero() {
			return corruptDataError("map created recipe revision", domain.ErrInvariant)
		}
		created, err = loadRecipeAggregate(ctx, queries, recipeID)
		return err
	})
	return created, err
}

func (s *Store) PublishRecipeRevision(ctx context.Context, input PublishRecipeRevisionInput) (recipedomain.Revision, error) {
	if err := validatePublishRecipeRevisionInput(input); err != nil {
		return recipedomain.Revision{}, err
	}
	var published recipedomain.Revision
	err := s.withWriteQueries(ctx, "publish recipe revision", func(queries *sqlcgen.Queries) error {
		current, err := loadRecipeAggregate(ctx, queries, input.RecipeID)
		if err != nil {
			return err
		}
		if !current.UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: recipe aggregate version changed", domain.ErrStale)
		}
		if current.CurrentRevision().Number() != input.ExpectedLatestRevision {
			return fmt.Errorf("%w: latest recipe revision changed", domain.ErrStale)
		}
		if current.IsArchived() {
			return fmt.Errorf("%w: archived recipe cannot publish", domain.ErrConflict)
		}
		if input.Revision.CreatedAt.Compare(current.CurrentRevision().CreatedAt()) <= 0 {
			return domain.Invalid("revision_created_at", domain.ViolationInvariant, "REC-005")
		}
		if _, err := loadActiveProducibleRecipeItem(ctx, queries, current.OutputItemID()); err != nil {
			return err
		}
		components, err := resolveRecipeComponents(ctx, queries, current.OutputItemID(), input.Revision.Components)
		if err != nil {
			return err
		}
		expected := input.ExpectedLatestRevision.Int64()
		if expected == math.MaxInt64 {
			return domain.ErrOverflow
		}
		rows, err := queries.AdvanceRecipeVersion(ctx, sqlcgen.AdvanceRecipeVersionParams{
			UpdatedAtMs: input.Revision.CreatedAt.UnixMilli(), ID: input.RecipeID.Int64(),
			ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyRecipeMutationMiss(ctx, queries, input.RecipeID, input.ExpectedUpdatedAt, false)
		}
		revisionID, err := insertRecipeRevision(ctx, queries, input.RecipeID, expected+1, expected, input.Revision, components)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: latest recipe revision changed", domain.ErrStale)
		}
		if err != nil {
			return err
		}
		published, err = loadRecipeRevision(ctx, queries, revisionID)
		return err
	})
	return published, err
}

func (s *Store) RenameRecipe(ctx context.Context, input RenameRecipeInput) (recipedomain.Recipe, error) {
	if input.ID.IsZero() {
		return recipedomain.Recipe{}, domain.Invalid("recipe_id", domain.ViolationRequired, "")
	}
	if input.Name.Display() == "" || input.Name.Key() == "" {
		return recipedomain.Recipe{}, domain.Invalid("name", domain.ViolationRequired, "CAT-005")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return recipedomain.Recipe{}, err
	}
	var renamed recipedomain.Recipe
	err := s.withWriteQueries(ctx, "rename recipe", func(queries *sqlcgen.Queries) error {
		current, err := loadRecipeAggregate(ctx, queries, input.ID)
		if err != nil {
			return err
		}
		if !current.UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: recipe version changed", domain.ErrStale)
		}
		if current.IsArchived() {
			return fmt.Errorf("%w: archived recipe cannot be renamed", domain.ErrConflict)
		}
		if _, err := recipedomain.New(recipedomain.Params{
			ID: current.ID(), Name: input.Name, OutputItemID: current.OutputItemID(),
			CreatedAt: current.CreatedAt(), UpdatedAt: input.UpdatedAt,
			ArchivedAt: domain.None[domain.UTCInstant](), CurrentRevision: current.CurrentRevision(),
		}); err != nil {
			return err
		}
		rows, err := queries.RenameRecipe(ctx, sqlcgen.RenameRecipeParams{
			Name: input.Name.Display(), NormalizedName: input.Name.Key(),
			UpdatedAtMs: input.UpdatedAt.UnixMilli(), ID: input.ID.Int64(),
			ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyRecipeMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, false)
		}
		renamed, err = loadRecipeAggregate(ctx, queries, input.ID)
		return err
	})
	return renamed, err
}

func (s *Store) ArchiveRecipe(ctx context.Context, input ArchiveRecipeInput) (recipedomain.Recipe, error) {
	if input.ID.IsZero() {
		return recipedomain.Recipe{}, domain.Invalid("recipe_id", domain.ViolationRequired, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.ArchivedAt); err != nil {
		return recipedomain.Recipe{}, err
	}
	var archived recipedomain.Recipe
	err := s.withWriteQueries(ctx, "archive recipe", func(queries *sqlcgen.Queries) error {
		current, err := loadRecipeAggregate(ctx, queries, input.ID)
		if err != nil {
			return err
		}
		if !current.UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: recipe version changed", domain.ErrStale)
		}
		if current.IsArchived() {
			return fmt.Errorf("%w: recipe is already archived", domain.ErrConflict)
		}
		if _, err := recipedomain.New(recipedomain.Params{
			ID: current.ID(), Name: current.Name(), OutputItemID: current.OutputItemID(),
			CreatedAt: current.CreatedAt(), UpdatedAt: input.ArchivedAt,
			ArchivedAt: domain.Some(input.ArchivedAt), CurrentRevision: current.CurrentRevision(),
		}); err != nil {
			return err
		}
		rows, err := queries.ArchiveRecipe(ctx, sqlcgen.ArchiveRecipeParams{
			ArchivedAtMs: input.ArchivedAt.UnixMilli(), UpdatedAtMs: input.ArchivedAt.UnixMilli(),
			ID: input.ID.Int64(), ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyRecipeMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, false)
		}
		archived, err = loadRecipeAggregate(ctx, queries, input.ID)
		return err
	})
	return archived, err
}

func (s *Store) RestoreRecipe(ctx context.Context, input RestoreRecipeInput) (recipedomain.Recipe, error) {
	if input.ID.IsZero() {
		return recipedomain.Recipe{}, domain.Invalid("recipe_id", domain.ViolationRequired, "")
	}
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return recipedomain.Recipe{}, err
	}
	var restored recipedomain.Recipe
	err := s.withWriteQueries(ctx, "restore recipe", func(queries *sqlcgen.Queries) error {
		current, err := loadRecipeAggregate(ctx, queries, input.ID)
		if err != nil {
			return err
		}
		if !current.UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: recipe version changed", domain.ErrStale)
		}
		if !current.IsArchived() {
			return fmt.Errorf("%w: recipe is already active", domain.ErrConflict)
		}
		if err := validateRecipeRestoreEligibility(ctx, queries, current); err != nil {
			return err
		}
		if _, err := recipedomain.New(recipedomain.Params{
			ID: current.ID(), Name: current.Name(), OutputItemID: current.OutputItemID(),
			CreatedAt: current.CreatedAt(), UpdatedAt: input.UpdatedAt,
			ArchivedAt: domain.None[domain.UTCInstant](), CurrentRevision: current.CurrentRevision(),
		}); err != nil {
			return err
		}
		rows, err := queries.RestoreRecipe(ctx, sqlcgen.RestoreRecipeParams{
			UpdatedAtMs: input.UpdatedAt.UnixMilli(), ID: input.ID.Int64(),
			ExpectedUpdatedAtMs: input.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return classifyRecipeMutationMiss(ctx, queries, input.ID, input.ExpectedUpdatedAt, true)
		}
		restored, err = loadRecipeAggregate(ctx, queries, input.ID)
		return err
	})
	return restored, err
}

type resolvedRecipeComponent struct {
	input                RecipeComponentInput
	enteredUnit          domain.UnitCode
	enteredPackagingName domain.Option[domain.NonEmptyText]
	conversion           domain.UnitConversion
}

func resolveRecipeComponents(
	ctx context.Context,
	queries *sqlcgen.Queries,
	outputItemID domain.ItemID,
	inputs []RecipeComponentInput,
) ([]resolvedRecipeComponent, error) {
	result := make([]resolvedRecipeComponent, 0, len(inputs))
	for _, input := range inputs {
		if input.ItemID == outputItemID {
			return nil, domain.Invalid("components.item_id", domain.ViolationInvariant, "REC-004")
		}
		item, err := loadActiveRecipeComponentItem(ctx, queries, input.ItemID)
		if err != nil {
			return nil, err
		}
		resolved := resolvedRecipeComponent{input: input}
		switch input.Source.kind {
		case recipeComponentSourceUnit:
			if input.Source.unit.String() == "" {
				return nil, domain.Invalid("component_unit", domain.ViolationRequired, "UNIT-005")
			}
			unit, err := loadRequiredUnit(ctx, queries, input.Source.unit)
			if err != nil {
				return nil, err
			}
			if err := catalog.ValidateCompatibleDimensions(item.BaseUnit().Dimension(), unit.Dimension()); err != nil {
				return nil, err
			}
			resolved.enteredUnit = unit.Code()
			resolved.enteredPackagingName = domain.None[domain.NonEmptyText]()
			resolved.conversion = unit.Conversion()
		case recipeComponentSourcePackaging:
			if input.Source.packagingID.IsZero() {
				return nil, domain.Invalid("component_packaging", domain.ViolationRequired, "UNIT-005")
			}
			packaging, err := loadPackagingAggregate(ctx, queries, input.Source.packagingID.Int64())
			if errors.Is(err, sql.ErrNoRows) {
				return nil, wrapClassifiedError("load recipe packaging", domain.ErrInvalidReference, err)
			}
			if err != nil {
				return nil, recipePersistedReferenceError(err)
			}
			if packaging.Packaging().ItemID() != input.ItemID || packaging.Packaging().IsArchived() {
				return nil, fmt.Errorf("%w: recipe packaging is unavailable for component", domain.ErrInvalidReference)
			}
			if err := catalog.ValidateCompatibleDimensions(item.BaseUnit().Dimension(), packaging.EnteredUnit().Dimension()); err != nil {
				return nil, err
			}
			name, err := domain.NewNonEmptyText(packaging.Packaging().Name().Display())
			if err != nil {
				return nil, corruptDataError("map recipe packaging name", err)
			}
			resolved.enteredUnit = packaging.Packaging().EnteredUnit()
			resolved.enteredPackagingName = domain.Some(name)
			resolved.conversion = packaging.Packaging().Conversion()
		default:
			return nil, domain.Invalid("component_source", domain.ViolationRequired, "UNIT-005")
		}
		result = append(result, resolved)
	}
	return result, nil
}

func insertRecipeRevision(
	ctx context.Context,
	queries *sqlcgen.Queries,
	recipeID domain.RecipeID,
	revisionNumber int64,
	expectedLatest int64,
	input RecipeRevisionInput,
	components []resolvedRecipeComponent,
) (domain.RecipeRevisionID, error) {
	revisionIDValue, err := queries.InsertRecipeRevision(ctx, sqlcgen.InsertRecipeRevisionParams{
		RecipeID: recipeID.Int64(), RevisionNumber: revisionNumber,
		StandardYieldQuantityAtomic: input.StandardYield.Int64(), Instructions: input.Instructions,
		PreparationTimeMinutes:   input.PreparationTime.Int64(),
		EstimatedDirectCostMicro: recipeNullableInventoryValue(input.EstimatedDirectCost),
		CreatedAtMs:              input.CreatedAt.UnixMilli(), ExpectedLatestRevisionNumber: expectedLatest,
	})
	if err != nil {
		return domain.RecipeRevisionID{}, err
	}
	revisionID, err := domain.NewRecipeRevisionID(revisionIDValue)
	if err != nil {
		return domain.RecipeRevisionID{}, corruptDataError("map inserted recipe revision id", err)
	}
	for _, component := range components {
		componentID, err := queries.InsertRecipeRevisionComponent(ctx, sqlcgen.InsertRecipeRevisionComponentParams{
			RecipeRevisionID: revisionID.Int64(), ComponentOrder: component.input.Order.Int64(),
			ItemID: component.input.ItemID.Int64(), QuantityAtomic: component.input.Quantity.Int64(),
			EnteredUnitCode:           component.enteredUnit.String(),
			EnteredPackagingName:      recipeNullableText(component.enteredPackagingName),
			ConversionNumeratorAtomic: component.conversion.NumeratorAtomic(),
			ConversionDenominator:     component.conversion.Denominator(),
			CreatedAtMs:               input.CreatedAt.UnixMilli(),
		})
		if err != nil {
			return domain.RecipeRevisionID{}, err
		}
		if _, err := domain.NewRecipeComponentID(componentID); err != nil {
			return domain.RecipeRevisionID{}, corruptDataError("map inserted recipe component id", err)
		}
	}
	return revisionID, nil
}

func loadActiveProducibleRecipeItem(ctx context.Context, queries *sqlcgen.Queries, id domain.ItemID) (ItemAggregate, error) {
	item, err := loadRecipeReferencedItem(ctx, queries, id)
	if err != nil {
		return ItemAggregate{}, err
	}
	if item.Item().IsArchived() || !item.Item().Capabilities().Producible() {
		return ItemAggregate{}, fmt.Errorf("%w: recipe output must be active and producible", domain.ErrInvalidReference)
	}
	return item, nil
}

func loadActiveRecipeComponentItem(ctx context.Context, queries *sqlcgen.Queries, id domain.ItemID) (ItemAggregate, error) {
	item, err := loadRecipeReferencedItem(ctx, queries, id)
	if err != nil {
		return ItemAggregate{}, err
	}
	if item.Item().IsArchived() {
		return ItemAggregate{}, fmt.Errorf("%w: recipe component item is archived", domain.ErrInvalidReference)
	}
	return item, nil
}

func loadRecipeReferencedItem(ctx context.Context, queries *sqlcgen.Queries, id domain.ItemID) (ItemAggregate, error) {
	if id.IsZero() {
		return ItemAggregate{}, domain.Invalid("item_id", domain.ViolationRequired, "")
	}
	item, err := loadItemAggregate(ctx, queries, id.Int64())
	if errors.Is(err, sql.ErrNoRows) {
		return ItemAggregate{}, wrapClassifiedError("load recipe item", domain.ErrInvalidReference, err)
	}
	if err != nil {
		return ItemAggregate{}, recipePersistedReferenceError(err)
	}
	return item, nil
}

func recipePersistedReferenceError(err error) error {
	if errors.Is(err, domain.ErrCorruptData) {
		return err
	}
	if errors.Is(err, domain.ErrValidation) || errors.Is(err, domain.ErrInvariant) {
		return domain.Corrupt(err)
	}
	return err
}

func validateRecipeRestoreEligibility(ctx context.Context, queries *sqlcgen.Queries, current recipedomain.Recipe) error {
	if _, err := loadActiveProducibleRecipeItem(ctx, queries, current.OutputItemID()); err != nil {
		return err
	}
	for _, component := range current.CurrentRevision().Components() {
		item, err := loadActiveRecipeComponentItem(ctx, queries, component.ItemID())
		if err != nil {
			return err
		}
		unit, err := loadRequiredUnit(ctx, queries, component.EnteredUnit())
		if err != nil {
			return err
		}
		if err := catalog.ValidateCompatibleDimensions(item.BaseUnit().Dimension(), unit.Dimension()); err != nil {
			return err
		}
	}
	return nil
}

func validateCreateRecipeInput(input CreateRecipeInput) error {
	violations := make([]domain.Violation, 0, 5)
	if input.Name.Display() == "" || input.Name.Key() == "" {
		violations = append(violations, domain.Violation{Field: "name", Code: domain.ViolationRequired})
	}
	if input.OutputItemID.IsZero() {
		violations = append(violations, domain.Violation{Field: "output_item_id", Code: domain.ViolationRequired})
	}
	if input.CreatedAt.IsZero() {
		violations = append(violations, domain.Violation{Field: "created_at", Code: domain.ViolationRequired})
	}
	if !input.Revision.CreatedAt.IsZero() && !input.CreatedAt.IsZero() && input.Revision.CreatedAt.Before(input.CreatedAt) {
		violations = append(violations, domain.Violation{Field: "revision_created_at", Code: domain.ViolationInvariant, InvariantID: "REC-002"})
	}
	if err := validateRecipeRevisionInput(input.Revision); err != nil {
		if validation, ok := err.(*domain.ValidationError); ok {
			violations = append(violations, validation.Violations()...)
		} else {
			return err
		}
	}
	return domain.NewValidationError(violations...)
}

func validatePublishRecipeRevisionInput(input PublishRecipeRevisionInput) error {
	violations := make([]domain.Violation, 0, 4)
	if input.RecipeID.IsZero() {
		violations = append(violations, domain.Violation{Field: "recipe_id", Code: domain.ViolationRequired})
	}
	if input.ExpectedLatestRevision.IsZero() {
		violations = append(violations, domain.Violation{Field: "expected_latest_revision", Code: domain.ViolationRequired, InvariantID: "REC-005"})
	}
	if input.ExpectedUpdatedAt.IsZero() {
		violations = append(violations, domain.Violation{Field: "expected_updated_at", Code: domain.ViolationRequired, InvariantID: "REC-005"})
	} else if !input.Revision.CreatedAt.IsZero() && input.Revision.CreatedAt.Compare(input.ExpectedUpdatedAt) <= 0 {
		violations = append(violations, domain.Violation{Field: "revision_created_at", Code: domain.ViolationInvariant, InvariantID: "REC-005"})
	}
	if err := validateRecipeRevisionInput(input.Revision); err != nil {
		if validation, ok := err.(*domain.ValidationError); ok {
			violations = append(violations, validation.Violations()...)
		} else {
			return err
		}
	}
	return domain.NewValidationError(violations...)
}

func validateRecipeRevisionInput(input RecipeRevisionInput) error {
	violations := make([]domain.Violation, 0, 8)
	if input.StandardYield.Int64() <= 0 {
		violations = append(violations, domain.Violation{Field: "standard_yield", Code: domain.ViolationNotPositive, InvariantID: "REC-003"})
	}
	if !utf8.ValidString(input.Instructions) {
		violations = append(violations, domain.Violation{Field: "instructions", Code: domain.ViolationInvalidFormat})
	}
	if input.CreatedAt.IsZero() {
		violations = append(violations, domain.Violation{Field: "revision_created_at", Code: domain.ViolationRequired})
	}
	if len(input.Components) == 0 {
		violations = append(violations, domain.Violation{Field: "components", Code: domain.ViolationRequired, InvariantID: "REC-002"})
	}
	seenOrders := make(map[int64]struct{}, len(input.Components))
	seenItems := make(map[int64]struct{}, len(input.Components))
	for _, component := range input.Components {
		if component.Order.IsZero() {
			violations = append(violations, domain.Violation{Field: "components.order", Code: domain.ViolationNotPositive, InvariantID: "REC-002"})
		}
		if component.ItemID.IsZero() {
			violations = append(violations, domain.Violation{Field: "components.item_id", Code: domain.ViolationRequired})
		}
		if component.Quantity.Int64() <= 0 {
			violations = append(violations, domain.Violation{Field: "components.quantity", Code: domain.ViolationNotPositive, InvariantID: "REC-003"})
		}
		if component.Source.kind != recipeComponentSourceUnit && component.Source.kind != recipeComponentSourcePackaging {
			violations = append(violations, domain.Violation{Field: "components.source", Code: domain.ViolationRequired, InvariantID: "UNIT-005"})
		} else if component.Source.kind == recipeComponentSourceUnit && component.Source.unit.String() == "" {
			violations = append(violations, domain.Violation{Field: "components.unit", Code: domain.ViolationRequired, InvariantID: "UNIT-005"})
		} else if component.Source.kind == recipeComponentSourcePackaging && component.Source.packagingID.IsZero() {
			violations = append(violations, domain.Violation{Field: "components.packaging", Code: domain.ViolationRequired, InvariantID: "UNIT-005"})
		}
		if _, found := seenOrders[component.Order.Int64()]; found {
			violations = append(violations, domain.Violation{Field: "components.order", Code: domain.ViolationDuplicate, InvariantID: "REC-002"})
		}
		seenOrders[component.Order.Int64()] = struct{}{}
		if _, found := seenItems[component.ItemID.Int64()]; found {
			violations = append(violations, domain.Violation{Field: "components.item_id", Code: domain.ViolationDuplicate, InvariantID: "REC-002"})
		}
		seenItems[component.ItemID.Int64()] = struct{}{}
	}
	return domain.NewValidationError(violations...)
}

func recipeNullableInventoryValue(value domain.Option[domain.InventoryValue]) sql.NullInt64 {
	amount, ok := value.Get()
	return sql.NullInt64{Int64: amount.Int64(), Valid: ok}
}

func recipeNullableText(value domain.Option[domain.NonEmptyText]) sql.NullString {
	text, ok := value.Get()
	return sql.NullString{String: text.String(), Valid: ok}
}
