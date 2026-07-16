package wails

import (
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type ReferenceDataHandler struct {
	service *application.ReferenceDataService
}

func NewReferenceDataHandler(service *application.ReferenceDataService) *ReferenceDataHandler {
	if service == nil {
		panic("reference data handler requires a service")
	}
	return &ReferenceDataHandler{service: service}
}

func (h *ReferenceDataHandler) GetMeasurementUnit(code string) (dto.MeasurementUnitResponse, error) {
	unitCode, err := domain.NewUnitCode(code)
	if err != nil {
		return dto.MeasurementUnitResponse{}, fmt.Errorf("unit code: %w", err)
	}
	unit, err := h.service.GetMeasurementUnit(handlerContext(), unitCode)
	if err != nil {
		return dto.MeasurementUnitResponse{}, fmt.Errorf("get measurement unit: %w", err)
	}
	return mapMeasurementUnit(unit), nil
}

func (h *ReferenceDataHandler) ListMeasurementUnits() ([]dto.MeasurementUnitResponse, error) {
	units, err := h.service.ListMeasurementUnits(handlerContext())
	if err != nil {
		return nil, fmt.Errorf("list measurement units: %w", err)
	}
	response := make([]dto.MeasurementUnitResponse, 0, len(units))
	for _, unit := range units {
		response = append(response, mapMeasurementUnit(unit))
	}
	return response, nil
}

func mapMeasurementUnit(unit catalog.MeasurementUnit) dto.MeasurementUnitResponse {
	return dto.MeasurementUnitResponse{
		Code:            unit.Code().String(),
		Name:            unit.Name().String(),
		Symbol:          unit.Symbol().String(),
		Dimension:       unit.Dimension().String(),
		NumeratorAtomic: unit.Conversion().NumeratorAtomic(),
		Denominator:     unit.Conversion().Denominator(),
		IsItemBase:      unit.IsItemBase(),
		IsSeeded:        unit.IsSeeded(),
	}
}
