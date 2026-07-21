package application

import (
	"context"

	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteProductionStore struct {
	store *sqlite.Store
}

func NewSQLiteProductionStore(store *sqlite.Store) ProductionStore {
	if store == nil {
		panic("sqlite production store requires a store")
	}
	return &sqliteProductionStore{store: store}
}

func (s *sqliteProductionStore) PostProduction(ctx context.Context, input productionPostStoreInput) (ProductionDocument, error) {
	inputs := make([]sqlite.PostProductionComponentInput, 0, len(input.Inputs))
	for _, line := range input.Inputs {
		inputs = append(inputs, sqlite.PostProductionComponentInput{
			ItemID:               line.ItemID,
			Quantity:             line.Quantity,
			EnteredUnit:          line.EnteredUnit,
			EnteredPackagingName: line.EnteredPackagingName,
			Conversion:           line.Conversion,
			LotID:                line.LotID,
		})
	}
	posted, err := s.store.PostProduction(ctx, sqlite.PostProductionInput{
		IdempotencyKey:   input.IdempotencyKey,
		RecipeRevisionID: input.RecipeRevisionID,
		OccurredOn:       input.OccurredOn,
		PostedAt:         input.PostedAt,
		DirectCost:       input.DirectCost,
		Notes:            input.Notes,
		Output: sqlite.PostProductionOutputInput{
			Quantity:             input.Output.Quantity,
			EnteredUnit:          input.Output.EnteredUnit,
			EnteredPackagingName: input.Output.EnteredPackagingName,
			Conversion:           input.Output.Conversion,
			LotCode:              input.Output.LotCode,
			ExpiresOn:            input.Output.ExpiresOn,
		},
		Inputs: inputs,
	})
	if err != nil {
		return ProductionDocument{}, err
	}
	return mapSQLitePostedProduction(posted)
}

func mapSQLitePostedProduction(posted sqlite.PostedProductionDocument) (ProductionDocument, error) {
	outputLine, err := mapSQLiteProductionLine(posted.OutputLine())
	if err != nil {
		return ProductionDocument{}, err
	}
	sourceLines := posted.InputLines()
	inputLines := make([]PostedProductionLine, 0, len(sourceLines))
	for _, line := range sourceLines {
		mapped, err := mapSQLiteProductionLine(line)
		if err != nil {
			return ProductionDocument{}, err
		}
		inputLines = append(inputLines, mapped)
	}
	return NewProductionDocument(
		posted.ID(),
		posted.IdempotencyKey(),
		posted.PostingSequence(),
		posted.RecipeRevisionID(),
		posted.OutputItemID(),
		posted.OccurredOn(),
		posted.PostedAt(),
		posted.Currency(),
		posted.DirectCost(),
		posted.Notes(),
		outputLine,
		inputLines,
	)
}

func mapSQLiteProductionLine(line sqlite.PostedProductionLine) (PostedProductionLine, error) {
	allocations, err := mapSQLiteProductionAllocations(line.Allocations())
	if err != nil {
		return PostedProductionLine{}, err
	}
	return NewPostedProductionLine(
		line.ID(),
		line.LineOrder(),
		line.ItemID(),
		line.Direction(),
		line.Quantity(),
		line.EnteredUnit(),
		line.EnteredPackagingName(),
		line.Conversion(),
		line.InventoryValue(),
		line.LotID(),
		line.LotCode(),
		line.OriginatedOn(),
		line.ExpiresOn(),
		allocations,
	)
}

func mapSQLiteProductionAllocations(source []sqlite.ProductionAllocation) ([]ProductionAllocation, error) {
	allocations := make([]ProductionAllocation, 0, len(source))
	for _, allocation := range source {
		mapped, err := NewProductionAllocation(
			allocation.ID(),
			allocation.LotID(),
			allocation.Quantity(),
		)
		if err != nil {
			return nil, err
		}
		allocations = append(allocations, mapped)
	}
	return allocations, nil
}
