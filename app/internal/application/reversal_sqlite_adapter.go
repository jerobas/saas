package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteReversalStore struct {
	store *sqlite.Store
}

func NewSQLiteReversalStore(store *sqlite.Store) ReversalStore {
	if store == nil {
		panic("sqlite reversal store requires a store")
	}
	return &sqliteReversalStore{store: store}
}

func (s *sqliteReversalStore) PostReversal(ctx context.Context, input reversalPostStoreInput) (ReversalDocument, error) {
	posted, err := s.store.PostReversal(ctx, sqlite.PostReversalInput{
		IdempotencyKey:   input.IdempotencyKey,
		TargetDocumentID: input.TargetDocumentID,
		OccurredOn:       input.OccurredOn,
		PostedAt:         input.PostedAt,
		Notes:            input.Notes,
	})
	if err != nil {
		return ReversalDocument{}, err
	}
	return mapSQLitePostedReversal(posted)
}

func mapSQLitePostedReversal(posted sqlite.PostedReversalDocument) (ReversalDocument, error) {
	sourceLines := posted.Lines()
	lines := make([]PostedReversalLine, 0, len(sourceLines))
	for _, line := range sourceLines {
		allocations, err := mapSQLiteReversalAllocations(line.Allocations())
		if err != nil {
			return ReversalDocument{}, err
		}
		mapped, err := NewPostedReversalLine(
			line.ID(),
			line.LineOrder(),
			line.ItemID(),
			line.Direction(),
			line.Quantity(),
			line.EnteredUnit(),
			line.EnteredPackagingName(),
			line.Conversion(),
			line.InventoryValue(),
			line.CommercialTotal(),
			line.ReversesLineID(),
			allocations,
		)
		if err != nil {
			return ReversalDocument{}, err
		}
		lines = append(lines, mapped)
	}
	return NewReversalDocument(
		posted.ID(),
		posted.IdempotencyKey(),
		posted.PostingSequence(),
		posted.TargetDocumentID(),
		posted.OccurredOn(),
		posted.PostedAt(),
		posted.Currency(),
		domain.ReasonExactReversal,
		posted.Notes(),
		lines,
	)
}

func mapSQLiteReversalAllocations(source []sqlite.ReversalAllocation) ([]ReversalAllocation, error) {
	allocations := make([]ReversalAllocation, 0, len(source))
	for _, allocation := range source {
		mapped, err := NewReversalAllocation(
			allocation.ID(),
			allocation.LotID(),
			allocation.Quantity(),
			allocation.RestoresAllocationID(),
		)
		if err != nil {
			return nil, err
		}
		allocations = append(allocations, mapped)
	}
	return allocations, nil
}
