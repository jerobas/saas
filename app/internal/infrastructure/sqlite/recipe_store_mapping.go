package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	recipedomain "github.com/jerobas/saas/internal/domain/recipe"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

func loadRecipeAggregate(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id domain.RecipeID,
) (recipedomain.Recipe, error) {
	row, err := queries.GetCurrentRecipe(ctx, id.Int64())
	if errors.Is(err, sql.ErrNoRows) {
		if _, headerErr := queries.GetRecipe(ctx, id.Int64()); headerErr == nil {
			return recipedomain.Recipe{}, domain.Corrupt(domain.Invalid("current_revision", domain.ViolationRequired, "REC-002"))
		} else if !errors.Is(headerErr, sql.ErrNoRows) {
			return recipedomain.Recipe{}, headerErr
		}
	}
	if err != nil {
		return recipedomain.Recipe{}, err
	}
	if row.RecipeID != id.Int64() {
		return recipedomain.Recipe{}, domain.Corrupt(domain.ErrInvariant)
	}
	if err := validateRecipeRevisionChain(
		row.RevisionNumber,
		row.RevisionCount,
		row.MinimumRevisionNumber,
	); err != nil {
		return recipedomain.Recipe{}, domain.Corrupt(err)
	}
	revision, err := mapRecipeRevisionWithComponents(ctx, queries, sqlcgen.RecipeRevision{
		ID: row.RevisionID, RecipeID: row.RecipeID, RevisionNumber: row.RevisionNumber,
		StandardYieldQuantityAtomic: row.StandardYieldQuantityAtomic,
		Instructions:                row.Instructions, PreparationTimeMinutes: row.PreparationTimeMinutes,
		EstimatedDirectCostMicro: row.EstimatedDirectCostMicro,
		CreatedAtMs:              row.RevisionCreatedAtMs,
	})
	if err != nil {
		return recipedomain.Recipe{}, err
	}
	idValue, err := domain.NewRecipeID(row.RecipeID)
	if err != nil {
		return recipedomain.Recipe{}, domain.Corrupt(err)
	}
	name, err := domain.RestoreUniqueName(row.RecipeName, row.RecipeNormalizedName)
	if err != nil {
		return recipedomain.Recipe{}, err
	}
	if name.Display() != row.RecipeName {
		return recipedomain.Recipe{}, domain.Corrupt(domain.Invalid("recipe_name", domain.ViolationInvariant, "CAT-005"))
	}
	outputItemID, err := domain.NewItemID(row.OutputItemID)
	if err != nil {
		return recipedomain.Recipe{}, domain.Corrupt(err)
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.RecipeCreatedAtMs)
	if err != nil {
		return recipedomain.Recipe{}, domain.Corrupt(err)
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(row.RecipeUpdatedAtMs)
	if err != nil {
		return recipedomain.Recipe{}, domain.Corrupt(err)
	}
	archivedAt, err := restoreOptionalInstant(row.RecipeArchivedAtMs)
	if err != nil {
		return recipedomain.Recipe{}, domain.Corrupt(err)
	}
	value, err := recipedomain.New(recipedomain.Params{
		ID: idValue, Name: name, OutputItemID: outputItemID,
		CreatedAt: createdAt, UpdatedAt: updatedAt, ArchivedAt: archivedAt,
		CurrentRevision: revision,
	})
	if err != nil {
		return recipedomain.Recipe{}, domain.Corrupt(err)
	}
	return value, nil
}

func loadRecipeRevision(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id domain.RecipeRevisionID,
) (recipedomain.Revision, error) {
	row, err := queries.GetRecipeRevision(ctx, id.Int64())
	if err != nil {
		return recipedomain.Revision{}, err
	}
	if err := validateRecipeRevisionChain(
		row.LatestRevisionNumber,
		row.RevisionCount,
		row.MinimumRevisionNumber,
	); err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	return mapRecipeRevisionWithComponents(ctx, queries, sqlcgen.RecipeRevision{
		ID: row.ID, RecipeID: row.RecipeID, RevisionNumber: row.RevisionNumber,
		StandardYieldQuantityAtomic: row.StandardYieldQuantityAtomic,
		Instructions:                row.Instructions,
		PreparationTimeMinutes:      row.PreparationTimeMinutes,
		EstimatedDirectCostMicro:    row.EstimatedDirectCostMicro,
		CreatedAtMs:                 row.CreatedAtMs,
	})
}

func mapRecipeRevisionWithComponents(
	ctx context.Context,
	queries *sqlcgen.Queries,
	row sqlcgen.RecipeRevision,
) (recipedomain.Revision, error) {
	componentRows, err := queries.ListRecipeRevisionComponents(ctx, row.ID)
	if err != nil {
		return recipedomain.Revision{}, err
	}
	components := make([]recipedomain.Component, 0, len(componentRows))
	for _, componentRow := range componentRows {
		component, mapErr := mapRecipeComponent(componentRow)
		if mapErr != nil {
			return recipedomain.Revision{}, domain.Corrupt(mapErr)
		}
		components = append(components, component)
	}
	id, err := domain.NewRecipeRevisionID(row.ID)
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	recipeID, err := domain.NewRecipeID(row.RecipeID)
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	number, err := domain.NewRevisionNumber(row.RevisionNumber)
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	yield, err := domain.NewPositiveAtomicQuantity(row.StandardYieldQuantityAtomic)
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	preparationTime, err := domain.NewPreparationMinutes(row.PreparationTimeMinutes)
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	estimatedCost, err := restoreRecipeOptionalInventoryValue(row.EstimatedDirectCostMicro)
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.CreatedAtMs)
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	value, err := recipedomain.NewRevision(recipedomain.RevisionParams{
		ID: id, RecipeID: recipeID, Number: number, StandardYield: yield,
		Instructions: row.Instructions, PreparationTime: preparationTime,
		EstimatedDirectCost: estimatedCost, CreatedAt: createdAt, Components: components,
	})
	if err != nil {
		return recipedomain.Revision{}, domain.Corrupt(err)
	}
	return value, nil
}

