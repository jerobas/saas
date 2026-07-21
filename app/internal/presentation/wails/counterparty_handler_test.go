package wails

import (
	"testing"

	"github.com/jerobas/saas/internal/domain"
	counterpartydomain "github.com/jerobas/saas/internal/domain/counterparty"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

func TestParseCounterpartyListRequestDefaultsToActiveFirstPage(t *testing.T) {
	input, err := parseCounterpartyListRequest(dto.CounterpartyListRequest{})
	if err != nil {
		t.Fatalf("parse list request: %v", err)
	}
	if input.Archive != domain.ArchiveActive {
		t.Fatalf("archive filter = %q", input.Archive)
	}
	if input.PageSize != 50 {
		t.Fatalf("page size = %d", input.PageSize)
	}
}

func TestMapCounterpartyIncludesRolesAndOptionalValues(t *testing.T) {
	instant := must(domain.UTCInstantFromUnixMilli(2_000))
	archivedAt := must(domain.UTCInstantFromUnixMilli(3_000))
	value := mustCounterparty(counterpartydomain.New(counterpartydomain.Params{
		ID:         must(domain.NewCounterpartyID(7)),
		Name:       must(domain.NewDisplayName("Supplier Prime")),
		Phone:      domain.Some(must(domain.NewNonEmptyText("+55 11 99999-0000"))),
		Roles:      mustCounterpartyRoles(counterpartydomain.NewRoleSet(domain.RoleSupplier, domain.RoleCustomer)),
		CreatedAt:  instant,
		UpdatedAt:  archivedAt,
		ArchivedAt: domain.Some(archivedAt),
	}))

	response := mapCounterparty(value)
	if response.ID != 7 || response.Name != "Supplier Prime" {
		t.Fatalf("identity = %d/%q", response.ID, response.Name)
	}
	if response.Phone == nil || *response.Phone != "+55 11 99999-0000" {
		t.Fatalf("phone = %#v", response.Phone)
	}
	if len(response.Roles) != 2 || response.Roles[0] != "SUPPLIER" || response.Roles[1] != "CUSTOMER" {
		t.Fatalf("roles = %#v", response.Roles)
	}
	if response.ArchivedAtMs == nil || *response.ArchivedAtMs != archivedAt.UnixMilli() {
		t.Fatalf("archived at = %#v", response.ArchivedAtMs)
	}
}

func mustCounterparty(value counterpartydomain.Counterparty, err error) counterpartydomain.Counterparty {
	if err != nil {
		panic(err)
	}
	return value
}

func mustCounterpartyRoles(value counterpartydomain.RoleSet, err error) counterpartydomain.RoleSet {
	if err != nil {
		panic(err)
	}
	return value
}
