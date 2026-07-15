package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
	domainsettings "github.com/jerobas/saas/internal/domain/settings"
	"github.com/jerobas/saas/internal/infrastructure/sqlite/sqlcgen"
)

type UpdateSettingsInput struct {
	BusinessName       domain.DisplayName
	Locale             domain.Locale
	Timezone           domain.BusinessTimezone
	Currency           domain.Currency
	HourlyLaborCost    domain.Option[domain.MinorAmount]
	DefaultGrossMargin domain.Option[domain.BasisPoints]
	ExpectedUpdatedAt  domain.UTCInstant
	UpdatedAt          domain.UTCInstant
}

func (s *Store) GetSettings(ctx context.Context) (domainsettings.Settings, error) {
	row, err := s.queries.GetAppSettings(ctx)
	if err != nil {
		return domainsettings.Settings{}, classifyError("get settings", err)
	}
	value, err := mapSettings(row)
	if err != nil {
		return domainsettings.Settings{}, corruptDataError("map settings", err)
	}
	return value, nil
}

func (s *Store) UpdateSettings(ctx context.Context, input UpdateSettingsInput) (domainsettings.Settings, error) {
	if err := validateVersionAdvance(input.ExpectedUpdatedAt, input.UpdatedAt); err != nil {
		return domainsettings.Settings{}, err
	}

	var updated domainsettings.Settings
	err := s.withWriteQueries(ctx, "update settings", func(queries *sqlcgen.Queries) error {
		currentRow, err := queries.GetAppSettings(ctx)
		if err != nil {
			return err
		}
		current, err := mapSettings(currentRow)
		if err != nil {
			return corruptDataError("map current settings", err)
		}
		if !current.UpdatedAt().Equal(input.ExpectedUpdatedAt) {
			return fmt.Errorf("%w: settings version changed", domain.ErrStale)
		}
		if !sameCurrencySnapshot(input.Currency, current.Currency()) {
			selected, err := domain.NewCurrency(input.Currency.Code().String())
			if err != nil {
				return err
			}
			if selected.MinorDigits().Int() != input.Currency.MinorDigits().Int() {
				return domain.Invalid("currency_minor_digits", domain.ViolationInvariant, "SET-002")
			}
		}

		desired, err := domainsettings.New(domainsettings.Params{
			BusinessName:       input.BusinessName,
			Locale:             input.Locale,
			Timezone:           input.Timezone,
			Currency:           input.Currency,
			HourlyLaborCost:    input.HourlyLaborCost,
			DefaultGrossMargin: input.DefaultGrossMargin,
			CreatedAt:          current.CreatedAt(),
			UpdatedAt:          input.UpdatedAt,
		})
		if err != nil {
			return err
		}

		row, err := queries.UpdateAppSettings(ctx, sqlcgen.UpdateAppSettingsParams{
			BusinessName:                  desired.BusinessName().String(),
			LocaleCode:                    desired.Locale().String(),
			TimezoneName:                  desired.Timezone().Name(),
			CurrencyCode:                  desired.Currency().Code().String(),
			CurrencyMinorDigits:           int64(desired.Currency().MinorDigits().Int()),
			HourlyLaborCostMinor:          nullableMinorAmount(desired.HourlyLaborCost()),
			DefaultGrossMarginBasisPoints: nullableBasisPoints(desired.DefaultGrossMargin()),
			UpdatedAtMs:                   desired.UpdatedAt().UnixMilli(),
			ExpectedUpdatedAtMs:           input.ExpectedUpdatedAt.UnixMilli(),
		})
		if errors.Is(err, sql.ErrNoRows) {
			return classifySettingsMutationMiss(ctx, queries, input.ExpectedUpdatedAt)
		}
		if err != nil {
			return err
		}
		updated, err = mapSettings(row)
		if err != nil {
			return corruptDataError("map updated settings", err)
		}
		return nil
	})
	return updated, err
}

