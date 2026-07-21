package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type ProductionHandler struct {
	service *application.ProductionService
}

func NewProductionHandler(service *application.ProductionService) *ProductionHandler {
	if service == nil {
		panic("production handler requires a service")
	}
	return &ProductionHandler{service: service}
}

func (h *ProductionHandler) PostProduction(req dto.ProductionPostRequest) (dto.ProductionDocumentResponse, error) {
	input, err := parseProductionPostRequest(req)
	if err != nil {
		return dto.ProductionDocumentResponse{}, err
	}
	posted, err := h.service.PostProduction(handlerContext(), input)
	if err != nil {
		return dto.ProductionDocumentResponse{}, fmt.Errorf("post production: %w", err)
	}
	return mapProductionDocument(posted), nil
}

func parseProductionPostRequest(req dto.ProductionPostRequest) (application.ProductionPostInput, error) {
	idempotencyKey, err := domain.NewIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		return application.ProductionPostInput{}, fmt.Errorf("idempotency key: %w", err)
	}
	revisionID, err := domain.NewRecipeRevisionID(req.RecipeRevisionID)
	if err != nil {
		return application.ProductionPostInput{}, fmt.Errorf("recipe revision id: %w", err)
	}
	occurredOn, err := domain.ParseBusinessDate(req.OccurredOn)
	if err != nil {
		return application.ProductionPostInput{}, fmt.Errorf("occurred on: %w", err)
	}
	directCost, err := domain.NewInventoryValue(req.DirectCostMicro)
	if err != nil {
		return application.ProductionPostInput{}, fmt.Errorf("direct cost: %w", err)
	}
	notes, err := optionalNonEmptyText(req.Notes)
	if err != nil {
		return application.ProductionPostInput{}, fmt.Errorf("notes: %w", err)
	}
	output, err := parseProductionOutputRequest(req.Output)
	if err != nil {
		return application.ProductionPostInput{}, fmt.Errorf("output: %w", err)
	}
	inputs := make([]application.ProductionComponentInput, 0, len(req.Inputs))
	for index, line := range req.Inputs {
		parsed, err := parseProductionComponentRequest(line)
		if err != nil {
			return application.ProductionPostInput{}, fmt.Errorf("input %d: %w", index+1, err)
		}
		inputs = append(inputs, parsed)
	}
	return application.ProductionPostInput{
		IdempotencyKey:   idempotencyKey,
		RecipeRevisionID: revisionID,
		OccurredOn:       occurredOn,
		DirectCost:       directCost,
		Notes:            notes,
		Output:           output,
		Inputs:           inputs,
	}, nil
}

func parseProductionOutputRequest(req dto.ProductionOutputRequest) (application.ProductionOutputInput, error) {
	quantity, err := domain.NewPositiveAtomicQuantity(req.QuantityAtomic)
	if err != nil {
		return application.ProductionOutputInput{}, fmt.Errorf("quantity: %w", err)
	}
	enteredUnit, err := domain.NewUnitCode(req.EnteredUnitCode)
	if err != nil {
		return application.ProductionOutputInput{}, fmt.Errorf("entered unit: %w", err)
	}
	enteredPackagingName, err := optionalNonEmptyText(req.EnteredPackagingName)
	if err != nil {
		return application.ProductionOutputInput{}, fmt.Errorf("entered packaging name: %w", err)
	}
	conversion, err := domain.NewUnitConversion(req.ConversionNumeratorAtomic, req.ConversionDenominator)
	if err != nil {
		return application.ProductionOutputInput{}, fmt.Errorf("conversion: %w", err)
	}
	lotCode, err := optionalNonEmptyText(req.LotCode)
	if err != nil {
		return application.ProductionOutputInput{}, fmt.Errorf("lot code: %w", err)
	}
	expiresOn, err := optionalBusinessDateFromString(req.ExpiresOn)
	if err != nil {
		return application.ProductionOutputInput{}, fmt.Errorf("expires on: %w", err)
	}
	return application.ProductionOutputInput{
		Quantity:             quantity,
		EnteredUnit:          enteredUnit,
		EnteredPackagingName: enteredPackagingName,
		Conversion:           conversion,
		LotCode:              lotCode,
		ExpiresOn:            expiresOn,
	}, nil
}

