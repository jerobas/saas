package domain_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/jerobas/saas/internal/domain"
)

func TestIDsOptionsAndDeterministicValidation(t *testing.T) {
	id, err := domain.NewItemID(42)
	if err != nil || id.Int64() != 42 || id.IsZero() || id.String() != "42" {
		t.Fatalf("unexpected item ID: %#v, %v", id, err)
	}
	if _, err := domain.NewItemID(0); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("zero ID error = %v", err)
	}

	none := domain.None[int64]()
	if _, ok := none.Get(); ok || !none.IsNone() {
		t.Fatal("None reported a value")
	}
	some := domain.Some[int64](0)
	if value, ok := some.Get(); !ok || value != 0 || !some.IsSome() {
		t.Fatal("Some did not preserve a zero value")
	}

	err = domain.NewValidationError(
		domain.Violation{Field: "z", Code: domain.ViolationRequired},
		domain.Violation{Field: "a", Code: domain.ViolationOutOfRange},
		domain.Violation{Field: "a", Code: domain.ViolationInvalidFormat},
	)
	if got, want := err.Error(), "domain validation failed: a:invalid_format, a:out_of_range, z:required"; got != want {
		t.Fatalf("validation error = %q, want %q", got, want)
	}
	var validation *domain.ValidationError
	if !errors.As(err, &validation) || !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("validation classification lost: %v", err)
	}
	violations := validation.Violations()
	violations[0].Field = "mutated"
	if validation.Violations()[0].Field != "a" {
		t.Fatal("Violations exposed mutable internal storage")
	}

	corrupt := domain.Corrupt(err)
	if !errors.Is(corrupt, domain.ErrCorruptData) || !errors.Is(corrupt, domain.ErrValidation) {
		t.Fatalf("corrupt error did not retain both classifications: %v", corrupt)
	}
}

func TestUnicodeIdentityNormalization(t *testing.T) {
	first, err := domain.NewUniqueName(" \u00c9clair  cake ")
	if err != nil {
		t.Fatal(err)
	}
	second, err := domain.NewUniqueName("E\u0301CLAIR  CAKE")
	if err != nil {
		t.Fatal(err)
	}
	if first.Display() != "\u00c9clair  cake" {
		t.Fatalf("display = %q", first.Display())
	}
	if first.Key() != second.Key() {
		t.Fatalf("canonically equivalent keys differ: %q != %q", first.Key(), second.Key())
	}
	sharpS, _ := domain.NewUniqueName("Stra\u00dfe")
	upper, _ := domain.NewUniqueName("STRASSE")
	if sharpS.Key() != upper.Key() {
		t.Fatalf("full fold was not applied: %q != %q", sharpS.Key(), upper.Key())
	}
	display, key, err := domain.NormalizeDisplayAndKey(first.Display())
	if err != nil || display != first.Display() || key != first.Key() {
		t.Fatalf("normalization is not idempotent: %q %q %v", display, key, err)
	}
	if _, err := domain.NewDisplayName(string([]byte{0xff})); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid UTF-8 accepted: %v", err)
	}
	if _, err := domain.RestoreUniqueName("Cake", "wrong"); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("persisted key mismatch not corrupt: %v", err)
	}
}

func TestFractionDecimalAndConversionAreExact(t *testing.T) {
	fraction, err := domain.NewFraction(6, 8)
	if err != nil || fraction.Numerator() != 3 || fraction.Denominator() != 4 {
		t.Fatalf("fraction was not reduced: %v, %v", fraction, err)
	}
	decimal, err := domain.ParseDecimalFraction(" +0012.50 ")
	if err != nil || decimal.Numerator() != 25 || decimal.Denominator() != 2 {
		t.Fatalf("decimal parse = %v, %v", decimal, err)
	}
	for _, raw := range []string{"", "-1", ".5", "5.", "1e2", "1,5", "1.2.3"} {
		if _, err := domain.ParseDecimalFraction(raw); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("ParseDecimalFraction(%q) error = %v", raw, err)
		}
	}
	if _, err := domain.ParseDecimalFraction("9223372036854775808"); !errors.Is(err, domain.ErrOverflow) {
		t.Fatalf("decimal overflow error = %v", err)
	}

	large, _ := domain.NewFraction(math.MaxInt64, 2)
	reciprocal, _ := domain.NewFraction(2, math.MaxInt64)
	product, err := large.Multiply(reciprocal)
	if err != nil || product.String() != "1" {
		t.Fatalf("cross-cancelled multiplication = %v, %v", product, err)
	}

	conversion, err := domain.NewUnitConversion(1000, 1)
	if err != nil {
		t.Fatal(err)
	}
	atomic, err := conversion.ToAtomic(decimal)
	if err != nil || atomic.Int64() != 12500 {
		t.Fatalf("atomic conversion = %v, %v", atomic, err)
	}
	back, err := conversion.FromAtomic(atomic)
	if err != nil || back != decimal {
		t.Fatalf("reverse conversion = %v, %v", back, err)
	}
	inexact, _ := domain.NewUnitConversion(1, 3)
	one, _ := domain.NewFraction(1, 1)
	if _, err := inexact.ToAtomic(one); !errors.Is(err, domain.ErrInexactConversion) {
		t.Fatalf("inexact conversion error = %v", err)
	}
}

