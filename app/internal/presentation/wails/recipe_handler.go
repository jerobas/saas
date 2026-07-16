package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	recipedomain "github.com/jerobas/saas/internal/domain/recipe"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type RecipeHandler struct {
	service *application.RecipeService
}

func NewRecipeHandler(service *application.RecipeService) *RecipeHandler {
	if service == nil {
		panic("recipe handler requires a service")
	}
	return &RecipeHandler{service: service}
}

func (h *RecipeHandler) GetRecipe(id int64) (dto.RecipeResponse, error) {
	recipeID, err := domain.NewRecipeID(id)
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("recipe id: %w", err)
	}
	value, err := h.service.GetRecipe(handlerContext(), recipeID)
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("get recipe: %w", err)
	}
	return mapRecipe(value), nil
}

func (h *RecipeHandler) GetRecipeRevision(id int64) (dto.RecipeRevisionResponse, error) {
	revisionID, err := domain.NewRecipeRevisionID(id)
	if err != nil {
		return dto.RecipeRevisionResponse{}, fmt.Errorf("recipe revision id: %w", err)
	}
	value, err := h.service.GetRecipeRevision(handlerContext(), revisionID)
	if err != nil {
		return dto.RecipeRevisionResponse{}, fmt.Errorf("get recipe revision: %w", err)
	}
	return mapRecipeRevision(value), nil
}

func (h *RecipeHandler) ListRecipeRevisions(recipeIDValue int64) ([]dto.RecipeRevisionResponse, error) {
	recipeID, err := domain.NewRecipeID(recipeIDValue)
	if err != nil {
		return nil, fmt.Errorf("recipe id: %w", err)
	}
	values, err := h.service.ListRecipeRevisions(handlerContext(), recipeID)
	if err != nil {
		return nil, fmt.Errorf("list recipe revisions: %w", err)
	}
	response := make([]dto.RecipeRevisionResponse, 0, len(values))
	for _, value := range values {
		response = append(response, mapRecipeRevision(value))
	}
	return response, nil
}

func (h *RecipeHandler) ListRecipes(req dto.RecipeListRequest) (dto.RecipePageResponse, error) {
	input, err := parseRecipeListRequest(req)
	if err != nil {
		return dto.RecipePageResponse{}, err
	}
	page, err := h.service.ListRecipes(handlerContext(), input)
	if err != nil {
		return dto.RecipePageResponse{}, fmt.Errorf("list recipes: %w", err)
	}
	return mapRecipePage(page), nil
}

func (h *RecipeHandler) CreateRecipe(req dto.RecipeCreateRequest) (dto.RecipeResponse, error) {
	input, err := parseRecipeCreateRequest(req)
	if err != nil {
		return dto.RecipeResponse{}, err
	}
	value, err := h.service.CreateRecipe(handlerContext(), input)
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("create recipe: %w", err)
	}
	return mapRecipe(value), nil
}

func (h *RecipeHandler) PublishRecipeRevision(id int64, req dto.RecipePublishRevisionRequest) (dto.RecipeRevisionResponse, error) {
	recipeID, err := domain.NewRecipeID(id)
	if err != nil {
		return dto.RecipeRevisionResponse{}, fmt.Errorf("recipe id: %w", err)
	}
	revision, err := parseRecipeRevisionWriteRequest(req.Revision)
	if err != nil {
		return dto.RecipeRevisionResponse{}, err
	}
	expectedLatest, err := domain.NewRevisionNumber(req.ExpectedLatestRevision)
	if err != nil {
		return dto.RecipeRevisionResponse{}, fmt.Errorf("expected latest revision: %w", err)
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return dto.RecipeRevisionResponse{}, fmt.Errorf("expected updated at: %w", err)
	}
	value, err := h.service.PublishRecipeRevision(handlerContext(), application.RecipePublishRevisionInput{
		RecipeID:               recipeID,
		ExpectedLatestRevision: expectedLatest,
		ExpectedUpdatedAt:      expectedUpdatedAt,
		Revision:               revision,
	})
	if err != nil {
		return dto.RecipeRevisionResponse{}, fmt.Errorf("publish recipe revision: %w", err)
	}
	return mapRecipeRevision(value), nil
}

