package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type SaleHandler struct {
	service *application.SaleService
}

func NewSaleHandler(service *application.SaleService) *SaleHandler {
	if service == nil {
		panic("sale handler requires a service")
	}
	return &SaleHandler{service: service}
}

func (h *SaleHandler) PostSale(req dto.SalePostRequest) (dto.SaleDocumentResponse, error) {
	input, err := parseSalePostRequest(req)
	if err != nil {
		return dto.SaleDocumentResponse{}, err
	}
	posted, err := h.service.PostSale(handlerContext(), input)
	if err != nil {
		return dto.SaleDocumentResponse{}, fmt.Errorf("post sale: %w", err)
	}
	return mapSaleDocument(posted), nil
}

func parseSalePostRequest(req dto.SalePostRequest) (application.SalePostInput, error) {
	idempotencyKey, err := domain.NewIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		return application.SalePostInput{}, fmt.Errorf("idempotency key: %w", err)
	}
	counterpartyID := domain.None[domain.CounterpartyID]()
	if req.CounterpartyID != nil {
		parsed, err := domain.NewCounterpartyID(*req.CounterpartyID)
		if err != nil {
			return application.SalePostInput{}, fmt.Errorf("counterparty id: %w", err)
		}
		counterpartyID = domain.Some(parsed)
	}
	occurredOn, err := domain.ParseBusinessDate(req.OccurredOn)
	if err != nil {
		return application.SalePostInput{}, fmt.Errorf("occurred on: %w", err)
	}
	reason, err := optionalSaleReason(req.ReasonCode)
	if err != nil {
		return application.SalePostInput{}, err
	}
	notes, err := optionalNonEmptyText(req.Notes)
	if err != nil {
		return application.SalePostInput{}, fmt.Errorf("notes: %w", err)
	}
	lines := make([]application.SaleLineInput, 0, len(req.Lines))
	for index, line := range req.Lines {
		parsed, err := parseSaleLineRequest(line)
		if err != nil {
			return application.SalePostInput{}, fmt.Errorf("line %d: %w", index+1, err)
		}
		lines = append(lines, parsed)
	}
	return application.SalePostInput{
		IdempotencyKey: idempotencyKey,
		CounterpartyID: counterpartyID,
		OccurredOn:     occurredOn,
		Reason:         reason,
		Notes:          notes,
		Lines:          lines,
	}, nil
}

func parseSaleLineRequest(req dto.SaleLineRequest) (application.SaleLineInput, error) {
	itemID, err := domain.NewItemID(req.ItemID)
	if err != nil {
		return application.SaleLineInput{}, fmt.Errorf("item id: %w", err)
	}
	quantity, err := domain.NewPositiveAtomicQuantity(req.QuantityAtomic)
	if err != nil {
		return application.SaleLineInput{}, fmt.Errorf("quantity: %w", err)
	}
	enteredUnit, err := domain.NewUnitCode(req.EnteredUnitCode)
	if err != nil {
		return application.SaleLineInput{}, fmt.Errorf("entered unit: %w", err)
	}
	enteredPackagingName, err := optionalNonEmptyText(req.EnteredPackagingName)
	if err != nil {
		return application.SaleLineInput{}, fmt.Errorf("entered packaging name: %w", err)
	}
	conversion, err := domain.NewUnitConversion(req.ConversionNumeratorAtomic, req.ConversionDenominator)
	if err != nil {
		return application.SaleLineInput{}, fmt.Errorf("conversion: %w", err)
	}
	commercialTotal, err := domain.NewMinorAmount(req.CommercialTotalMinor)
	if err != nil {
		return application.SaleLineInput{}, fmt.Errorf("commercial total: %w", err)
	}
	lotID := domain.None[domain.InventoryLotID]()
	if req.LotID != nil {
		parsed, err := domain.NewInventoryLotID(*req.LotID)
		if err != nil {
			return application.SaleLineInput{}, fmt.Errorf("lot id: %w", err)
		}
		lotID = domain.Some(parsed)
	}
	return application.SaleLineInput{
		ItemID:               itemID,
		Quantity:             quantity,
		EnteredUnit:          enteredUnit,
		EnteredPackagingName: enteredPackagingName,
		Conversion:           conversion,
		CommercialTotal:      commercialTotal,
		LotID:                lotID,
	}, nil
}

func optionalSaleReason(value *string) (domain.Option[domain.DocumentReason], error) {
	if value == nil {
		return domain.None[domain.DocumentReason](), nil
	}
	return domain.ParseDocumentReason(domain.DocumentSale, *value)
}

func mapSaleDocument(document application.SaleDocument) dto.SaleDocumentResponse {
	lines := document.Lines()
	response := dto.SaleDocumentResponse{
		ID:                  document.ID().Int64(),
		IdempotencyKey:      document.IdempotencyKey().String(),
		PostingSequence:     document.PostingSequence().Int64(),
		CounterpartyID:      optionalCounterpartyIDValue(document.CounterpartyID()),
		OccurredOn:          document.OccurredOn().String(),
		PostedAtMs:          document.PostedAt().UnixMilli(),
		CurrencyCode:        document.Currency().Code().String(),
		CurrencyMinorDigits: int64(document.Currency().MinorDigits().Int()),
		ReasonCode:          optionalDocumentReasonValue(document.Reason()),
		Notes:               optionalText(document.Notes()),
		Lines:               make([]dto.SaleLineResponse, 0, len(lines)),
	}
	for _, line := range lines {
		response.Lines = append(response.Lines, mapSaleLine(line))
	}
	return response
}

func mapSaleLine(line application.PostedSaleLine) dto.SaleLineResponse {
	allocations := line.Allocations()
	response := dto.SaleLineResponse{
		ID:                        line.ID().Int64(),
		LineOrder:                 line.LineOrder().Int64(),
		ItemID:                    line.ItemID().Int64(),
		Direction:                 domain.DirectionOut.String(),
		QuantityAtomic:            line.Quantity().Int64(),
		EnteredUnitCode:           line.EnteredUnit().String(),
		EnteredPackagingName:      optionalText(line.EnteredPackagingName()),
		ConversionNumeratorAtomic: line.Conversion().NumeratorAtomic(),
		ConversionDenominator:     line.Conversion().Denominator(),
		InventoryValueMicro:       line.InventoryValue().Int64(),
		CommercialTotalMinor:      line.CommercialTotal().Int64(),
		Allocations:               make([]dto.SaleAllocationResponse, 0, len(allocations)),
	}
	for _, allocation := range allocations {
		response.Allocations = append(response.Allocations, dto.SaleAllocationResponse{
			ID:             allocation.ID().Int64(),
			LotID:          allocation.LotID().Int64(),
			QuantityAtomic: allocation.Quantity().Int64(),
		})
	}
	return response
}
