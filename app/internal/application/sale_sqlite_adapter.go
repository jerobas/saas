package application

import (
	"context"

	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteSaleStore struct {
	store *sqlite.Store
}

func NewSQLiteSaleStore(store *sqlite.Store) SaleStore {
	if store == nil {
		panic("sqlite sale store requires a store")
	}
	return &sqliteSaleStore{store: store}
}

func (s *sqliteSaleStore) PostSale(ctx context.Context, input salePostStoreInput) (SaleDocument, error) {
	lines := make([]sqlite.PostSaleLineInput, 0, len(input.Lines))
	for _, line := range input.Lines {
		lines = append(lines, sqlite.PostSaleLineInput{
			ItemID:               line.ItemID,
			Quantity:             line.Quantity,
			EnteredUnit:          line.EnteredUnit,
			EnteredPackagingName: line.EnteredPackagingName,
			Conversion:           line.Conversion,
			CommercialTotal:      line.CommercialTotal,
			LotID:                line.LotID,
		})
	}
	posted, err := s.store.PostSale(ctx, sqlite.PostSaleInput{
		IdempotencyKey: input.IdempotencyKey,
		CounterpartyID: input.CounterpartyID,
		OccurredOn:     input.OccurredOn,
		PostedAt:       input.PostedAt,
		Reason:         input.Reason,
		Notes:          input.Notes,
		Lines:          lines,
	})
	if err != nil {
		return SaleDocument{}, err
	}
	return mapSQLitePostedSale(posted)
}

func mapSQLitePostedSale(posted sqlite.PostedSaleDocument) (SaleDocument, error) {
	sourceLines := posted.Lines()
	lines := make([]PostedSaleLine, 0, len(sourceLines))
	for _, line := range sourceLines {
		allocations, err := mapSQLiteSaleAllocations(line.Allocations())
		if err != nil {
			return SaleDocument{}, err
		}
		mapped, err := NewPostedSaleLine(
			line.ID(),
			line.LineOrder(),
			line.ItemID(),
			line.Quantity(),
			line.EnteredUnit(),
			line.EnteredPackagingName(),
			line.Conversion(),
			line.InventoryValue(),
			line.CommercialTotal(),
			allocations,
		)
		if err != nil {
			return SaleDocument{}, err
		}
		lines = append(lines, mapped)
	}
	return NewSaleDocument(
		posted.ID(),
		posted.IdempotencyKey(),
		posted.PostingSequence(),
		posted.CounterpartyID(),
		posted.OccurredOn(),
		posted.PostedAt(),
		posted.Currency(),
		posted.Reason(),
		posted.Notes(),
		lines,
	)
}

func mapSQLiteSaleAllocations(source []sqlite.SaleAllocation) ([]SaleAllocation, error) {
	allocations := make([]SaleAllocation, 0, len(source))
	for _, allocation := range source {
		mapped, err := NewSaleAllocation(
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
