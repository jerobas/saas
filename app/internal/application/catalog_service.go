package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

type CatalogStore interface {
	GetItem(ctx context.Context, id domain.ItemID) (ItemAggregate, error)
	ListItems(ctx context.Context, input ItemListInput) (ItemPage, error)
	CreateItem(ctx context.Context, input itemCreateStoreInput) (ItemAggregate, error)
	UpdateItem(ctx context.Context, input itemUpdateStoreInput) (ItemAggregate, error)
	ArchiveItem(ctx context.Context, input itemArchiveStoreInput) (ItemAggregate, error)
	RestoreItem(ctx context.Context, input itemRestoreStoreInput) (ItemAggregate, error)
	GetItemPackaging(ctx context.Context, id domain.PackagingID) (PackagingAggregate, error)
	CreatePackaging(ctx context.Context, input packagingCreateStoreInput) (PackagingAggregate, error)
	UpdatePackaging(ctx context.Context, input packagingUpdateStoreInput) (PackagingAggregate, error)
	ArchivePackaging(ctx context.Context, input packagingArchiveStoreInput) (PackagingAggregate, error)
	ReconfigureArchivedPackaging(ctx context.Context, input packagingReconfigureStoreInput) (PackagingAggregate, error)
	RestorePackaging(ctx context.Context, input packagingRestoreStoreInput) (PackagingAggregate, error)
}

type ItemCursor struct {
	Name domain.UniqueName
	ID   domain.ItemID
}

type ItemListInput struct {
	Archive             domain.ArchiveFilter
	RequireCapabilities catalog.Capabilities
	Search              domain.Option[domain.NonEmptyText]
	After               domain.Option[ItemCursor]
	PageSize            int
}

type ItemPage struct {
	items []catalog.ItemSummary
	next  domain.Option[ItemCursor]
}

func NewItemPage(items []catalog.ItemSummary, next domain.Option[ItemCursor]) ItemPage {
	cloned := make([]catalog.ItemSummary, len(items))
	copy(cloned, items)
	return ItemPage{items: cloned, next: next}
}

func (p ItemPage) Items() []catalog.ItemSummary {
	items := make([]catalog.ItemSummary, len(p.items))
	copy(items, p.items)
	return items
}

func (p ItemPage) Next() domain.Option[ItemCursor] { return p.next }

type PackagingAggregate struct {
	packaging   catalog.ItemPackaging
	baseUnit    catalog.MeasurementUnit
	enteredUnit catalog.MeasurementUnit
}

func NewPackagingAggregate(packaging catalog.ItemPackaging, baseUnit, enteredUnit catalog.MeasurementUnit) PackagingAggregate {
	return PackagingAggregate{packaging: packaging, baseUnit: baseUnit, enteredUnit: enteredUnit}
}

func (a PackagingAggregate) Packaging() catalog.ItemPackaging     { return a.packaging }
func (a PackagingAggregate) BaseUnit() catalog.MeasurementUnit    { return a.baseUnit }
func (a PackagingAggregate) EnteredUnit() catalog.MeasurementUnit { return a.enteredUnit }

type ItemAggregate struct {
	item       catalog.Item
	baseUnit   catalog.MeasurementUnit
	packagings []PackagingAggregate
}

func NewItemAggregate(item catalog.Item, baseUnit catalog.MeasurementUnit, packagings []PackagingAggregate) ItemAggregate {
	cloned := make([]PackagingAggregate, len(packagings))
	copy(cloned, packagings)
	return ItemAggregate{item: item, baseUnit: baseUnit, packagings: cloned}
}

func (a ItemAggregate) Item() catalog.Item                { return a.item }
func (a ItemAggregate) BaseUnit() catalog.MeasurementUnit { return a.baseUnit }
func (a ItemAggregate) Packagings() []PackagingAggregate {
	packagings := make([]PackagingAggregate, len(a.packagings))
	copy(packagings, a.packagings)
	return packagings
}

type ItemWriteInput struct {
	Name             domain.UniqueName
	SKU              domain.Option[domain.SKU]
	Description      domain.Option[domain.NonEmptyText]
	BaseUnit         domain.UnitCode
	Capabilities     catalog.Capabilities
	DefaultSalePrice domain.Option[domain.MinorAmount]
	ReorderQuantity  domain.Option[domain.AtomicQuantity]
}

