package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	counterpartydomain "github.com/jerobas/saas/internal/domain/counterparty"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type CounterpartyHandler struct {
	service *application.CounterpartyService
}

func NewCounterpartyHandler(service *application.CounterpartyService) *CounterpartyHandler {
	if service == nil {
		panic("counterparty handler requires a service")
	}
	return &CounterpartyHandler{service: service}
}

func (h *CounterpartyHandler) GetCounterparty(id int64) (dto.CounterpartyResponse, error) {
	counterpartyID, err := domain.NewCounterpartyID(id)
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("counterparty id: %w", err)
	}
	value, err := h.service.GetCounterparty(handlerContext(), counterpartyID)
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("get counterparty: %w", err)
	}
	return mapCounterparty(value), nil
}

func (h *CounterpartyHandler) ListCounterparties(req dto.CounterpartyListRequest) (dto.CounterpartyPageResponse, error) {
	input, err := parseCounterpartyListRequest(req)
	if err != nil {
		return dto.CounterpartyPageResponse{}, err
	}
	page, err := h.service.ListCounterparties(handlerContext(), input)
	if err != nil {
		return dto.CounterpartyPageResponse{}, fmt.Errorf("list counterparties: %w", err)
	}
	return mapCounterpartyPage(page), nil
}

func (h *CounterpartyHandler) CreateCounterparty(req dto.CounterpartyWriteRequest) (dto.CounterpartyResponse, error) {
	input, err := parseCounterpartyWriteRequest(req)
	if err != nil {
		return dto.CounterpartyResponse{}, err
	}
	value, err := h.service.CreateCounterparty(handlerContext(), application.CounterpartyCreateInput{
		Name: input.Name, Phone: input.Phone, Email: input.Email, Notes: input.Notes, Roles: input.Roles,
	})
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("create counterparty: %w", err)
	}
	return mapCounterparty(value), nil
}

func (h *CounterpartyHandler) UpdateCounterparty(id int64, req dto.CounterpartyUpdateRequest) (dto.CounterpartyResponse, error) {
	counterpartyID, err := domain.NewCounterpartyID(id)
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("counterparty id: %w", err)
	}
	input, err := parseCounterpartyWriteRequest(req.CounterpartyWriteRequest)
	if err != nil {
		return dto.CounterpartyResponse{}, err
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("expected updated at: %w", err)
	}
	value, err := h.service.UpdateCounterparty(handlerContext(), application.CounterpartyUpdateInput{
		ID: counterpartyID, Name: input.Name, Phone: input.Phone, Email: input.Email,
		Notes: input.Notes, Roles: input.Roles, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("update counterparty: %w", err)
	}
	return mapCounterparty(value), nil
}

