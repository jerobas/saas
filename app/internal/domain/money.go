package domain

import (
	"math"
	"strings"

	"golang.org/x/text/currency"
)

const InventoryValueScale int64 = 1_000_000

type CurrencyCode struct{ value string }

func (c CurrencyCode) String() string { return c.value }

// Currency is the persisted currency snapshot: ISO code plus the minor-unit
// scale in force when the value was recorded.
type Currency struct {
	code        CurrencyCode
	minorDigits MinorDigits
}

// NewCurrency validates a new user selection against the x/text ISO registry.
func NewCurrency(raw string) (Currency, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if err := validateCurrencyShape(code); err != nil {
		return Currency{}, err
	}
	unit, err := currency.ParseISO(code)
	if err != nil || unit == currency.XXX {
		return Currency{}, Invalid("currency_code", ViolationInvalidFormat, "SET-002")
	}
	digits, _ := currency.Standard.Rounding(unit)
	minorDigits, err := NewMinorDigits(digits)
	if err != nil {
		return Currency{}, err
	}
	return Currency{code: CurrencyCode{value: code}, minorDigits: minorDigits}, nil
}

// RestoreCurrency validates only the stable persisted representation. This is
// intentionally independent from a future x/text registry so historical
// currency snapshots remain readable.
func RestoreCurrency(code string, minorDigits int) (Currency, error) {
	if err := validateCurrencyShape(code); err != nil {
		return Currency{}, Corrupt(err)
	}
	digits, err := NewMinorDigits(minorDigits)
	if err != nil {
		return Currency{}, Corrupt(err)
	}
	return Currency{code: CurrencyCode{value: code}, minorDigits: digits}, nil
}

func (c Currency) Code() CurrencyCode       { return c.code }
func (c Currency) MinorDigits() MinorDigits { return c.minorDigits }
func (c Currency) IsZero() bool             { return c.code.value == "" }

func validateCurrencyShape(code string) error {
	if len(code) != 3 {
		return Invalid("currency_code", ViolationInvalidFormat, "SET-002")
	}
	for _, character := range code {
		if character < 'A' || character > 'Z' {
			return Invalid("currency_code", ViolationInvalidFormat, "SET-002")
		}
	}
	return nil
}

type MinorAmount struct{ value int64 }

func NewMinorAmount(value int64) (MinorAmount, error) {
	if value < 0 {
		return MinorAmount{}, Invalid("amount_minor", ViolationOutOfRange, "INV-001")
	}
	return MinorAmount{value: value}, nil
}

func (m MinorAmount) Int64() int64 { return m.value }
func (m MinorAmount) IsZero() bool { return m.value == 0 }

func (m MinorAmount) Add(other MinorAmount) (MinorAmount, error) {
	if m.value > math.MaxInt64-other.value {
		return MinorAmount{}, ErrOverflow
	}
	return MinorAmount{value: m.value + other.value}, nil
}

func (m MinorAmount) Sub(other MinorAmount) (MinorAmount, error) {
	if other.value > m.value {
		return MinorAmount{}, ErrNegativeResult
	}
	return MinorAmount{value: m.value - other.value}, nil
}

func (m MinorAmount) ToInventoryValue(snapshot Currency) (InventoryValue, error) {
	if snapshot.IsZero() {
		return InventoryValue{}, ErrInvariant
	}
	factor := int64(1)
	for digits := snapshot.MinorDigits().Int(); digits < 6; digits++ {
		if factor > math.MaxInt64/10 {
			return InventoryValue{}, ErrOverflow
		}
		factor *= 10
	}
	value, err := multiplyNonnegative(m.value, factor)
	if err != nil {
		return InventoryValue{}, err
	}
	return InventoryValue{value: value}, nil
}

type InventoryValue struct{ value int64 }

func NewInventoryValue(value int64) (InventoryValue, error) {
	if value < 0 {
		return InventoryValue{}, Invalid("inventory_value_micro", ViolationOutOfRange, "INV-002")
	}
	return InventoryValue{value: value}, nil
}

func (v InventoryValue) Int64() int64 { return v.value }
func (v InventoryValue) IsZero() bool { return v.value == 0 }

func (v InventoryValue) Add(other InventoryValue) (InventoryValue, error) {
	if v.value > math.MaxInt64-other.value {
		return InventoryValue{}, ErrOverflow
	}
	return InventoryValue{value: v.value + other.value}, nil
}

func (v InventoryValue) Sub(other InventoryValue) (InventoryValue, error) {
	if other.value > v.value {
		return InventoryValue{}, ErrNegativeResult
	}
	return InventoryValue{value: v.value - other.value}, nil
}

func (v InventoryValue) Per(quantity AtomicQuantity) (Fraction, bool) {
	if quantity.IsZero() {
		return Fraction{}, false
	}
	fraction, err := NewFraction(v.value, quantity.Int64())
	return fraction, err == nil
}
