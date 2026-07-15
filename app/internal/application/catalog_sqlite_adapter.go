package application

import (
	"context"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	counterpartydomain "github.com/jerobas/saas/internal/domain/counterparty"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type sqliteCatalogStore struct {
	store *sqlite.Store
}

func NewSQLiteReferenceDataStore(store *sqlite.Store) ReferenceDataStore {
	if store == nil {
		panic("sqlite reference data store requires a store")
	}
	return &sqliteCatalogStore{store: store}
}

func NewSQLiteCounterpartyStore(store *sqlite.Store) CounterpartyStore {
	if store == nil {
		panic("sqlite counterparty store requires a store")
	}
	return &sqliteCatalogStore{store: store}
}

func NewSQLiteCatalogStore(store *sqlite.Store) CatalogStore {
	if store == nil {
		panic("sqlite catalog store requires a store")
	}
	return &sqliteCatalogStore{store: store}
}

func (s *sqliteCatalogStore) GetMeasurementUnit(ctx context.Context, code domain.UnitCode) (catalog.MeasurementUnit, error) {
	return s.store.GetMeasurementUnit(ctx, code)
}

func (s *sqliteCatalogStore) ListMeasurementUnits(ctx context.Context) ([]catalog.MeasurementUnit, error) {
	return s.store.ListMeasurementUnits(ctx)
}

func (s *sqliteCatalogStore) GetCounterparty(ctx context.Context, id domain.CounterpartyID) (counterpartydomain.Counterparty, error) {
	return s.store.GetCounterparty(ctx, id)
}

func (s *sqliteCatalogStore) ListCounterparties(ctx context.Context, input CounterpartyListInput) (CounterpartyPage, error) {
	pageSize, err := sqlite.NewCounterpartyPageSize(input.PageSize)
	if err != nil {
		return CounterpartyPage{}, err
	}
	after := domain.None[sqlite.CounterpartyCursor]()
	if cursor, ok := input.After.Get(); ok {
		after = domain.Some(sqlite.CounterpartyCursor{Name: cursor.Name, ID: cursor.ID})
	}
	page, err := s.store.ListCounterparties(ctx, sqlite.CounterpartyListFilter{
		Archive:  input.Archive,
		Role:     input.Role,
		Search:   input.Search,
		After:    after,
		PageSize: pageSize,
	})
	if err != nil {
		return CounterpartyPage{}, err
	}
	next := domain.None[CounterpartyCursor]()
	if cursor, ok := page.Next().Get(); ok {
		next = domain.Some(CounterpartyCursor{Name: cursor.Name, ID: cursor.ID})
	}
	return NewCounterpartyPage(page.Items(), next), nil
}

func (s *sqliteCatalogStore) CreateCounterparty(ctx context.Context, input counterpartyCreateStoreInput) (counterpartydomain.Counterparty, error) {
	return s.store.CreateCounterparty(ctx, sqlite.CreateCounterpartyInput{
		Name:      input.Name,
		Phone:     input.Phone,
		Email:     input.Email,
		Notes:     input.Notes,
		Roles:     input.Roles,
		CreatedAt: input.CreatedAt,
	})
}

func (s *sqliteCatalogStore) UpdateCounterparty(ctx context.Context, input counterpartyUpdateStoreInput) (counterpartydomain.Counterparty, error) {
	return s.store.UpdateCounterparty(ctx, sqlite.UpdateCounterpartyInput{
		ID:                input.ID,
		Name:              input.Name,
		Phone:             input.Phone,
		Email:             input.Email,
		Notes:             input.Notes,
		Roles:             input.Roles,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		UpdatedAt:         input.UpdatedAt,
	})
}

func (s *sqliteCatalogStore) ArchiveCounterparty(ctx context.Context, input counterpartyArchiveStoreInput) (counterpartydomain.Counterparty, error) {
	return s.store.ArchiveCounterparty(ctx, input.ID, input.ExpectedUpdatedAt, input.ArchivedAt)
}

func (s *sqliteCatalogStore) RestoreCounterparty(ctx context.Context, input counterpartyRestoreStoreInput) (counterpartydomain.Counterparty, error) {
	return s.store.RestoreCounterparty(ctx, input.ID, input.ExpectedUpdatedAt, input.UpdatedAt)
}

func (s *sqliteCatalogStore) GetItem(ctx context.Context, id domain.ItemID) (ItemAggregate, error) {
	item, err := s.store.GetItem(ctx, id)
	if err != nil {
		return ItemAggregate{}, err
	}
	return mapSQLiteItemAggregate(item), nil
}

func (s *sqliteCatalogStore) ListItems(ctx context.Context, input ItemListInput) (ItemPage, error) {
	pageSize, err := sqlite.NewItemPageSize(input.PageSize)
	if err != nil {
		return ItemPage{}, err
	}
	after := domain.None[sqlite.ItemCursor]()
	if cursor, ok := input.After.Get(); ok {
		after = domain.Some(sqlite.ItemCursor{Name: cursor.Name, ID: cursor.ID})
	}
	page, err := s.store.ListItems(ctx, sqlite.ItemListFilter{
		Archive:             input.Archive,
		RequireCapabilities: input.RequireCapabilities,
		Search:              input.Search,
		After:               after,
		PageSize:            pageSize,
	})
	if err != nil {
		return ItemPage{}, err
	}
	next := domain.None[ItemCursor]()
	if cursor, ok := page.Next().Get(); ok {
		next = domain.Some(ItemCursor{Name: cursor.Name, ID: cursor.ID})
	}
	return NewItemPage(page.Items(), next), nil
}

func (s *sqliteCatalogStore) CreateItem(ctx context.Context, input itemCreateStoreInput) (ItemAggregate, error) {
	item, err := s.store.CreateItem(ctx, sqlite.CreateItemInput{
		Name:             input.Name,
		SKU:              input.SKU,
		Description:      input.Description,
		BaseUnit:         input.BaseUnit,
		Capabilities:     input.Capabilities,
		DefaultSalePrice: input.DefaultSalePrice,
		ReorderQuantity:  input.ReorderQuantity,
		CreatedAt:        input.CreatedAt,
		UpdatedAt:        input.UpdatedAt,
	})
	if err != nil {
		return ItemAggregate{}, err
	}
	return mapSQLiteItemAggregate(item), nil
}

func (s *sqliteCatalogStore) UpdateItem(ctx context.Context, input itemUpdateStoreInput) (ItemAggregate, error) {
	item, err := s.store.UpdateItem(ctx, sqlite.UpdateItemInput{
		ID:                input.ID,
		Name:              input.Name,
		SKU:               input.SKU,
		Description:       input.Description,
		BaseUnit:          input.BaseUnit,
		Capabilities:      input.Capabilities,
		DefaultSalePrice:  input.DefaultSalePrice,
		ReorderQuantity:   input.ReorderQuantity,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		UpdatedAt:         input.UpdatedAt,
	})
	if err != nil {
		return ItemAggregate{}, err
	}
	return mapSQLiteItemAggregate(item), nil
}

func (s *sqliteCatalogStore) ArchiveItem(ctx context.Context, input itemArchiveStoreInput) (ItemAggregate, error) {
	item, err := s.store.ArchiveItem(ctx, sqlite.ArchiveItemInput{
		ID:                input.ID,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		ArchivedAt:        input.ArchivedAt,
	})
	if err != nil {
		return ItemAggregate{}, err
	}
	return mapSQLiteItemAggregate(item), nil
}

func (s *sqliteCatalogStore) RestoreItem(ctx context.Context, input itemRestoreStoreInput) (ItemAggregate, error) {
	item, err := s.store.RestoreItem(ctx, sqlite.RestoreItemInput{
		ID:                input.ID,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		UpdatedAt:         input.UpdatedAt,
	})
	if err != nil {
		return ItemAggregate{}, err
	}
	return mapSQLiteItemAggregate(item), nil
}

func (s *sqliteCatalogStore) GetItemPackaging(ctx context.Context, id domain.PackagingID) (PackagingAggregate, error) {
	packaging, err := s.store.GetItemPackaging(ctx, id)
	if err != nil {
		return PackagingAggregate{}, err
	}
	return mapSQLitePackagingAggregate(packaging), nil
}

func (s *sqliteCatalogStore) CreatePackaging(ctx context.Context, input packagingCreateStoreInput) (PackagingAggregate, error) {
	packaging, err := s.store.CreatePackaging(ctx, sqlite.CreatePackagingInput{
		ItemID:      input.ItemID,
		Name:        input.Name,
		EnteredUnit: input.EnteredUnit,
		Conversion:  input.Conversion,
		CreatedAt:   input.CreatedAt,
		UpdatedAt:   input.UpdatedAt,
	})
	if err != nil {
		return PackagingAggregate{}, err
	}
	return mapSQLitePackagingAggregate(packaging), nil
}

func (s *sqliteCatalogStore) UpdatePackaging(ctx context.Context, input packagingUpdateStoreInput) (PackagingAggregate, error) {
	packaging, err := s.store.UpdatePackaging(ctx, sqlite.UpdatePackagingInput{
		ID:                input.ID,
		Name:              input.Name,
		EnteredUnit:       input.EnteredUnit,
		Conversion:        input.Conversion,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		UpdatedAt:         input.UpdatedAt,
	})
	if err != nil {
		return PackagingAggregate{}, err
	}
	return mapSQLitePackagingAggregate(packaging), nil
}

func (s *sqliteCatalogStore) ArchivePackaging(ctx context.Context, input packagingArchiveStoreInput) (PackagingAggregate, error) {
	packaging, err := s.store.ArchivePackaging(ctx, sqlite.ArchivePackagingInput{
		ID:                input.ID,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		ArchivedAt:        input.ArchivedAt,
	})
	if err != nil {
		return PackagingAggregate{}, err
	}
	return mapSQLitePackagingAggregate(packaging), nil
}

func (s *sqliteCatalogStore) ReconfigureArchivedPackaging(ctx context.Context, input packagingReconfigureStoreInput) (PackagingAggregate, error) {
	packaging, err := s.store.ReconfigureArchivedPackaging(ctx, sqlite.ReconfigureArchivedPackagingInput{
		ID:                input.ID,
		Name:              input.Name,
		EnteredUnit:       input.EnteredUnit,
		Conversion:        input.Conversion,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		UpdatedAt:         input.UpdatedAt,
	})
	if err != nil {
		return PackagingAggregate{}, err
	}
	return mapSQLitePackagingAggregate(packaging), nil
}

func (s *sqliteCatalogStore) RestorePackaging(ctx context.Context, input packagingRestoreStoreInput) (PackagingAggregate, error) {
	packaging, err := s.store.RestorePackaging(ctx, sqlite.RestorePackagingInput{
		ID:                input.ID,
		ExpectedUpdatedAt: input.ExpectedUpdatedAt,
		UpdatedAt:         input.UpdatedAt,
	})
	if err != nil {
		return PackagingAggregate{}, err
	}
	return mapSQLitePackagingAggregate(packaging), nil
}

func mapSQLiteItemAggregate(item sqlite.ItemAggregate) ItemAggregate {
	packagings := item.Packagings()
	mappedPackagings := make([]PackagingAggregate, 0, len(packagings))
	for _, packaging := range packagings {
		mappedPackagings = append(mappedPackagings, mapSQLitePackagingAggregate(packaging))
	}
	return NewItemAggregate(item.Item(), item.BaseUnit(), mappedPackagings)
}

func mapSQLitePackagingAggregate(packaging sqlite.PackagingAggregate) PackagingAggregate {
	return NewPackagingAggregate(packaging.Packaging(), packaging.BaseUnit(), packaging.EnteredUnit())
}
