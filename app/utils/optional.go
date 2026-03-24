package utils

import (
	"bytes"
	"encoding/json"
)

type Optional[T any] struct {
	set   bool // field was present in JSON
	valid bool // field was not null
	value T
}

func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	o.set = true

	// null
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		o.valid = false
		var zero T
		o.value = zero
		return nil
	}

	// value
	o.valid = true
	return json.Unmarshal(b, &o.value)
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
    return nil, fmt.Errorf("Optional is input-only; do not marshal")
}

func (o Optional[T]) IsSet() bool {
	return o.set
}

func (o Optional[T]) IsValid() bool {
	return o.valid
}

func (o Optional[T]) GetValue() bool {
	return o.value
}

// func (o Optional[T]) MarshalJSON() ([]byte, error) {
// 	// If you ever send Optional back to frontend:
// 	// - not set: omit via `omitempty` on the struct field (see below)
// 	// - set but invalid: return null
// 	if !o.Set {
// 		// This won't be used if you add `omitempty` and keep Set=false,
// 		// but returning null is a safe default.
// 		return []byte("null"), nil
// 	}
// 	if !o.Valid {
// 		return []byte("null"), nil
// 	}
// 	return json.Marshal(o.Value)
// }