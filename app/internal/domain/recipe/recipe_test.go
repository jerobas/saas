package recipe_test

import (
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/recipe"
)

func TestImmutableRecipeAggregateAndSummary(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1000))
	recipeID := must(domain.NewRecipeID(1))
	revisionID := must(domain.NewRecipeRevisionID(10))
	component := must(recipe.NewComponent(recipe.ComponentParams{
		ID: must(domain.NewRecipeComponentID(100)), RevisionID: revisionID,
		Order: must(domain.NewComponentOrder(1)), ItemID: must(domain.NewItemID(2)),
		Quantity: must(domain.NewPositiveAtomicQuantity(1500)), EnteredUnit: must(domain.NewUnitCode("g")),
		Conversion: must(domain.NewUnitConversion(1000, 1)), CreatedAt: created,
	}))
	components := []recipe.Component{component}
	revision, err := recipe.NewRevision(recipe.RevisionParams{
		ID: revisionID, RecipeID: recipeID, Number: must(domain.NewRevisionNumber(1)),
		StandardYield: must(domain.NewPositiveAtomicQuantity(1000)), Instructions: "",
		PreparationTime: must(domain.NewPreparationMinutes(20)), CreatedAt: created,
		Components: components,
	})
	if err != nil {
		t.Fatal(err)
	}
	components[0] = recipe.Component{}
	if revision.Components()[0].ID().Int64() != 100 {
		t.Fatal("revision retained caller-owned component slice")
	}
	returned := revision.Components()
	returned[0] = recipe.Component{}
	if revision.Components()[0].ID().Int64() != 100 {
		t.Fatal("component accessor exposed revision slice")
	}

	aggregate, err := recipe.New(recipe.Params{
		ID: recipeID, Name: must(domain.NewUniqueName("Chocolate cake")),
		OutputItemID: must(domain.NewItemID(1)), CreatedAt: created, UpdatedAt: created,
		CurrentRevision: revision,
	})
	if err != nil || aggregate.CurrentRevision().Number().Int64() != 1 {
		t.Fatalf("recipe aggregate = %#v, %v", aggregate, err)
	}

	current := must(recipe.NewCurrentRevisionSummary(recipe.CurrentRevisionSummaryParams{
		ID: revisionID, Number: must(domain.NewRevisionNumber(1)),
		StandardYield: must(domain.NewPositiveAtomicQuantity(1000)),
	}))
	summary, err := recipe.NewRecipeSummary(recipe.RecipeSummaryParams{
		ID: recipeID, Name: aggregate.Name(), OutputItemID: aggregate.OutputItemID(),
		OutputItemName: must(domain.NewDisplayName("Cake item")),
		CreatedAt:      created, UpdatedAt: created, CurrentRevision: current,
	})
	if err != nil || summary.CurrentRevision().ID() != revisionID || summary.OutputItemName().String() != "Cake item" {
		t.Fatalf("recipe summary = %#v, %v", summary, err)
	}
}

func TestRevisionRejectsDuplicatesAndForeignComponents(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1000))
	recipeID := must(domain.NewRecipeID(1))
	revisionID := must(domain.NewRecipeRevisionID(10))
	otherRevisionID := must(domain.NewRecipeRevisionID(11))
	itemID := must(domain.NewItemID(2))
	makeComponent := func(id int64, owner domain.RecipeRevisionID, order int64) recipe.Component {
		return must(recipe.NewComponent(recipe.ComponentParams{
			ID: must(domain.NewRecipeComponentID(id)), RevisionID: owner,
			Order: must(domain.NewComponentOrder(order)), ItemID: itemID,
			Quantity: must(domain.NewPositiveAtomicQuantity(1)), EnteredUnit: must(domain.NewUnitCode("g")),
			Conversion: must(domain.NewUnitConversion(1, 1)), CreatedAt: created,
		}))
	}
	_, err := recipe.NewRevision(recipe.RevisionParams{
		ID: revisionID, RecipeID: recipeID, Number: must(domain.NewRevisionNumber(1)),
		StandardYield:   must(domain.NewPositiveAtomicQuantity(1)),
		PreparationTime: must(domain.NewPreparationMinutes(0)), CreatedAt: created,
		Components: []recipe.Component{makeComponent(1, revisionID, 1), makeComponent(2, otherRevisionID, 1)},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid revision error = %v", err)
	}
	var validation *domain.ValidationError
	if !errors.As(err, &validation) || len(validation.Violations()) < 3 {
		t.Fatalf("expected ownership/order/item violations: %v", err)
	}
}