func (s *Store) GetMeasurementUnit(ctx context.Context, code domain.UnitCode) (catalog.MeasurementUnit, error) {
	if code.String() == "" {
		return catalog.MeasurementUnit{}, domain.Invalid("unit_code", domain.ViolationRequired, "")
	}
	row, err := s.queries.GetMeasurementUnit(ctx, code.String())
	if err != nil {
		return catalog.MeasurementUnit{}, classifyError("get measurement unit", err)
	}
	unit, err := mapMeasurementUnit(row)
	if err != nil {
		return catalog.MeasurementUnit{}, corruptDataError("map measurement unit", err)
	}
	return unit, nil
}

func (s *Store) ListMeasurementUnits(ctx context.Context) ([]catalog.MeasurementUnit, error) {
	rows, err := s.queries.ListMeasurementUnits(ctx)
	if err != nil {
		return nil, classifyError("list measurement units", err)
	}
	units := make([]catalog.MeasurementUnit, 0, len(rows))
	for _, row := range rows {
		unit, err := mapMeasurementUnit(row)
		if err != nil {
			return nil, corruptDataError("map measurement unit", err)
		}
		units = append(units, unit)
	}
	return units, nil
}

func mapSettings(row sqlcgen.AppSetting) (domainsettings.Settings, error) {
	businessName, err := domain.NewDisplayName(row.BusinessName)
	if err != nil || businessName.String() != row.BusinessName {
		if err == nil {
			err = domain.Invalid("business_name", domain.ViolationInvariant, "SET-001")
		}
		return domainsettings.Settings{}, err
	}
	locale, err := domain.NewLocale(row.LocaleCode)
	if err != nil || locale.String() != row.LocaleCode {
		if err == nil {
			err = domain.Invalid("locale_code", domain.ViolationInvariant, "SET-002")
		}
		return domainsettings.Settings{}, err
	}
	timezone, err := domain.NewBusinessTimezone(row.TimezoneName)
	if err != nil || timezone.Name() != row.TimezoneName {
		if err == nil {
			err = domain.Invalid("timezone_name", domain.ViolationInvariant, "SET-002")
		}
		return domainsettings.Settings{}, err
	}
	currency, err := domain.RestoreCurrency(row.CurrencyCode, int(row.CurrencyMinorDigits))
	if err != nil {
		return domainsettings.Settings{}, err
	}
	hourlyLaborCost, err := restoreOptionalMinorAmount(row.HourlyLaborCostMinor)
	if err != nil {
		return domainsettings.Settings{}, err
	}
	defaultMargin, err := restoreOptionalBasisPoints(row.DefaultGrossMarginBasisPoints)
	if err != nil {
		return domainsettings.Settings{}, err
	}
	createdAt, err := domain.UTCInstantFromUnixMilli(row.CreatedAtMs)
	if err != nil {
		return domainsettings.Settings{}, err
	}
	updatedAt, err := domain.UTCInstantFromUnixMilli(row.UpdatedAtMs)
	if err != nil {
		return domainsettings.Settings{}, err
	}
	return domainsettings.New(domainsettings.Params{
		BusinessName:       businessName,
		Locale:             locale,
		Timezone:           timezone,
		Currency:           currency,
		HourlyLaborCost:    hourlyLaborCost,
		DefaultGrossMargin: defaultMargin,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	})
}

