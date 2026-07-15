package counterparty

import "github.com/jerobas/saas/internal/domain"

const (
	supplierRole uint8 = 1 << iota
	customerRole
)

type RoleSet struct{ bits uint8 }

func NewRoleSet(roles ...domain.CounterpartyRole) (RoleSet, error) {
	var bits uint8
	for _, role := range roles {
		switch role {
		case domain.RoleSupplier:
			bits |= supplierRole
		case domain.RoleCustomer:
			bits |= customerRole
		default:
			return RoleSet{}, domain.Invalid("counterparty_role", domain.ViolationInvalidEnum, "CPY-001")
		}
	}
	return RoleSet{bits: bits}, nil
}

func (r RoleSet) IsEmpty() bool { return r.bits == 0 }
func (r RoleSet) Contains(role domain.CounterpartyRole) bool {
	switch role {
	case domain.RoleSupplier:
		return r.bits&supplierRole != 0
	case domain.RoleCustomer:
		return r.bits&customerRole != 0
	default:
		return false
	}
}
func (r RoleSet) Roles() []domain.CounterpartyRole {
	roles := make([]domain.CounterpartyRole, 0, 2)
	if r.Contains(domain.RoleSupplier) {
		roles = append(roles, domain.RoleSupplier)
	}
	if r.Contains(domain.RoleCustomer) {
		roles = append(roles, domain.RoleCustomer)
	}
	return roles
}

type Params struct {
	ID         domain.CounterpartyID
	Name       domain.DisplayName
	Phone      domain.Option[domain.NonEmptyText]
	Email      domain.Option[domain.NonEmptyText]
	Notes      domain.Option[domain.NonEmptyText]
	Roles      RoleSet
	CreatedAt  domain.UTCInstant
	UpdatedAt  domain.UTCInstant
	ArchivedAt domain.Option[domain.UTCInstant]
}

// Counterparty is the identity plus its complete eligibility role set. Roles
// are never exposed as independently mutable rows.
type Counterparty struct {
	id         domain.CounterpartyID
	name       domain.DisplayName
	phone      domain.Option[domain.NonEmptyText]
	email      domain.Option[domain.NonEmptyText]
	notes      domain.Option[domain.NonEmptyText]
	roles      RoleSet
	createdAt  domain.UTCInstant
	updatedAt  domain.UTCInstant
	archivedAt domain.Option[domain.UTCInstant]
}

func New(params Params) (Counterparty, error) {
	violations := make([]domain.Violation, 0, 6)
	if params.ID.IsZero() {
		violations = append(violations, domain.Violation{Field: "counterparty_id", Code: domain.ViolationRequired})
	}
	if params.Name.String() == "" {
		violations = append(violations, domain.Violation{Field: "name", Code: domain.ViolationRequired})
	}
	optionalText := []struct {
		field string
		value domain.Option[domain.NonEmptyText]
	}{{"phone", params.Phone}, {"email", params.Email}, {"notes", params.Notes}}
	for _, candidate := range optionalText {
		if value, ok := candidate.value.Get(); ok && value.String() == "" {
			violations = append(violations, domain.Violation{Field: candidate.field, Code: domain.ViolationRequired})
		}
	}
	if params.ArchivedAt.IsNone() && params.Roles.IsEmpty() {
		violations = append(violations, domain.Violation{Field: "roles", Code: domain.ViolationRequired, InvariantID: "CPY-001"})
	}
	if err := domain.ValidateTimestampOrder(params.CreatedAt, params.UpdatedAt, params.ArchivedAt); err != nil {
		if validation, ok := err.(*domain.ValidationError); ok {
			violations = append(violations, validation.Violations()...)
		}
	}
	if err := domain.NewValidationError(violations...); err != nil {
		return Counterparty{}, err
	}
	return Counterparty{
		id: params.ID, name: params.Name, phone: params.Phone, email: params.Email,
		notes: params.Notes, roles: params.Roles, createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt, archivedAt: params.ArchivedAt,
	}, nil
}

func (c Counterparty) ID() domain.CounterpartyID                    { return c.id }
func (c Counterparty) Name() domain.DisplayName                     { return c.name }
func (c Counterparty) Phone() domain.Option[domain.NonEmptyText]    { return c.phone }
func (c Counterparty) Email() domain.Option[domain.NonEmptyText]    { return c.email }
func (c Counterparty) Notes() domain.Option[domain.NonEmptyText]    { return c.notes }
func (c Counterparty) Roles() RoleSet                               { return c.roles }
func (c Counterparty) CreatedAt() domain.UTCInstant                 { return c.createdAt }
func (c Counterparty) UpdatedAt() domain.UTCInstant                 { return c.updatedAt }
func (c Counterparty) ArchivedAt() domain.Option[domain.UTCInstant] { return c.archivedAt }
func (c Counterparty) IsArchived() bool                             { return c.archivedAt.IsSome() }
