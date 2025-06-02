package flat

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewData(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		wantType DataType
	}{
		{
			name:     "object",
			input:    map[string]any{"key": "value"},
			wantType: DataTypeObject,
		},
		{
			name:     "array",
			input:    []any{"value1", "value2"},
			wantType: DataTypeArray,
		},
		{
			name:     "array of objects",
			input:    []map[string]any{{"key": "value"}},
			wantType: DataTypeArrayOfObjects,
		},
		{
			name:     "string",
			input:    "test",
			wantType: DataTypeString,
		},
		{
			name:     "float",
			input:    float64(123.45),
			wantType: DataTypeFloat,
		},
		{
			name:     "int",
			input:    42,
			wantType: DataTypeInt,
		},
		{
			name:     "boolean",
			input:    true,
			wantType: DataTypeBoolean,
		},
		{
			name:     "null",
			input:    nil,
			wantType: DataTypeNull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newDynamicValue(tt.input)
			if data.DataType() != tt.wantType {
				t.Errorf("NewData() type = %v, want %v", data.DataType(), tt.wantType)
			}
		})
	}
}

func TestNewFromJSONReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType DataType
		wantErr  bool
	}{
		{
			name:     "valid object",
			input:    `{"key": "value"}`,
			wantType: DataTypeObject,
			wantErr:  false,
		},
		{
			name:     "valid array",
			input:    `["value1", "value2"]`,
			wantType: DataTypeArray,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			input:    `{invalid}`,
			wantType: DataTypeNull,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			data := ReadJSONFromReader(reader)

			if tt.wantErr && data.err == nil {
				t.Error("NewFromJSONReader() expected error, got nil")
			}
			if !tt.wantErr && data.err != nil {
				t.Errorf("NewFromJSONReader() unexpected error: %v", data.err)
			}
			if data.DataType() != tt.wantType {
				t.Errorf("NewFromJSONReader() type = %v, want %v", data.DataType(), tt.wantType)
			}
		})
	}
}

func TestDataVal(t *testing.T) {
	tests := []struct {
		name    string
		data    *DynamicValue
		want    string
		wantErr bool
	}{
		{
			name:    "string value",
			data:    newDynamicValue("test"),
			want:    "test",
			wantErr: false,
		},
		{
			name:    "float value",
			data:    newDynamicValue(float64(123.45)),
			want:    "123.45",
			wantErr: false,
		},
		{
			name:    "object value",
			data:    newDynamicValue(map[string]any{"key": "value"}),
			want:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name:    "error data",
			data:    errorDynamicValue(fmt.Errorf("test error")),
			want:    errorStrValue,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.data.strVal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Data.val() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Data.val() = %v, want %v", got, tt.want)
			}
		})
	}
}