func (h *RecipeHandler) RenameRecipe(id int64, req dto.RecipeRenameRequest) (dto.RecipeResponse, error) {
	recipeID, err := domain.NewRecipeID(id)
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("recipe id: %w", err)
	}
	name, err := domain.NewUniqueName(req.Name)
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("name: %w", err)
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("expected updated at: %w", err)
	}
	value, err := h.service.RenameRecipe(handlerContext(), application.RecipeRenameInput{
		ID: recipeID, Name: name, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("rename recipe: %w", err)
	}
	return mapRecipe(value), nil
}

func (h *RecipeHandler) ArchiveRecipe(id int64, req dto.VersionedRequest) (dto.RecipeResponse, error) {
	recipeID, expectedUpdatedAt, err := parseVersionedRecipe(id, req)
	if err != nil {
		return dto.RecipeResponse{}, err
	}
	value, err := h.service.ArchiveRecipe(handlerContext(), application.RecipeArchiveInput{
		ID: recipeID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("archive recipe: %w", err)
	}
	return mapRecipe(value), nil
}

func (h *RecipeHandler) RestoreRecipe(id int64, req dto.VersionedRequest) (dto.RecipeResponse, error) {
	recipeID, expectedUpdatedAt, err := parseVersionedRecipe(id, req)
	if err != nil {
		return dto.RecipeResponse{}, err
	}
	value, err := h.service.RestoreRecipe(handlerContext(), application.RecipeRestoreInput{
		ID: recipeID, ExpectedUpdatedAt: expectedUpdatedAt,
	})
	if err != nil {
		return dto.RecipeResponse{}, fmt.Errorf("restore recipe: %w", err)
	}
	return mapRecipe(value), nil
}

func parseRecipeListRequest(req dto.RecipeListRequest) (application.RecipeListInput, error) {
	archive := domain.ArchiveActive
	if req.ArchiveFilter != "" {
		parsed, err := domain.ParseArchiveFilter(req.ArchiveFilter)
		if err != nil {
			return application.RecipeListInput{}, err
		}
		archive = parsed
	}
	search, err := optionalNonEmptyText(req.Search)
	if err != nil {
		return application.RecipeListInput{}, fmt.Errorf("search: %w", err)
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 50
	}
	after := domain.None[application.RecipeCursor]()
	if req.After != nil {
		name, err := domain.NewUniqueName(req.After.Name)
		if err != nil {
			return application.RecipeListInput{}, fmt.Errorf("cursor name: %w", err)
		}
		id, err := domain.NewRecipeID(req.After.ID)
		if err != nil {
			return application.RecipeListInput{}, fmt.Errorf("cursor id: %w", err)
		}
		after = domain.Some(application.RecipeCursor{Name: name, ID: id})
	}
	return application.RecipeListInput{Archive: archive, Search: search, After: after, PageSize: pageSize}, nil
}

func parseRecipeCreateRequest(req dto.RecipeCreateRequest) (application.RecipeCreateInput, error) {
	name, err := domain.NewUniqueName(req.Name)
	if err != nil {
		return application.RecipeCreateInput{}, fmt.Errorf("name: %w", err)
	}
	outputItemID, err := domain.NewItemID(req.OutputItemID)
	if err != nil {
		return application.RecipeCreateInput{}, fmt.Errorf("output item id: %w", err)
	}
	revision, err := parseRecipeRevisionWriteRequest(req.Revision)
	if err != nil {
		return application.RecipeCreateInput{}, err
	}
	return application.RecipeCreateInput{Name: name, OutputItemID: outputItemID, Revision: revision}, nil
}

func parseRecipeRevisionWriteRequest(req dto.RecipeRevisionWriteRequest) (application.RecipeRevisionWriteInput, error) {
	standardYield, err := domain.NewPositiveAtomicQuantity(req.StandardYieldQuantity)
	if err != nil {
		return application.RecipeRevisionWriteInput{}, fmt.Errorf("standard yield: %w", err)
	}
	preparationTime, err := domain.NewPreparationMinutes(req.PreparationTimeMinutes)
	if err != nil {
		return application.RecipeRevisionWriteInput{}, fmt.Errorf("preparation time: %w", err)
	}
	estimatedCost := domain.None[domain.InventoryValue]()
	if req.EstimatedDirectCostMicro != nil {
		value, err := domain.NewInventoryValue(*req.EstimatedDirectCostMicro)
		if err != nil {
			return application.RecipeRevisionWriteInput{}, fmt.Errorf("estimated direct cost: %w", err)
		}
		estimatedCost = domain.Some(value)
	}
	components := make([]application.RecipeComponentInput, 0, len(req.Components))
	for index, component := range req.Components {
		parsed, err := parseRecipeComponentRequest(component)
		if err != nil {
			return application.RecipeRevisionWriteInput{}, fmt.Errorf("component %d: %w", index+1, err)
		}
		components = append(components, parsed)
	}
	return application.RecipeRevisionWriteInput{
		StandardYield:       standardYield,
		Instructions:        req.Instructions,
		PreparationTime:     preparationTime,
		EstimatedDirectCost: estimatedCost,
		Components:          components,
	}, nil
}

func parseRecipeComponentRequest(req dto.RecipeComponentRequest) (application.RecipeComponentInput, error) {
	order, err := domain.NewComponentOrder(req.Order)
	if err != nil {
		return application.RecipeComponentInput{}, fmt.Errorf("order: %w", err)
	}
	itemID, err := domain.NewItemID(req.ItemID)
	if err != nil {
		return application.RecipeComponentInput{}, fmt.Errorf("item id: %w", err)
	}
	quantity, err := domain.NewPositiveAtomicQuantity(req.QuantityAtomic)
	if err != nil {
		return application.RecipeComponentInput{}, fmt.Errorf("quantity: %w", err)
	}
	source, err := parseRecipeComponentSource(req)
	if err != nil {
		return application.RecipeComponentInput{}, err
	}
	return application.RecipeComponentInput{Order: order, ItemID: itemID, Quantity: quantity, Source: source}, nil
}

func parseRecipeComponentSource(req dto.RecipeComponentRequest) (application.RecipeComponentSource, error) {
	switch application.RecipeComponentSourceKind(req.SourceType) {
	case application.RecipeComponentSourceUnit:
		if req.UnitCode == nil {
			return application.RecipeComponentSource{}, domain.Invalid("unit_code", domain.ViolationRequired, "UNIT-005")
		}
		unit, err := domain.NewUnitCode(*req.UnitCode)
		if err != nil {
			return application.RecipeComponentSource{}, fmt.Errorf("unit code: %w", err)
		}
		return application.NewRecipeUnitComponentSource(unit), nil
	case application.RecipeComponentSourcePackaging:
		if req.PackagingID == nil {
			return application.RecipeComponentSource{}, domain.Invalid("packaging_id", domain.ViolationRequired, "UNIT-005")
		}
		packagingID, err := domain.NewPackagingID(*req.PackagingID)
		if err != nil {
			return application.RecipeComponentSource{}, fmt.Errorf("packaging id: %w", err)
		}
		return application.NewRecipePackagingComponentSource(packagingID), nil
	default:
		return application.RecipeComponentSource{}, domain.Invalid("component_source", domain.ViolationInvalidEnum, "UNIT-005")
	}
}

func parseVersionedRecipe(id int64, req dto.VersionedRequest) (domain.RecipeID, domain.UTCInstant, error) {
	recipeID, err := domain.NewRecipeID(id)
	if err != nil {
		return domain.RecipeID{}, domain.UTCInstant{}, fmt.Errorf("recipe id: %w", err)
	}
	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return domain.RecipeID{}, domain.UTCInstant{}, fmt.Errorf("expected updated at: %w", err)
	}
	return recipeID, expectedUpdatedAt, nil
}

