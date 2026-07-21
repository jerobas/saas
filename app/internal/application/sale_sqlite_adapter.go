package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
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

func (s *sqliteSaleStore) GetSale(ctx context.Context, id domain.StockDocumentID) (SaleDocument, error) {
	posted, err := s.store.GetPostedSale(ctx, id)
	if err != nil {
		return SaleDocument{}, err
	}
	return mapSQLitePostedSale(posted)
}

func (s *sqliteSaleStore) ListSales(ctx context.Context, input SaleListInput) (SalePage, error) {
	after := domain.None[sqlite.SaleCursor]()
	if cursor, ok := input.After.Get(); ok {
		after = domain.Some(sqlite.SaleCursor{
			PostingSequence: cursor.PostingSequence,
			ID:              cursor.ID,
		})
	}
	page, err := s.store.ListPostedSales(ctx, sqlite.SaleListFilter{
		After:    after,
		PageSize: input.PageSize,
	})
	if err != nil {
		return SalePage{}, err
	}
	sourceItems := page.Items()
	items := make([]SaleDocument, 0, len(sourceItems))
	for _, posted := range sourceItems {
		mapped, err := mapSQLitePostedSale(posted)
		if err != nil {
			return SalePage{}, err
		}
		items = append(items, mapped)
	}
	next := domain.None[SaleCursor]()
	if cursor, ok := page.Next().Get(); ok {
		next = domain.Some(SaleCursor{
			PostingSequence: cursor.PostingSequence,
			ID:              cursor.ID,
		})
	}
	return NewSalePage(items, next), nil
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
