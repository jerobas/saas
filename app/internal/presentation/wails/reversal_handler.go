package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type ReversalHandler struct {
	service *application.ReversalService
}

func NewReversalHandler(service *application.ReversalService) *ReversalHandler {
	if service == nil {
		panic("reversal handler requires a service")
	}
	return &ReversalHandler{service: service}
}

func (h *ReversalHandler) PostReversal(req dto.ReversalPostRequest) (dto.ReversalDocumentResponse, error) {
	input, err := parseReversalPostRequest(req)
	if err != nil {
		return dto.ReversalDocumentResponse{}, err
	}
	posted, err := h.service.PostReversal(handlerContext(), input)
	if err != nil {
		return dto.ReversalDocumentResponse{}, fmt.Errorf("post reversal: %w", err)
	}
	return mapReversalDocument(posted), nil
}

func parseReversalPostRequest(req dto.ReversalPostRequest) (application.ReversalPostInput, error) {
	idempotencyKey, err := domain.NewIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		return application.ReversalPostInput{}, fmt.Errorf("idempotency key: %w", err)
	}
	targetDocumentID, err := domain.NewStockDocumentID(req.TargetDocumentID)
	if err != nil {
		return application.ReversalPostInput{}, fmt.Errorf("target document id: %w", err)
	}
	occurredOn, err := domain.ParseBusinessDate(req.OccurredOn)
	if err != nil {
		return application.ReversalPostInput{}, fmt.Errorf("occurred on: %w", err)
	}
	notes, err := optionalNonEmptyText(req.Notes)
	if err != nil {
		return application.ReversalPostInput{}, fmt.Errorf("notes: %w", err)
	}
	return application.ReversalPostInput{
		IdempotencyKey:   idempotencyKey,
		TargetDocumentID: targetDocumentID,
		OccurredOn:       occurredOn,
		Notes:            notes,
	}, nil
}

func mapReversalDocument(document application.ReversalDocument) dto.ReversalDocumentResponse {
	lines := document.Lines()
	response := dto.ReversalDocumentResponse{
		ID:                  document.ID().Int64(),
		IdempotencyKey:      document.IdempotencyKey().String(),
		PostingSequence:     document.PostingSequence().Int64(),
		TargetDocumentID:    document.TargetDocumentID().Int64(),
		OccurredOn:          document.OccurredOn().String(),
		PostedAtMs:          document.PostedAt().UnixMilli(),
		CurrencyCode:        document.Currency().Code().String(),
		CurrencyMinorDigits: int64(document.Currency().MinorDigits().Int()),
		ReasonCode:          document.Reason().String(),
		Notes:               optionalText(document.Notes()),
		Lines:               make([]dto.ReversalLineResponse, 0, len(lines)),
	}
	for _, line := range lines {
		response.Lines = append(response.Lines, mapReversalLine(line))
	}
	return response
}

func mapReversalLine(line application.PostedReversalLine) dto.ReversalLineResponse {
	allocations := line.Allocations()
	response := dto.ReversalLineResponse{
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
		CommercialTotalMinor:      optionalMinorAmountValue(line.CommercialTotal()),
		ReversesLineID:            line.ReversesLineID().Int64(),
		Allocations:               make([]dto.ReversalAllocationResponse, 0, len(allocations)),
	}
	for _, allocation := range allocations {
		response.Allocations = append(response.Allocations, dto.ReversalAllocationResponse{
			ID:                   allocation.ID().Int64(),
			LotID:                allocation.LotID().Int64(),
			QuantityAtomic:       allocation.Quantity().Int64(),
			RestoresAllocationID: optionalLotAllocationIDValue(allocation.RestoresAllocationID()),
		})
	}
	return response
}

func optionalMinorAmountValue(value domain.Option[domain.MinorAmount]) *int64 {
	amount, ok := value.Get()
	if !ok {
		return nil
	}
	raw := amount.Int64()
	return &raw
}

func optionalLotAllocationIDValue(value domain.Option[domain.LotAllocationID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}