func mapRecipeComponent(row sqlcgen.RecipeRevisionComponent) (recipedomain.Component, error) {
	id, err := domain.NewRecipeComponentID(row.ID)
	if err != nil {
		return recipedomain.Component{}, err
	}
	revisionID, err := domain.NewRecipeRevisionID(row.RecipeRevisionID)
	if err != nil {
		return recipedomain.Component{}, err
	}
	order, err := domain.NewComponentOrder(row.ComponentOrder)
	if err != nil {
		return recipedomain.Component{}, err
	}
	itemID, err := domain.NewItemID(row.ItemID)
	if err != nil {
		return recipedomain.Component{}, err
	}
	quantity, err := domain.NewPositiveAtomicQuantity(row.QuantityAtomic)
	if err != nil {
		return recipedomain.Component{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.EnteredUnitCode)
	if err != nil {
		return recipedomain.Component{}, err
	}
	if enteredUnit.String() != row.EnteredUnitCode {
		return recipedomain.Component{}, domain.Invalid("entered_unit_code", domain.ViolationInvariant, "UNIT-005")
	}
	packagingName, err := restoreOptionalText(row.EnteredPackagingName)
	if err != nil {
		return recipedomain.Component{}, err
	}
	conversion, err := domain.NewUnitConversion(row.ConversionNumeratorAtomic, row.ConversionDenominator)
	if err != nil {
		return recipedomain.Component{}, err
	}
	if conversion.NumeratorAtomic() != row.ConversionNumeratorAtomic || conversion.Denominator() != row.ConversionDenominator {
		return recipedomain.Component{}, domain.Invalid("component_conversion", domain.ViolationInvariant, "UNIT-003")
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.CreatedAtMs)
	if err != nil {
		return recipedomain.Component{}, err
	}
	return recipedomain.NewComponent(recipedomain.ComponentParams{
		ID: id, RevisionID: revisionID, Order: order, ItemID: itemID,
		Quantity: quantity, EnteredUnit: enteredUnit,
		EnteredPackagingName: packagingName, Conversion: conversion, CreatedAt: createdAt,
	})
}

func mapRecipeSummary(row sqlcgen.ListRecipesRow) (recipedomain.RecipeSummary, error) {
	id, err := domain.NewRecipeID(row.ID)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	name, err := domain.RestoreUniqueName(row.Name, row.NormalizedName)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	if name.Display() != row.Name {
		return recipedomain.RecipeSummary{}, domain.Invalid("recipe_name", domain.ViolationInvariant, "CAT-005")
	}
	outputItemID, err := domain.NewItemID(row.OutputItemID)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	outputItemName, err := domain.NewDisplayName(row.OutputItemName)
	if err != nil || outputItemName.String() != row.OutputItemName {
		if err == nil {
			err = domain.ErrInvariant
		}
		return recipedomain.RecipeSummary{}, err
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.CreatedAtMs)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(row.UpdatedAtMs)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	archivedAt, err := restoreOptionalInstant(row.ArchivedAtMs)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	if !row.CurrentRevisionID.Valid || !row.CurrentRevisionNumber.Valid || !row.CurrentStandardYieldQuantityAtomic.Valid {
		return recipedomain.RecipeSummary{}, domain.Invalid("current_revision", domain.ViolationRequired, "REC-002")
	}
	if err := validateRecipeRevisionChain(
		row.CurrentRevisionNumber.Int64,
		row.RevisionCount,
		row.MinimumRevisionNumber,
	); err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	revisionID, err := domain.NewRecipeRevisionID(row.CurrentRevisionID.Int64)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	revisionNumber, err := domain.NewRevisionNumber(row.CurrentRevisionNumber.Int64)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	standardYield, err := domain.NewPositiveAtomicQuantity(row.CurrentStandardYieldQuantityAtomic.Int64)
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	currentRevision, err := recipedomain.NewCurrentRevisionSummary(recipedomain.CurrentRevisionSummaryParams{
		ID: revisionID, Number: revisionNumber, StandardYield: standardYield,
	})
	if err != nil {
		return recipedomain.RecipeSummary{}, err
	}
	return recipedomain.NewRecipeSummary(recipedomain.RecipeSummaryParams{
		ID: id, Name: name, OutputItemID: outputItemID, OutputItemName: outputItemName,
		CreatedAt: createdAt, UpdatedAt: updatedAt, ArchivedAt: archivedAt,
		CurrentRevision: currentRevision,
	})
}

func validateRecipeRevisionChain(currentNumber, count, minimumNumber int64) error {
	if currentNumber <= 0 || count != currentNumber || minimumNumber != 1 {
		return domain.Invalid("revision_chain", domain.ViolationInvariant, "REC-005")
	}
	return nil
}

func recipeListQuery(filter RecipeListFilter) (sqlcgen.ListRecipesParams, error) {
	if filter.PageSize.value < 1 || filter.PageSize.value > maximumRecipePageSize {
		return sqlcgen.ListRecipesParams{}, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	archive, err := recipeArchiveFilter(filter.Archive)
	if err != nil {
		return sqlcgen.ListRecipesParams{}, err
	}
	query := sqlcgen.ListRecipesParams{
		ArchiveFilter: archive, LimitCount: filter.PageSize.value + 1,
	}
	if search, ok := filter.Search.Get(); ok {
		_, key, normalizeErr := domain.NormalizeDisplayAndKey(search.String())
		if normalizeErr != nil {
			return sqlcgen.ListRecipesParams{}, normalizeErr
		}
		query.SearchKey = key
	}
	if cursor, ok := filter.After.Get(); ok {
		if cursor.Name.Key() == "" || cursor.ID.IsZero() {
			return sqlcgen.ListRecipesParams{}, domain.Invalid("recipe_cursor", domain.ViolationInvalidFormat, "")
		}
		query.AfterNormalizedName = cursor.Name.Key()
		query.AfterID = cursor.ID.Int64()
	}
	return query, nil
}

func recipeArchiveFilter(filter domain.ArchiveFilter) (int64, error) {
	switch filter {
	case "", domain.ArchiveActive:
		return 0, nil
	case domain.ArchiveArchived:
		return 1, nil
	case domain.ArchiveAll:
		return 2, nil
	default:
		return 0, domain.Invalid("archive_filter", domain.ViolationInvalidEnum, "ARC-001")
	}
}

func classifyRecipeMutationMiss(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id domain.RecipeID,
	expected domain.UTCInstant,
	wantArchived bool,
) error {
	row, err := queries.GetRecipe(ctx, id.Int64())
	if err != nil {
		return err
	}
	if row.UpdatedAtMs != expected.UnixMilli() {
		return fmt.Errorf("%w: recipe version changed", domain.ErrStale)
	}
	if row.ArchivedAtMs.Valid != wantArchived {
		return fmt.Errorf("%w: recipe archive state changed", domain.ErrConflict)
	}
	return fmt.Errorf("%w: recipe update matched no row", domain.ErrConflict)
}

func restoreRecipeOptionalInventoryValue(value sql.NullInt64) (domain.Option[domain.InventoryValue], error) {
	if !value.Valid {
		return domain.None[domain.InventoryValue](), nil
	}
	amount, err := domain.NewInventoryValue(value.Int64)
	if err != nil {
		return domain.None[domain.InventoryValue](), err
	}
	return domain.Some(amount), nil
}
