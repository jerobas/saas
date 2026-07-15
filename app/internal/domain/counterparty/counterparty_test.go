package counterparty_test

import (
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/counterparty"
)

func TestCounterpartyOwnsCompleteRoleSet(t *testing.T) {
	roles, err := counterparty.NewRoleSet(domain.RoleCustomer, domain.RoleSupplier, domain.RoleCustomer)
	if err != nil || len(roles.Roles()) != 2 || roles.Roles()[0] != domain.RoleSupplier {
		t.Fatalf("role set = %#v, %v", roles.Roles(), err)
	}
	instant := must(domain.UTCInstantFromUnixMilli(1000))
	entity, err := counterparty.New(counterparty.Params{
		ID: must(domain.NewCounterpartyID(1)), Name: must(domain.NewDisplayName("Market")),
		Email: domain.Some(must(domain.NewNonEmptyText("buyer@example.test"))),
		Roles: roles, CreatedAt: instant, UpdatedAt: instant,
	})
	if err != nil || !entity.Roles().Contains(domain.RoleCustomer) || entity.IsArchived() {
		t.Fatalf("counterparty = %#v, %v", entity, err)
	}
}

func TestActiveCounterpartyRequiresRoleButArchivedSnapshotDoesNot(t *testing.T) {
	instant := must(domain.UTCInstantFromUnixMilli(1000))
	empty, _ := counterparty.NewRoleSet()
	params := counterparty.Params{
		ID: must(domain.NewCounterpartyID(1)), Name: must(domain.NewDisplayName("Market")),
		Roles: empty, CreatedAt: instant, UpdatedAt: instant,
	}
	if _, err := counterparty.New(params); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("active empty roles error = %v", err)
	}
	params.ArchivedAt = domain.Some(instant)
	entity, err := counterparty.New(params)
	if err != nil || !entity.IsArchived() {
		t.Fatalf("archived counterparty = %#v, %v", entity, err)
	}
	if _, err := counterparty.NewRoleSet(domain.CounterpartyRole("UNKNOWN")); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("unknown role error = %v", err)
	}
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
