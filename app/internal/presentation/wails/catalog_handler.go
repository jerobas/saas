package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type CatalogHandler struct {
	service *application.CatalogService
}

func NewCatalogHandler(service *application.CatalogService) *CatalogHandler {
	if service == nil {
		panic("catalog handler requires a service")
	}
	return &CatalogHandler{service: service}
}

func (h *CatalogHandler) GetItem(id int64) (dto.ItemResponse, error) {
	itemID, err := domain.NewItemID(id)
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("item id: %w", err)
	}
	item, err := h.service.GetItem(handlerContext(), itemID)
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("get item: %w", err)
	}
	return mapItem(item), nil
}

func (h *CatalogHandler) ListItems(req dto.ItemListRequest) (dto.ItemPageResponse, error) {
	input, err := parseItemListRequest(req)
	if err != nil {
		return dto.ItemPageResponse{}, err
	}
	page, err := h.service.ListItems(handlerContext(), input)
	if err != nil {
		return dto.ItemPageResponse{}, fmt.Errorf("list items: %w", err)
	}
	return mapItemPage(page), nil
}

func (h *CatalogHandler) CreateItem(req dto.ItemWriteRequest) (dto.ItemResponse, error) {
	input, err := parseItemWriteRequest(req)
	if err != nil {
		return dto.ItemResponse{}, err
	}
	item, err := h.service.CreateItem(handlerContext(), application.ItemCreateInput{ItemWriteInput: input})
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("create item: %w", err)
	}
	return mapItem(item), nil
}

func (h *CatalogHandler) UpdateItem(id int64, req dto.ItemUpdateRequest) (dto.ItemResponse, error) {
	itemID, err := domain.NewItemID(id)
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("item id: %w", err)
	}
	input, err := parseItemWriteRequest(req.ItemWriteRequest)
	if err != nil {
		return dto.ItemResponse{}, err
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("expected updated at: %w", err)
	}
	item, err := h.service.UpdateItem(handlerContext(), application.ItemUpdateInput{
		ID: itemID, ItemWriteInput: input, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("update item: %w", err)
	}
	return mapItem(item), nil
}

func (h *CatalogHandler) ArchiveItem(id int64, req dto.VersionedRequest) (dto.ItemResponse, error) {
	itemID, expectedUpdatedAt, err := parseVersionedItem(id, req)
	if err != nil {
		return dto.ItemResponse{}, err
	}
	item, err := h.service.ArchiveItem(handlerContext(), application.ItemArchiveInput{
		ID: itemID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("archive item: %w", err)
	}
	return mapItem(item), nil
}

func (h *CatalogHandler) RestoreItem(id int64, req dto.VersionedRequest) (dto.ItemResponse, error) {
	itemID, expectedUpdatedAt, err := parseVersionedItem(id, req)
	if err != nil {
		return dto.ItemResponse{}, err
	}
	item, err := h.service.RestoreItem(handlerContext(), application.ItemRestoreInput{
		ID: itemID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.ItemResponse{}, fmt.Errorf("restore item: %w", err)
	}
	return mapItem(item), nil
}

func (h *CatalogHandler) GetItemPackaging(id int64) (dto.PackagingResponse, error) {
	packagingID, err := domain.NewPackagingID(id)
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("packaging id: %w", err)
	}
	packaging, err := h.service.GetItemPackaging(handlerContext(), packagingID)
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("get item packaging: %w", err)
	}
	return mapPackaging(packaging), nil
}

func (h *CatalogHandler) CreateItemPackaging(req dto.PackagingCreateRequest) (dto.PackagingResponse, error) {
	itemID, err := domain.NewItemID(req.ItemID)
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("item id: %w", err)
	}
	input, err := parsePackagingWriteRequest(req.PackagingWriteRequest)
	if err != nil {
		return dto.PackagingResponse{}, err
	}
	packaging, err := h.service.CreatePackaging(handlerContext(), application.PackagingCreateInput{
		ItemID: itemID, PackagingWriteInput: input,
	})
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("create item packaging: %w", err)
	}
	return mapPackaging(packaging), nil
}

