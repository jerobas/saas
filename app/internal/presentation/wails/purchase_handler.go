package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type PurchaseHandler struct {
	service *application.PurchaseService
}

func NewPurchaseHandler(service *application.PurchaseService) *PurchaseHandler {
	if service == nil {
		panic("purchase handler requires a service")
	}
	return &PurchaseHandler{service: service}
}

func (h *PurchaseHandler) PostPurchase(req dto.PurchasePostRequest) (dto.PurchaseDocumentResponse, error) {
	input, err := parsePurchasePostRequest(req)
	if err != nil {
		return dto.PurchaseDocumentResponse{}, err
	}
	posted, err := h.service.PostPurchase(handlerContext(), input)
	if err != nil {
		return dto.PurchaseDocumentResponse{}, fmt.Errorf("post purchase: %w", err)
	}
	return mapPurchaseDocument(posted), nil
}

func parsePurchasePostRequest(req dto.PurchasePostRequest) (application.PurchasePostInput, error) {
	idempotencyKey, err := domain.NewIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		return application.PurchasePostInput{}, fmt.Errorf("idempotency key: %w", err)
	}
	counterpartyID := domain.None[domain.CounterpartyID]()
	if req.CounterpartyID != nil {
		parsed, err := domain.NewCounterpartyID(*req.CounterpartyID)
		if err != nil {
			return application.PurchasePostInput{}, fmt.Errorf("counterparty id: %w", err)
		}
		counterpartyID = domain.Some(parsed)
	}
	occurredOn, err := domain.ParseBusinessDate(req.OccurredOn)
	if err != nil {
		return application.PurchasePostInput{}, fmt.Errorf("occurred on: %w", err)
	}
	reason, err := optionalPurchaseReason(req.ReasonCode)
	if err != nil {
		return application.PurchasePostInput{}, err
	}
	notes, err := optionalNonEmptyText(req.Notes)
	if err != nil {
		return application.PurchasePostInput{}, fmt.Errorf("notes: %w", err)
	}
	lines := make([]application.PurchaseLineInput, 0, len(req.Lines))
	for index, line := range req.Lines {
		parsed, err := parsePurchaseLineRequest(line)
		if err != nil {
			return application.PurchasePostInput{}, fmt.Errorf("line %d: %w", index+1, err)
		}
		lines = append(lines, parsed)
	}
	return application.PurchasePostInput{
		IdempotencyKey: idempotencyKey,
		CounterpartyID: counterpartyID,
		OccurredOn:     occurredOn,
		Reason:         reason,
		Notes:          notes,
		Lines:          lines,
	}, nil
}

func parsePurchaseLineRequest(req dto.PurchaseLineRequest) (application.PurchaseLineInput, error) {
	itemID, err := domain.NewItemID(req.ItemID)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("item id: %w", err)
	}
	quantity, err := domain.NewPositiveAtomicQuantity(req.QuantityAtomic)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("quantity: %w", err)
	}
	enteredUnit, err := domain.NewUnitCode(req.EnteredUnitCode)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("entered unit: %w", err)
	}
	enteredPackagingName, err := optionalNonEmptyText(req.EnteredPackagingName)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("entered packaging name: %w", err)
	}
	conversion, err := domain.NewUnitConversion(req.ConversionNumeratorAtomic, req.ConversionDenominator)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("conversion: %w", err)
	}
	commercialTotal, err := domain.NewMinorAmount(req.CommercialTotalMinor)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("commercial total: %w", err)
	}
	lotCode, err := optionalNonEmptyText(req.LotCode)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("lot code: %w", err)
	}
	expiresOn, err := optionalBusinessDateFromString(req.ExpiresOn)
	if err != nil {
		return application.PurchaseLineInput{}, fmt.Errorf("expires on: %w", err)
	}
	return application.PurchaseLineInput{
		ItemID:               itemID,
		Quantity:             quantity,
		EnteredUnit:          enteredUnit,
		EnteredPackagingName: enteredPackagingName,
		Conversion:           conversion,
		CommercialTotal:      commercialTotal,
		LotCode:              lotCode,
		ExpiresOn:            expiresOn,
	}, nil
}

func optionalPurchaseReason(value *string) (domain.Option[domain.DocumentReason], error) {
	if value == nil {
		return domain.None[domain.DocumentReason](), nil
	}
	return domain.ParseDocumentReason(domain.DocumentPurchase, *value)
}

func optionalBusinessDateFromString(value *string) (domain.Option[domain.BusinessDate], error) {
	if value == nil {
		return domain.None[domain.BusinessDate](), nil
	}
	parsed, err := domain.ParseBusinessDate(*value)
	if err != nil {
		return domain.None[domain.BusinessDate](), err
	}
	return domain.Some(parsed), nil
}

func mapPurchaseDocument(document application.PurchaseDocument) dto.PurchaseDocumentResponse {
	counterpartyID := optionalCounterpartyIDValue(document.CounterpartyID())
	reason := optionalDocumentReasonValue(document.Reason())
	notes := optionalText(document.Notes())
	lines := document.Lines()
	response := dto.PurchaseDocumentResponse{
		ID:                  document.ID().Int64(),
		IdempotencyKey:      document.IdempotencyKey().String(),
		PostingSequence:     document.PostingSequence().Int64(),
		CounterpartyID:      counterpartyID,
		OccurredOn:          document.OccurredOn().String(),
		PostedAtMs:          document.PostedAt().UnixMilli(),
		CurrencyCode:        document.Currency().Code().String(),
		CurrencyMinorDigits: int64(document.Currency().MinorDigits().Int()),
		ReasonCode:          reason,
		Notes:               notes,
		Lines:               make([]dto.PurchaseLineResponse, 0, len(lines)),
	}
	for _, line := range lines {
		response.Lines = append(response.Lines, mapPurchaseLine(line))
	}
	return response
}

func mapPurchaseLine(line application.PostedPurchaseLine) dto.PurchaseLineResponse {
	return dto.PurchaseLineResponse{
		ID:                        line.ID().Int64(),
		LineOrder:                 line.LineOrder().Int64(),
		ItemID:                    line.ItemID().Int64(),
		QuantityAtomic:            line.Quantity().Int64(),
		EnteredUnitCode:           line.EnteredUnit().String(),
		EnteredPackagingName:      optionalText(line.EnteredPackagingName()),
		ConversionNumeratorAtomic: line.Conversion().NumeratorAtomic(),
		ConversionDenominator:     line.Conversion().Denominator(),
		InventoryValueMicro:       line.InventoryValue().Int64(),
		CommercialTotalMinor:      line.CommercialTotal().Int64(),
		LotID:                     line.LotID().Int64(),
		LotCode:                   optionalText(line.LotCode()),
		OriginatedOn:              line.OriginatedOn().String(),
		ExpiresOn:                 optionalBusinessDateValue(line.ExpiresOn()),
	}
}

func optionalCounterpartyIDValue(value domain.Option[domain.CounterpartyID]) *int64 {
	id, ok := value.Get()
	if !ok {
		return nil
	}
	raw := id.Int64()
	return &raw
}

func optionalDocumentReasonValue(value domain.Option[domain.DocumentReason]) *string {
	reason, ok := value.Get()
	if !ok {
		return nil
	}
	raw := reason.String()
	return &raw
}

func optionalBusinessDateValue(value domain.Option[domain.BusinessDate]) *string {
	date, ok := value.Get()
	if !ok {
		return nil
	}
	raw := date.String()
	return &raw
}
