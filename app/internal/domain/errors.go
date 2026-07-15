package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrValidation        = errors.New("domain validation failed")
	ErrOverflow          = errors.New("checked integer overflow")
	ErrNegativeResult    = errors.New("operation would produce a negative value")
	ErrInexactConversion = errors.New("quantity conversion is not exact")
	ErrInvariant         = errors.New("domain invariant violated")
	ErrNotFound          = errors.New("domain object not found")
	ErrConflict          = errors.New("domain conflict")
	ErrStale             = errors.New("stale domain snapshot")
	ErrInvalidReference  = errors.New("invalid domain reference")
	ErrBusy              = errors.New("domain write is busy")
	ErrCorruptData       = errors.New("persisted data violates the domain contract")
)

const (
	ViolationRequired              = "required"
	ViolationInvalidFormat         = "invalid_format"
	ViolationInvalidEnum           = "invalid_enum"
	ViolationOutOfRange            = "out_of_range"
	ViolationNotPositive           = "not_positive"
	ViolationIncompatibleDimension = "incompatible_dimension"
	ViolationDuplicate             = "duplicate"
	ViolationInvariant             = "invariant"
)

// Violation is stable, machine-readable validation metadata. Human-facing
// copy belongs to a presentation layer.
type Violation struct {
	Field       string
	Code        string
	InvariantID string
}

// ValidationError holds a deterministic set of validation failures.
type ValidationError struct {
	violations []Violation
}

// CorruptDataError lets persistence adapters preserve the precise validation
// cause while classifying an invalid loaded snapshot as ErrCorruptData.
type CorruptDataError struct{ cause error }

func Corrupt(cause error) error {
	if cause == nil || errors.Is(cause, ErrCorruptData) {
		return cause
	}
	return &CorruptDataError{cause: cause}
}

func (e *CorruptDataError) Error() string {
	return fmt.Sprintf("%s: %v", ErrCorruptData, e.cause)
}

func (e *CorruptDataError) Unwrap() []error {
	return []error{ErrCorruptData, e.cause}
}

func NewValidationError(violations ...Violation) error {
	if len(violations) == 0 {
		return nil
	}
	items := append([]Violation(nil), violations...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Field != items[j].Field {
			return items[i].Field < items[j].Field
		}
		if items[i].Code != items[j].Code {
			return items[i].Code < items[j].Code
		}
		return items[i].InvariantID < items[j].InvariantID
	})
	return &ValidationError{violations: items}
}

func Invalid(field, code, invariantID string) error {
	return NewValidationError(Violation{Field: field, Code: code, InvariantID: invariantID})
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.violations))
	for _, violation := range e.violations {
		part := violation.Field + ":" + violation.Code
		if violation.InvariantID != "" {
			part += "[" + violation.InvariantID + "]"
		}
		parts = append(parts, part)
	}
	return fmt.Sprintf("%s: %s", ErrValidation, strings.Join(parts, ", "))
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

func (e *ValidationError) Violations() []Violation {
	return append([]Violation(nil), e.violations...)
}
