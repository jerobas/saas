package domain

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// Fraction is a reduced, nonnegative rational number backed exclusively by
// checked int64 arithmetic.
type Fraction struct {
	numerator   int64
	denominator int64
}

func NewFraction(numerator, denominator int64) (Fraction, error) {
	if numerator < 0 {
		return Fraction{}, Invalid("numerator", ViolationOutOfRange, "UNIT-001")
	}
	if denominator <= 0 {
		return Fraction{}, Invalid("denominator", ViolationNotPositive, "UNIT-003")
	}
	if numerator == 0 {
		return Fraction{denominator: 1}, nil
	}
	gcd := greatestCommonDivisor(numerator, denominator)
	return Fraction{numerator: numerator / gcd, denominator: denominator / gcd}, nil
}

func (f Fraction) Numerator() int64   { return f.numerator }
func (f Fraction) Denominator() int64 { return f.denominator }
func (f Fraction) IsZero() bool       { return f.numerator == 0 && f.denominator != 0 }
func (f Fraction) IsValid() bool      { return f.numerator >= 0 && f.denominator > 0 }
func (f Fraction) String() string {
	if f.denominator == 1 {
		return strconv.FormatInt(f.numerator, 10)
	}
	return strconv.FormatInt(f.numerator, 10) + "/" + strconv.FormatInt(f.denominator, 10)
}

func (f Fraction) Multiply(other Fraction) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvariant
	}
	a, b := f.numerator, f.denominator
	c, d := other.numerator, other.denominator
	g1 := greatestCommonDivisor(a, d)
	a, d = a/g1, d/g1
	g2 := greatestCommonDivisor(c, b)
	c, b = c/g2, b/g2
	numerator, err := multiplyNonnegative(a, c)
	if err != nil {
		return Fraction{}, err
	}
	denominator, err := multiplyPositive(b, d)
	if err != nil {
		return Fraction{}, err
	}
	return NewFraction(numerator, denominator)
}

func (f Fraction) Divide(other Fraction) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvariant
	}
	if other.numerator == 0 {
		return Fraction{}, Invalid("divisor", ViolationNotPositive, "UNIT-003")
	}
	a, b := f.numerator, f.denominator
	c, d := other.numerator, other.denominator
	g1 := greatestCommonDivisor(a, c)
	a, c = a/g1, c/g1
	g2 := greatestCommonDivisor(d, b)
	d, b = d/g2, b/g2
	numerator, err := multiplyNonnegative(a, d)
	if err != nil {
		return Fraction{}, err
	}
	denominator, err := multiplyPositive(b, c)
	if err != nil {
		return Fraction{}, err
	}
	return NewFraction(numerator, denominator)
}

func (f Fraction) Int64Exact() (int64, error) {
	if !f.IsValid() {
		return 0, ErrInvariant
	}
	if f.denominator != 1 {
		return 0, ErrInexactConversion
	}
	return f.numerator, nil
}

// ParseDecimalFraction accepts a plain, locale-independent decimal using a
// dot separator. Signs other than an optional leading plus, exponents,
// commas, and incomplete decimal forms are rejected.
func ParseDecimalFraction(raw string) (Fraction, error) {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "+")
	if value == "" {
		return Fraction{}, Invalid("quantity", ViolationRequired, "UNIT-001")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		return Fraction{}, Invalid("quantity", ViolationInvalidFormat, "UNIT-001")
	}
	digits := parts[0]
	decimalPlaces := 0
	if len(parts) == 2 {
		digits += parts[1]
		decimalPlaces = len(parts[1])
	}
	numerator, err := parseDecimalDigits(digits)
	if err != nil {
		if errors.Is(err, ErrOverflow) {
			return Fraction{}, err
		}
		return Fraction{}, Invalid("quantity", ViolationInvalidFormat, "UNIT-001")
	}
	denominator := int64(1)
	for range decimalPlaces {
		if denominator > math.MaxInt64/10 {
			return Fraction{}, ErrOverflow
		}
		denominator *= 10
	}
	return NewFraction(numerator, denominator)
}

type UnitConversion struct{ fraction Fraction }

func NewUnitConversion(numeratorAtomic, denominator int64) (UnitConversion, error) {
	if numeratorAtomic <= 0 {
		return UnitConversion{}, Invalid("conversion_numerator_atomic", ViolationNotPositive, "UNIT-003")
	}
	if denominator <= 0 {
		return UnitConversion{}, Invalid("conversion_denominator", ViolationNotPositive, "UNIT-003")
	}
	fraction, err := NewFraction(numeratorAtomic, denominator)
	if err != nil {
		return UnitConversion{}, err
	}
	return UnitConversion{fraction: fraction}, nil
}

func (c UnitConversion) NumeratorAtomic() int64 { return c.fraction.Numerator() }
func (c UnitConversion) Denominator() int64     { return c.fraction.Denominator() }
func (c UnitConversion) IsZero() bool           { return !c.fraction.IsValid() }

func (c UnitConversion) ToAtomic(entered Fraction) (AtomicQuantity, error) {
	if c.IsZero() {
		return AtomicQuantity{}, ErrInvariant
	}
	result, err := entered.Multiply(c.fraction)
	if err != nil {
		return AtomicQuantity{}, err
	}
	value, err := result.Int64Exact()
	if err != nil {
		return AtomicQuantity{}, err
	}
	return NewAtomicQuantity(value)
}

func (c UnitConversion) FromAtomic(quantity AtomicQuantity) (Fraction, error) {
	if c.IsZero() {
		return Fraction{}, ErrInvariant
	}
	atomic, _ := NewFraction(quantity.Int64(), 1)
	return atomic.Divide(c.fraction)
}

func greatestCommonDivisor(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

func multiplyNonnegative(a, b int64) (int64, error) {
	if a < 0 || b < 0 {
		return 0, ErrInvariant
	}
	if a != 0 && b > math.MaxInt64/a {
		return 0, ErrOverflow
	}
	return a * b, nil
}

func multiplyPositive(a, b int64) (int64, error) {
	if a <= 0 || b <= 0 {
		return 0, ErrInvariant
	}
	return multiplyNonnegative(a, b)
}

func parseDecimalDigits(value string) (int64, error) {
	var result int64
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, ErrValidation
		}
		value := int64(digit - '0')
		if result > (math.MaxInt64-value)/10 {
			return 0, ErrOverflow
		}
		result = result*10 + value
	}
	return result, nil
}
