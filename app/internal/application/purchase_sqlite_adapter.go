package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqlitePurchaseStore struct {
	store *sqlite.Store
}

func NewSQLitePurchaseStore(store *sqlite.Store) PurchaseStore {
	if store == nil {
		panic("sqlite purchase store requires a store")
	}
	return &sqlitePurchaseStore{store: store}
}

func (s *sqlitePurchaseStore) GetPurchase(ctx context.Context, id domain.StockDocumentID) (PurchaseDocument, error) {
	posted, err := s.store.GetPostedPurchase(ctx, id)
	if err != nil {
		return PurchaseDocument{}, err
	}
	return mapSQLitePostedPurchase(posted)
}

func (s *sqlitePurchaseStore) ListPurchases(ctx context.Context, input PurchaseListInput) (PurchasePage, error) {
	after := domain.None[sqlite.PurchaseCursor]()
	if cursor, ok := input.After.Get(); ok {
		after = domain.Some(sqlite.PurchaseCursor{
			PostingSequence: cursor.PostingSequence,
			ID:              cursor.ID,
		})
	}
	page, err := s.store.ListPostedPurchases(ctx, sqlite.PurchaseListFilter{
		After:    after,
		PageSize: input.PageSize,
	})
	if err != nil {
		return PurchasePage{}, err
	}
	sourceItems := page.Items()
	items := make([]PurchaseDocument, 0, len(sourceItems))
	for _, posted := range sourceItems {
		mapped, err := mapSQLitePostedPurchase(posted)
		if err != nil {
			return PurchasePage{}, err
		}
		items = append(items, mapped)
	}
	next := domain.None[PurchaseCursor]()
	if cursor, ok := page.Next().Get(); ok {
		next = domain.Some(PurchaseCursor{
			PostingSequence: cursor.PostingSequence,
			ID:              cursor.ID,
		})
	}
	return NewPurchasePage(items, next), nil
}

func (s *sqlitePurchaseStore) PostPurchase(ctx context.Context, input purchasePostStoreInput) (PurchaseDocument, error) {
	lines := make([]sqlite.PostPurchaseLineInput, 0, len(input.Lines))
	for _, line := range input.Lines {
		lines = append(lines, sqlite.PostPurchaseLineInput{
			ItemID:               line.ItemID,
			Quantity:             line.Quantity,
			EnteredUnit:          line.EnteredUnit,
			EnteredPackagingName: line.EnteredPackagingName,
			Conversion:           line.Conversion,
			CommercialTotal:      line.CommercialTotal,
			LotCode:              line.LotCode,
			ExpiresOn:            line.ExpiresOn,
		})
	}
	posted, err := s.store.PostPurchase(ctx, sqlite.PostPurchaseInput{
		IdempotencyKey: input.IdempotencyKey,
		CounterpartyID: input.CounterpartyID,
		OccurredOn:     input.OccurredOn,
		PostedAt:       input.PostedAt,
		Reason:         input.Reason,
		Notes:          input.Notes,
		Lines:          lines,
	})
	if err != nil {
		return PurchaseDocument{}, err
	}
	return mapSQLitePostedPurchase(posted)
}

func mapSQLitePostedPurchase(posted sqlite.PostedPurchaseDocument) (PurchaseDocument, error) {
	sourceLines := posted.Lines()
	lines := make([]PostedPurchaseLine, 0, len(sourceLines))
	for _, line := range sourceLines {
		mapped, err := NewPostedPurchaseLine(
			line.ID(),
			line.LineOrder(),
			line.ItemID(),
			line.Quantity(),
			line.EnteredUnit(),
			line.EnteredPackagingName(),
			line.Conversion(),
			line.InventoryValue(),
			line.CommercialTotal(),
			line.LotID(),
			line.LotCode(),
			line.OriginatedOn(),
			line.ExpiresOn(),
		)
		if err != nil {
			return PurchaseDocument{}, err
		}
		lines = append(lines, mapped)
	}
	return NewPurchaseDocument(
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
