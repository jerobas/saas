package catalog_test

import (
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/domain/catalog"
)

func TestMeasurementUnitAndDimensionCompatibility(t *testing.T) {
	unit, err := catalog.NewMeasurementUnit(catalog.MeasurementUnitParams{
		Code: must(domain.NewUnitCode("g")), Name: must(domain.NewDisplayName("gram")),
		Symbol: must(domain.NewNonEmptyText("g")), Dimension: domain.DimensionMass,
		Conversion: must(domain.NewUnitConversion(1000, 1)), ItemBase: true, Seeded: true,
	})
	if err != nil || unit.Code().String() != "g" || !unit.IsItemBase() || !unit.IsSeeded() {
		t.Fatalf("measurement unit = %#v, %v", unit, err)
	}
	if err := catalog.ValidateCompatibleDimensions(domain.DimensionMass, domain.DimensionVolume); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("dimension mismatch error = %v", err)
	}
}

func TestItemAggregateValidatesAndCopiesPackagings(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1000))
	updated := must(domain.UTCInstantFromUnixMilli(2000))
	itemID := must(domain.NewItemID(1))
	packaging := must(catalog.NewItemPackaging(catalog.ItemPackagingParams{
		ID: must(domain.NewPackagingID(10)), ItemID: itemID,
		Name: must(domain.NewUniqueName("5 kg bag")), EnteredUnit: must(domain.NewUnitCode("kg")),
		Conversion: must(domain.NewUnitConversion(5_000_000, 1)),
		CreatedAt:  created, UpdatedAt: updated, ArchivedAt: domain.None[domain.UTCInstant](),
	}))
	packagings := []catalog.ItemPackaging{packaging}
	item, err := catalog.NewItem(catalog.ItemParams{
		ID: itemID, Name: must(domain.NewUniqueName("Flour")), SKU: domain.Some(must(domain.NewSKU("FL-1"))),
		Description: domain.Some(must(domain.NewNonEmptyText("Wheat flour"))), BaseUnit: must(domain.NewUnitCode("g")),
		Capabilities:    catalog.NewCapabilities(true, false, false),
		ReorderQuantity: domain.Some(must(domain.NewAtomicQuantity(20_000))),
		CreatedAt:       created, UpdatedAt: updated, Packagings: packagings,
	})
	if err != nil {
		t.Fatal(err)
	}
	packagings[0] = catalog.ItemPackaging{}
	if item.Packagings()[0].ID().Int64() != 10 {
		t.Fatal("constructor retained caller-owned packaging slice")
	}
	copyOfPackagings := item.Packagings()
	copyOfPackagings[0] = catalog.ItemPackaging{}
	if item.Packagings()[0].ID().Int64() != 10 {
		t.Fatal("packaging accessor exposed aggregate slice")
	}

	price := domain.Some(must(domain.NewMinorAmount(100)))
	_, err = catalog.NewItem(catalog.ItemParams{
		ID: itemID, Name: must(domain.NewUniqueName("Invalid")), BaseUnit: must(domain.NewUnitCode("g")),
		Capabilities: catalog.NewCapabilities(true, false, false), DefaultSalePrice: price,
		CreatedAt: created, UpdatedAt: updated,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("non-sellable price error = %v", err)
	}

	_, err = catalog.NewItem(catalog.ItemParams{
		ID: itemID, Name: must(domain.NewUniqueName("No capability")), BaseUnit: must(domain.NewUnitCode("g")),
		Capabilities: catalog.NewCapabilities(false, false, false), CreatedAt: created, UpdatedAt: updated,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("active capability error = %v", err)
	}
}

func TestItemRejectsPackagingOwnershipAndDuplicateIdentity(t *testing.T) {
	created := must(domain.UTCInstantFromUnixMilli(1000))
	itemID := must(domain.NewItemID(1))
	otherID := must(domain.NewItemID(2))
	name := must(domain.NewUniqueName("box"))
	conversion := must(domain.NewUnitConversion(12_000, 1))
	first := must(catalog.NewItemPackaging(catalog.ItemPackagingParams{
		ID: must(domain.NewPackagingID(1)), ItemID: itemID, Name: name,
		EnteredUnit: must(domain.NewUnitCode("each")), Conversion: conversion,
		CreatedAt: created, UpdatedAt: created,
	}))
	second := must(catalog.NewItemPackaging(catalog.ItemPackagingParams{
		ID: must(domain.NewPackagingID(2)), ItemID: otherID, Name: name,
		EnteredUnit: must(domain.NewUnitCode("each")), Conversion: conversion,
		CreatedAt: created, UpdatedAt: created,
	}))
	_, err := catalog.NewItem(catalog.ItemParams{
		ID: itemID, Name: must(domain.NewUniqueName("Cookies")), BaseUnit: must(domain.NewUnitCode("each")),
		Capabilities: catalog.NewCapabilities(false, true, true),
		CreatedAt:    created, UpdatedAt: created, Packagings: []catalog.ItemPackaging{first, second},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid packaging aggregate error = %v", err)
	}
	var validation *domain.ValidationError
	if !errors.As(err, &validation) || len(validation.Violations()) < 2 {
		t.Fatalf("expected ownership and duplicate violations: %v", err)
	}
}

func TestItemSummaryDoesNotRequirePackagingAggregate(t *testing.T) {
	instant := must(domain.UTCInstantFromUnixMilli(1000))
	summary, err := catalog.NewItemSummary(catalog.ItemSummaryParams{
		ID: must(domain.NewItemID(3)), Name: must(domain.NewUniqueName("Cake")),
		BaseUnit: must(domain.NewUnitCode("each")), Capabilities: catalog.NewCapabilities(false, true, true),
		CreatedAt: instant, UpdatedAt: instant,
	})
	if err != nil || summary.ID().Int64() != 3 || !summary.Capabilities().Sellable() {
		t.Fatalf("item summary = %#v, %v", summary, err)
	}
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
