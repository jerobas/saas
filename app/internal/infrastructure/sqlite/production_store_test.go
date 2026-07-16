package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
)

func TestProductionStorePostsProductionConsumesFEFOAndCreatesOutputLot(t *testing.T) {
	store := recipeTestStore(t, "production.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Cake", false, true)
	componentID := recipeTestItem(t, store, "Flour", true, false)
	recipeValue, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Cake recipe"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "mix", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 500, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	firstPurchase := postAdjustmentTestPurchase(t, store, componentID, "production-late", "FLOUR-LATE", "2026-12-31", 600, 600)
	secondPurchase := postAdjustmentTestPurchase(t, store, componentID, "production-early", "FLOUR-EARLY", "2026-08-01", 400, 400)

	directCost := mustInventoryValue(t, 2_500_000)
	posted, err := store.PostProduction(ctx, PostProductionInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "production-1"),
		RecipeRevisionID: recipeValue.CurrentRevision().ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-15"),
		PostedAt:         recipeInstant(t, 5_000),
		DirectCost:       directCost,
		Output: PostProductionOutputInput{
			Quantity:    recipeQuantity(t, 250),
			EnteredUnit: recipeUnit(t, "g"),
			Conversion:  recipeConversion(t, 1_000, 1),
			LotCode:     domain.Some(recipeText(t, "CAKE-001")),
			ExpiresOn:   domain.Some(mustPurchaseDate(t, "2026-07-20")),
		},
		Inputs: []PostProductionComponentInput{
			{
				ItemID:      componentID,
				Quantity:    recipeQuantity(t, 500),
				EnteredUnit: recipeUnit(t, "g"),
				Conversion:  recipeConversion(t, 1_000, 1),
			},
		},
	})
	if err != nil {
		t.Fatalf("post production: %v", err)
	}
	if posted.ID().IsZero() || posted.PostingSequence().Int64() != 3 || posted.OutputItemID() != outputID {
		t.Fatalf("posted production = %#v", posted)
	}
	if posted.DirectCost().Int64() != directCost.Int64() || posted.RecipeRevisionID() != recipeValue.CurrentRevision().ID() {
		t.Fatalf("production metadata = %#v", posted)
	}

	inputs := posted.InputLines()
	if len(inputs) != 1 || inputs[0].Direction() != domain.DirectionOut || inputs[0].InventoryValue().Int64() != 5_000_000 {
		t.Fatalf("input lines = %#v", inputs)
	}
	allocations := inputs[0].Allocations()
	if len(allocations) != 2 {
		t.Fatalf("allocations = %#v, want 2", allocations)
	}
	if allocations[0].LotID() != secondPurchase.Lines()[0].LotID() || allocations[0].Quantity().Int64() != 400 ||
		allocations[1].LotID() != firstPurchase.Lines()[0].LotID() || allocations[1].Quantity().Int64() != 100 {
		t.Fatalf("allocations = %#v, want early lot then late lot", allocations)
	}
	output := posted.OutputLine()
	outputLotID, hasOutputLot := output.LotID().Get()
	if output.Direction() != domain.DirectionIn || output.ItemID() != outputID || output.InventoryValue().Int64() != 7_500_000 ||
		!hasOutputLot || outputLotID.IsZero() {
		t.Fatalf("output line = %#v", output)
	}

	var componentQuantity, componentValue, outputQuantity, outputValue int64
	if err := store.database.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro
		FROM inventory_balances WHERE item_id = ?`, componentID.Int64()).Scan(&componentQuantity, &componentValue); err != nil {
		t.Fatal(err)
	}
	if err := store.database.QueryRowContext(ctx, `
		SELECT quantity_atomic, inventory_value_micro
		FROM inventory_balances WHERE item_id = ?`, outputID.Int64()).Scan(&outputQuantity, &outputValue); err != nil {
		t.Fatal(err)
	}
	if componentQuantity != 500 || componentValue != 5_000_000 || outputQuantity != 250 || outputValue != 7_500_000 {
		t.Fatalf("balances component=%d/%d output=%d/%d", componentQuantity, componentValue, outputQuantity, outputValue)
	}

	replayed, err := store.PostProduction(ctx, PostProductionInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "production-1"),
		RecipeRevisionID: recipeValue.CurrentRevision().ID(),
		OccurredOn:       mustPurchaseDate(t, "2026-07-15"),
		PostedAt:         recipeInstant(t, 9_000),
		Output: PostProductionOutputInput{
			Quantity:    recipeQuantity(t, 250),
			EnteredUnit: recipeUnit(t, "g"),
			Conversion:  recipeConversion(t, 1_000, 1),
		},
		Inputs: []PostProductionComponentInput{
			{
				ItemID:      componentID,
				Quantity:    recipeQuantity(t, 500),
				EnteredUnit: recipeUnit(t, "g"),
				Conversion:  recipeConversion(t, 1_000, 1),
			},
		},
	})
	if err != nil {
		t.Fatalf("replay production: %v", err)
	}
	if replayed.ID() != posted.ID() || !replayed.PostedAt().Equal(posted.PostedAt()) || replayed.OutputLine().InventoryValue() != posted.OutputLine().InventoryValue() {
		t.Fatalf("replayed = %#v, want original %#v", replayed, posted)
	}
}

func TestProductionStoreAllowsManualLotOverride(t *testing.T) {
	store := recipeTestStore(t, "production-override.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Tart", false, true)
	componentID := recipeTestItem(t, store, "Jam", true, false)
	recipeValue, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Tart recipe"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "mix", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 50, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	earlyPurchase := postAdjustmentTestPurchase(t, store, componentID, "production-override-early", "JAM-EARLY", "2026-08-01", 100, 100)
	latePurchase := postAdjustmentTestPurchase(t, store, componentID, "production-override-late", "JAM-LATE", "2026-12-31", 100, 100)

	input := productionInputFixture(t, recipeValue.CurrentRevision().ID(), componentID, 50)
	input.IdempotencyKey = mustPurchaseIdempotencyKey(t, "production-override")
	input.Inputs[0].LotID = domain.Some(latePurchase.Lines()[0].LotID())
	posted, err := store.PostProduction(ctx, input)
	if err != nil {
		t.Fatalf("post production with override: %v", err)
	}
	allocations := posted.InputLines()[0].Allocations()
	if len(allocations) != 1 || allocations[0].LotID() != latePurchase.Lines()[0].LotID() || allocations[0].Quantity().Int64() != 50 {
		t.Fatalf("override allocations = %#v", allocations)
	}
	earlyAvailable := productionLotAvailableQuantity(t, store, earlyPurchase.Lines()[0].LotID())
	lateAvailable := productionLotAvailableQuantity(t, store, latePurchase.Lines()[0].LotID())
	if earlyAvailable != 100 || lateAvailable != 50 {
		t.Fatalf("lot availability early/late = %d/%d, want 100/50", earlyAvailable, lateAvailable)
	}
}

func TestProductionStoreRejectsInsufficientStock(t *testing.T) {
	store := recipeTestStore(t, "production-insufficient.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Mousse", false, true)
	componentID := recipeTestItem(t, store, "Cream", true, false)
	recipeValue, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Mousse recipe"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "mix", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 100, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.PostProduction(ctx, productionInputFixture(t, recipeValue.CurrentRevision().ID(), componentID, 1))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("insufficient stock error = %v, want validation", err)
	}
}

func productionLotAvailableQuantity(t *testing.T, store *Store, lotID domain.InventoryLotID) int64 {
	t.Helper()
	var available int64
	if err := store.database.QueryRowContext(context.Background(), `
		SELECT lot.initial_quantity_atomic
			- COALESCE(SUM(
				CASE WHEN allocation.restores_allocation_id IS NULL
					THEN allocation.quantity_atomic ELSE 0 END
			), 0)
			+ COALESCE(SUM(
				CASE WHEN allocation.restores_allocation_id IS NOT NULL
					THEN allocation.quantity_atomic ELSE 0 END
			), 0)
		FROM inventory_lots lot
		LEFT JOIN lot_allocations allocation ON allocation.lot_id = lot.id
		WHERE lot.id = ?
		GROUP BY lot.id, lot.initial_quantity_atomic
	`, lotID.Int64()).Scan(&available); err != nil {
		t.Fatal(err)
	}
	return available
}

func TestProductionStoreRejectsExpiredLots(t *testing.T) {
	store := recipeTestStore(t, "production-expired.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Cookie", false, true)
	componentID := recipeTestItem(t, store, "Eggs", true, false)
	recipeValue, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Cookie recipe"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "mix", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 100, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	postAdjustmentTestPurchase(t, store, componentID, "production-expired-stock", "EGG-OLD", "2026-07-01", 100, 100)

	_, err = store.PostProduction(ctx, productionInputFixture(t, recipeValue.CurrentRevision().ID(), componentID, 100))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expired lot error = %v, want validation", err)
	}
}

func TestProductionStoreRejectsOutputItemAsInput(t *testing.T) {
	store := recipeTestStore(t, "production-output-as-input.db")
	ctx := context.Background()
	outputID := recipeTestItem(t, store, "Ganache", true, true)
	componentID := recipeTestItem(t, store, "Chocolate chips", true, false)
	recipeValue, err := store.CreateRecipe(ctx, CreateRecipeInput{
		Name: recipeName(t, "Ganache recipe"), OutputItemID: outputID,
		CreatedAt: recipeInstant(t, 1_000),
		Revision: recipeRevisionInput(t, 1_000, "mix", []RecipeComponentInput{
			recipeComponentInput(t, 1, componentID, 100, recipeUnitSource(t, "g")),
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	postAdjustmentTestPurchase(t, store, outputID, "production-output-stock", "GANACHE-RAW", "2026-12-31", 100, 100)

	input := productionInputFixture(t, recipeValue.CurrentRevision().ID(), outputID, 50)
	_, err = store.PostProduction(ctx, input)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("output-as-input error = %v, want validation", err)
	}
}

func productionInputFixture(
	t *testing.T,
	revisionID domain.RecipeRevisionID,
	componentID domain.ItemID,
	componentQuantity int64,
) PostProductionInput {
	t.Helper()
	return PostProductionInput{
		IdempotencyKey:   mustPurchaseIdempotencyKey(t, "production-fixture"),
		RecipeRevisionID: revisionID,
		OccurredOn:       mustPurchaseDate(t, "2026-07-15"),
		PostedAt:         recipeInstant(t, 5_000),
		Output: PostProductionOutputInput{
			Quantity:    recipeQuantity(t, 100),
			EnteredUnit: recipeUnit(t, "g"),
			Conversion:  recipeConversion(t, 1_000, 1),
		},
		Inputs: []PostProductionComponentInput{
			{
				ItemID:      componentID,
				Quantity:    recipeQuantity(t, componentQuantity),
				EnteredUnit: recipeUnit(t, "g"),
				Conversion:  recipeConversion(t, 1_000, 1),
			},
		},
	}
}
