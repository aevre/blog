// Package field
package field

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Optional is a generic optional value that tracks presence via a set flag and an underlying pointer.
type Optional[T any] struct {
	// Set indicates whether the optional value is set.
	Set bool
	// Value is a pointer to the underlying optional value, valid when Set is true.
	Value *T
}

// ValidatorValue returns the underlying value as an any, or nil if the optional is not set.
func (o Optional[T]) ValidatorValue() any {
	if !o.Set || o.Value == nil {
		return nil
	}

	return *o.Value
}

// UnmarshalJSON implements json.Unmarshaler for Optional, decoding JSON data and handling null by setting the value as absent.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = false
	o.Value = nil

	data = bytes.TrimSpace(data)

	if bytes.Equal(data, []byte("null")) {
		o.Set = true
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("decode optional field: %w", err)
	}

	o.Set = true
	o.Value = &value
	return nil
}
