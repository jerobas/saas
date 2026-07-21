package application

import (
	"context"
	"testing"
	"time"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	counterpartydomain "github.com/jerobas/saas/internal/domain/counterparty"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

type mutableClock struct {
	now domain.UTCInstant
}

func (c *mutableClock) Now() (domain.UTCInstant, error) {
	return c.now, nil
}

func TestReferenceDataServiceListsMeasurementUnits(t *testing.T) {
	db := newApplicationTestDatabase(t)
	store := sqlite.NewStore(db)
	service := NewReferenceDataService(NewSQLiteReferenceDataStore(store))

	units, err := service.ListMeasurementUnits(context.Background())
	if err != nil {
		t.Fatalf("list measurement units: %v", err)
	}
	if len(units) == 0 {
		t.Fatal("expected seeded measurement units")
	}
}

func TestCounterpartyServiceRunsMutationLifecycleWithApplicationClock(t *testing.T) {
	db := newApplicationTestDatabase(t)
	store := sqlite.NewStore(db)
	clock := &mutableClock{now: mustInstant(2_000)}
	service := NewCounterpartyService(NewSQLiteCounterpartyStore(store), clock)
	ctx := context.Background()

	roles := mustRoleSet(counterpartydomain.NewRoleSet(domain.RoleSupplier))
	created, err := service.CreateCounterparty(ctx, CounterpartyCreateInput{
		Name:  must(domain.NewDisplayName("Supplier One")),
		Phone: domain.Some(must(domain.NewNonEmptyText("+55 11 99999-0000"))),
		Roles: roles,
	})
	if err != nil {
		t.Fatalf("create counterparty: %v", err)
	}
	if !created.CreatedAt().Equal(clock.now) || !created.UpdatedAt().Equal(clock.now) {
		t.Fatalf("created timestamps = %d/%d", created.CreatedAt().UnixMilli(), created.UpdatedAt().UnixMilli())
	}

	clock.now = created.UpdatedAt()
	updated, err := service.UpdateCounterparty(ctx, CounterpartyUpdateInput{
		ID: created.ID(), Name: must(domain.NewDisplayName("Supplier Prime")),
		Roles:             mustRoleSet(counterpartydomain.NewRoleSet(domain.RoleSupplier, domain.RoleCustomer)),
		ExpectedUpdatedAt: created.UpdatedAt(),
	})
	if err != nil {
		t.Fatalf("update counterparty: %v", err)
	}
	expectedAdvanced := mustInstant(created.UpdatedAt().UnixMilli() + 1)
	if !updated.UpdatedAt().Equal(expectedAdvanced) || !updated.Roles().Contains(domain.RoleCustomer) {
		t.Fatalf("updated counterparty = %#v", updated)
	}

	page, err := service.ListCounterparties(ctx, CounterpartyListInput{
		Archive:  domain.ArchiveActive,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list counterparties: %v", err)
	}
	if len(page.Items()) != 1 {
		t.Fatalf("counterparty count = %d, want 1", len(page.Items()))
	}

	clock.now = mustInstant(4_000)
	archived, err := service.ArchiveCounterparty(ctx, CounterpartyArchiveInput{
		ID: updated.ID(), ExpectedUpdatedAt: updated.UpdatedAt(),
	})
	if err != nil {
		t.Fatalf("archive counterparty: %v", err)
	}
	archivedAt, ok := archived.ArchivedAt().Get()
	if !ok || !archivedAt.Equal(clock.now) {
		t.Fatalf("archived at = %#v", archived.ArchivedAt())
	}

	clock.now = mustInstant(5_000)
	restored, err := service.RestoreCounterparty(ctx, CounterpartyRestoreInput{
		ID: archived.ID(), ExpectedUpdatedAt: archived.UpdatedAt(),
	})
	if err != nil {
		t.Fatalf("restore counterparty: %v", err)
	}
	if restored.IsArchived() || !restored.UpdatedAt().Equal(clock.now) {
		t.Fatalf("restored counterparty = %#v", restored)
	}
}

func newApplicationTestDatabase(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.NewDatabaseWithOptions(":memory:", database.OpenOptions{BusyTimeout: time.Second})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})
	return db
}

func mustInstant(ms int64) domain.UTCInstant {
	return must(domain.UTCInstantFromUnixMilli(ms))
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func mustRoleSet(value counterpartydomain.RoleSet, err error) counterpartydomain.RoleSet {
	if err != nil {
		panic(err)
	}
	return value
}