func (h *CounterpartyHandler) ArchiveCounterparty(id int64, req dto.VersionedCounterpartyRequest) (dto.CounterpartyResponse, error) {
	counterpartyID, expectedUpdatedAt, err := parseVersionedCounterparty(id, req)
	if err != nil {
		return dto.CounterpartyResponse{}, err
	}
	value, err := h.service.ArchiveCounterparty(handlerContext(), application.CounterpartyArchiveInput{
		ID: counterpartyID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("archive counterparty: %w", err)
	}
	return mapCounterparty(value), nil
}

func (h *CounterpartyHandler) RestoreCounterparty(id int64, req dto.VersionedCounterpartyRequest) (dto.CounterpartyResponse, error) {
	counterpartyID, expectedUpdatedAt, err := parseVersionedCounterparty(id, req)
	if err != nil {
		return dto.CounterpartyResponse{}, err
	}
	value, err := h.service.RestoreCounterparty(handlerContext(), application.CounterpartyRestoreInput{
		ID: counterpartyID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.CounterpartyResponse{}, fmt.Errorf("restore counterparty: %w", err)
	}
	return mapCounterparty(value), nil
}

type parsedCounterpartyWrite struct {
	Name  domain.DisplayName
	Phone domain.Option[domain.NonEmptyText]
	Email domain.Option[domain.NonEmptyText]
	Notes domain.Option[domain.NonEmptyText]
	Roles counterpartydomain.RoleSet
}

func parseCounterpartyWriteRequest(req dto.CounterpartyWriteRequest) (parsedCounterpartyWrite, error) {
	name, err := domain.NewDisplayName(req.Name)
	if err != nil {
		return parsedCounterpartyWrite{}, fmt.Errorf("name: %w", err)
	}
	phone, err := optionalNonEmptyText(req.Phone)
	if err != nil {
		return parsedCounterpartyWrite{}, fmt.Errorf("phone: %w", err)
	}
	email, err := optionalNonEmptyText(req.Email)
	if err != nil {
		return parsedCounterpartyWrite{}, fmt.Errorf("email: %w", err)
	}
	notes, err := optionalNonEmptyText(req.Notes)
	if err != nil {
		return parsedCounterpartyWrite{}, fmt.Errorf("notes: %w", err)
	}
	roles := make([]domain.CounterpartyRole, 0, len(req.Roles))
	for _, raw := range req.Roles {
		role, err := domain.ParseCounterpartyRole(raw)
		if err != nil {
			return parsedCounterpartyWrite{}, fmt.Errorf("role: %w", err)
		}
		roles = append(roles, role)
	}
	roleSet, err := counterpartydomain.NewRoleSet(roles...)
	if err != nil {
		return parsedCounterpartyWrite{}, fmt.Errorf("roles: %w", err)
	}
	return parsedCounterpartyWrite{Name: name, Phone: phone, Email: email, Notes: notes, Roles: roleSet}, nil
}

func parseCounterpartyListRequest(req dto.CounterpartyListRequest) (application.CounterpartyListInput, error) {
	archive := domain.ArchiveActive
	if req.ArchiveFilter != "" {
		parsed, err := domain.ParseArchiveFilter(req.ArchiveFilter)
		if err != nil {
			return application.CounterpartyListInput{}, err
		}
		archive = parsed
	}
	role := domain.None[domain.CounterpartyRole]()
	if req.Role != nil {
		parsed, err := domain.ParseCounterpartyRole(*req.Role)
		if err != nil {
			return application.CounterpartyListInput{}, err
		}
		role = domain.Some(parsed)
	}
	search, err := optionalNonEmptyText(req.Search)
	if err != nil {
		return application.CounterpartyListInput{}, fmt.Errorf("search: %w", err)
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 50
	}
	after := domain.None[application.CounterpartyCursor]()
	if req.After != nil {
		name, err := domain.NewDisplayName(req.After.Name)
		if err != nil {
			return application.CounterpartyListInput{}, fmt.Errorf("cursor name: %w", err)
		}
		id, err := domain.NewCounterpartyID(req.After.ID)
		if err != nil {
			return application.CounterpartyListInput{}, fmt.Errorf("cursor id: %w", err)
		}
		after = domain.Some(application.CounterpartyCursor{Name: name, ID: id})
	}
	return application.CounterpartyListInput{
		Archive:  archive,
		Role:     role,
		Search:   search,
		After:    after,
		PageSize: pageSize,
	}, nil
}

func parseVersionedCounterparty(id int64, req dto.VersionedCounterpartyRequest) (domain.CounterpartyID, domain.UTCInstant, error) {
	counterpartyID, err := domain.NewCounterpartyID(id)
	if err != nil {
		return domain.CounterpartyID{}, domain.UTCInstant{}, fmt.Errorf("counterparty id: %w", err)
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return domain.CounterpartyID{}, domain.UTCInstant{}, fmt.Errorf("expected updated at: %w", err)
	}
	return counterpartyID, expectedUpdatedAt, nil
}

func mapCounterpartyPage(page application.CounterpartyPage) dto.CounterpartyPageResponse {
	items := page.Items()
	response := dto.CounterpartyPageResponse{Items: make([]dto.CounterpartyResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapCounterparty(item))
	}
	if cursor, ok := page.Next().Get(); ok {
		response.Next = &dto.CounterpartyCursorResponse{Name: cursor.Name.String(), ID: cursor.ID.Int64()}
	}
	return response
}

func mapCounterparty(value counterpartydomain.Counterparty) dto.CounterpartyResponse {
	roles := value.Roles().Roles()
	roleValues := make([]string, 0, len(roles))
	for _, role := range roles {
		roleValues = append(roleValues, role.String())
	}
	return dto.CounterpartyResponse{
		ID:           value.ID().Int64(),
		Name:         value.Name().String(),
		Phone:        optionalText(value.Phone()),
		Email:        optionalText(value.Email()),
		Notes:        optionalText(value.Notes()),
		Roles:        roleValues,
		CreatedAtMs:  value.CreatedAt().UnixMilli(),
		UpdatedAtMs:  value.UpdatedAt().UnixMilli(),
		ArchivedAtMs: optionalInstant(value.ArchivedAt()),
	}
}

func optionalNonEmptyText(value *string) (domain.Option[domain.NonEmptyText], error) {
	if value == nil {
		return domain.None[domain.NonEmptyText](), nil
	}
	text, err := domain.NewNonEmptyText(*value)
	if err != nil {
		return domain.None[domain.NonEmptyText](), err
	}
	return domain.Some(text), nil
}

func optionalText(value domain.Option[domain.NonEmptyText]) *string {
	text, ok := value.Get()
	if !ok {
		return nil
	}
	raw := text.String()
	return &raw
}

func optionalInstant(value domain.Option[domain.UTCInstant]) *int64 {
	instant, ok := value.Get()
	if !ok {
		return nil
	}
	raw := instant.UnixMilli()
	return &raw
}
