package sqlite

import (
	"context"
	"database/sql"

	"github.com/jerobas/saas/internal/domain"
	counterpartydomain "github.com/jerobas/saas/internal/domain/counterparty"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

const (
	maximumCounterpartyPageSize = 100
)

type CounterpartyPageSize struct{ value int64 }

func NewCounterpartyPageSize(value int) (CounterpartyPageSize, error) {
	if value < 1 || value > maximumCounterpartyPageSize {
		return CounterpartyPageSize{}, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	return CounterpartyPageSize{value: int64(value)}, nil
}

func (s CounterpartyPageSize) Int() int { return int(s.value) }

type CounterpartyCursor struct {
	Name domain.DisplayName
	ID   domain.CounterpartyID
}

type CounterpartyListFilter struct {
	Archive  domain.ArchiveFilter
	Role     domain.Option[domain.CounterpartyRole]
	Search   domain.Option[domain.NonEmptyText]
	After    domain.Option[CounterpartyCursor]
	PageSize CounterpartyPageSize
}

type CounterpartyPage struct {
	items []counterpartydomain.Counterparty
	next  domain.Option[CounterpartyCursor]
}

func (p CounterpartyPage) Items() []counterpartydomain.Counterparty {
	items := make([]counterpartydomain.Counterparty, len(p.items))
	copy(items, p.items)
	return items
}

func (p CounterpartyPage) Next() domain.Option[CounterpartyCursor] { return p.next }

type CreateCounterpartyInput struct {
	Name      domain.DisplayName
	Phone     domain.Option[domain.NonEmptyText]
	Email     domain.Option[domain.NonEmptyText]
	Notes     domain.Option[domain.NonEmptyText]
	Roles     counterpartydomain.RoleSet
	CreatedAt domain.UTCInstant
}

type UpdateCounterpartyInput struct {
	ID                domain.CounterpartyID
	Name              domain.DisplayName
	Phone             domain.Option[domain.NonEmptyText]
	Email             domain.Option[domain.NonEmptyText]
	Notes             domain.Option[domain.NonEmptyText]
	Roles             counterpartydomain.RoleSet
	ExpectedUpdatedAt domain.UTCInstant
	UpdatedAt         domain.UTCInstant
}

func (s *Store) GetCounterparty(ctx context.Context, id domain.CounterpartyID) (counterpartydomain.Counterparty, error) {
	if id.IsZero() {
		return counterpartydomain.Counterparty{}, domain.Invalid("counterparty_id", domain.ViolationRequired, "")
	}
	value, err := loadCounterparty(ctx, s.queries, id)
	if err != nil {
		return counterpartydomain.Counterparty{}, classifyError("get counterparty", err)
	}
	return value, nil
}

func (s *Store) ListCounterparties(ctx context.Context, filter CounterpartyListFilter) (CounterpartyPage, error) {
	query, err := counterpartyListQuery(filter)
	if err != nil {
		return CounterpartyPage{}, err
	}
	rows, err := s.queries.ListCounterparties(ctx, query)
	if err != nil {
		return CounterpartyPage{}, classifyError("list counterparties", err)
	}
	hasMore := int64(len(rows)) > filter.PageSize.value
	if hasMore {
		rows = rows[:filter.PageSize.value]
	}
	result := make([]counterpartydomain.Counterparty, 0, len(rows))
	for _, row := range rows {
		value, _, mapErr := mapCounterpartyRecord(counterpartyRecordFromList(row))
		if mapErr != nil {
			return CounterpartyPage{}, corruptDataError("map counterparty list", mapErr)
		}
		result = append(result, value)
	}
	next := domain.None[CounterpartyCursor]()
	if hasMore && len(result) > 0 {
		last := result[len(result)-1]
		next = domain.Some(CounterpartyCursor{Name: last.Name(), ID: last.ID()})
	}
	return CounterpartyPage{items: result, next: next}, nil
}

func (s *Store) CreateCounterparty(ctx context.Context, params CreateCounterpartyInput) (counterpartydomain.Counterparty, error) {
	if err := validateCounterpartyCreate(params); err != nil {
		return counterpartydomain.Counterparty{}, err
	}
	var created counterpartydomain.Counterparty
	err := s.withWriteQueries(ctx, "create counterparty", func(queries *sqlcgen.Queries) error {
		idValue, err := queries.InsertCounterparty(ctx, sqlcgen.InsertCounterpartyParams{
			Name:        params.Name.String(),
			Phone:       counterpartyNullText(params.Phone),
			Email:       counterpartyNullText(params.Email),
			Notes:       counterpartyNullText(params.Notes),
			CreatedAtMs: params.CreatedAt.UnixMilli(),
			UpdatedAtMs: params.CreatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		id, err := domain.NewCounterpartyID(idValue)
		if err != nil {
			return corruptDataError("map created counterparty id", err)
		}
		if err := replaceCounterpartyRoles(ctx, queries, id, params.Roles, nil, params.CreatedAt); err != nil {
			return err
		}
		created, err = loadCounterparty(ctx, queries, id)
		return err
	})
	return created, err
}

func (s *Store) UpdateCounterparty(ctx context.Context, params UpdateCounterpartyInput) (counterpartydomain.Counterparty, error) {
	if err := validateCounterpartyUpdate(params); err != nil {
		return counterpartydomain.Counterparty{}, err
	}
	var updated counterpartydomain.Counterparty
	err := s.withWriteQueries(ctx, "update counterparty", func(queries *sqlcgen.Queries) error {
		current, roleSnapshots, err := loadCounterpartySnapshot(ctx, queries, params.ID)
		if err != nil {
			return err
		}
		if !current.UpdatedAt().Equal(params.ExpectedUpdatedAt) {
			return domain.ErrStale
		}
		if current.IsArchived() {
			return domain.ErrConflict
		}
		if _, err := counterpartydomain.New(counterpartydomain.Params{
			ID: params.ID, Name: params.Name, Phone: params.Phone, Email: params.Email,
			Notes: params.Notes, Roles: params.Roles, CreatedAt: current.CreatedAt(),
			UpdatedAt: params.UpdatedAt, ArchivedAt: domain.None[domain.UTCInstant](),
		}); err != nil {
			return err
		}
		rows, err := queries.UpdateCounterparty(ctx, sqlcgen.UpdateCounterpartyParams{
			Name: params.Name.String(), Phone: counterpartyNullText(params.Phone),
			Email: counterpartyNullText(params.Email), Notes: counterpartyNullText(params.Notes),
			UpdatedAtMs: params.UpdatedAt.UnixMilli(), ID: params.ID.Int64(),
			ExpectedUpdatedAtMs: params.ExpectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows != 1 {
			return classifyCounterpartyMiss(ctx, queries, params.ID, params.ExpectedUpdatedAt, false)
		}
		if err := replaceCounterpartyRoles(ctx, queries, params.ID, params.Roles, roleSnapshots, params.UpdatedAt); err != nil {
			return err
		}
		updated, err = loadCounterparty(ctx, queries, params.ID)
		return err
	})
	return updated, err
}

func (s *Store) ArchiveCounterparty(
	ctx context.Context,
	id domain.CounterpartyID,
	expectedUpdatedAt domain.UTCInstant,
	archivedAt domain.UTCInstant,
) (counterpartydomain.Counterparty, error) {
	if err := validateCounterpartyVersion(id, expectedUpdatedAt, archivedAt); err != nil {
		return counterpartydomain.Counterparty{}, err
	}
	var archived counterpartydomain.Counterparty
	err := s.withWriteQueries(ctx, "archive counterparty", func(queries *sqlcgen.Queries) error {
		current, _, err := loadCounterpartySnapshot(ctx, queries, id)
		if err != nil {
			return err
		}
		if !current.UpdatedAt().Equal(expectedUpdatedAt) {
			return domain.ErrStale
		}
		if current.IsArchived() {
			return domain.ErrConflict
		}
		rows, err := queries.ArchiveCounterparty(ctx, sqlcgen.ArchiveCounterpartyParams{
			ArchivedAtMs: archivedAt.UnixMilli(), UpdatedAtMs: archivedAt.UnixMilli(),
			ID: id.Int64(), ExpectedUpdatedAtMs: expectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows != 1 {
			return classifyCounterpartyMiss(ctx, queries, id, expectedUpdatedAt, false)
		}
		archived, err = loadCounterparty(ctx, queries, id)
		return err
	})
	return archived, err
}

func (s *Store) RestoreCounterparty(
	ctx context.Context,
	id domain.CounterpartyID,
	expectedUpdatedAt domain.UTCInstant,
	restoredAt domain.UTCInstant,
) (counterpartydomain.Counterparty, error) {
	if err := validateCounterpartyVersion(id, expectedUpdatedAt, restoredAt); err != nil {
		return counterpartydomain.Counterparty{}, err
	}
	var restored counterpartydomain.Counterparty
	err := s.withWriteQueries(ctx, "restore counterparty", func(queries *sqlcgen.Queries) error {
		current, err := loadCounterparty(ctx, queries, id)
		if err != nil {
			return err
		}
		if !current.UpdatedAt().Equal(expectedUpdatedAt) {
			return domain.ErrStale
		}
		if !current.IsArchived() {
			return domain.ErrConflict
		}
		if current.Roles().IsEmpty() {
			return domain.ErrCorruptData
		}
		rows, err := queries.RestoreCounterparty(ctx, sqlcgen.RestoreCounterpartyParams{
			UpdatedAtMs: restoredAt.UnixMilli(), ID: id.Int64(),
			ExpectedUpdatedAtMs: expectedUpdatedAt.UnixMilli(),
		})
		if err != nil {
			return err
		}
		if rows != 1 {
			return classifyCounterpartyMiss(ctx, queries, id, expectedUpdatedAt, true)
		}
		restored, err = loadCounterparty(ctx, queries, id)
		return err
	})
	return restored, err
}

func loadCounterparty(ctx context.Context, queries *sqlcgen.Queries, id domain.CounterpartyID) (counterpartydomain.Counterparty, error) {
	value, _, err := loadCounterpartySnapshot(ctx, queries, id)
	return value, err
}

type counterpartyRoleSnapshot struct {
	role      domain.CounterpartyRole
	createdAt domain.UTCInstant
}

type counterpartyRecord struct {
	id                      int64
	name                    string
	phone                   sql.NullString
	email                   sql.NullString
	notes                   sql.NullString
	createdAtMs             int64
	updatedAtMs             int64
	archivedAtMs            sql.NullInt64
	supplierRoleCreatedAtMs int64
	customerRoleCreatedAtMs int64
}

func loadCounterpartySnapshot(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id domain.CounterpartyID,
) (counterpartydomain.Counterparty, []counterpartyRoleSnapshot, error) {
	row, err := queries.GetCounterparty(ctx, id.Int64())
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	if row.ID != id.Int64() {
		return counterpartydomain.Counterparty{}, nil, corruptDataError("map counterparty identity", domain.ErrInvariant)
	}
	value, roles, err := mapCounterpartyRecord(counterpartyRecordFromGet(row))
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, corruptDataError("map counterparty", err)
	}
	return value, roles, nil
}

func counterpartyRecordFromGet(row sqlcgen.GetCounterpartyRow) counterpartyRecord {
	return counterpartyRecord{
		id: row.ID, name: row.Name, phone: row.Phone, email: row.Email, notes: row.Notes,
		createdAtMs: row.CreatedAtMs, updatedAtMs: row.UpdatedAtMs, archivedAtMs: row.ArchivedAtMs,
		supplierRoleCreatedAtMs: row.SupplierRoleCreatedAtMs,
		customerRoleCreatedAtMs: row.CustomerRoleCreatedAtMs,
	}
}

func counterpartyRecordFromList(row sqlcgen.ListCounterpartiesRow) counterpartyRecord {
	return counterpartyRecord{
		id: row.ID, name: row.Name, phone: row.Phone, email: row.Email, notes: row.Notes,
		createdAtMs: row.CreatedAtMs, updatedAtMs: row.UpdatedAtMs, archivedAtMs: row.ArchivedAtMs,
		supplierRoleCreatedAtMs: row.SupplierRoleCreatedAtMs,
		customerRoleCreatedAtMs: row.CustomerRoleCreatedAtMs,
	}
}

func mapCounterpartyRecord(row counterpartyRecord) (counterpartydomain.Counterparty, []counterpartyRoleSnapshot, error) {
	id, err := domain.NewCounterpartyID(row.id)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	name, err := domain.NewDisplayName(row.name)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	if name.String() != row.name {
		return counterpartydomain.Counterparty{}, nil, domain.ErrInvariant
	}
	phone, err := counterpartyOptionalText(row.phone)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	email, err := counterpartyOptionalText(row.email)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	notes, err := counterpartyOptionalText(row.notes)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.createdAtMs)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(row.updatedAtMs)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	archivedAt, err := counterpartyOptionalInstant(row.archivedAtMs)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}

	roleValues := make([]domain.CounterpartyRole, 0, 2)
	roleSnapshots := make([]counterpartyRoleSnapshot, 0, 2)
	for _, persisted := range []struct {
		role        domain.CounterpartyRole
		createdAtMs int64
	}{
		{role: domain.RoleSupplier, createdAtMs: row.supplierRoleCreatedAtMs},
		{role: domain.RoleCustomer, createdAtMs: row.customerRoleCreatedAtMs},
	} {
		if persisted.createdAtMs == -1 {
			continue
		}
		roleCreatedAt, roleErr := domain.UTCInstantFromUnixMilli(persisted.createdAtMs)
		if roleErr != nil || roleCreatedAt.Before(createdAt) || updatedAt.Before(roleCreatedAt) {
			if roleErr != nil {
				return counterpartydomain.Counterparty{}, nil, roleErr
			}
			return counterpartydomain.Counterparty{}, nil, domain.ErrInvariant
		}
		roleValues = append(roleValues, persisted.role)
		roleSnapshots = append(roleSnapshots, counterpartyRoleSnapshot{role: persisted.role, createdAt: roleCreatedAt})
	}
	roleSet, err := counterpartydomain.NewRoleSet(roleValues...)
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	value, err := counterpartydomain.New(counterpartydomain.Params{
		ID: id, Name: name, Phone: phone, Email: email, Notes: notes, Roles: roleSet,
		CreatedAt: createdAt, UpdatedAt: updatedAt, ArchivedAt: archivedAt,
	})
	if err != nil {
		return counterpartydomain.Counterparty{}, nil, err
	}
	return value, roleSnapshots, nil
}

func replaceCounterpartyRoles(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id domain.CounterpartyID,
	roles counterpartydomain.RoleSet,
	previous []counterpartyRoleSnapshot,
	changedAt domain.UTCInstant,
) error {
	if roles.IsEmpty() {
		return domain.Invalid("roles", domain.ViolationRequired, "CPY-001")
	}
	createdByRole := make(map[domain.CounterpartyRole]int64, len(previous))
	for _, snapshot := range previous {
		createdByRole[snapshot.role] = snapshot.createdAt.UnixMilli()
	}
	if _, err := queries.DeleteCounterpartyRoles(ctx, id.Int64()); err != nil {
		return err
	}
	for _, role := range roles.Roles() {
		createdAt := changedAt.UnixMilli()
		if previousCreatedAt, ok := createdByRole[role]; ok {
			createdAt = previousCreatedAt
		}
		if err := queries.InsertCounterpartyRole(ctx, sqlcgen.InsertCounterpartyRoleParams{
			CounterpartyID: id.Int64(), Role: role.String(), CreatedAtMs: createdAt,
		}); err != nil {
			return err
		}
	}
	return nil
}

func classifyCounterpartyMiss(
	ctx context.Context,
	queries *sqlcgen.Queries,
	id domain.CounterpartyID,
	expected domain.UTCInstant,
	wantArchived bool,
) error {
	row, err := queries.GetCounterparty(ctx, id.Int64())
	if err != nil {
		return err
	}
	if row.UpdatedAtMs != expected.UnixMilli() {
		return domain.ErrStale
	}
	if row.ArchivedAtMs.Valid != wantArchived {
		return domain.ErrConflict
	}
	return domain.ErrConflict
}

func counterpartyListQuery(filter CounterpartyListFilter) (sqlcgen.ListCounterpartiesParams, error) {
	archive, err := counterpartyArchiveFilter(filter.Archive)
	if err != nil {
		return sqlcgen.ListCounterpartiesParams{}, err
	}
	if filter.PageSize.value < 1 || filter.PageSize.value > maximumCounterpartyPageSize {
		return sqlcgen.ListCounterpartiesParams{}, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	query := sqlcgen.ListCounterpartiesParams{
		ArchiveFilter: archive,
		LimitCount:    filter.PageSize.value + 1,
	}
	if role, ok := filter.Role.Get(); ok {
		if _, parseErr := domain.ParseCounterpartyRole(role.String()); parseErr != nil {
			return sqlcgen.ListCounterpartiesParams{}, parseErr
		}
		query.RoleFilter = role.String()
	}
	if search, ok := filter.Search.Get(); ok {
		if search.String() == "" {
			return sqlcgen.ListCounterpartiesParams{}, domain.Invalid("search", domain.ViolationRequired, "")
		}
		query.SearchText = search.String()
	}
	if cursor, ok := filter.After.Get(); ok {
		if cursor.Name.String() == "" || cursor.ID.IsZero() {
			return sqlcgen.ListCounterpartiesParams{}, domain.Invalid("cursor", domain.ViolationInvalidFormat, "")
		}
		query.AfterName = cursor.Name.String()
		query.AfterID = cursor.ID.Int64()
	}
	return query, nil
}

func counterpartyArchiveFilter(filter domain.ArchiveFilter) (int64, error) {
	switch filter {
	case "", domain.ArchiveActive:
		return 0, nil
	case domain.ArchiveArchived:
		return 1, nil
	case domain.ArchiveAll:
		return 2, nil
	default:
		return 0, domain.Invalid("archive_filter", domain.ViolationInvalidEnum, "ARC-001")
	}
}

func validateCounterpartyCreate(params CreateCounterpartyInput) error {
	violations := counterpartyContentViolations(params.Name, params.Phone, params.Email, params.Notes, params.Roles)
	if params.CreatedAt.IsZero() {
		violations = append(violations, domain.Violation{Field: "created_at", Code: domain.ViolationRequired})
	}
	return domain.NewValidationError(violations...)
}

func counterpartyContentViolations(
	name domain.DisplayName,
	phone domain.Option[domain.NonEmptyText],
	email domain.Option[domain.NonEmptyText],
	notes domain.Option[domain.NonEmptyText],
	roles counterpartydomain.RoleSet,
) []domain.Violation {
	violations := make([]domain.Violation, 0, 5)
	if name.String() == "" {
		violations = append(violations, domain.Violation{Field: "name", Code: domain.ViolationRequired})
	}
	if roles.IsEmpty() {
		violations = append(violations, domain.Violation{Field: "roles", Code: domain.ViolationRequired, InvariantID: "CPY-001"})
	}
	for _, candidate := range []struct {
		field string
		value domain.Option[domain.NonEmptyText]
	}{
		{field: "phone", value: phone},
		{field: "email", value: email},
		{field: "notes", value: notes},
	} {
		if text, ok := candidate.value.Get(); ok && text.String() == "" {
			violations = append(violations, domain.Violation{Field: candidate.field, Code: domain.ViolationRequired})
		}
	}
	return violations
}

func validateCounterpartyUpdate(params UpdateCounterpartyInput) error {
	violations := counterpartyContentViolations(params.Name, params.Phone, params.Email, params.Notes, params.Roles)
	if params.ID.IsZero() {
		violations = append(violations, domain.Violation{Field: "counterparty_id", Code: domain.ViolationRequired})
	}
	if params.ExpectedUpdatedAt.IsZero() {
		violations = append(violations, domain.Violation{Field: "expected_updated_at", Code: domain.ViolationRequired})
	}
	if params.UpdatedAt.IsZero() || !params.ExpectedUpdatedAt.Before(params.UpdatedAt) {
		violations = append(violations, domain.Violation{Field: "updated_at", Code: domain.ViolationOutOfRange})
	}
	return domain.NewValidationError(violations...)
}

func validateCounterpartyVersion(id domain.CounterpartyID, expected, next domain.UTCInstant) error {
	violations := make([]domain.Violation, 0, 3)
	if id.IsZero() {
		violations = append(violations, domain.Violation{Field: "counterparty_id", Code: domain.ViolationRequired})
	}
	if expected.IsZero() {
		violations = append(violations, domain.Violation{Field: "expected_updated_at", Code: domain.ViolationRequired})
	}
	if next.IsZero() || !expected.Before(next) {
		violations = append(violations, domain.Violation{Field: "updated_at", Code: domain.ViolationOutOfRange})
	}
	return domain.NewValidationError(violations...)
}

func counterpartyNullText(value domain.Option[domain.NonEmptyText]) sql.NullString {
	text, ok := value.Get()
	return sql.NullString{String: text.String(), Valid: ok}
}

func counterpartyOptionalText(value sql.NullString) (domain.Option[domain.NonEmptyText], error) {
	if !value.Valid {
		return domain.None[domain.NonEmptyText](), nil
	}
	text, err := domain.NewNonEmptyText(value.String)
	if err != nil {
		return domain.None[domain.NonEmptyText](), err
	}
	if text.String() != value.String {
		return domain.None[domain.NonEmptyText](), domain.ErrInvariant
	}
	return domain.Some(text), nil
}

func counterpartyOptionalInstant(value sql.NullInt64) (domain.Option[domain.UTCInstant], error) {
	if !value.Valid {
		return domain.None[domain.UTCInstant](), nil
	}
	instant, err := domain.UTCInstantFromUnixMilli(value.Int64)
	if err != nil {
		return domain.None[domain.UTCInstant](), err
	}
	return domain.Some(instant), nil
}