func TestRecipeRejectsDirectSelfConsumptionAndMissingCurrentRevision(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1000))
	recipeID := must(domain.NewRecipeID(1))
	revisionID := must(domain.NewRecipeRevisionID(10))
	outputItemID := must(domain.NewItemID(1))
	component := must(recipe.NewComponent(recipe.ComponentParams{
		ID: must(domain.NewRecipeComponentID(1)), RevisionID: revisionID,
		Order: must(domain.NewComponentOrder(1)), ItemID: outputItemID,
		Quantity: must(domain.NewPositiveAtomicQuantity(1)), EnteredUnit: must(domain.NewUnitCode("each")),
		Conversion: must(domain.NewUnitConversion(1000, 1)), CreatedAt: created,
	}))
	revision := must(recipe.NewRevision(recipe.RevisionParams{
		ID: revisionID, RecipeID: recipeID, Number: must(domain.NewRevisionNumber(1)),
		StandardYield:   must(domain.NewPositiveAtomicQuantity(1000)),
		PreparationTime: must(domain.NewPreparationMinutes(0)), CreatedAt: created,
		Components: []recipe.Component{component},
	}))
	_, err := recipe.New(recipe.Params{
		ID: recipeID, Name: must(domain.NewUniqueName("Loop")), OutputItemID: outputItemID,
		CreatedAt: created, UpdatedAt: created, CurrentRevision: revision,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("self consumption error = %v", err)
	}

	_, err = recipe.NewRecipeSummary(recipe.RecipeSummaryParams{
		ID: recipeID, Name: must(domain.NewUniqueName("Incomplete")), OutputItemID: outputItemID,
		OutputItemName: must(domain.NewDisplayName("Output")), CreatedAt: created, UpdatedAt: created,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing current revision error = %v", err)
	}
}

func TestRecipeVersionCannotPrecedeCurrentRevision(t *testing.T) {
	recipeID := must(domain.NewRecipeID(1))
	revisionID := must(domain.NewRecipeRevisionID(10))
	created := must(domain.UTCInstantFromUnixMilli(1_000))
	revised := must(domain.UTCInstantFromUnixMilli(2_000))
	component := must(recipe.NewComponent(recipe.ComponentParams{
		ID: must(domain.NewRecipeComponentID(1)), RevisionID: revisionID,
		Order: must(domain.NewComponentOrder(1)), ItemID: must(domain.NewItemID(2)),
		Quantity: must(domain.NewPositiveAtomicQuantity(1)), EnteredUnit: must(domain.NewUnitCode("g")),
		Conversion: must(domain.NewUnitConversion(1, 1)), CreatedAt: revised,
	}))
	revision := must(recipe.NewRevision(recipe.RevisionParams{
		ID: revisionID, RecipeID: recipeID, Number: must(domain.NewRevisionNumber(2)),
		StandardYield:   must(domain.NewPositiveAtomicQuantity(1)),
		PreparationTime: must(domain.NewPreparationMinutes(0)), CreatedAt: revised,
		Components: []recipe.Component{component},
	}))
	_, err := recipe.New(recipe.Params{
		ID: recipeID, Name: must(domain.NewUniqueName("Stale version")),
		OutputItemID: must(domain.NewItemID(3)), CreatedAt: created, UpdatedAt: created,
		CurrentRevision: revision,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("version before revision error = %v, want ErrValidation", err)
	}
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
