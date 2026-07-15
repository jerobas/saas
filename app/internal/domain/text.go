package domain

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var unicodeFold = cases.Fold()

func NormalizeDisplay(raw string) (string, error) {
	if !utf8.ValidString(raw) {
		return "", Invalid("text", ViolationInvalidFormat, "")
	}
	display := norm.NFC.String(strings.TrimSpace(raw))
	if display == "" {
		return "", Invalid("text", ViolationRequired, "")
	}
	return display, nil
}

func NormalizeDisplayAndKey(raw string) (display string, key string, err error) {
	display, err = NormalizeDisplay(raw)
	if err != nil {
		return "", "", err
	}
	key = norm.NFC.String(unicodeFold.String(display))
	return display, key, nil
}

type UniqueName struct {
	display string
	key     string
}

func NewUniqueName(raw string) (UniqueName, error) {
	display, key, err := NormalizeDisplayAndKey(raw)
	if err != nil {
		return UniqueName{}, err
	}
	return UniqueName{display: display, key: key}, nil
}

func RestoreUniqueName(display, key string) (UniqueName, error) {
	name, err := NewUniqueName(display)
	if err != nil {
		return UniqueName{}, Corrupt(err)
	}
	if key != name.key {
		return UniqueName{}, Corrupt(Invalid("normalized_name", ViolationInvariant, "CAT-005"))
	}
	return name, nil
}

func (n UniqueName) Display() string { return n.display }
func (n UniqueName) Key() string     { return n.key }

type SKU struct {
	display string
	key     string
}

func NewSKU(raw string) (SKU, error) {
	display, key, err := NormalizeDisplayAndKey(raw)
	if err != nil {
		return SKU{}, err
	}
	return SKU{display: display, key: key}, nil
}

func RestoreSKU(display, key string) (SKU, error) {
	sku, err := NewSKU(display)
	if err != nil {
		return SKU{}, Corrupt(err)
	}
	if key != sku.key {
		return SKU{}, Corrupt(Invalid("normalized_sku", ViolationInvariant, "CAT-008"))
	}
	return sku, nil
}

func (s SKU) Display() string { return s.display }
func (s SKU) Key() string     { return s.key }

type DisplayName struct{ value string }

func NewDisplayName(raw string) (DisplayName, error) {
	value, err := NormalizeDisplay(raw)
	if err != nil {
		return DisplayName{}, err
	}
	return DisplayName{value: value}, nil
}

func (n DisplayName) String() string { return n.value }

type NonEmptyText struct{ value string }

func NewNonEmptyText(raw string) (NonEmptyText, error) {
	value, err := NormalizeDisplay(raw)
	if err != nil {
		return NonEmptyText{}, err
	}
	return NonEmptyText{value: value}, nil
}

func (t NonEmptyText) String() string { return t.value }

type UnitCode struct{ value string }

func NewUnitCode(raw string) (UnitCode, error) {
	value, err := NormalizeDisplay(raw)
	if err != nil {
		return UnitCode{}, err
	}
	return UnitCode{value: value}, nil
}

func (c UnitCode) String() string { return c.value }

type IdempotencyKey struct{ value string }

func NewIdempotencyKey(raw string) (IdempotencyKey, error) {
	value, err := NormalizeDisplay(raw)
	if err != nil {
		return IdempotencyKey{}, err
	}
	return IdempotencyKey{value: value}, nil
}

func (k IdempotencyKey) String() string { return k.value }
