package wails

import (
	"context"
	"fmt"

	"github.com/jerobas/saas/internal/application"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/presentation/wails/dto"
)

type SettingsHandler struct {
	service *application.SettingsService
}

func NewSettingsHandler(service *application.SettingsService) *SettingsHandler {
	if service == nil {
		panic("settings handler requires a service")
	}
	return &SettingsHandler{service: service}
}

func (h *SettingsHandler) GetSettings(ctx context.Context) (dto.SettingsResponse, error) {
	settingsValue, err := h.service.GetSettings(ctx)
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("get settings: %w", err)
	}
	return mapSettings(settingsValue), nil
}

func (h *SettingsHandler) UpdateSettings(ctx context.Context, req dto.SettingsUpdateRequest) (dto.SettingsResponse, error) {
	businessName, err := domain.NewDisplayName(req.BusinessName)
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("business name: %w", err)
	}
	locale, err := domain.NewLocale(req.Locale)
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("locale: %w", err)
	}
	timezone, err := domain.NewBusinessTimezone(req.Timezone)
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("timezone: %w", err)
	}
	currency, err := domain.RestoreCurrency(req.CurrencyCode, int(req.CurrencyMinorDigits))
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("currency: %w", err)
	}

	var hourlyLaborCost domain.Option[domain.MinorAmount]
	if req.HourlyLaborCost != nil {
		minor, err := domain.NewMinorAmount(*req.HourlyLaborCost)
		if err != nil {
			return dto.SettingsResponse{}, fmt.Errorf("hourly labor cost: %w", err)
		}
		hourlyLaborCost = domain.Some(minor)
	} else {
		hourlyLaborCost = domain.None[domain.MinorAmount]()
	}

	var defaultGrossMargin domain.Option[domain.BasisPoints]
	if req.DefaultGrossMargin != nil {
		basisPoints, err := domain.NewBasisPoints(*req.DefaultGrossMargin)
		if err != nil {
			return dto.SettingsResponse{}, fmt.Errorf("default gross margin: %w", err)
		}
		defaultGrossMargin = domain.Some(basisPoints)
	} else {
		defaultGrossMargin = domain.None[domain.BasisPoints]()
	}

	expectedUpdatedAt, err := domain.UTCInstantFromUnixMilli(req.ExpectedUpdatedAtMs)
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("expected updated at: %w", err)
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(req.UpdatedAtMs)
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("updated at: %w", err)
	}

	updatedSettings, err := h.service.UpdateSettings(ctx, application.SettingsUpdateInput{
		BusinessName:       businessName,
		Locale:             locale,
		Timezone:           timezone,
		Currency:           currency,
		HourlyLaborCost:    hourlyLaborCost,
		DefaultGrossMargin: defaultGrossMargin,
		ExpectedUpdatedAt:  expectedUpdatedAt,
		UpdatedAt:          updatedAt,
	})
	if err != nil {
		return dto.SettingsResponse{}, fmt.Errorf("update settings: %w", err)
	}
	return mapSettings(updatedSettings), nil
}

func mapSettings(settingsValue interface {
	BusinessName() domain.DisplayName
	Locale() domain.Locale
	Timezone() domain.BusinessTimezone
	Currency() domain.Currency
	HourlyLaborCost() domain.Option[domain.MinorAmount]
	DefaultGrossMargin() domain.Option[domain.BasisPoints]
	CreatedAt() domain.UTCInstant
	UpdatedAt() domain.UTCInstant
}) dto.SettingsResponse {
	hourlyLaborCost := optionalMinorAmount(settingsValue.HourlyLaborCost())
	defaultGrossMargin := optionalBasisPoints(settingsValue.DefaultGrossMargin())
	return dto.SettingsResponse{
		BusinessName:        settingsValue.BusinessName().String(),
		Locale:              settingsValue.Locale().String(),
		Timezone:            settingsValue.Timezone().Name(),
		CurrencyCode:        settingsValue.Currency().Code().String(),
		CurrencyMinorDigits: int64(settingsValue.Currency().MinorDigits().Int()),
		HourlyLaborCost:     hourlyLaborCost,
		DefaultGrossMargin:  defaultGrossMargin,
		CreatedAtMs:         settingsValue.CreatedAt().UnixMilli(),
		UpdatedAtMs:         settingsValue.UpdatedAt().UnixMilli(),
	}
}

func optionalMinorAmount(value domain.Option[domain.MinorAmount]) *int64 {
	minor, ok := value.Get()
	if !ok {
		return nil
	}
	raw := minor.Int64()
	return &raw
}

func optionalBasisPoints(value domain.Option[domain.BasisPoints]) *int64 {
	points, ok := value.Get()
	if !ok {
		return nil
	}
	raw := points.Int64()
	return &raw
}
