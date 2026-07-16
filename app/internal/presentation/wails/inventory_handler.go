package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	inventorydomain "github.com/jerobas/saas/internal/domain/inventory"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type InventoryHandler struct {
	service *application.InventoryService
}

func NewInventoryHandler(service *application.InventoryService) *InventoryHandler {
	if service == nil {
		panic("inventory handler requires a service")
	}
	return &InventoryHandler{service: service}
}

func (h *InventoryHandler) GetInventoryBalance(itemID int64) (dto.InventoryBalanceResponse, error) {
	id, err := domain.NewItemID(itemID)
	if err != nil {
		return dto.InventoryBalanceResponse{}, fmt.Errorf("item id: %w", err)
	}
	snapshot, err := h.service.GetInventoryBalance(handlerContext(), id)
	if err != nil {
		return dto.InventoryBalanceResponse{}, fmt.Errorf("get inventory balance: %w", err)
	}
	return mapInventoryBalanceSnapshot(snapshot, nil), nil
}

func (h *InventoryHandler) ListInventoryBalances(req dto.InventoryBalanceListRequest) (dto.InventoryBalancePageResponse, error) {
	input, err := parseInventoryBalanceListRequest(req)
	if err != nil {
		return dto.InventoryBalancePageResponse{}, err
	}
	page, err := h.service.ListInventoryBalances(handlerContext(), input)
	if err != nil {
		return dto.InventoryBalancePageResponse{}, fmt.Errorf("list inventory balances: %w", err)
	}
	return mapInventoryBalancePage(page), nil
}

func (h *InventoryHandler) ListItemLotFacts(itemID int64) ([]dto.LotResponse, error) {
	id, err := domain.NewItemID(itemID)
	if err != nil {
		return nil, fmt.Errorf("item id: %w", err)
	}
	lots, err := h.service.ListItemLotFacts(handlerContext(), id)
	if err != nil {
		return nil, fmt.Errorf("list item lot facts: %w", err)
	}
	return mapLots(lots), nil
}

func (h *InventoryHandler) ListEligibleFEFOLots(itemID int64, businessDate string) ([]dto.LotResponse, error) {
	id, err := domain.NewItemID(itemID)
	if err != nil {
		return nil, fmt.Errorf("item id: %w", err)
	}
	on, err := domain.ParseBusinessDate(businessDate)
	if err != nil {
		return nil, fmt.Errorf("business date: %w", err)
	}
	lots, err := h.service.ListEligibleFEFOLots(handlerContext(), id, on)
	if err != nil {
		return nil, fmt.Errorf("list eligible FEFO lots: %w", err)
	}
	return mapLots(lots), nil
}

func (h *InventoryHandler) ListItemLedgerPage(req dto.ItemLedgerPageRequest) (dto.LedgerEntryPageResponse, error) {
	input, err := parseItemLedgerPageRequest(req)
	if err != nil {
		return dto.LedgerEntryPageResponse{}, err
	}
	page, err := h.service.ListItemLedgerPage(handlerContext(), input)
	if err != nil {
		return dto.LedgerEntryPageResponse{}, fmt.Errorf("list item ledger page: %w", err)
	}
	return mapLedgerEntryPage(page), nil
}

func (h *InventoryHandler) ListLineAllocations(lineID int64) ([]dto.AllocationResponse, error) {
	id, err := domain.NewStockDocumentLineID(lineID)
	if err != nil {
		return nil, fmt.Errorf("line id: %w", err)
	}
	allocations, err := h.service.ListLineAllocations(handlerContext(), id)
	if err != nil {
		return nil, fmt.Errorf("list line allocations: %w", err)
	}
	response := make([]dto.AllocationResponse, 0, len(allocations))
	for _, allocation := range allocations {
		response = append(response, mapAllocation(allocation))
	}
	return response, nil
}

func parseInventoryBalanceListRequest(req dto.InventoryBalanceListRequest) (application.InventoryBalanceListInput, error) {
	search, err := optionalNonEmptyText(req.Search)
	if err != nil {
		return application.InventoryBalanceListInput{}, fmt.Errorf("search: %w", err)
	}
	after := domain.None[application.InventoryBalanceCursor]()
	if req.After != nil {
		name, err := domain.NewUniqueName(req.After.ItemName)
		if err != nil {
			return application.InventoryBalanceListInput{}, fmt.Errorf("cursor item name: %w", err)
		}
		id, err := domain.NewItemID(req.After.ItemID)
		if err != nil {
			return application.InventoryBalanceListInput{}, fmt.Errorf("cursor item id: %w", err)
		}
		after = domain.Some(application.InventoryBalanceCursor{ItemName: name, ItemID: id})
	}
	return application.InventoryBalanceListInput{
		IncludeArchived: req.IncludeArchived,
		Search:          search,
		After:           after,
		PageSize:        req.PageSize,
	}, nil
}