type ItemCreateInput struct {
	ItemWriteInput
}

type ItemUpdateInput struct {
	ID domain.ItemID
	ItemWriteInput
	ExpectedUpdatedAt domain.UTCInstant
}

type ItemArchiveInput struct {
	ID                domain.ItemID
	ExpectedUpdatedAt domain.UTCInstant
}

type ItemRestoreInput struct {
	ID                domain.ItemID
	ExpectedUpdatedAt domain.UTCInstant
}

type itemCreateStoreInput struct {
	ItemCreateInput
	CreatedAt domain.UTCInstant
	UpdatedAt domain.UTCInstant
}

type itemUpdateStoreInput struct {
	ItemUpdateInput
	UpdatedAt domain.UTCInstant
}

type itemArchiveStoreInput struct {
	ItemArchiveInput
	ArchivedAt domain.UTCInstant
}

type itemRestoreStoreInput struct {
	ItemRestoreInput
	UpdatedAt domain.UTCInstant
}

type PackagingWriteInput struct {
	Name        domain.UniqueName
	EnteredUnit domain.UnitCode
	Conversion  domain.UnitConversion
}

type PackagingCreateInput struct {
	ItemID domain.ItemID
	PackagingWriteInput
}

type PackagingUpdateInput struct {
	ID domain.PackagingID
	PackagingWriteInput
	ExpectedUpdatedAt domain.UTCInstant
}

type PackagingArchiveInput struct {
	ID                domain.PackagingID
	ExpectedUpdatedAt domain.UTCInstant
}

type PackagingReconfigureInput struct {
	ID domain.PackagingID
	PackagingWriteInput
	ExpectedUpdatedAt domain.UTCInstant
}

type PackagingRestoreInput struct {
	ID                domain.PackagingID
	ExpectedUpdatedAt domain.UTCInstant
}

type packagingCreateStoreInput struct {
	PackagingCreateInput
	CreatedAt domain.UTCInstant
	UpdatedAt domain.UTCInstant
}

type packagingUpdateStoreInput struct {
	PackagingUpdateInput
	UpdatedAt domain.UTCInstant
}

type packagingArchiveStoreInput struct {
	PackagingArchiveInput
	ArchivedAt domain.UTCInstant
}

type packagingReconfigureStoreInput struct {
	PackagingReconfigureInput
	UpdatedAt domain.UTCInstant
}

type packagingRestoreStoreInput struct {
	PackagingRestoreInput
	UpdatedAt domain.UTCInstant
}

type CatalogService struct {
	store CatalogStore
	clock Clock
}

func NewCatalogService(store CatalogStore, clock Clock) *CatalogService {
	if store == nil {
		panic("catalog service requires a store")
	}
	if clock == nil {
		panic("catalog service requires a clock")
	}
	return &CatalogService{store: store, clock: clock}
}

func (s *CatalogService) GetItem(ctx context.Context, id domain.ItemID) (ItemAggregate, error) {
	item, err := s.store.GetItem(ctx, id)
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("get item: %w", err)
	}
	return item, nil
}

func (s *CatalogService) ListItems(ctx context.Context, input ItemListInput) (ItemPage, error) {
	page, err := s.store.ListItems(ctx, input)
	if err != nil {
		return ItemPage{}, fmt.Errorf("list items: %w", err)
	}
	return page, nil
}

func (s *CatalogService) CreateItem(ctx context.Context, input ItemCreateInput) (ItemAggregate, error) {
	now, err := s.clock.Now()
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	item, err := s.store.CreateItem(ctx, itemCreateStoreInput{
		ItemCreateInput: input,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("create item: %w", err)
	}
	if !item.Item().CreatedAt().Equal(now) || !item.Item().UpdatedAt().Equal(now) {
		return ItemAggregate{}, domain.ErrInvariant
	}
	return item, nil
}

func (s *CatalogService) UpdateItem(ctx context.Context, input ItemUpdateInput) (ItemAggregate, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	item, err := s.store.UpdateItem(ctx, itemUpdateStoreInput{ItemUpdateInput: input, UpdatedAt: now})
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("update item: %w", err)
	}
	if !item.Item().UpdatedAt().Equal(now) {
		return ItemAggregate{}, domain.ErrInvariant
	}
	return item, nil
}

