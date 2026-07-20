package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteAdjustmentStore struct {
	store *sqlite.Store
}

func NewSQLiteAdjustmentStore(store *sqlite.Store) AdjustmentStore {
	if store == nil {
		panic("sqlite adjustment store requires a store")
	}
	return &sqliteAdjustmentStore{store: store}
}

func (s *sqliteAdjustmentStore) ListAdjustments(ctx context.Context, input AdjustmentListInput) (AdjustmentPage, error) {
	after := domain.None[sqlite.AdjustmentCursor]()
	if cursor, ok := input.After.Get(); ok {
		after = domain.Some(sqlite.AdjustmentCursor{
			PostingSequence: cursor.PostingSequence,
			ID:              cursor.ID,
		})
	}
	page, err := s.store.ListPostedAdjustments(ctx, sqlite.AdjustmentListFilter{
		After:    after,
		PageSize: input.PageSize,
	})
	if err != nil {
		return AdjustmentPage{}, err
	}
	sourceItems := page.Items()
	items := make([]AdjustmentDocument, 0, len(sourceItems))
	for _, posted := range sourceItems {
		mapped, err := mapSQLitePostedAdjustment(posted)
		if err != nil {
			return AdjustmentPage{}, err
		}
		items = append(items, mapped)
	}
	next := domain.None[AdjustmentCursor]()
	if cursor, ok := page.Next().Get(); ok {
		next = domain.Some(AdjustmentCursor{
			PostingSequence: cursor.PostingSequence,
			ID:              cursor.ID,
		})
	}
	return NewAdjustmentPage(items, next), nil
}

func (s *sqliteAdjustmentStore) PostAdjustment(ctx context.Context, input adjustmentPostStoreInput) (AdjustmentDocument, error) {
	lines := make([]sqlite.PostAdjustmentLineInput, 0, len(input.Lines))
	for _, line := range input.Lines {
		lines = append(lines, sqlite.PostAdjustmentLineInput{
			ItemID:               line.ItemID,
			Direction:            line.Direction,
			Quantity:             line.Quantity,
			EnteredUnit:          line.EnteredUnit,
			EnteredPackagingName: line.EnteredPackagingName,
			Conversion:           line.Conversion,
			InventoryValue:       line.InventoryValue,
			LotCode:              line.LotCode,
			ExpiresOn:            line.ExpiresOn,
		})
	}
	posted, err := s.store.PostAdjustment(ctx, sqlite.PostAdjustmentInput{
		IdempotencyKey: input.IdempotencyKey,
		OccurredOn:     input.OccurredOn,
		PostedAt:       input.PostedAt,
		Reason:         input.Reason,
		Notes:          input.Notes,
		Lines:          lines,
	})
	if err != nil {
		return AdjustmentDocument{}, err
	}
	return mapSQLitePostedAdjustment(posted)
}

func mapSQLitePostedAdjustment(posted sqlite.PostedAdjustmentDocument) (AdjustmentDocument, error) {
	sourceLines := posted.Lines()
	lines := make([]PostedAdjustmentLine, 0, len(sourceLines))
	for _, line := range sourceLines {
		allocations, err := mapSQLiteAdjustmentAllocations(line.Allocations())
		if err != nil {
			return AdjustmentDocument{}, err
		}
		mapped, err := NewPostedAdjustmentLine(
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
		if err != nil {
			return AdjustmentDocument{}, err
		}
		lines = append(lines, mapped)
	}
	return NewAdjustmentDocument(
		posted.ID(),
		posted.IdempotencyKey(),
		posted.PostingSequence(),
		posted.OccurredOn(),
		posted.PostedAt(),
		posted.Currency(),
		posted.Reason(),
		posted.Notes(),
		lines,
	)
}

func mapSQLiteAdjustmentAllocations(source []sqlite.AdjustmentAllocation) ([]AdjustmentAllocation, error) {
	allocations := make([]AdjustmentAllocation, 0, len(source))
	for _, allocation := range source {
		mapped, err := NewAdjustmentAllocation(
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
