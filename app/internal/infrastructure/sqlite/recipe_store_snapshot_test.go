package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

func TestRecipeRevisionListStaysOnSnapshotAcrossExternalComponentInsert(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recipe-snapshot.db")
	reader := newAdapterTestStore(t, path, database.DefaultOpenOptions())
	writer := newAdapterTestStore(t, path, database.DefaultOpenOptions())
	ctx := context.Background()
	outputID := recipeTestItem(t, reader, "Snapshot output", false, true)
	firstComponentID := recipeTestItem(t, reader, "Snapshot first component", true, false)
	secondComponentID := recipeTestItem(t, reader, "Snapshot second component", true, false)
	created, err := reader.CreateRecipe(ctx, CreateRecipeInput{
		Name:         recipeName(t, "Snapshot recipe"),
		OutputItemID: outputID,
		CreatedAt:    recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "snapshot", []RecipeComponentInput{
			recipeComponentInput(t, 1, firstComponentID, 1_000, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	hookCalled := false
	reader.recipeReadHook = func(stage recipeReadStage) error {
		if stage != recipeRevisionRowsLoaded {
			return errors.New("unexpected recipe read stage")
		}
		hookCalled = true
		reader.recipeReadHook = nil
		return writer.withWriteQueries(ctx, "insert external recipe component", func(queries *sqlcgen.Queries) error {
			_, err := queries.InsertRecipeRevisionComponent(ctx, sqlcgen.InsertRecipeRevisionComponentParams{
				RecipeRevisionID:          created.CurrentRevision().ID().Int64(),
				ComponentOrder:            2,
				ItemID:                    secondComponentID.Int64(),
				QuantityAtomic:            2_000,
				EnteredUnitCode:           "g",
				ConversionNumeratorAtomic: 1_000,
				ConversionDenominator:     1,
				CreatedAtMs:               1_000,
			})
			return err
		})
	}

	snapshot, err := reader.ListRecipeRevisions(ctx, created.ID())
	if err != nil {
		t.Fatalf("ListRecipeRevisions snapshot: %v", err)
	}
	if !hookCalled {
		t.Fatal("external writer hook was not called")
	}
	if len(snapshot) != 1 || len(snapshot[0].Components()) != 1 {
		t.Fatalf("snapshot revision/components = %d/%d, want 1/1", len(snapshot), len(snapshot[0].Components()))
	}

	latest, err := reader.ListRecipeRevisions(ctx, created.ID())
	if err != nil {
		t.Fatalf("ListRecipeRevisions latest: %v", err)
	}
	if len(latest) != 1 || len(latest[0].Components()) != 2 {
		t.Fatalf("latest revision/components = %d/%d, want 1/2", len(latest), len(latest[0].Components()))
	}
}
