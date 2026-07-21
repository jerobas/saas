package domain

import (
	"math"
	"strconv"
)

// AtomicQuantity stores a canonical quantity in the smallest unit used by the
// inventory ledger. It is deliberately nonnegative; movement direction is a
// separate domain value.
type AtomicQuantity struct{ value int64 }

func NewAtomicQuantity(value int64) (AtomicQuantity, error) {
	if value < 0 {
		return AtomicQuantity{}, Invalid("quantity_atomic", ViolationOutOfRange, "INV-002")
	}
	return AtomicQuantity{value: value}, nil
}

func NewPositiveAtomicQuantity(value int64) (AtomicQuantity, error) {
	if value <= 0 {
		return AtomicQuantity{}, Invalid("quantity_atomic", ViolationNotPositive, "REC-003")
	}
	return AtomicQuantity{value: value}, nil
}

func (q AtomicQuantity) Int64() int64   { return q.value }
func (q AtomicQuantity) IsZero() bool   { return q.value == 0 }
func (q AtomicQuantity) String() string { return strconv.FormatInt(q.value, 10) }

func (q AtomicQuantity) Add(other AtomicQuantity) (AtomicQuantity, error) {
	if q.value > math.MaxInt64-other.value {
		return AtomicQuantity{}, ErrOverflow
	}
	return AtomicQuantity{value: q.value + other.value}, nil
}

func (q AtomicQuantity) Sub(other AtomicQuantity) (AtomicQuantity, error) {
	if other.value > q.value {
		return AtomicQuantity{}, ErrNegativeResult
	}
	return AtomicQuantity{value: q.value - other.value}, nil
}
