package flat

import (
	"encoding/json"
	"fmt"
	"io"
)

type DataType int

const (
	DataTypeObject DataType = iota
	DataTypeArray
	DataTypeArrayOfObjects
	DataTypeStreamOfObjects
	DataTypeString
	DataTypeFloat
	DataTypeInt
	DataTypeBoolean
	DataTypeNull
)

const errorStrValue = "<ERROR>"

type DynamicValue struct {
	dataType DataType
	value    any
	err      error
}

var DynamicValueNull = &DynamicValue{dataType: DataTypeNull, value: nil}

func getDataTypeFromValue(v any) DataType {
	switch v.(type) {
	case map[string]any:
		return DataTypeObject
	case []any:
		return DataTypeArray
	case []map[string]any:
		return DataTypeArrayOfObjects
	case string:
		return DataTypeString
	case float64:
		return DataTypeFloat
	case int:
		return DataTypeInt
	case bool:
		return DataTypeBoolean
	case io.Reader:
		return DataTypeStreamOfObjects
	default:
		return DataTypeNull
	}
}

func newDynamicValue(d any) *DynamicValue {
	return &DynamicValue{dataType: getDataTypeFromValue(d), value: d}
}

func isValidDataType[T any]() bool {
	return getDataTypeFromType[T]() != DataTypeNull
}

func getDataTypeFromType[T any]() DataType {
	return getDataTypeFromValue(*new(T))
}

func errorDynamicValue(err error) *DynamicValue {
	return &DynamicValue{
		dataType: DataTypeNull,
		value:    nil,
		err:      err,
	}
}

// ReadJSONFromReader creates a new DynamicValue instance from a io.Reader containing JSON data.
// It expect the that IO.Reader contains a single JSON object, array or array of objects.
func ReadJSONFromReader(r io.Reader) *DynamicValue {
	var data any
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&data); err != nil {
		return errorDynamicValue(fmt.Errorf("failed to decode JSON: %w", err))
	}
	return newDynamicValue(data)
}

// StreamJSONFromReader creates a new DynamicValue instance from a io.Reader containing a stream of JSON objects.
// It expects the IO.Reader to contain a stream of JSON objects, where each object is separated by a newline.
// This is useful for processing large datasets where each line is a separate JSON object.
func StreamJSONFromReader(r io.Reader) *DynamicValue {
	return newDynamicValue(r)
}

// DataType returns the type of data contained in the Data instance.
func (d *DynamicValue) DataType() DataType {
	return d.dataType
}

// Value returns the underlying data of the Value instance.
func (d *DynamicValue) Value() any {
	return d.value
}

// Error returns any error associated with the Data instance.
func (d *DynamicValue) Error() error {
	return d.err
}

// strVal returns the string representation of the data based on its type.
// If the data type is not supported or an error occurs, it returns an error.
func (d *DynamicValue) strVal() (string, error) {
	if d.err != nil {
		return errorStrValue, fmt.Errorf("data contains error: %w", d.err)
	}

	switch d.dataType {
	case DataTypeObject, DataTypeArray, DataTypeArrayOfObjects:
		strObj, err := json.Marshal(d.value)
		if err != nil {
			return errorStrValue, fmt.Errorf("failed to marshal data: %w", err)
		}
		return string(strObj), nil
	case DataTypeString:
		str := d.value.(string)
		return str, nil
	case DataTypeFloat:
		num := d.value.(float64)
		return fmt.Sprintf("%g", num), nil
	case DataTypeInt:
		num := d.value.(int)
		return fmt.Sprintf("%d", num), nil
	case DataTypeBoolean:
		boolean := d.value.(bool)
		return fmt.Sprintf("%t", boolean), nil
	case DataTypeStreamOfObjects:
		return errorStrValue, fmt.Errorf("data is a stream of objects, cannot convert to string")
	case DataTypeNull:
		return "", nil
	default:
		return errorStrValue, fmt.Errorf("unknown data type: %v", d.dataType)
	}
}

// rootKey retrieves a value from a Data instances holding a object.
// If the data is not an object or the key does not exist, it returns NullData.
func (d *DynamicValue) rootKey(key string) *DynamicValue {
	// Return null if data is not an object
	if d.dataType != DataTypeObject {
		return DynamicValueNull
	}

	if obj, ok := d.value.(map[string]any); ok {
		if value, exists := obj[key]; exists {
			return newDynamicValue(value)
		}
	}

	// Return null if key does not exist
	return DynamicValueNull
}

// Key retrieves a value from a Data instance using a sequence of keys.
// If no keys are provided, it returns NullData.
func (d *DynamicValue) Key(keys ...string) *DynamicValue {
	// Return null if no keys provided
	if len(keys) == 0 {
		return DynamicValueNull
	}

	// If only one key, return the value for that key
	if len(keys) == 1 {
		return d.rootKey(keys[0])
	}

	// Recursively get the value for the next key
	return d.rootKey(keys[0]).Key(keys[1:]...)
}

// Format applies a transformation function to the Data instance.
// If the function is nil, it returns the original Data instance.
func (d *DynamicValue) Format(formatterFunc Formatter) *DynamicValue {
	if formatterFunc == nil {
		return d
	}

	data, err := formatterFunc(d)
	if err != nil {
		return errorDynamicValue(fmt.Errorf("error transforming data: %w", err))
	}

	return data
}

// Idx retrieves an element from a Data instance that holds an array or an array of objects.
// If the index is out of bounds or the data type is not an array, it returns NullData.
func (d *DynamicValue) Idx(index int) *DynamicValue {
	if d.dataType != DataTypeArray && d.dataType != DataTypeArrayOfObjects {
		return DynamicValueNull // Return null if not an array
	}

	if arr, ok := d.value.([]any); ok {
		if index >= 0 && index < len(arr) {
			return newDynamicValue(arr[index])
		}
	}

	if arr, ok := d.value.([]map[string]any); ok {
		if index >= 0 && index < len(arr) {
			return newDynamicValue(arr[index])
		}
	}

	return DynamicValueNull
}

// GetCSV generates a CSV representation of the DynamicValue instance.
// It applies the provided mapper function to each item in the DynamicValue instance.
// The mapper function takes a Source and a Dest as arguments, allowing it to write data to the CSV.
// If the Data instance contains an error or is not an array or array of objects, it returns a CSV with an error.
func (d *DynamicValue) GetCSV(f flattener) *CSV {
	return newCsv(d, f)
}
