package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	recipedomain "github.com/jerobas/saas/internal/domain/recipe"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

func TestRecipeStoreCreateReadPublishAndImmutableHistory(t *testing.T) {
	store := recipeTestStore(t, "lifecycle.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Cake", false, true)
	componentID := recipeTestItem(t, store, "Flour", true, false)
	unitSource := recipeUnitSource(t, "g")

	created, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Chocolate cake"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 500),
		Revision: recipeRevisionInput(t, 1_000, "version one", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 500_000, unitSource),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID().IsZero() || created.CurrentRevision().Number().Int64() != 1 || created.UpdatedAt().UnixMilli() != 1_000 {
		t.Fatalf("created recipe = %#v", created)
	}
	if components := created.CurrentRevision().Components(); components == nil || len(components) != 1 {
		t.Fatalf("created components = %#v", components)
	}
	loaded, err := store.GetRecipe(ctx, created.ID())
	if err != nil || loaded.CurrentRevision().Instructions() != "version one" {
		t.Fatalf("loaded recipe = %#v, %v", loaded, err)
	}
	firstRevisionID := loaded.CurrentRevision().ID()

	published, err := store.PublishRecipeRevision(ctx, PublishRecipeRevisionInput{
		RecipeID: created.ID(), ExpectedLatestRevision: loaded.CurrentRevision().Number(),
		ExpectedUpdatedAt: loaded.UpdatedAt(),
		Revision: recipeRevisionInput(t, 2_000, "version two", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 600_000, unitSource),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if published.Number().Int64() != 2 || published.Instructions() != "version two" {
		t.Fatalf("published revision = %#v", published)
	}
	currentAfterPublish, err := store.GetRecipe(ctx, created.ID())
	if err != nil || currentAfterPublish.UpdatedAt().UnixMilli() != 2_000 || currentAfterPublish.CurrentRevision().Number().Int64() != 2 {
		t.Fatalf("published aggregate version = %#v, %v", currentAfterPublish, err)
	}
	_, err = store.RenameRecipe(ctx, RenameRecipeInput{
		ID: created.ID(), Name: recipeName(t, "Lost update"),
		ExpectedUpdatedAt: loaded.UpdatedAt(), UpdatedAt: recipeInstant(t, 2_500),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("rename across published revision error = %v, want ErrStale", err)
	}

	revisions, err := store.ListRecipeRevisions(ctx, created.ID())
	if err != nil {
		t.Fatal(err)
	}
	if revisions == nil || len(revisions) != 2 || revisions[0].Number().Int64() != 2 || revisions[1].Number().Int64() != 1 {
		t.Fatalf("revision order = %#v", revisions)
	}
	first, err := store.GetRecipeRevision(ctx, firstRevisionID)
	if err != nil || first.Instructions() != "version one" || first.Components()[0].Quantity().Int64() != 500_000 {
		t.Fatalf("immutable first revision = %#v, %v", first, err)
	}

	_, err = store.PublishRecipeRevision(ctx, PublishRecipeRevisionInput{
		RecipeID: created.ID(), ExpectedLatestRevision: currentAfterPublish.CurrentRevision().Number(),
		ExpectedUpdatedAt: loaded.UpdatedAt(),
		Revision: recipeRevisionInput(t, 3_000, "stale", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1, unitSource),
		}),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale aggregate CAS error = %v, want ErrStale", err)
	}
	_, err = store.PublishRecipeRevision(ctx, PublishRecipeRevisionInput{
		RecipeID: created.ID(), ExpectedLatestRevision: loaded.CurrentRevision().Number(),
		ExpectedUpdatedAt: currentAfterPublish.UpdatedAt(),
		Revision: recipeRevisionInput(t, 4_000, "stale revision", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1, unitSource),
		}),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale revision error = %v, want ErrStale", err)
	}
	revisions, err = store.ListRecipeRevisions(ctx, created.ID())
	if err != nil || len(revisions) != 2 {
		t.Fatalf("stale publish changed history: %d, %v", len(revisions), err)
	}

	if _, err := store.database.ExecContext(ctx, `UPDATE recipe_revisions SET instructions = 'mutated' WHERE id = ?`, firstRevisionID.Int64()); err == nil {
		t.Fatal("SQLite allowed mutation of an immutable revision")
	}
}

func TestRecipeStoreProtectsActiveOutputCatalogState(t *testing.T) {
	store := recipeTestStore(t, "active-output.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Protected output", true, true)
	componentID := recipeTestItem(t, store, "Protected input", true, false)
	recipeValue, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Protected recipe"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "v1", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1_000, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	output, err := store.GetItem(ctx, outputID)
	if err != nil {
		t.Fatal(err)
	}
	item := output.Item()

	_, err = store.UpdateItem(ctx, UpdateItemInput{
		ID: item.ID(), Name: item.Name(), SKU: item.SKU(), Description: item.Description(),
		BaseUnit: item.BaseUnit(), Capabilities: catalog.NewCapabilities(true, false, false),
		DefaultSalePrice: item.DefaultSalePrice(), ReorderQuantity: item.ReorderQuantity(),
		ExpectedUpdatedAt: item.UpdatedAt(), UpdatedAt: recipeInstant(t, 2_000),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("disable active recipe output error = %v, want ErrConflict", err)
	}
	_, err = store.ArchiveItem(ctx, ArchiveItemInput{
		ID: item.ID(), ExpectedUpdatedAt: item.UpdatedAt(), ArchivedAt: recipeInstant(t, 2_500),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("archive active recipe output error = %v, want ErrConflict", err)
	}

	archivedRecipe, err := store.ArchiveRecipe(ctx, ArchiveRecipeInput{
		ID: recipeValue.ID(), ExpectedUpdatedAt: recipeValue.UpdatedAt(), ArchivedAt: recipeInstant(t, 3_000),
	})
	if err != nil || !archivedRecipe.IsArchived() {
		t.Fatalf("archive recipe before output = %#v, %v", archivedRecipe, err)
	}
	archivedOutput, err := store.ArchiveItem(ctx, ArchiveItemInput{
		ID: item.ID(), ExpectedUpdatedAt: item.UpdatedAt(), ArchivedAt: recipeInstant(t, 4_000),
	})
	if err != nil || !archivedOutput.Item().IsArchived() {
		t.Fatalf("archive output after recipe = %#v, %v", archivedOutput.Item(), err)
	}
}

func TestRecipeStoreRejectsDiscontinuousRevisionChainBeforeMutation(t *testing.T) {
	store := recipeTestStore(t, "revision-gap.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Gap output", false, true)
	componentID := recipeTestItem(t, store, "Gap input", true, false)
	name := recipeName(t, "Gap recipe")
	recipeIDValue, err := store.queries.InsertRecipe(ctx, sqlcgen.InsertRecipeParams{
		Name: name.Display(), NormalizedName: name.Key(), OutputItemID: outputID.Int64(),
		CreatedAtMs: 1_000, UpdatedAtMs: 1_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.database.ExecContext(ctx, `
		INSERT INTO recipe_revisions (
			recipe_id, revision_number, standard_yield_quantity_atomic,
			instructions, preparation_time_minutes, created_at_ms
		) VALUES (?, 2, 1000, 'gap', 1, 1000)
	`, recipeIDValue); err == nil {
		t.Fatal("SQLite accepted a discontinuous recipe revision")
	}

	// Temporarily remove the version-two sequencing guard to model a damaged or
	// externally authored file and exercise adapter defense in depth.
	if _, err := store.database.ExecContext(ctx, `DROP TRIGGER recipe_revisions_require_next_number`); err != nil {
		t.Fatal(err)
	}
	result, err := store.database.ExecContext(ctx, `
		INSERT INTO recipe_revisions (
			recipe_id, revision_number, standard_yield_quantity_atomic,
			instructions, preparation_time_minutes, created_at_ms
		) VALUES (?, 2, 1000, 'gap', 1, 1000)
	`, recipeIDValue)
	if err != nil {
		t.Fatal(err)
	}
	revisionID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.queries.InsertRecipeRevisionComponent(ctx, sqlcgen.InsertRecipeRevisionComponentParams{
		RecipeRevisionID: revisionID, ComponentOrder: 1, ItemID: componentID.Int64(),
		QuantityAtomic: 1_000, EnteredUnitCode: "g", ConversionNumeratorAtomic: 1_000,
		ConversionDenominator: 1, CreatedAtMs: 1_000,
	}); err != nil {
		t.Fatal(err)
	}
	id := recipeID(t, recipeIDValue)
	if _, err := store.GetRecipe(ctx, id); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("GetRecipe discontinuous-chain error = %v, want ErrCorruptData", err)
	}
	corruptRevisionID, err := domain.NewRecipeRevisionID(revisionID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetRecipeRevision(ctx, corruptRevisionID); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("GetRecipeRevision discontinuous-chain error = %v, want ErrCorruptData", err)
	}
	if _, err := store.ListRecipes(ctx, RecipeListFilter{
		Archive: domain.ArchiveAll, PageSize: recipePageSize(t, 10),
	}); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("ListRecipes discontinuous-chain error = %v, want ErrCorruptData", err)
	}
	expectedLatest, err := domain.NewRevisionNumber(2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.PublishRecipeRevision(ctx, PublishRecipeRevisionInput{
		RecipeID: id, ExpectedLatestRevision: expectedLatest,
		ExpectedUpdatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 2_000, "must not publish", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1_000, recipeUnitSource(t, "g")),
		}),
	})
	if !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("publish discontinuous-chain error = %v, want ErrCorruptData", err)
	}
	var updatedAt, revisionCount, latestRevision int64
	if err := store.database.QueryRowContext(ctx, `
		SELECT recipe.updated_at_ms, COUNT(revision.id), MAX(revision.revision_number)
		FROM recipes recipe
		JOIN recipe_revisions revision ON revision.recipe_id = recipe.id
		WHERE recipe.id = ?
		GROUP BY recipe.id
	`, id.Int64()).Scan(&updatedAt, &revisionCount, &latestRevision); err != nil {
		t.Fatal(err)
	}
	if updatedAt != 1_000 || revisionCount != 1 || latestRevision != 2 {
		t.Fatalf("failed publish mutated corrupt chain: updated=%d count=%d latest=%d", updatedAt, revisionCount, latestRevision)
	}
}

