package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/inventory"
)

type InventoryStore interface {
	GetInventoryBalance(ctx context.Context, itemID domain.ItemID) (inventory.BalanceSnapshot, error)
	ListInventoryBalances(ctx context.Context, input InventoryBalanceListInput) (InventoryBalancePage, error)
	ListItemLotFacts(ctx context.Context, itemID domain.ItemID) ([]inventory.LotView, error)
	ListEligibleFEFOLots(ctx context.Context, itemID domain.ItemID, on domain.BusinessDate) ([]inventory.LotView, error)
	ListItemLedgerPage(ctx context.Context, input ItemLedgerPageInput) (ItemLedgerPage, error)
	ListLineAllocations(ctx context.Context, lineID domain.StockDocumentLineID) ([]inventory.AllocationView, error)
}

type InventoryBalanceCursor struct {
	ItemName domain.UniqueName
	ItemID   domain.ItemID
}

type InventoryBalanceListInput struct {
	IncludeArchived bool
	Search          domain.Option[domain.NonEmptyText]
	After           domain.Option[InventoryBalanceCursor]
	PageSize        int
}

type InventoryBalancePage struct {
	items []inventory.BalanceListItem
	next  domain.Option[InventoryBalanceCursor]
}

func NewInventoryBalancePage(items []inventory.BalanceListItem, next domain.Option[InventoryBalanceCursor]) InventoryBalancePage {
	cloned := make([]inventory.BalanceListItem, len(items))
	copy(cloned, items)
	return InventoryBalancePage{items: cloned, next: next}
}

func (p InventoryBalancePage) Items() []inventory.BalanceListItem {
	items := make([]inventory.BalanceListItem, len(p.items))
	copy(items, p.items)
	return items
}

func (p InventoryBalancePage) Next() domain.Option[InventoryBalanceCursor] { return p.next }

type ItemLedgerPageInput struct {
	ItemID   domain.ItemID
	After    domain.Option[inventory.LedgerCursor]
	PageSize int
}

type ItemLedgerPage struct {
	items []inventory.LedgerEntryView
	next  domain.Option[inventory.LedgerCursor]
}

func NewItemLedgerPage(items []inventory.LedgerEntryView, next domain.Option[inventory.LedgerCursor]) ItemLedgerPage {
	cloned := make([]inventory.LedgerEntryView, len(items))
	copy(cloned, items)
	return ItemLedgerPage{items: cloned, next: next}
}

func (p ItemLedgerPage) Items() []inventory.LedgerEntryView {
	items := make([]inventory.LedgerEntryView, len(p.items))
	copy(items, p.items)
	return items
}

func (p ItemLedgerPage) Next() domain.Option[inventory.LedgerCursor] { return p.next }

type InventoryService struct {
	store InventoryStore
}

func NewInventoryService(store InventoryStore) *InventoryService {
	if store == nil {
		panic("inventory service requires a store")
	}
	return &InventoryService{store: store}
}

func (s *InventoryService) GetInventoryBalance(ctx context.Context, itemID domain.ItemID) (inventory.BalanceSnapshot, error) {
	snapshot, err := s.store.GetInventoryBalance(ctx, itemID)
	if err != nil {
		return inventory.BalanceSnapshot{}, fmt.Errorf("get inventory balance: %w", err)
	}
	return snapshot, nil
}

func (s *InventoryService) ListInventoryBalances(ctx context.Context, input InventoryBalanceListInput) (InventoryBalancePage, error) {
	page, err := s.store.ListInventoryBalances(ctx, input)
	if err != nil {
		return InventoryBalancePage{}, fmt.Errorf("list inventory balances: %w", err)
	}
	return page, nil
}

func (s *InventoryService) ListItemLotFacts(ctx context.Context, itemID domain.ItemID) ([]inventory.LotView, error) {
	lots, err := s.store.ListItemLotFacts(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("list item lot facts: %w", err)
	}
	return lots, nil
}

func (s *InventoryService) ListEligibleFEFOLots(ctx context.Context, itemID domain.ItemID, on domain.BusinessDate) ([]inventory.LotView, error) {
	lots, err := s.store.ListEligibleFEFOLots(ctx, itemID, on)
	if err != nil {
		return nil, fmt.Errorf("list eligible FEFO lots: %w", err)
	}
	return lots, nil
}

func (s *InventoryService) ListItemLedgerPage(ctx context.Context, input ItemLedgerPageInput) (ItemLedgerPage, error) {
	page, err := s.store.ListItemLedgerPage(ctx, input)
	if err != nil {
		return ItemLedgerPage{}, fmt.Errorf("list item ledger page: %w", err)
	}
	return page, nil
}

func (s *InventoryService) ListLineAllocations(ctx context.Context, lineID domain.StockDocumentLineID) ([]inventory.AllocationView, error) {
	allocations, err := s.store.ListLineAllocations(ctx, lineID)
	if err != nil {
		return nil, fmt.Errorf("list line allocations: %w", err)
	}
	return allocations, nil
}