func TestCheckedQuantitiesAndMoney(t *testing.T) {
	maximum, _ := domain.NewAtomicQuantity(math.MaxInt64)
	one, _ := domain.NewAtomicQuantity(1)
	if _, err := maximum.Add(one); !errors.Is(err, domain.ErrOverflow) {
		t.Fatalf("quantity overflow = %v", err)
	}
	zero, _ := domain.NewAtomicQuantity(0)
	if _, err := zero.Sub(one); !errors.Is(err, domain.ErrNegativeResult) {
		t.Fatalf("negative quantity = %v", err)
	}

	brl, err := domain.NewCurrency(" brl ")
	if err != nil || brl.Code().String() != "BRL" || brl.MinorDigits().Int() != 2 {
		t.Fatalf("BRL snapshot = %#v, %v", brl, err)
	}
	historical, err := domain.RestoreCurrency("ZZZ", 4)
	if err != nil || historical.Code().String() != "ZZZ" || historical.MinorDigits().Int() != 4 {
		t.Fatalf("shape-only historical snapshot = %#v, %v", historical, err)
	}
	if _, err := domain.RestoreCurrency("brl", 2); !errors.Is(err, domain.ErrCorruptData) {
		t.Fatalf("invalid persisted currency error = %v", err)
	}
	minor, _ := domain.NewMinorAmount(123)
	value, err := minor.ToInventoryValue(brl)
	if err != nil || value.Int64() != 1_230_000 {
		t.Fatalf("minor-to-micro = %d, %v", value.Int64(), err)
	}
	quantity, _ := domain.NewAtomicQuantity(2)
	average, ok := value.Per(quantity)
	if !ok || average.Numerator() != 615_000 || average.Denominator() != 1 {
		t.Fatalf("average = %v, %t", average, ok)
	}
	if _, ok := value.Per(zero); ok {
		t.Fatal("average over zero quantity was reported")
	}
}

func TestDatesInstantsTimezoneAndLocale(t *testing.T) {
	leap, err := domain.ParseBusinessDate("2024-02-29")
	if err != nil || leap.String() != "2024-02-29" {
		t.Fatalf("leap date = %v, %v", leap, err)
	}
	if _, err := domain.ParseBusinessDate("2023-02-29"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid date error = %v", err)
	}
	next, _ := domain.ParseBusinessDate("2024-03-01")
	if !leap.Before(next) || next.Compare(leap) != 1 {
		t.Fatal("business date ordering failed")
	}

	instant, err := domain.NewUTCInstant(time.Date(2026, 7, 14, 12, 0, 0, 123456789, time.FixedZone("test", -3*60*60)))
	if err != nil || instant.Time().Location() != time.UTC || instant.Time().Nanosecond()%int(time.Millisecond) != 0 {
		t.Fatalf("canonical UTC instant = %v, %v", instant.Time(), err)
	}
	epoch, err := domain.UTCInstantFromUnixMilli(0)
	if err != nil || epoch.UnixMilli() != 0 {
		t.Fatalf("epoch = %v, %v", epoch.Time(), err)
	}
	if _, err := domain.UTCInstantFromUnixMilli(-1); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("negative instant error = %v", err)
	}
	timezone, err := domain.NewBusinessTimezone("America/Sao_Paulo")
	if err != nil || timezone.Name() != "America/Sao_Paulo" || timezone.Location() == nil {
		t.Fatalf("timezone = %#v, %v", timezone, err)
	}
	if _, err := domain.NewBusinessTimezone("Local"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Local timezone accepted: %v", err)
	}
	locale, err := domain.NewLocale("pt-BR")
	if err != nil || locale.String() != "pt-BR" {
		t.Fatalf("locale = %s, %v", locale.String(), err)
	}
}

func TestClosedEnumsAndReasonCompatibility(t *testing.T) {
	if _, err := domain.ParseDimension("MASS"); err != nil {
		t.Fatal(err)
	}
	if _, err := domain.ParseDimension("mass"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid dimension error = %v", err)
	}
	reason, err := domain.ParseDocumentReason(domain.DocumentSale, "PROMOTION")
	if err != nil || reason.IsNone() {
		t.Fatalf("sale promotion = %#v, %v", reason, err)
	}
	if _, err := domain.ParseDocumentReason(domain.DocumentProduction, "WASTE"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid production reason error = %v", err)
	}
	if _, err := domain.ParseDocumentReason(domain.DocumentKind("UNKNOWN"), ""); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("invalid kind with blank reason error = %v", err)
	}
	if _, err := domain.ParseArchiveFilter("ALL"); err != nil {
		t.Fatal(err)
	}
}

func TestArchivedTimestampIsTheOptimisticVersion(t *testing.T) {
	created := mustInstantValue(t, 1_000)
	updated := mustInstantValue(t, 2_000)
	laterArchive := mustInstantValue(t, 3_000)
	if err := domain.ValidateTimestampOrder(created, updated, domain.Some(laterArchive)); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("archive after optimistic version error = %v, want ErrValidation", err)
	}
	if err := domain.ValidateTimestampOrder(created, updated, domain.Some(updated)); err != nil {
		t.Fatalf("archive equal to optimistic version: %v", err)
	}
}

func mustInstantValue(t *testing.T, milliseconds int64) domain.UTCInstant {
	t.Helper()
	value, err := domain.UTCInstantFromUnixMilli(milliseconds)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