func TestRecipeStoreCreationRollsBackHeaderRevisionAndComponents(t *testing.T) {
	store := recipeTestStore(t, "rollback.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Output", false, true)
	componentID := recipeTestItem(t, store, "Input", true, false)
	name := recipeName(t, "Rollback recipe")
	conversion := recipeConversion(t, 1000, 1)
	createdAt := recipeInstant(t, 1_000)
	order := recipeOrder(t, 1)
	quantity := recipeQuantity(t, 1000)
	input := RecipeRevisionInput{
		StandardYield: recipeQuantity(t, 1000), PreparationTime: recipePreparation(t, 1),
		CreatedAt: createdAt,
	}
	duplicate := resolvedRecipeComponent{
		input:       RecipeComponentInput{Order: order, ItemID: componentID, Quantity: quantity},
		enteredUnit: recipeUnit(t, "g"), enteredPackagingName: domain.None[domain.NonEmptyText](),
		conversion: conversion,
	}

	err := store.withWriteQueries(ctx, "create invalid recipe aggregate", func(queries *sqlcgen.Queries) error {
		idValue, err := queries.InsertRecipe(ctx, sqlcgen.InsertRecipeParams{
			Name: name.Display(), NormalizedName: name.Key(), OutputItemID: outputID.Int64(),
			CreatedAtMs: createdAt.UnixMilli(), UpdatedAtMs: createdAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		id, err := domain.NewRecipeID(idValue)
		if err != nil {
			return err
		}
		_, err = insertRecipeRevision(ctx, queries, id, 1, 0, input, []resolvedRecipeComponent{duplicate, duplicate})
		return err
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate component error = %v, want ErrConflict", err)
	}
	var count int
	if err := store.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM recipes WHERE normalized_name = ?`, name.Key()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rolled-back recipe headers = %d, want 0", count)
	}
}

func TestRecipeStoreUnicodeSearchConflictAndKeyset(t *testing.T) {
	store := recipeTestStore(t, "list.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Output", false, true)
	componentID := recipeTestItem(t, store, "Input", true, false)
	source := recipeUnitSource(t, "g")

	create := func(name string, at int64) recipedomain.Recipe {
		t.Helper()
		value, err := store.CreateRecipe(ctx, CreateRecipeInput{
			Name: recipeName(t, name), OutputItemID: outputID, CreatedAt: recipeInstant(t, at),
			Revision: recipeRevisionInput(t, at, name, []RecipeComponentInput{
				recipeComponentInput(t, 1, componentID, 1, source),
			}),
		})
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	create("\u00c9clair", 1_000)
	create("Mousse", 2_000)
	create("Zulu", 3_000)

	_, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "E\u0301CLAIR"), OutputItemID: outputID, CreatedAt: recipeInstant(t, 4_000),
		Revision: recipeRevisionInput(t, 4_000, "duplicate", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1, source),
		}),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Unicode-equivalent name error = %v, want ErrConflict", err)
	}

	searchPage, err := store.ListRecipes(ctx, RecipeListFilter{
		Archive: domain.ArchiveAll, Search: domain.Some(recipeText(t, "e\u0301C")),
		PageSize: recipePageSize(t, 10),
	})
	if err != nil {
		t.Fatal(err)
	}
	if items := searchPage.Items(); len(items) != 1 || items[0].Name().Display() != "\u00c9clair" {
		t.Fatalf("Unicode search results = %#v", items)
	}

	firstPage, err := store.ListRecipes(ctx, RecipeListFilter{Archive: domain.ArchiveAll, PageSize: recipePageSize(t, 2)})
	if err != nil {
		t.Fatal(err)
	}
	if len(firstPage.Items()) != 2 || firstPage.Next().IsNone() {
		t.Fatalf("first page = %#v, next=%#v", firstPage.Items(), firstPage.Next())
	}
	secondPage, err := store.ListRecipes(ctx, RecipeListFilter{
		Archive: domain.ArchiveAll, After: firstPage.Next(), PageSize: recipePageSize(t, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if items := secondPage.Items(); items == nil || len(items) != 1 || items[0].Name().Display() != "Éclair" {
		t.Fatalf("second page = %#v", items)
	}
}

func TestRecipeStoreDerivesUnitAndPackagingSnapshots(t *testing.T) {
	store := recipeTestStore(t, "snapshots.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Output", false, true)
	unitItemID := recipeTestItem(t, store, "Unit input", true, false)
	packagingItemID := recipeTestItem(t, store, "Packaged input", true, false)

	packagingIDValue, err := store.queries.InsertItemPackaging(ctx, sqlcgen.InsertItemPackagingParams{
		ItemID: packagingItemID.Int64(), Name: "5 kg sack", NormalizedName: "5 kg sack",
		EnteredUnitCode: "kg", ConversionNumeratorAtomic: 5_000_000,
		ConversionDenominator: 1, CreatedAtMs: 1_000, UpdatedAtMs: 1_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	packagingID := recipePackagingID(t, packagingIDValue)

	created, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Snapshot recipe"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 2_000),
		Revision: recipeRevisionInput(t, 2_000, "snapshots", []RecipeComponentInput{
			recipeComponentInput(t, 1, unitItemID, 1_000_000, recipeUnitSource(t, "kg")),
			recipeComponentInput(t, 2, packagingItemID, 5_000_000, recipePackagingSource(t, packagingID)),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	components := created.CurrentRevision().Components()
	if len(components) != 2 {
		t.Fatalf("components = %#v", components)
	}
	if components[0].EnteredUnit().String() != "kg" || components[0].Conversion().NumeratorAtomic() != 1_000_000 || components[0].EnteredPackagingName().IsSome() {
		t.Fatalf("controlled-unit snapshot = %#v", components[0])
	}
	packagingName, ok := components[1].EnteredPackagingName().Get()
	if !ok || packagingName.String() != "5 kg sack" || components[1].EnteredUnit().String() != "kg" || components[1].Conversion().NumeratorAtomic() != 5_000_000 {
		t.Fatalf("packaging snapshot = %#v", components[1])
	}

	otherItemID := recipeTestItem(t, store, "Other", true, false)
	_, err = store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Wrong packaging"), OutputItemID: outputID, CreatedAt: recipeInstant(t, 3_000),
		Revision: recipeRevisionInput(t, 3_000, "wrong", []RecipeComponentInput{
			recipeComponentInput(t, 1, otherItemID, 1, recipePackagingSource(t, packagingID)),
		}),
	})
	if !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("foreign packaging error = %v, want ErrInvalidReference", err)
	}
	_, err = store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Wrong dimension"), OutputItemID: outputID, CreatedAt: recipeInstant(t, 4_000),
		Revision: recipeRevisionInput(t, 4_000, "wrong unit", []RecipeComponentInput{
			recipeComponentInput(t, 1, unitItemID, 1, recipeUnitSource(t, "l")),
		}),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("incompatible unit error = %v, want validation", err)
	}
}

func TestRecipeStoreArchiveRestoreAndArchivedReferenceRules(t *testing.T) {
	store := recipeTestStore(t, "archive.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Output", false, true)
	componentID := recipeTestItem(t, store, "Input", true, false)
	source := recipeUnitSource(t, "g")
	created, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Archive recipe"), OutputItemID: outputID, CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "v1", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1, source),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	renamed, err := store.RenameRecipe(ctx, RenameRecipeInput{
		ID: created.ID(), Name: recipeName(t, "Renamed recipe"),
		ExpectedUpdatedAt: created.UpdatedAt(), UpdatedAt: recipeInstant(t, 1_500),
	})
	if err != nil || renamed.Name().Display() != "Renamed recipe" {
		t.Fatalf("renamed recipe = %#v, %v", renamed, err)
	}
	_, err = store.RenameRecipe(ctx, RenameRecipeInput{
		ID: created.ID(), Name: recipeName(t, "Stale rename"),
		ExpectedUpdatedAt: created.UpdatedAt(), UpdatedAt: recipeInstant(t, 1_750),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale rename error = %v, want ErrStale", err)
	}
	archived, err := store.ArchiveRecipe(ctx, ArchiveRecipeInput{
		ID: renamed.ID(), ExpectedUpdatedAt: renamed.UpdatedAt(), ArchivedAt: recipeInstant(t, 2_000),
	})
	if err != nil || !archived.IsArchived() {
		t.Fatalf("archived recipe = %#v, %v", archived, err)
	}
	activePage, err := store.ListRecipes(ctx, RecipeListFilter{
		Archive: domain.ArchiveActive, PageSize: recipePageSize(t, 10),
	})
	if err != nil || len(activePage.Items()) != 0 {
		t.Fatalf("active recipes after archive = %#v, %v", activePage.Items(), err)
	}
	archivedPage, err := store.ListRecipes(ctx, RecipeListFilter{
		Archive: domain.ArchiveArchived, PageSize: recipePageSize(t, 10),
	})
	if err != nil || len(archivedPage.Items()) != 1 {
		t.Fatalf("archived recipe page = %#v, %v", archivedPage.Items(), err)
	}
	_, err = store.RenameRecipe(ctx, RenameRecipeInput{
		ID: archived.ID(), Name: recipeName(t, "No rename"),
		ExpectedUpdatedAt: archived.UpdatedAt(), UpdatedAt: recipeInstant(t, 3_000),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("rename archived error = %v, want ErrConflict", err)
	}
	restored, err := store.RestoreRecipe(ctx, RestoreRecipeInput{
		ID: archived.ID(), ExpectedUpdatedAt: archived.UpdatedAt(), UpdatedAt: recipeInstant(t, 3_000),
	})
	if err != nil || restored.IsArchived() {
		t.Fatalf("restored recipe = %#v, %v", restored, err)
	}
	archived, err = store.ArchiveRecipe(ctx, ArchiveRecipeInput{
		ID: restored.ID(), ExpectedUpdatedAt: restored.UpdatedAt(), ArchivedAt: recipeInstant(t, 4_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.database.ExecContext(ctx, `UPDATE items SET archived_at_ms = 5000, updated_at_ms = 5000 WHERE id = ?`, componentID.Int64()); err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Archived component"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 5_500),
		Revision: recipeRevisionInput(t, 5_500, "invalid", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1, source),
		}),
	})
	if !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("create with archived component error = %v, want ErrInvalidReference", err)
	}
	_, err = store.RestoreRecipe(ctx, RestoreRecipeInput{
		ID: archived.ID(), ExpectedUpdatedAt: archived.UpdatedAt(), UpdatedAt: recipeInstant(t, 6_000),
	})
	if !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("restore with archived component error = %v, want ErrInvalidReference", err)
	}
	stillArchived, err := store.GetRecipe(ctx, archived.ID())
	if err != nil || !stillArchived.IsArchived() {
		t.Fatalf("failed restore changed recipe: %#v, %v", stillArchived, err)
	}
}

func TestRecipeStoreCorruptSnapshotsMissingRowsAndCancellation(t *testing.T) {
	store := recipeTestStore(t, "errors.db")
	ctx := context.Background()
	missingID := recipeID(t, 999)
	if _, err := store.GetRecipe(ctx, missingID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing recipe error = %v, want ErrNotFound", err)
	}

	outputID := recipeTestItem(t, store, "Output", false, true)
	componentID := recipeTestItem(t, store, "Input", true, false)
	created, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Corrupt recipe"), OutputItemID: outputID, CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "v1", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 1, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.database.ExecContext(ctx, `UPDATE recipes SET name = CAST(X'FF' AS TEXT) WHERE id = ?`, created.ID().Int64()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetRecipe(ctx, created.ID()); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("malformed recipe error = %v, want ErrCorruptData", err)
	}

	incompleteName := recipeName(t, "Incomplete")
	incompleteID, err := store.queries.InsertRecipe(ctx, sqlcgen.InsertRecipeParams{
		Name: incompleteName.Display(), NormalizedName: incompleteName.Key(), OutputItemID: outputID.Int64(),
		CreatedAtMs: 2_000, UpdatedAtMs: 2_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetRecipe(ctx, recipeID(t, incompleteID)); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("missing revision error = %v, want ErrCorruptData", err)
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = store.ListRecipes(canceled, RecipeListFilter{Archive: domain.ArchiveAll, PageSize: recipePageSize(t, 10)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v, want context.Canceled", err)
	}
}

func TestRecipeStoreRejectsNoncanonicalPersistedSnapshots(t *testing.T) {
	tests := []struct {
		name             string
		recipeDisplay    string
		recipeKey        string
		enteredUnit      string
		conversionNumber int64
		conversionDenom  int64
		addSpacedUnit    bool
	}{
		{
			name: "noncanonical recipe display", recipeDisplay: " Spaced recipe ",
			recipeKey: "spaced recipe", enteredUnit: "g", conversionNumber: 1000, conversionDenom: 1,
		},
		{
			name: "noncanonical entered unit", recipeDisplay: "Unit snapshot",
			recipeKey: "unit snapshot", enteredUnit: " g ", conversionNumber: 1000, conversionDenom: 1,
			addSpacedUnit: true,
		},
		{
			name: "non-reduced conversion", recipeDisplay: "Ratio snapshot",
			recipeKey: "ratio snapshot", enteredUnit: "g", conversionNumber: 2000, conversionDenom: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := recipeTestStore(t, "noncanonical.db")
			ctx := context.Background()
			if test.addSpacedUnit {
				if _, err := store.database.ExecContext(ctx, `
					INSERT INTO measurement_units (
						code, name, symbol, dimension, atomic_numerator, atomic_denominator,
						is_item_base, is_seeded
					) VALUES (' g ', 'spaced gram', 'sg', 'MASS', 1000, 1, 0, 0)
				`); err != nil {
					t.Fatal(err)
				}
			}
			id := insertRawRecipeSnapshot(
				t, store, test.recipeDisplay, test.recipeKey, test.enteredUnit,
				test.conversionNumber, test.conversionDenom,
			)
			if _, err := store.GetRecipe(ctx, id); !errors.Is(err, domain.ErrCorruptData) {
				t.Fatalf("GetRecipe error = %v, want ErrCorruptData", err)
			}
			if test.name == "noncanonical recipe display" {
				_, err := store.ListRecipes(ctx, RecipeListFilter{
					Archive: domain.ArchiveAll, PageSize: recipePageSize(t, 10),
				})
				if !errors.Is(err, domain.ErrCorruptData) {
					t.Fatalf("ListRecipes error = %v, want ErrCorruptData", err)
				}
			}
		})
	}
}

func insertRawRecipeSnapshot(
	t *testing.T,
	store *Store,
	display string,
	normalized string,
	enteredUnit string,
	conversionNumerator int64,
	conversionDenominator int64,
) domain.RecipeID {
	t.Helper()
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Raw output", false, true)
	componentID := recipeTestItem(t, store, "Raw input", true, false)
	recipeIDValue, err := store.queries.InsertRecipe(ctx, sqlcgen.InsertRecipeParams{
		Name: display, NormalizedName: normalized, OutputItemID: outputID.Int64(),
		CreatedAtMs: 1_000, UpdatedAtMs: 1_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	revisionID, err := store.queries.InsertRecipeRevision(ctx, sqlcgen.InsertRecipeRevisionParams{
		RecipeID: recipeIDValue, RevisionNumber: 1, StandardYieldQuantityAtomic: 1_000,
		Instructions: "raw", PreparationTimeMinutes: 0, CreatedAtMs: 1_000,
		ExpectedLatestRevisionNumber: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.queries.InsertRecipeRevisionComponent(ctx, sqlcgen.InsertRecipeRevisionComponentParams{
		RecipeRevisionID: revisionID, ComponentOrder: 1, ItemID: componentID.Int64(),
		QuantityAtomic: 1_000, EnteredUnitCode: enteredUnit,
		ConversionNumeratorAtomic: conversionNumerator,
		ConversionDenominator:     conversionDenominator, CreatedAtMs: 1_000,
	}); err != nil {
		t.Fatal(err)
	}
	return recipeID(t, recipeIDValue)
}

func recipeTestStore(t *testing.T, name string) *Store {
	t.Helper()
	return newAdapterTestStore(t, filepath.Join(t.TempDir(), name), database.DefaultOpenOptions())
}

func recipeTestItem(t *testing.T, store *Store, name string, purchasable, producible bool) domain.ItemID {
	t.Helper()
	unique := recipeName(t, name)
	params := testItemParams(unique.Display(), unique.Key())
	params.IsPurchasable = boolInteger(purchasable)
	params.IsProducible = boolInteger(producible)
	id, err := store.queries.InsertItem(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	return recipeItemID(t, id)
}

func recipeRevisionInput(t *testing.T, createdAt int64, instructions string, components []RecipeComponentInput) RecipeRevisionInput {
	t.Helper()
	return RecipeRevisionInput{
		StandardYield: recipeQuantity(t, 1_000), Instructions: instructions,
		PreparationTime: recipePreparation(t, 10), Components: components,
		CreatedAt: recipeInstant(t, createdAt),
	}
}

func recipeComponentInput(t *testing.T, order int64, itemID domain.ItemID, quantity int64, source RecipeComponentSource) RecipeComponentInput {
	t.Helper()
	return RecipeComponentInput{
		Order: recipeOrder(t, order), ItemID: itemID,
		Quantity: recipeQuantity(t, quantity), Source: source,
	}
}

func recipeName(t *testing.T, value string) domain.UniqueName {
	t.Helper()
	name, err := domain.NewUniqueName(value)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func recipeText(t *testing.T, value string) domain.NonEmptyText {
	t.Helper()
	text, err := domain.NewNonEmptyText(value)
	if err != nil {
		t.Fatal(err)
	}
	return text
}

func recipeUnit(t *testing.T, value string) domain.UnitCode {
	t.Helper()
	unit, err := domain.NewUnitCode(value)
	if err != nil {
		t.Fatal(err)
	}
	return unit
}

func recipeUnitSource(t *testing.T, value string) RecipeComponentSource {
	t.Helper()
	source, err := NewRecipeUnitSource(recipeUnit(t, value))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func recipePackagingSource(t *testing.T, id domain.PackagingID) RecipeComponentSource {
	t.Helper()
	source, err := NewRecipePackagingSource(id)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func recipeQuantity(t *testing.T, value int64) domain.AtomicQuantity {
	t.Helper()
	quantity, err := domain.NewPositiveAtomicQuantity(value)
	if err != nil {
		t.Fatal(err)
	}
	return quantity
}

func recipeOrder(t *testing.T, value int64) domain.ComponentOrder {
	t.Helper()
	order, err := domain.NewComponentOrder(value)
	if err != nil {
		t.Fatal(err)
	}
	return order
}

func recipePreparation(t *testing.T, value int64) domain.PreparationMinutes {
	t.Helper()
	preparation, err := domain.NewPreparationMinutes(value)
	if err != nil {
		t.Fatal(err)
	}
	return preparation
}

func recipeInstant(t *testing.T, value int64) domain.UTCInstant {
	t.Helper()
	instant, err := domain.UTCInstantFromUnixMilli(value)
	if err != nil {
		t.Fatal(err)
	}
	return instant
}

func recipeConversion(t *testing.T, numerator, denominator int64) domain.UnitConversion {
	t.Helper()
	conversion, err := domain.NewUnitConversion(numerator, denominator)
	if err != nil {
		t.Fatal(err)
	}
	return conversion
}

func recipePageSize(t *testing.T, value int) RecipePageSize {
	t.Helper()
	size, err := NewRecipePageSize(value)
	if err != nil {
		t.Fatal(err)
	}
	return size
}

func recipeID(t *testing.T, value int64) domain.RecipeID {
	t.Helper()
	id, err := domain.NewRecipeID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func recipeItemID(t *testing.T, value int64) domain.ItemID {
	t.Helper()
	id, err := domain.NewItemID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func recipePackagingID(t *testing.T, value int64) domain.PackagingID {
	t.Helper()
	id, err := domain.NewPackagingID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