func parseItemLedgerPageRequest(req dto.ItemLedgerPageRequest) (application.ItemLedgerPageInput, error) {
	itemID, err := domain.NewItemID(req.ItemID)
	if err != nil {
		return application.ItemLedgerPageInput{}, fmt.Errorf("item id: %w", err)
	}
	after := domain.None[inventorydomain.LedgerCursor]()
	if req.After != nil {
		postingSequence, err := domain.NewPostingSequence(req.After.PostingSequence)
		if err != nil {
			return application.ItemLedgerPageInput{}, fmt.Errorf("cursor posting sequence: %w", err)
		}
		lineOrder, err := domain.NewLineOrder(req.After.LineOrder)
		if err != nil {
			return application.ItemLedgerPageInput{}, fmt.Errorf("cursor line order: %w", err)
		}
		lineID, err := domain.NewStockDocumentLineID(req.After.LineID)
		if err != nil {
			return application.ItemLedgerPageInput{}, fmt.Errorf("cursor line id: %w", err)
		}
		cursor, err := inventorydomain.NewLedgerCursor(postingSequence, lineOrder, lineID)
		if err != nil {
			return application.ItemLedgerPageInput{}, fmt.Errorf("cursor: %w", err)
		}
		after = domain.Some(cursor)
	}
	return application.ItemLedgerPageInput{ItemID: itemID, After: after, PageSize: req.PageSize}, nil
}

func mapInventoryBalancePage(page application.InventoryBalancePage) dto.InventoryBalancePageResponse {
	items := page.Items()
	response := dto.InventoryBalancePageResponse{Items: make([]dto.InventoryBalanceResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapInventoryBalanceSnapshot(item.Snapshot(), &item))
	}
	if cursor, ok := page.Next().Get(); ok {
		response.Next = &dto.InventoryBalanceCursorResponse{
			ItemName: cursor.ItemName.Display(),
			ItemID:   cursor.ItemID.Int64(),
		}
	}
	return response
}

func mapInventoryBalanceSnapshot(snapshot inventorydomain.BalanceSnapshot, listItem *inventorydomain.BalanceListItem) dto.InventoryBalanceResponse {
	balance := snapshot.Balance()
	response := dto.InventoryBalanceResponse{
		ItemID:              balance.ItemID().Int64(),
		ItemName:            snapshot.ItemName().Display(),
		BaseUnitCode:        snapshot.BaseUnit().String(),
		ItemArchivedAtMs:    optionalInstant(snapshot.ItemArchivedAt()),
		QuantityAtomic:      balance.Quantity().Int64(),
		InventoryValueMicro: balance.Value().Int64(),
		LastDocumentID:      invOptionalStockDocumentID(balance.LastDocumentID()),
		UpdatedAtMs:         balance.UpdatedAt().UnixMilli(),
	}
	if listItem != nil {
		response.Capabilities = mapCapabilities(listItem.Capabilities())
		response.ReorderQuantity = optionalAtomicQuantityValue(listItem.ReorderQuantity())
	}
	return response
}

func mapLots(lots []inventorydomain.LotView) []dto.LotResponse {
	response := make([]dto.LotResponse, 0, len(lots))
	for _, lot := range lots {
		response = append(response, mapLot(lot))
	}
	return response
}

func mapLot(view inventorydomain.LotView) dto.LotResponse {
	lot := view.Lot()
	return dto.LotResponse{
		ID:                    lot.ID().Int64(),
		ItemID:                lot.ItemID().Int64(),
		SourceLineID:          lot.SourceLineID().Int64(),
		SourcePostingSequence: lot.SourcePostingSequence().Int64(),
		InitialQuantity:       lot.InitialQuantity().Int64(),
		ConsumedQuantity:      lot.ConsumedQuantity().Int64(),
		RestoredQuantity:      lot.RestoredQuantity().Int64(),
		AvailableQuantity:     lot.AvailableQuantity().Int64(),
		LotCode:               optionalText(lot.LotCode()),
		OriginatedOn:          lot.OriginatedOn().String(),
		ExpiresOn:             optionalBusinessDateValue(lot.ExpiresOn()),
		CreatedAtMs:           lot.CreatedAt().UnixMilli(),
		SourceDocumentID:      view.SourceDocumentID().Int64(),
		SourceKind:            view.SourceKind().String(),
		SourceOccurredOn:      view.SourceOccurredOn().String(),
	}
}

