package application

import (
	"context"

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
