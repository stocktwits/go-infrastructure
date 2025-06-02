package flat

import (
	"fmt"
)

// Formatter is a function type that formats a DynamicValue.
type Formatter func(*DynamicValue) (*DynamicValue, error)

// NewFormatter creates a Formatter that formats data of type T to type S.
// It takes a function that accepts a value of type T and returns a value of type S and an error.
func NewFormatter[T, S any](f func(T) (S, error)) Formatter {
	return func(dv *DynamicValue) (*DynamicValue, error) {
		if dv == nil || dv.Value() == nil {
			return dv, nil
		}

		if !isValidDataType[T]() || !isValidDataType[S]() {
			return nil, fmt.Errorf("invalid data types: %T to %T", *new(T), *new(S))
		}

		if getDataTypeFromType[T]() != dv.DataType() {
			return nil, fmt.Errorf("formatter function type mismatch with data type")
		}

		newVal, err := f(dv.Value().(T))
		if err != nil {
			return nil, fmt.Errorf("error formatting data: %w", err)
		}

		return newDynamicValue(newVal), nil
	}
}

// NewSafeFormatter creates a Formatter that safely formats data of type T to type S.
// it takes a function that accepts a value of type T and returns a value of type S.
func NewSafeFormatter[T, S any](f func(T) S) Formatter {
	return NewFormatter(func(d T) (S, error) {
		return f(d), nil
	})
}