func parseProductionComponentRequest(req dto.ProductionComponentRequest) (application.ProductionComponentInput, error) {
	itemID, err := domain.NewItemID(req.ItemID)
	if err != nil {
		return application.ProductionComponentInput{}, fmt.Errorf("item id: %w", err)
	}
	quantity, err := domain.NewPositiveAtomicQuantity(req.QuantityAtomic)
	if err != nil {
		return application.ProductionComponentInput{}, fmt.Errorf("quantity: %w", err)
	}
	enteredUnit, err := domain.NewUnitCode(req.EnteredUnitCode)
	if err != nil {
		return application.ProductionComponentInput{}, fmt.Errorf("entered unit: %w", err)
	}
	enteredPackagingName, err := optionalNonEmptyText(req.EnteredPackagingName)
	if err != nil {
		return application.ProductionComponentInput{}, fmt.Errorf("entered packaging name: %w", err)
	}
	conversion, err := domain.NewUnitConversion(req.ConversionNumeratorAtomic, req.ConversionDenominator)
	if err != nil {
		return application.ProductionComponentInput{}, fmt.Errorf("conversion: %w", err)
	}
	lotID := domain.None[domain.InventoryLotID]()
	if req.LotID != nil {
		parsed, err := domain.NewInventoryLotID(*req.LotID)
		if err != nil {
			return application.ProductionComponentInput{}, fmt.Errorf("lot id: %w", err)
		}
		lotID = domain.Some(parsed)
	}
	return application.ProductionComponentInput{
		ItemID:               itemID,
		Quantity:             quantity,
		EnteredUnit:          enteredUnit,
		EnteredPackagingName: enteredPackagingName,
		Conversion:           conversion,
		LotID:                lotID,
	}, nil
}

func mapProductionDocument(document application.ProductionDocument) dto.ProductionDocumentResponse {
	inputLines := document.InputLines()
	response := dto.ProductionDocumentResponse{
		ID:                  document.ID().Int64(),
		IdempotencyKey:      document.IdempotencyKey().String(),
		PostingSequence:     document.PostingSequence().Int64(),
		RecipeRevisionID:    document.RecipeRevisionID().Int64(),
		OutputItemID:        document.OutputItemID().Int64(),
		OccurredOn:          document.OccurredOn().String(),
		PostedAtMs:          document.PostedAt().UnixMilli(),
		CurrencyCode:        document.Currency().Code().String(),
		CurrencyMinorDigits: int64(document.Currency().MinorDigits().Int()),
		DirectCostMicro:     document.DirectCost().Int64(),
		Notes:               optionalText(document.Notes()),
		OutputLine:          mapProductionLine(document.OutputLine()),
		InputLines:          make([]dto.ProductionLineResponse, 0, len(inputLines)),
	}
	for _, line := range inputLines {
		response.InputLines = append(response.InputLines, mapProductionLine(line))
	}
	return response
}

func mapProductionLine(line application.PostedProductionLine) dto.ProductionLineResponse {
	allocations := line.Allocations()
	response := dto.ProductionLineResponse{
		ID:                        line.ID().Int64(),
		LineOrder:                 line.LineOrder().Int64(),
		ItemID:                    line.ItemID().Int64(),
		Direction:                 line.Direction().String(),
		QuantityAtomic:            line.Quantity().Int64(),
		EnteredUnitCode:           line.EnteredUnit().String(),
		EnteredPackagingName:      optionalText(line.EnteredPackagingName()),
		ConversionNumeratorAtomic: line.Conversion().NumeratorAtomic(),
		ConversionDenominator:     line.Conversion().Denominator(),
		InventoryValueMicro:       line.InventoryValue().Int64(),
		LotID:                     optionalInventoryLotID(line.LotID()),
		LotCode:                   optionalText(line.LotCode()),
		OriginatedOn:              optionalBusinessDateValue(line.OriginatedOn()),
		ExpiresOn:                 optionalBusinessDateValue(line.ExpiresOn()),
		Allocations:               make([]dto.ProductionAllocationResponse, 0, len(allocations)),
	}
	for _, allocation := range allocations {
		response.Allocations = append(response.Allocations, dto.ProductionAllocationResponse{
			ID:             allocation.ID().Int64(),
			LotID:          allocation.LotID().Int64(),
			QuantityAtomic: allocation.Quantity().Int64(),
		})
	}
	return response
}