func mapRecipePage(page application.RecipePage) dto.RecipePageResponse {
	items := page.Items()
	response := dto.RecipePageResponse{Items: make([]dto.RecipeSummaryResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, mapRecipeSummary(item))
	}
	if cursor, ok := page.Next().Get(); ok {
		response.Next = &dto.RecipeCursorResponse{Name: cursor.Name.Display(), ID: cursor.ID.Int64()}
	}
	return response
}

func mapRecipeSummary(summary recipedomain.RecipeSummary) dto.RecipeSummaryResponse {
	current := summary.CurrentRevision()
	return dto.RecipeSummaryResponse{
		ID:             summary.ID().Int64(),
		Name:           summary.Name().Display(),
		OutputItemID:   summary.OutputItemID().Int64(),
		OutputItemName: summary.OutputItemName().String(),
		CreatedAtMs:    summary.CreatedAt().UnixMilli(),
		UpdatedAtMs:    summary.UpdatedAt().UnixMilli(),
		ArchivedAtMs:   optionalInstant(summary.ArchivedAt()),
		CurrentRevision: dto.CurrentRecipeRevisionSummaryResponse{
			ID:                    current.ID().Int64(),
			Number:                current.Number().Int64(),
			StandardYieldQuantity: current.StandardYield().Int64(),
		},
	}
}

