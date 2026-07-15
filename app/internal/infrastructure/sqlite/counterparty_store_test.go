package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	counterpartydomain "github.com/jerobas/saas/internal/domain/counterparty"
)

func TestCounterpartyStoreLifecycleAndOptimisticVersion(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "counterparties.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	createdAt := counterpartyInstant(t, 1_000)
	supplier := counterpartyRoles(t, domain.RoleSupplier)

	created, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name: counterpartyName(t, " Acme Foods "), Phone: counterpartyOptionalTextValue(t, "+55 11 0000-0000"),
		Roles: supplier, CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID().Int64() <= 0 || created.Name().String() != "Acme Foods" {
		t.Fatalf("created counterparty = id %s, name %q", created.ID(), created.Name().String())
	}
	if !created.Roles().Contains(domain.RoleSupplier) || created.Roles().Contains(domain.RoleCustomer) {
		t.Fatalf("created roles = %v", created.Roles().Roles())
	}

	updatedAt := counterpartyInstant(t, 2_000)
	bothRoles := counterpartyRoles(t, domain.RoleSupplier, domain.RoleCustomer)
	updated, err := store.UpdateCounterparty(ctx, UpdateCounterpartyInput{
		ID: created.ID(), Name: counterpartyName(t, "Acme Brasil"),
		Email: domain.Some(counterpartyText(t, "orders@example.test")), Roles: bothRoles,
		ExpectedUpdatedAt: created.UpdatedAt(), UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name().String() != "Acme Brasil" || !updated.Roles().Contains(domain.RoleCustomer) {
		t.Fatalf("updated counterparty = %q, roles %v", updated.Name().String(), updated.Roles().Roles())
	}

	_, err = store.UpdateCounterparty(ctx, UpdateCounterpartyInput{
		ID: created.ID(), Name: counterpartyName(t, "Stale overwrite"), Roles: supplier,
		ExpectedUpdatedAt: created.UpdatedAt(), UpdatedAt: counterpartyInstant(t, 3_000),
	})
	if !errors.Is(err, domain.ErrStale) {
		t.Fatalf("stale update error = %v, want ErrStale", err)
	}
	unchanged, err := store.GetCounterparty(ctx, created.ID())
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Name().String() != "Acme Brasil" || len(unchanged.Roles().Roles()) != 2 {
		t.Fatalf("stale update changed aggregate: %q, %v", unchanged.Name().String(), unchanged.Roles().Roles())
	}

	archived, err := store.ArchiveCounterparty(ctx, updated.ID(), updated.UpdatedAt(), counterpartyInstant(t, 3_000))
	if err != nil {
		t.Fatal(err)
	}
	if !archived.IsArchived() {
		t.Fatal("archived counterparty is active")
	}
	pageSize := counterpartyPageSize(t, 100)
	active, err := store.ListCounterparties(ctx, CounterpartyListFilter{PageSize: pageSize})
	if err != nil {
		t.Fatal(err)
	}
	if active.Items() == nil || len(active.Items()) != 0 {
		t.Fatalf("active counterparties = %#v, want nonnil empty slice", active.Items())
	}

	_, err = store.UpdateCounterparty(ctx, UpdateCounterpartyInput{
		ID: archived.ID(), Name: archived.Name(), Roles: archived.Roles(),
		ExpectedUpdatedAt: archived.UpdatedAt(), UpdatedAt: counterpartyInstant(t, 4_000),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("update archived error = %v, want ErrConflict", err)
	}

	restored, err := store.RestoreCounterparty(ctx, archived.ID(), archived.UpdatedAt(), counterpartyInstant(t, 4_000))
	if err != nil {
		t.Fatal(err)
	}
	if restored.IsArchived() || len(restored.Roles().Roles()) != 2 {
		t.Fatalf("restored counterparty archived=%v roles=%v", restored.IsArchived(), restored.Roles().Roles())
	}
}

func TestCounterpartyStoreListFiltersAndKeyset(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "counterparty-list.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	supplier := counterpartyRoles(t, domain.RoleSupplier)
	customer := counterpartyRoles(t, domain.RoleCustomer)
	names := []struct {
		name  string
		roles counterpartydomain.RoleSet
	}{
		{"Alpha", supplier},
		{"Alpha", customer},
		{"Beta", supplier},
	}
	created := make([]counterpartydomain.Counterparty, 0, len(names))
	for index, fixture := range names {
		value, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
			Name: counterpartyName(t, fixture.name), Roles: fixture.roles,
			CreatedAt: counterpartyInstant(t, int64(index+1)*1_000),
		})
		if err != nil {
			t.Fatal(err)
		}
		created = append(created, value)
	}
	pageSize := counterpartyPageSize(t, 2)

	first, err := store.ListCounterparties(ctx, CounterpartyListFilter{Archive: domain.ArchiveAll, PageSize: pageSize})
	if err != nil {
		t.Fatal(err)
	}
	firstItems := first.Items()
	if len(firstItems) != 2 || firstItems[0].Name().String() != "Alpha" || firstItems[1].Name().String() != "Alpha" {
		t.Fatalf("first page = %#v", firstItems)
	}
	firstNext, ok := first.Next().Get()
	if !ok {
		t.Fatal("first page did not return a next cursor")
	}
	second, err := store.ListCounterparties(ctx, CounterpartyListFilter{
		Archive:  domain.ArchiveAll,
		After:    domain.Some(firstNext),
		PageSize: pageSize,
	})
	if err != nil {
		t.Fatal(err)
	}
	secondItems := second.Items()
	if len(secondItems) != 1 || secondItems[0].Name().String() != "Beta" || second.Next().IsSome() {
		t.Fatalf("second page = %#v, next=%v", secondItems, second.Next())
	}

	suppliers, err := store.ListCounterparties(ctx, CounterpartyListFilter{
		Role: domain.Some(domain.RoleSupplier), Search: domain.Some(counterpartyText(t, "Alpha")),
		PageSize: pageSize,
	})
	if err != nil {
		t.Fatal(err)
	}
	supplierItems := suppliers.Items()
	if len(supplierItems) != 1 || supplierItems[0].ID() != created[0].ID() {
		t.Fatalf("supplier filter = %#v", supplierItems)
	}
	caseSensitive, err := store.ListCounterparties(ctx, CounterpartyListFilter{
		Search:   domain.Some(counterpartyText(t, "alpha")),
		PageSize: pageSize,
	})
	if err != nil {
		t.Fatal(err)
	}
	if caseSensitive.Items() == nil || len(caseSensitive.Items()) != 0 {
		t.Fatalf("lowercase display search = %#v, want nonnil empty", caseSensitive.Items())
	}
}

func TestCounterpartyStorePreservesRoleHistoryAndRemovesRoles(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "counterparty-role-history.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	created, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name: counterpartyName(t, "Role history"), Roles: counterpartyRoles(t, domain.RoleSupplier),
		CreatedAt: counterpartyInstant(t, 1_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	withCustomer, err := store.UpdateCounterparty(ctx, UpdateCounterpartyInput{
		ID: created.ID(), Name: created.Name(), Roles: counterpartyRoles(t, domain.RoleSupplier, domain.RoleCustomer),
		ExpectedUpdatedAt: created.UpdatedAt(), UpdatedAt: counterpartyInstant(t, 2_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertCounterpartyRoleTimes(t, store, created.ID(), map[string]int64{
		domain.RoleSupplier.String(): 1_000,
		domain.RoleCustomer.String(): 2_000,
	})

	_, err = store.UpdateCounterparty(ctx, UpdateCounterpartyInput{
		ID: withCustomer.ID(), Name: withCustomer.Name(), Roles: counterpartyRoles(t, domain.RoleCustomer),
		ExpectedUpdatedAt: withCustomer.UpdatedAt(), UpdatedAt: counterpartyInstant(t, 3_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertCounterpartyRoleTimes(t, store, created.ID(), map[string]int64{
		domain.RoleCustomer.String(): 2_000,
	})
}

func TestCounterpartyStoreRejectsCorruptRoleStateBeforeArchive(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "counterparty-corrupt-roles.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	created, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name: counterpartyName(t, "No roles"), Roles: counterpartyRoles(t, domain.RoleSupplier),
		CreatedAt: counterpartyInstant(t, 1_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.database.ExecContext(ctx, `DELETE FROM counterparty_roles WHERE counterparty_id = ?`, created.ID().Int64()); err != nil {
		t.Fatal(err)
	}

	_, err = store.ArchiveCounterparty(ctx, created.ID(), created.UpdatedAt(), counterpartyInstant(t, 2_000))
	if !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("archive corrupt counterparty error = %v, want ErrCorruptData", err)
	}
	var archivedCount int
	if err := store.database.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM counterparties WHERE id = ? AND archived_at_ms IS NOT NULL
	`, created.ID().Int64()).Scan(&archivedCount); err != nil {
		t.Fatal(err)
	}
	if archivedCount != 0 {
		t.Fatal("corrupt counterparty was archived")
	}
}

func TestCounterpartyStoreErrorsAndCorruptSnapshot(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "counterparty-errors.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	missingID, err := domain.NewCounterpartyID(999)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.GetCounterparty(ctx, missingID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing counterparty error = %v, want ErrNotFound", err)
	}

	created, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name: counterpartyName(t, "Corrupt me"), Roles: counterpartyRoles(t, domain.RoleSupplier),
		CreatedAt: counterpartyInstant(t, 1_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.database.ExecContext(ctx, `
		UPDATE counterparties SET name = CAST(X'FF' AS TEXT) WHERE id = ?
	`, created.ID().Int64()); err != nil {
		t.Fatal(err)
	}
	_, err = store.GetCounterparty(ctx, created.ID())
	if !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("corrupt counterparty error = %v, want ErrCorruptData", err)
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = store.ListCounterparties(canceled, CounterpartyListFilter{PageSize: counterpartyPageSize(t, 10)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v, want context.Canceled", err)
	}
}

func TestCounterpartyStoreRejectsCorruptRoleChronology(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "counterparty-role-time.db"), database.DefaultOpenOptions())
	ctx := context.Background()
	created, err := store.CreateCounterparty(ctx, CreateCounterpartyInput{
		Name: counterpartyName(t, "Future role"), Roles: counterpartyRoles(t, domain.RoleSupplier),
		CreatedAt: counterpartyInstant(t, 1_000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.database.ExecContext(ctx, `
		UPDATE counterparty_roles SET created_at_ms = 2000 WHERE counterparty_id = ?
	`, created.ID().Int64()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetCounterparty(ctx, created.ID()); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("future role timestamp error = %v, want ErrCorruptData", err)
	}
}

func TestCounterpartyUpdateValidationReportsUpdateFieldsTogether(t *testing.T) {
	store := newAdapterTestStore(t, filepath.Join(t.TempDir(), "counterparty-validation.db"), database.DefaultOpenOptions())
	_, err := store.UpdateCounterparty(context.Background(), UpdateCounterpartyInput{})
	var validation *domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("update validation error = %v, want ValidationError", err)
	}
	fields := make(map[string]bool)
	for _, violation := range validation.Violations() {
		fields[violation.Field] = true
	}
	for _, want := range []string{"counterparty_id", "expected_updated_at", "name", "roles", "updated_at"} {
		if !fields[want] {
			t.Errorf("missing validation field %q in %#v", want, validation.Violations())
		}
	}
	if fields["created_at"] {
		t.Errorf("update validation incorrectly reported created_at: %#v", validation.Violations())
	}
}

func assertCounterpartyRoleTimes(
	t *testing.T,
	store *Store,
	id domain.CounterpartyID,
	want map[string]int64,
) {
	t.Helper()
	rows, err := store.queries.ListCounterpartyRoles(context.Background(), id.Int64())
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]int64, len(rows))
	for _, row := range rows {
		got[row.Role] = row.CreatedAtMs
	}
	if len(got) != len(want) {
		t.Fatalf("role history = %#v, want %#v", got, want)
	}
	for role, createdAt := range want {
		if got[role] != createdAt {
			t.Fatalf("role %s created_at = %d, want %d", role, got[role], createdAt)
		}
	}
}

func counterpartyPageSize(t *testing.T, value int) CounterpartyPageSize {
	t.Helper()
	size, err := NewCounterpartyPageSize(value)
	if err != nil {
		t.Fatal(err)
	}
	return size
}

func counterpartyName(t *testing.T, value string) domain.DisplayName {
	t.Helper()
	name, err := domain.NewDisplayName(value)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func counterpartyText(t *testing.T, value string) domain.NonEmptyText {
	t.Helper()
	text, err := domain.NewNonEmptyText(value)
	if err != nil {
		t.Fatal(err)
	}
	return text
}

func counterpartyOptionalTextValue(t *testing.T, value string) domain.Option[domain.NonEmptyText] {
	t.Helper()
	return domain.Some(counterpartyText(t, value))
}

func counterpartyRoles(t *testing.T, roles ...domain.CounterpartyRole) counterpartydomain.RoleSet {
	t.Helper()
	set, err := counterpartydomain.NewRoleSet(roles...)
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func counterpartyInstant(t *testing.T, milliseconds int64) domain.UTCInstant {
	t.Helper()
	instant, err := domain.UTCInstantFromUnixMilli(milliseconds)
	if err != nil {
		t.Fatal(err)
	}
	return instant
}
