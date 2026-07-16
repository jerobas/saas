package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type AdjustmentHandler struct {
	service *application.AdjustmentService
}

func NewAdjustmentHandler(service *application.AdjustmentService) *AdjustmentHandler {
	if service == nil {
		panic("adjustment handler requires a service")
	}
	return &AdjustmentHandler{service: service}
}

func (h *AdjustmentHandler) PostAdjustment(req dto.AdjustmentPostRequest) (dto.AdjustmentDocumentResponse, error) {
	input, err := parseAdjustmentPostRequest(req)
	if err != nil {
		return dto.AdjustmentDocumentResponse{}, err
	}
	posted, err := h.service.PostAdjustment(handlerContext(), input)
	if err != nil {
		return dto.AdjustmentDocumentResponse{}, fmt.Errorf("post adjustment: %w", err)
	}
	return mapAdjustmentDocument(posted), nil
}

func parseAdjustmentPostRequest(req dto.AdjustmentPostRequest) (application.AdjustmentPostInput, error) {
	idempotencyKey, err := domain.NewIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		return application.AdjustmentPostInput{}, fmt.Errorf("idempotency key: %w", err)
	}
	occurredOn, err := domain.ParseBusinessDate(req.OccurredOn)
	if err != nil {
		return application.AdjustmentPostInput{}, fmt.Errorf("occurred on: %w", err)
	}
	reason, err := domain.ParseDocumentReason(domain.DocumentAdjustment, req.ReasonCode)
	if err != nil {
		return application.AdjustmentPostInput{}, fmt.Errorf("reason: %w", err)
	}
	parsedReason, ok := reason.Get()
	if !ok {
		return application.AdjustmentPostInput{}, fmt.Errorf("reason: %w", domain.ErrValidation)
	}
	notes, err := optionalNonEmptyText(req.Notes)
	if err != nil {
		return application.AdjustmentPostInput{}, fmt.Errorf("notes: %w", err)
	}
	lines := make([]application.AdjustmentLineInput, 0, len(req.Lines))
	for index, line := range req.Lines {
		parsed, err := parseAdjustmentLineRequest(line)
		if err != nil {
			return application.AdjustmentPostInput{}, fmt.Errorf("line %d: %w", index+1, err)
		}
		lines = append(lines, parsed)
	}
	return application.AdjustmentPostInput{
		IdempotencyKey: idempotencyKey,
		OccurredOn:     occurredOn,
		Reason:         parsedReason,
		Notes:          notes,
		Lines:          lines,
	}, nil
}

func parseAdjustmentLineRequest(req dto.AdjustmentLineRequest) (application.AdjustmentLineInput, error) {
	itemID, err := domain.NewItemID(req.ItemID)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("item id: %w", err)
	}
	direction, err := domain.ParseDirection(req.Direction)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("direction: %w", err)
	}
	quantity, err := domain.NewPositiveAtomicQuantity(req.QuantityAtomic)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("quantity: %w", err)
	}
	enteredUnit, err := domain.NewUnitCode(req.EnteredUnitCode)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("entered unit: %w", err)
	}
	enteredPackagingName, err := optionalNonEmptyText(req.EnteredPackagingName)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("entered packaging name: %w", err)
	}
	conversion, err := domain.NewUnitConversion(req.ConversionNumeratorAtomic, req.ConversionDenominator)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("conversion: %w", err)
	}
	inventoryValue := domain.None[domain.InventoryValue]()
	if req.InventoryValueMicro != nil {
		parsed, err := domain.NewInventoryValue(*req.InventoryValueMicro)
		if err != nil {
			return application.AdjustmentLineInput{}, fmt.Errorf("inventory value: %w", err)
		}
		inventoryValue = domain.Some(parsed)
	}
	lotCode, err := optionalNonEmptyText(req.LotCode)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("lot code: %w", err)
	}
	expiresOn, err := optionalBusinessDateFromString(req.ExpiresOn)
	if err != nil {
		return application.AdjustmentLineInput{}, fmt.Errorf("expires on: %w", err)
	}
	return application.AdjustmentLineInput{
		ItemID:               itemID,
		Direction:            direction,
		Quantity:             quantity,
		EnteredUnit:          enteredUnit,
		EnteredPackagingName: enteredPackagingName,
		Conversion:           conversion,
		InventoryValue:       inventoryValue,
		LotCode:              lotCode,
		ExpiresOn:            expiresOn,
	}, nil
}

func mapAdjustmentDocument(document application.AdjustmentDocument) dto.AdjustmentDocumentResponse {
	notes := optionalText(document.Notes())
	lines := document.Lines()
	response := dto.AdjustmentDocumentResponse{
		ID:                  document.ID().Int64(),
		IdempotencyKey:      document.IdempotencyKey().String(),
		PostingSequence:     document.PostingSequence().Int64(),
		OccurredOn:          document.OccurredOn().String(),
		PostedAtMs:          document.PostedAt().UnixMilli(),
		CurrencyCode:        document.Currency().Code().String(),
		CurrencyMinorDigits: int64(document.Currency().MinorDigits().Int()),
		ReasonCode:          document.Reason().String(),
		Notes:               notes,
		Lines:               make([]dto.AdjustmentLineResponse, 0, len(lines)),
	}
	for _, line := range lines {
		response.Lines = append(response.Lines, mapAdjustmentLine(line))
	}
	return response
}

func mapAdjustmentLine(line application.PostedAdjustmentLine) dto.AdjustmentLineResponse {
	allocations := line.Allocations()
	response := dto.AdjustmentLineResponse{
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
		Allocations:               make([]dto.AdjustmentAllocationResponse, 0, len(allocations)),
	}
	for _, allocation := range allocations {
		response.Allocations = append(response.Allocations, dto.AdjustmentAllocationResponse{
			ID:             allocation.ID().Int64(),
			LotID:          allocation.LotID().Int64(),
			QuantityAtomic: allocation.Quantity().Int64(),
		})
	}
	return response
}

func optionalInventoryLotID(value domain.Option[domain.InventoryLotID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}