func (h *CatalogHandler) UpdateItemPackaging(id int64, req dto.PackagingUpdateRequest) (dto.PackagingResponse, error) {
	packagingID, err := domain.NewPackagingID(id)
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("packaging id: %w", err)
	}
	input, err := parsePackagingWriteRequest(req.PackagingWriteRequest)
	if err != nil {
		return dto.PackagingResponse{}, err
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("expected updated at: %w", err)
	}
	packaging, err := h.service.UpdatePackaging(handlerContext(), application.PackagingUpdateInput{
		ID: packagingID, PackagingWriteInput: input, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("update item packaging: %w", err)
	}
	return mapPackaging(packaging), nil
}

func (h *CatalogHandler) ArchiveItemPackaging(id int64, req dto.VersionedRequest) (dto.PackagingResponse, error) {
	packagingID, expectedUpdatedAt, err := parseVersionedPackaging(id, req)
	if err != nil {
		return dto.PackagingResponse{}, err
	}
	packaging, err := h.service.ArchivePackaging(handlerContext(), application.PackagingArchiveInput{
		ID: packagingID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("archive item packaging: %w", err)
	}
	return mapPackaging(packaging), nil
}

func (h *CatalogHandler) ReconfigureArchivedItemPackaging(id int64, req dto.PackagingUpdateRequest) (dto.PackagingResponse, error) {
	packagingID, err := domain.NewPackagingID(id)
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("packaging id: %w", err)
	}
	input, err := parsePackagingWriteRequest(req.PackagingWriteRequest)
	if err != nil {
		return dto.PackagingResponse{}, err
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("expected updated at: %w", err)
	}
	packaging, err := h.service.ReconfigureArchivedPackaging(handlerContext(), application.PackagingReconfigureInput{
		ID: packagingID, PackagingWriteInput: input, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("reconfigure archived item packaging: %w", err)
	}
	return mapPackaging(packaging), nil
}

func (h *CatalogHandler) RestoreItemPackaging(id int64, req dto.VersionedRequest) (dto.PackagingResponse, error) {
	packagingID, expectedUpdatedAt, err := parseVersionedPackaging(id, req)
	if err != nil {
		return dto.PackagingResponse{}, err
	}
	packaging, err := h.service.RestorePackaging(handlerContext(), application.PackagingRestoreInput{
		ID: packagingID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.PackagingResponse{}, fmt.Errorf("restore item packaging: %w", err)
	}
	return mapPackaging(packaging), nil
}

func parseItemListRequest(req dto.ItemListRequest) (application.ItemListInput, error) {
	archive := domain.ArchiveActive
	if req.ArchiveFilter != "" {
		parsed, err := domain.ParseArchiveFilter(req.ArchiveFilter)
		if err != nil {
			return application.ItemListInput{}, err
		}
		archive = parsed
	}
	search, err := optionalNonEmptyText(req.Search)
	if err != nil {
		return application.ItemListInput{}, fmt.Errorf("search: %w", err)
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 50
	}
	after := domain.None[application.ItemCursor]()
	if req.After != nil {
		name, err := domain.NewUniqueName(req.After.Name)
		if err != nil {
			return application.ItemListInput{}, fmt.Errorf("cursor name: %w", err)
		}
		id, err := domain.NewItemID(req.After.ID)
		if err != nil {
			return application.ItemListInput{}, fmt.Errorf("cursor id: %w", err)
		}
		after = domain.Some(application.ItemCursor{Name: name, ID: id})
	}
	return application.ItemListInput{
		Archive:             archive,
		RequireCapabilities: parseCapabilities(req.RequireCapabilities),
		Search:              search,
		After:               after,
		PageSize:            pageSize,
	}, nil
}

func parseItemWriteRequest(req dto.ItemWriteRequest) (application.ItemWriteInput, error) {
	name, err := domain.NewUniqueName(req.Name)
	if err != nil {
		return application.ItemWriteInput{}, fmt.Errorf("name: %w", err)
	}
	sku, err := optionalSKU(req.SKU)
	if err != nil {
		return application.ItemWriteInput{}, fmt.Errorf("sku: %w", err)
	}
	description, err := optionalNonEmptyText(req.Description)
	if err != nil {
		return application.ItemWriteInput{}, fmt.Errorf("description: %w", err)
	}
	baseUnit, err := domain.NewUnitCode(req.BaseUnitCode)
	if err != nil {
		return application.ItemWriteInput{}, fmt.Errorf("base unit: %w", err)
	}
	defaultSalePrice, err := optionalMinorAmountInput(req.DefaultSalePrice)
	if err != nil {
		return application.ItemWriteInput{}, fmt.Errorf("default sale price: %w", err)
	}
	reorderQuantity, err := optionalAtomicQuantity(req.ReorderQuantity)
	if err != nil {
		return application.ItemWriteInput{}, fmt.Errorf("reorder quantity: %w", err)
	}
	return application.ItemWriteInput{
		Name:             name,
		SKU:              sku,
		Description:      description,
		BaseUnit:         baseUnit,
		Capabilities:     parseCapabilities(req.Capabilities),
		DefaultSalePrice: defaultSalePrice,
		ReorderQuantity:  reorderQuantity,
	}, nil
}

func parsePackagingWriteRequest(req dto.PackagingWriteRequest) (application.PackagingWriteInput, error) {
	name, err := domain.NewUniqueName(req.Name)
	if err != nil {
		return application.PackagingWriteInput{}, fmt.Errorf("name: %w", err)
	}
	enteredUnit, err := domain.NewUnitCode(req.EnteredUnitCode)
	if err != nil {
		return application.PackagingWriteInput{}, fmt.Errorf("entered unit: %w", err)
	}
	conversion, err := domain.NewUnitConversion(req.ConversionNumerator, req.ConversionDenominator)
	if err != nil {
		return application.PackagingWriteInput{}, fmt.Errorf("conversion: %w", err)
	}
	return application.PackagingWriteInput{Name: name, EnteredUnit: enteredUnit, Conversion: conversion}, nil
}

func parseCapabilities(req dto.CapabilitiesRequest) catalog.Capabilities {
	return catalog.NewCapabilities(req.Purchasable, req.Producible, req.Sellable)
}

func mapCapabilities(value catalog.Capabilities) dto.CapabilitiesResponse {
	return dto.CapabilitiesResponse{
		Purchasable: value.Purchasable(),
		Producible:  value.Producible(),
		Sellable:    value.Sellable(),
	}
}

func parseVersionedItem(id int64, req dto.VersionedRequest) (domain.ItemID, domain.UTCInstant, error) {
	itemID, err := domain.NewItemID(id)
	if err != nil {
		return domain.ItemID{}, domain.UTCInstant{}, fmt.Errorf("item id: %w", err)
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return domain.ItemID{}, domain.UTCInstant{}, fmt.Errorf("expected updated at: %w", err)
	}
	return itemID, expectedUpdatedAt, nil
}

func parseVersionedPackaging(id int64, req dto.VersionedRequest) (domain.PackagingID, domain.UTCInstant, error) {
	packagingID, err := domain.NewPackagingID(id)
	if err != nil {
		return domain.PackagingID{}, domain.UTCInstant{}, fmt.Errorf("packaging id: %w", err)
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return domain.PackagingID{}, domain.UTCInstant{}, fmt.Errorf("expected updated at: %w", err)
	}
	return packagingID, expectedUpdatedAt, nil
}

func mapItemPage(page application.ItemPage) dto.ItemPageResponse {
	items := page.Items()
	response := dto.ItemPageResponse{Items: make([]dto.ItemSummaryResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapItemSummary(item))
	}
	if cursor, ok := page.Next().Get(); ok {
		response.Next = &dto.ItemCursorResponse{Name: cursor.Name.Display(), ID: cursor.ID.Int64()}
	}
	return response
}

func mapItem(item application.ItemAggregate) dto.ItemResponse {
	itemValue := item.Item()
	packagings := item.Packagings()
	response := dto.ItemResponse{
		ItemSummaryResponse: mapCatalogItemFields(itemValue),
		BaseUnit:            mapMeasurementUnit(item.BaseUnit()),
		Packagings:          make([]dto.PackagingResponse, 0, len(packagings)),
	}
	for _, packaging := range packagings {
		response.Packagings = append(response.Packagings, mapPackaging(packaging))
	}
	return response
}

func mapItemSummary(item catalog.ItemSummary) dto.ItemSummaryResponse {
	return dto.ItemSummaryResponse{
		ID:               item.ID().Int64(),
		Name:             item.Name().Display(),
		SKU:              optionalSKUText(item.SKU()),
		Description:      optionalText(item.Description()),
		BaseUnitCode:     item.BaseUnit().String(),
		Capabilities:     mapCapabilities(item.Capabilities()),
		DefaultSalePrice: optionalMinorAmount(item.DefaultSalePrice()),
		ReorderQuantity:  optionalAtomicQuantityValue(item.ReorderQuantity()),
		CreatedAtMs:      item.CreatedAt().UnixMilli(),
		UpdatedAtMs:      item.UpdatedAt().UnixMilli(),
		ArchivedAtMs:     optionalInstant(item.ArchivedAt()),
	}
}

func mapCatalogItemFields(item catalog.Item) dto.ItemSummaryResponse {
	return dto.ItemSummaryResponse{
		ID:               item.ID().Int64(),
		Name:             item.Name().Display(),
		SKU:              optionalSKUText(item.SKU()),
		Description:      optionalText(item.Description()),
		BaseUnitCode:     item.BaseUnit().String(),
		Capabilities:     mapCapabilities(item.Capabilities()),
		DefaultSalePrice: optionalMinorAmount(item.DefaultSalePrice()),
		ReorderQuantity:  optionalAtomicQuantityValue(item.ReorderQuantity()),
		CreatedAtMs:      item.CreatedAt().UnixMilli(),
		UpdatedAtMs:      item.UpdatedAt().UnixMilli(),
		ArchivedAtMs:     optionalInstant(item.ArchivedAt()),
	}
}

func mapPackaging(packaging application.PackagingAggregate) dto.PackagingResponse {
	value := packaging.Packaging()
	return dto.PackagingResponse{
		ID:                    value.ID().Int64(),
		ItemID:                value.ItemID().Int64(),
		Name:                  value.Name().Display(),
		EnteredUnitCode:       value.EnteredUnit().String(),
		ConversionNumerator:   value.Conversion().NumeratorAtomic(),
		ConversionDenominator: value.Conversion().Denominator(),
		BaseUnit:              mapMeasurementUnit(packaging.BaseUnit()),
		EnteredUnit:           mapMeasurementUnit(packaging.EnteredUnit()),
		CreatedAtMs:           value.CreatedAt().UnixMilli(),
		UpdatedAtMs:           value.UpdatedAt().UnixMilli(),
		ArchivedAtMs:          optionalInstant(value.ArchivedAt()),
	}
}

func optionalSKU(value *string) (domain.Option[domain.SKU], error) {
	if value == nil {
		return domain.None[domain.SKU](), nil
	}
	sku, err := domain.NewSKU(*value)
	if err != nil {
		return domain.None[domain.SKU](), err
	}
	return domain.Some(sku), nil
}

func optionalMinorAmountInput(value *int64) (domain.Option[domain.MinorAmount], error) {
	if value == nil {
		return domain.None[domain.MinorAmount](), nil
	}
	amount, err := domain.NewMinorAmount(*value)
	if err != nil {
		return domain.None[domain.MinorAmount](), err
	}
	return domain.Some(amount), nil
}

func optionalAtomicQuantity(value *int64) (domain.Option[domain.AtomicQuantity], error) {
	if value == nil {
		return domain.None[domain.AtomicQuantity](), nil
	}
	quantity, err := domain.NewAtomicQuantity(*value)
	if err != nil {
		return domain.None[domain.AtomicQuantity](), err
	}
	return domain.Some(quantity), nil
}

func optionalSKUText(value domain.Option[domain.SKU]) *string {
	sku, ok := value.Get()
	if !ok {
		return nil
	}
	raw := sku.Display()
	return &raw
}

func optionalAtomicQuantityValue(value domain.Option[domain.AtomicQuantity]) *int64 {
	quantity, ok := value.Get()
	if !ok {
		return nil
	}
	raw := quantity.Int64()
	return &raw
}