func mapRecipe(value recipedomain.Recipe) dto.RecipeResponse {
	return dto.RecipeResponse{
		ID:              value.ID().Int64(),
		Name:            value.Name().Display(),
		OutputItemID:    value.OutputItemID().Int64(),
		CreatedAtMs:     value.CreatedAt().UnixMilli(),
		UpdatedAtMs:     value.UpdatedAt().UnixMilli(),
		ArchivedAtMs:    optionalInstant(value.ArchivedAt()),
		CurrentRevision: mapRecipeRevision(value.CurrentRevision()),
	}
}

func mapRecipeRevision(revision recipedomain.Revision) dto.RecipeRevisionResponse {
	components := revision.Components()
	response := dto.RecipeRevisionResponse{
		ID:                       revision.ID().Int64(),
		RecipeID:                 revision.RecipeID().Int64(),
		Number:                   revision.Number().Int64(),
		StandardYieldQuantity:    revision.StandardYield().Int64(),
		Instructions:             revision.Instructions(),
		PreparationTimeMinutes:   revision.PreparationTime().Int64(),
		EstimatedDirectCostMicro: optionalInventoryValueMicro(revision.EstimatedDirectCost()),
		CreatedAtMs:              revision.CreatedAt().UnixMilli(),
		Components:               make([]dto.RecipeComponentResponse, 0, len(components)),
	}
	for _, component := range components {
		response.Components = append(response.Components, mapRecipeComponent(component))
	}
	return response
}

func mapRecipeComponent(component recipedomain.Component) dto.RecipeComponentResponse {
	return dto.RecipeComponentResponse{
		ID:                        component.ID().Int64(),
		RevisionID:                component.RevisionID().Int64(),
		Order:                     component.Order().Int64(),
		ItemID:                    component.ItemID().Int64(),
		QuantityAtomic:            component.Quantity().Int64(),
		EnteredUnitCode:           component.EnteredUnit().String(),
		EnteredPackagingName:      optionalText(component.EnteredPackagingName()),
		ConversionNumeratorAtomic: component.Conversion().NumeratorAtomic(),
		ConversionDenominator:     component.Conversion().Denominator(),
		CreatedAtMs:               component.CreatedAt().UnixMilli(),
	}
}

func optionalInventoryValueMicro(value domain.Option[domain.InventoryValue]) *int64 {
	amount, ok := value.Get()
	if !ok {
		return nil
	}
	raw := amount.Int64()
	return &raw
}