func mapMeasurementUnit(row sqlcgen.MeasurementUnit) (catalog.MeasurementUnit, error) {
	code, err := domain.NewUnitCode(row.Code)
	if err != nil || code.String() != row.Code {
		if err == nil {
			err = domain.Invalid("unit_code", domain.ViolationInvariant, "UNIT-002")
		}
		return catalog.MeasurementUnit{}, err
	}
	name, err := domain.NewDisplayName(row.Name)
	if err != nil || name.String() != row.Name {
		if err == nil {
			err = domain.Invalid("unit_name", domain.ViolationInvariant, "UNIT-002")
		}
		return catalog.MeasurementUnit{}, err
	}
	symbol, err := domain.NewNonEmptyText(row.Symbol)
	if err != nil || symbol.String() != row.Symbol {
		if err == nil {
			err = domain.Invalid("unit_symbol", domain.ViolationInvariant, "UNIT-002")
		}
		return catalog.MeasurementUnit{}, err
	}
	dimension, err := domain.ParseDimension(row.Dimension)
	if err != nil {
		return catalog.MeasurementUnit{}, err
	}
	conversion, err := domain.NewUnitConversion(row.AtomicNumerator, row.AtomicDenominator)
	if err != nil {
		return catalog.MeasurementUnit{}, err
	}
	if conversion.NumeratorAtomic() != row.AtomicNumerator || conversion.Denominator() != row.AtomicDenominator {
		return catalog.MeasurementUnit{}, domain.Invalid("unit_conversion", domain.ViolationInvariant, "UNIT-003")
	}
	itemBase, err := restoreBoolean("is_item_base", row.IsItemBase)
	if err != nil {
		return catalog.MeasurementUnit{}, err
	}
	seeded, err := restoreBoolean("is_seeded", row.IsSeeded)
	if err != nil {
		return catalog.MeasurementUnit{}, err
	}
	return catalog.NewMeasurementUnit(catalog.MeasurementUnitParams{
		Code: code, Name: name, Symbol: symbol, Dimension: dimension,
		Conversion: conversion, ItemBase: itemBase, Seeded: seeded,
	})
}

func classifySettingsMutationMiss(ctx context.Context, queries *sqlcgen.Queries, expected domain.UTCInstant) error {
	row, err := queries.GetAppSettings(ctx)
	if err != nil {
		return classifyError("reload settings after missed update", err)
	}
	current, err := mapSettings(row)
	if err != nil {
		return corruptDataError("map settings after missed update", err)
	}
	if !current.UpdatedAt().Equal(expected) {
		return fmt.Errorf("%w: settings version changed", domain.ErrStale)
	}
	return fmt.Errorf("%w: settings update matched no row", domain.ErrConflict)
}

func validateVersionAdvance(expected, updated domain.UTCInstant) error {
	if expected.IsZero() {
		return domain.Invalid("expected_updated_at", domain.ViolationRequired, "SET-004")
	}
	if updated.IsZero() || updated.Compare(expected) <= 0 {
		return domain.Invalid("updated_at", domain.ViolationInvariant, "SET-004")
	}
	return nil
}

func sameCurrencySnapshot(left, right domain.Currency) bool {
	return left.Code().String() == right.Code().String() &&
		left.MinorDigits().Int() == right.MinorDigits().Int()
}

func nullableMinorAmount(value domain.Option[domain.MinorAmount]) sql.NullInt64 {
	amount, ok := value.Get()
	return sql.NullInt64{Int64: amount.Int64(), Valid: ok}
}

func nullableBasisPoints(value domain.Option[domain.BasisPoints]) sql.NullInt64 {
	points, ok := value.Get()
	return sql.NullInt64{Int64: points.Int64(), Valid: ok}
}

func restoreOptionalMinorAmount(value sql.NullInt64) (domain.Option[domain.MinorAmount], error) {
	if !value.Valid {
		return domain.None[domain.MinorAmount](), nil
	}
	amount, err := domain.NewMinorAmount(value.Int64)
	if err != nil {
		return domain.None[domain.MinorAmount](), err
	}
	return domain.Some(amount), nil
}

func restoreOptionalBasisPoints(value sql.NullInt64) (domain.Option[domain.BasisPoints], error) {
	if !value.Valid {
		return domain.None[domain.BasisPoints](), nil
	}
	points, err := domain.NewBasisPoints(value.Int64)
	if err != nil {
		return domain.None[domain.BasisPoints](), err
	}
	return domain.Some(points), nil
}

func restoreBoolean(field string, value int64) (bool, error) {
	switch value {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, domain.Invalid(field, domain.ViolationInvariant, "")
	}
}