func mapLedgerEntryPage(page application.ItemLedgerPage) dto.LedgerEntryPageResponse {
	items := page.Items()
	response := dto.LedgerEntryPageResponse{Items: make([]dto.LedgerEntryResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapLedgerEntry(item))
	}
	if cursor, ok := page.Next().Get(); ok {
		response.Next = &dto.LedgerCursorResponse{
			PostingSequence: cursor.PostingSequence().Int64(),
			LineOrder:       cursor.LineOrder().Int64(),
			LineID:          cursor.LineID().Int64(),
		}
	}
	return response
}

func mapLedgerEntry(view inventorydomain.LedgerEntryView) dto.LedgerEntryResponse {
	entry := view.Entry()
	return dto.LedgerEntryResponse{
		LineID:                    entry.LineID().Int64(),
		DocumentID:                entry.DocumentID().Int64(),
		PostingSequence:           entry.PostingSequence().Int64(),
		LineOrder:                 entry.LineOrder().Int64(),
		DocumentKind:              entry.Kind().String(),
		OccurredOn:                entry.OccurredOn().String(),
		PostedAtMs:                entry.PostedAt().UnixMilli(),
		ItemID:                    entry.ItemID().Int64(),
		Direction:                 entry.Direction().String(),
		QuantityAtomic:            entry.Quantity().Int64(),
		InventoryValueMicro:       entry.Value().Int64(),
		CommercialTotalMinor:      invOptionalMinorAmount(entry.CommercialTotal()),
		CurrencyCode:              entry.Currency().Code().String(),
		CurrencyMinorDigits:       int64(entry.Currency().MinorDigits().Int()),
		EnteredUnitCode:           view.EnteredUnit().String(),
		EnteredPackagingName:      optionalText(view.EnteredPackaging()),
		ConversionNumeratorAtomic: view.Conversion().NumeratorAtomic(),
		ConversionDenominator:     view.Conversion().Denominator(),
		ReversesLineID:            invOptionalStockDocumentLineID(view.ReversesLineID()),
		IdempotencyKey:            view.IdempotencyKey().String(),
		CounterpartyID:            optionalCounterpartyIDValue(view.CounterpartyID()),
		CounterpartyName:          invOptionalDisplayName(view.CounterpartyName()),
		ReasonCode:                optionalDocumentReasonValue(view.Reason()),
		Notes:                     optionalText(view.Notes()),
		ReversesDocumentID:        invOptionalStockDocumentID(view.ReversesDocumentID()),
	}
}

func mapAllocation(view inventorydomain.AllocationView) dto.AllocationResponse {
	allocation := view.Allocation()
	return dto.AllocationResponse{
		ID:                   allocation.ID().Int64(),
		LineID:               allocation.LineID().Int64(),
		LotID:                allocation.LotID().Int64(),
		QuantityAtomic:       allocation.Quantity().Int64(),
		Effect:               allocation.Effect().String(),
		RestoresAllocationID: invOptionalLotAllocationID(allocation.RestoresAllocationID()),
		CreatedAtMs:          allocation.CreatedAt().UnixMilli(),
		SourceLineID:         view.SourceLineID().Int64(),
		LotInitialQuantity:   view.LotInitialQuantity().Int64(),
		LotCode:              optionalText(view.LotCode()),
		OriginatedOn:         view.OriginatedOn().String(),
		ExpiresOn:            optionalBusinessDateValue(view.ExpiresOn()),
	}
}

func invOptionalStockDocumentID(value domain.Option[domain.StockDocumentID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}

func invOptionalStockDocumentLineID(value domain.Option[domain.StockDocumentLineID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}

func invOptionalLotAllocationID(value domain.Option[domain.LotAllocationID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}

func invOptionalMinorAmount(value domain.Option[domain.MinorAmount]) *int64 {
	amount, ok := value.Get()
	if !ok {
		return nil
	}
	raw := amount.Int64()
	return &raw
}

func invOptionalDisplayName(value domain.Option[domain.DisplayName]) *string {
	name, ok := value.Get()
	if !ok {
		return nil
	}
	raw := name.String()
	return &raw
}
