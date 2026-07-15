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
