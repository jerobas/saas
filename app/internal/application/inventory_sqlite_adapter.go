package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/inventory"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

const inventoryReadDefaultPageSize = 50
const inventoryReadMaximumPageSize = 100

type sqliteInventoryStore struct {
	store *sqlite.Store
}

func NewSQLiteInventoryStore(store *sqlite.Store) InventoryStore {
	if store == nil {
		panic("sqlite inventory store requires a store")
	}
	return &sqliteInventoryStore{store: store}
}

func (s *sqliteInventoryStore) GetInventoryBalance(ctx context.Context, itemID domain.ItemID) (inventory.BalanceSnapshot, error) {
	return s.store.GetInventoryBalance(ctx, itemID)
}

func (s *sqliteInventoryStore) ListInventoryBalances(ctx context.Context, input InventoryBalanceListInput) (InventoryBalancePage, error) {
	pageSize, limit, err := inventoryReadPageSize(input.PageSize)
	if err != nil {
		return InventoryBalancePage{}, err
	}
	after := domain.None[sqlite.InventoryBalanceCursor]()
	if cursor, ok := input.After.Get(); ok {
		mapped, err := sqlite.NewInventoryBalanceCursor(cursor.ItemName, cursor.ItemID)
		if err != nil {
			return InventoryBalancePage{}, err
		}
		after = domain.Some(mapped)
	}
	items, err := s.store.ListInventoryBalances(ctx, sqlite.InventoryBalanceListParams{
		IncludeArchived: input.IncludeArchived,
		Search:          input.Search,
		After:           after,
		Limit:           limit,
	})
	if err != nil {
		return InventoryBalancePage{}, err
	}
	next := domain.None[InventoryBalanceCursor]()
	if len(items) > int(pageSize) {
		items = items[:pageSize]
		last := items[len(items)-1].Snapshot()
		next = domain.Some(InventoryBalanceCursor{ItemName: last.ItemName(), ItemID: last.Balance().ItemID()})
	}
	return NewInventoryBalancePage(items, next), nil
}

func (s *sqliteInventoryStore) ListItemLotFacts(ctx context.Context, itemID domain.ItemID) ([]inventory.LotView, error) {
	return s.store.ListItemLotFacts(ctx, itemID)
}

func (s *sqliteInventoryStore) ListEligibleFEFOLots(ctx context.Context, itemID domain.ItemID, on domain.BusinessDate) ([]inventory.LotView, error) {
	return s.store.ListEligibleFEFOLots(ctx, itemID, on)
}

func (s *sqliteInventoryStore) ListItemLedgerPage(ctx context.Context, input ItemLedgerPageInput) (ItemLedgerPage, error) {
	pageSize, limit, err := inventoryReadPageSize(input.PageSize)
	if err != nil {
		return ItemLedgerPage{}, err
	}
	items, err := s.store.ListItemLedgerPage(ctx, sqlite.ItemLedgerPageParams{
		ItemID: input.ItemID,
		After:  input.After,
		Limit:  limit,
	})
	if err != nil {
		return ItemLedgerPage{}, err
	}
	next := domain.None[inventory.LedgerCursor]()
	if len(items) > int(pageSize) {
		items = items[:pageSize]
		next = domain.Some(items[len(items)-1].Entry().Cursor())
	}
	return NewItemLedgerPage(items, next), nil
}

func (s *sqliteInventoryStore) ListLineAllocations(ctx context.Context, lineID domain.StockDocumentLineID) ([]inventory.AllocationView, error) {
	return s.store.ListLineAllocations(ctx, lineID)
}

func inventoryReadPageSize(requested int) (int, int64, error) {
	if requested == 0 {
		requested = inventoryReadDefaultPageSize
	}
	if requested < 1 || requested > inventoryReadMaximumPageSize {
		return 0, 0, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	return requested, int64(requested + 1), nil
}
