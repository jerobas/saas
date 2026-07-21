package application

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

type ReferenceDataStore interface {
	GetMeasurementUnit(ctx context.Context, code domain.UnitCode) (catalog.MeasurementUnit, error)
	ListMeasurementUnits(ctx context.Context) ([]catalog.MeasurementUnit, error)
}

type ReferenceDataService struct {
	store ReferenceDataStore
}

func NewReferenceDataService(store ReferenceDataStore) *ReferenceDataService {
	if store == nil {
		panic("reference data service requires a store")
	}
	return &ReferenceDataService{store: store}
}

func (s *ReferenceDataService) GetMeasurementUnit(ctx context.Context, code domain.UnitCode) (catalog.MeasurementUnit, error) {
	unit, err := s.store.GetMeasurementUnit(ctx, code)
	if err != nil {
		return catalog.MeasurementUnit{}, fmt.Errorf("get measurement unit: %w", err)
	}
	return unit, nil
}

func (s *ReferenceDataService) ListMeasurementUnits(ctx context.Context) ([]catalog.MeasurementUnit, error) {
	units, err := s.store.ListMeasurementUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("list measurement units: %w", err)
	}
	return units, nil
}
