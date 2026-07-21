package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	counterpartydomain "github.com/jerobas/saas/internal/domain/counterparty"
)

type CounterpartyStore interface {
	GetCounterparty(ctx context.Context, id domain.CounterpartyID) (counterpartydomain.Counterparty, error)
	ListCounterparties(ctx context.Context, input CounterpartyListInput) (CounterpartyPage, error)
	CreateCounterparty(ctx context.Context, input counterpartyCreateStoreInput) (counterpartydomain.Counterparty, error)
	UpdateCounterparty(ctx context.Context, input counterpartyUpdateStoreInput) (counterpartydomain.Counterparty, error)
	ArchiveCounterparty(ctx context.Context, input counterpartyArchiveStoreInput) (counterpartydomain.Counterparty, error)
	RestoreCounterparty(ctx context.Context, input counterpartyRestoreStoreInput) (counterpartydomain.Counterparty, error)
}

type CounterpartyCursor struct {
	Name domain.DisplayName
	ID   domain.CounterpartyID
}

type CounterpartyListInput struct {
	Archive  domain.ArchiveFilter
	Role     domain.Option[domain.CounterpartyRole]
	Search   domain.Option[domain.NonEmptyText]
	After    domain.Option[CounterpartyCursor]
	PageSize int
}

type CounterpartyPage struct {
	items []counterpartydomain.Counterparty
	next  domain.Option[CounterpartyCursor]
}

func NewCounterpartyPage(items []counterpartydomain.Counterparty, next domain.Option[CounterpartyCursor]) CounterpartyPage {
	cloned := make([]counterpartydomain.Counterparty, len(items))
	copy(cloned, items)
	return CounterpartyPage{items: cloned, next: next}
}

func (p CounterpartyPage) Items() []counterpartydomain.Counterparty {
	items := make([]counterpartydomain.Counterparty, len(p.items))
	copy(items, p.items)
	return items
}

func (p CounterpartyPage) Next() domain.Option[CounterpartyCursor] { return p.next }

type CounterpartyCreateInput struct {
	Name  domain.DisplayName
	Phone domain.Option[domain.NonEmptyText]
	Email domain.Option[domain.NonEmptyText]
	Notes domain.Option[domain.NonEmptyText]
	Roles counterpartydomain.RoleSet
}

type CounterpartyUpdateInput struct {
	ID                domain.CounterpartyID
	Name              domain.DisplayName
	Phone             domain.Option[domain.NonEmptyText]
	Email             domain.Option[domain.NonEmptyText]
	Notes             domain.Option[domain.NonEmptyText]
	Roles             counterpartydomain.RoleSet
	ExpectedUpdatedAt domain.UTCInstant
}

type CounterpartyArchiveInput struct {
	ID                domain.CounterpartyID
	ExpectedUpdatedAt domain.UTCInstant
}

type CounterpartyRestoreInput struct {
	ID                domain.CounterpartyID
	ExpectedUpdatedAt domain.UTCInstant
}

type counterpartyCreateStoreInput struct {
	CounterpartyCreateInput
	CreatedAt domain.UTCInstant
}

type counterpartyUpdateStoreInput struct {
	CounterpartyUpdateInput
	UpdatedAt domain.UTCInstant
}

type counterpartyArchiveStoreInput struct {
	CounterpartyArchiveInput
	ArchivedAt domain.UTCInstant
}

type counterpartyRestoreStoreInput struct {
	CounterpartyRestoreInput
	UpdatedAt domain.UTCInstant
}

type CounterpartyService struct {
	store CounterpartyStore
	clock Clock
}

func NewCounterpartyService(store CounterpartyStore, clock Clock) *CounterpartyService {
	if store == nil {
		panic("counterparty service requires a store")
	}
	if clock == nil {
		panic("counterparty service requires a clock")
	}
	return &CounterpartyService{store: store, clock: clock}
}

func (s *CounterpartyService) GetCounterparty(ctx context.Context, id domain.CounterpartyID) (counterpartydomain.Counterparty, error) {
	value, err := s.store.GetCounterparty(ctx, id)
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("get counterparty: %w", err)
	}
	return value, nil
}

func (s *CounterpartyService) ListCounterparties(ctx context.Context, input CounterpartyListInput) (CounterpartyPage, error) {
	page, err := s.store.ListCounterparties(ctx, input)
	if err != nil {
		return CounterpartyPage{}, fmt.Errorf("list counterparties: %w", err)
	}
	return page, nil
}

func (s *CounterpartyService) CreateCounterparty(ctx context.Context, input CounterpartyCreateInput) (counterpartydomain.Counterparty, error) {
	now, err := s.clock.Now()
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("read clock: %w", err)
	}
	created, err := s.store.CreateCounterparty(ctx, counterpartyCreateStoreInput{
		CounterpartyCreateInput: input,
		CreatedAt:               now,
	})
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("create counterparty: %w", err)
	}
	if !created.CreatedAt().Equal(now) {
		return counterpartydomain.Counterparty{}, domain.ErrInvariant
	}
	return created, nil
}

func (s *CounterpartyService) UpdateCounterparty(ctx context.Context, input CounterpartyUpdateInput) (counterpartydomain.Counterparty, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("read clock: %w", err)
	}
	updated, err := s.store.UpdateCounterparty(ctx, counterpartyUpdateStoreInput{
		CounterpartyUpdateInput: input,
		UpdatedAt:               now,
	})
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("update counterparty: %w", err)
	}
	if !updated.UpdatedAt().Equal(now) {
		return counterpartydomain.Counterparty{}, domain.ErrInvariant
	}
	return updated, nil
}

func (s *CounterpartyService) ArchiveCounterparty(ctx context.Context, input CounterpartyArchiveInput) (counterpartydomain.Counterparty, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("read clock: %w", err)
	}
	archived, err := s.store.ArchiveCounterparty(ctx, counterpartyArchiveStoreInput{
		CounterpartyArchiveInput: input,
		ArchivedAt:               now,
	})
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("archive counterparty: %w", err)
	}
	archivedAt, ok := archived.ArchivedAt().Get()
	if !ok || !archivedAt.Equal(now) {
		return counterpartydomain.Counterparty{}, domain.ErrInvariant
	}
	return archived, nil
}

func (s *CounterpartyService) RestoreCounterparty(ctx context.Context, input CounterpartyRestoreInput) (counterpartydomain.Counterparty, error) {
	now, err := nextMutationInstant(s.clock, input.ExpectedUpdatedAt)
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("read clock: %w", err)
	}
	restored, err := s.store.RestoreCounterparty(ctx, counterpartyRestoreStoreInput{
		CounterpartyRestoreInput: input,
		UpdatedAt:                now,
	})
	if err != nil {
		return counterpartydomain.Counterparty{}, fmt.Errorf("restore counterparty: %w", err)
	}
	if restored.IsArchived() || !restored.UpdatedAt().Equal(now) {
		return counterpartydomain.Counterparty{}, domain.ErrInvariant
	}
	return restored, nil
}