func (s *CatalogService) ArchiveItem(ctx context.Context, input ItemArchiveInput) (ItemAggregate, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	item, err := s.store.ArchiveItem(ctx, itemArchiveStoreInput{ItemArchiveInput: input, ArchivedAt: now})
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("archive item: %w", err)
	}
	archivedAt, ok := item.Item().ArchivedAt().Get()
	if !ok || !archivedAt.Equal(now) {
		return ItemAggregate{}, domain.ErrInvariant
	}
	return item, nil
}

func (s *CatalogService) RestoreItem(ctx context.Context, input ItemRestoreInput) (ItemAggregate, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	item, err := s.store.RestoreItem(ctx, itemRestoreStoreInput{ItemRestoreInput: input, UpdatedAt: now})
	if err != nil {
		return ItemAggregate{}, fmt.Errorf("restore item: %w", err)
	}
	if item.Item().IsArchived() || !item.Item().UpdatedAt().Equal(now) {
		return ItemAggregate{}, domain.ErrInvariant
	}
	return item, nil
}

func (s *CatalogService) GetItemPackaging(ctx context.Context, id domain.PackagingID) (PackagingAggregate, error) {
	packaging, err := s.store.GetItemPackaging(ctx, id)
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("get item packaging: %w", err)
	}
	return packaging, nil
}

func (s *CatalogService) CreatePackaging(ctx context.Context, input PackagingCreateInput) (PackagingAggregate, error) {
	now, err := s.clock.Now()
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	packaging, err := s.store.CreatePackaging(ctx, packagingCreateStoreInput{
		PackagingCreateInput: input,
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("create item packaging: %w", err)
	}
	if !packaging.Packaging().CreatedAt().Equal(now) || !packaging.Packaging().UpdatedAt().Equal(now) {
		return PackagingAggregate{}, domain.ErrInvariant
	}
	return packaging, nil
}

func (s *CatalogService) UpdatePackaging(ctx context.Context, input PackagingUpdateInput) (PackagingAggregate, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	packaging, err := s.store.UpdatePackaging(ctx, packagingUpdateStoreInput{PackagingUpdateInput: input, UpdatedAt: now})
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("update item packaging: %w", err)
	}
	if !packaging.Packaging().UpdatedAt().Equal(now) {
		return PackagingAggregate{}, domain.ErrInvariant
	}
	return packaging, nil
}

func (s *CatalogService) ArchivePackaging(ctx context.Context, input PackagingArchiveInput) (PackagingAggregate, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	packaging, err := s.store.ArchivePackaging(ctx, packagingArchiveStoreInput{PackagingArchiveInput: input, ArchivedAt: now})
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("archive item packaging: %w", err)
	}
	archivedAt, ok := packaging.Packaging().ArchivedAt().Get()
	if !ok || !archivedAt.Equal(now) {
		return PackagingAggregate{}, domain.ErrInvariant
	}
	return packaging, nil
}

func (s *CatalogService) ReconfigureArchivedPackaging(ctx context.Context, input PackagingReconfigureInput) (PackagingAggregate, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	packaging, err := s.store.ReconfigureArchivedPackaging(ctx, packagingReconfigureStoreInput{
		PackagingReconfigureInput: input,
		UpdatedAt:                 now,
	})
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("reconfigure archived item packaging: %w", err)
	}
	archivedAt, ok := packaging.Packaging().ArchivedAt().Get()
	if !ok || !archivedAt.Equal(now) {
		return PackagingAggregate{}, domain.ErrInvariant
	}
	return packaging, nil
}

func (s *CatalogService) RestorePackaging(ctx context.Context, input PackagingRestoreInput) (PackagingAggregate, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("read clock: %w", err)
	}
	packaging, err := s.store.RestorePackaging(ctx, packagingRestoreStoreInput{PackagingRestoreInput: input, UpdatedAt: now})
	if err != nil {
		return PackagingAggregate{}, fmt.Errorf("restore item packaging: %w", err)
	}
	if packaging.Packaging().IsArchived() || !packaging.Packaging().UpdatedAt().Equal(now) {
		return PackagingAggregate{}, domain.ErrInvariant
	}
	return packaging, nil
}
