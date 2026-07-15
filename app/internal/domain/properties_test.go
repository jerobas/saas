package domain

import (
	"math"
	"strings"
	"testing"
	"testing/quick"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func TestFractionReductionProperty(t *testing.T) {
	property := func(rawNumerator, rawDenominator uint32) bool {
		numerator := int64(rawNumerator % 1_000_000)
		denominator := int64(rawDenominator%999_999) + 1
		fraction, err := NewFraction(numerator, denominator)
		if err != nil || !fraction.IsValid() {
			return false
		}
		if fraction.Numerator()*denominator != numerator*fraction.Denominator() {
			return false
		}
		if numerator == 0 {
			return fraction.Numerator() == 0 && fraction.Denominator() == 1
		}
		return greatestCommonDivisor(fraction.Numerator(), fraction.Denominator()) == 1
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 2_000}); err != nil {
		t.Fatal(err)
	}
}

func TestCheckedQuantityArithmeticProperty(t *testing.T) {
	property := func(leftRaw, rightRaw uint32) bool {
		left, err := NewAtomicQuantity(int64(leftRaw))
		if err != nil {
			return false
		}
		right, err := NewAtomicQuantity(int64(rightRaw))
		if err != nil {
			return false
		}
		sum, err := left.Add(right)
		if err != nil {
			return false
		}
		restored, err := sum.Sub(right)
		return err == nil && restored == left
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 2_000}); err != nil {
		t.Fatal(err)
	}

	maximum, _ := NewAtomicQuantity(math.MaxInt64)
	one, _ := NewAtomicQuantity(1)
	if _, err := maximum.Add(one); err != ErrOverflow {
		t.Fatalf("maximum + one error = %v, want ErrOverflow", err)
	}
}

func FuzzNormalizeDisplayAndKey(f *testing.F) {
	for _, seed := range []string{
		" Café ",
		"CAFE\u0301",
		"Straße",
		"İstanbul",
		"\u2003sugar\u00a0",
		"",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		display, key, err := NormalizeDisplayAndKey(raw)
		if err != nil {
			return
		}
		if display == "" || key == "" || !utf8.ValidString(display) || !utf8.ValidString(key) {
			t.Fatalf("invalid normalized result: display=%q key=%q", display, key)
		}
		if strings.TrimSpace(display) != display || !norm.NFC.IsNormalString(display) || !norm.NFC.IsNormalString(key) {
			t.Fatalf("result is not trimmed NFC: display=%q key=%q", display, key)
		}
		secondDisplay, secondKey, err := NormalizeDisplayAndKey(display)
		if err != nil || secondDisplay != display || secondKey != key {
			t.Fatalf("normalization is not idempotent: (%q,%q) then (%q,%q), err=%v", display, key, secondDisplay, secondKey, err)
		}
	})
}

func FuzzParseDecimalFraction(f *testing.F) {
	for _, seed := range []string{"0", "1", "1.25", "+0.001", "9223372036854775807", "-1", "1e3", "1,5", ""} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		fraction, err := ParseDecimalFraction(raw)
		if err != nil {
			return
		}
		if !fraction.IsValid() || fraction.Numerator() < 0 || fraction.Denominator() <= 0 {
			t.Fatalf("ParseDecimalFraction(%q) returned invalid %v", raw, fraction)
		}
		if fraction.Numerator() != 0 && greatestCommonDivisor(fraction.Numerator(), fraction.Denominator()) != 1 {
			t.Fatalf("ParseDecimalFraction(%q) was not reduced: %v", raw, fraction)
		}
	})
}
