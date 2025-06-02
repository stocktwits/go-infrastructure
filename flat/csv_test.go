package flat

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

// TestCSVExport tests the CSV export functionality
func TestCSVExport(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *CSV
		want    string
		wantErr bool
	}{
		{
			name: "simple export",
			setup: func() *CSV {
				data := newDynamicValue(map[string]any{
					"name": "John",
					"age":  float64(30),
				})
				return newCsv(data, func(s Source, d Dest) {
					d.Col("name", s.Key("name"))
					d.Col("age", s.Key("age"))
				})
			},
			want:    "name,age\nJohn,30\n",
			wantErr: false,
		},
		{
			name: "multiple rows",
			setup: func() *CSV {
				data := newDynamicValue([]map[string]any{
					{
						"name": "John",
						"age":  float64(30),
					},
					{
						"name": "Jane",
						"age":  float64(25),
					},
				})
				return newCsv(data, func(s Source, d Dest) {
					d.Col("name", s.Key("name"))
					d.Col("age", s.Key("age"))
				})
			},
			want:    "name,age\nJohn,30\nJane,25\n",
			wantErr: false,
		},
		{
			name: "missing values",
			setup: func() *CSV {
				data := newDynamicValue([]map[string]any{
					{
						"name": "John",
						"age":  float64(30),
					},
					{
						"name": "Jane",
					},
				})
				return newCsv(data, func(s Source, d Dest) {
					d.Col("name", s.Key("name"))
					d.Col("age", s.Key("age"))
				})
			},
			want:    "name,age\nJohn,30\nJane,\n",
			wantErr: false,
		},
		{
			name: "error CSV",
			setup: func() *CSV {
				return newErrorCsv(fmt.Errorf("test error"))
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csv := tt.setup()
			var buf bytes.Buffer
			err := csv.Export(&buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("CSV.Export() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := buf.String()
				if got != tt.want {
					t.Errorf("CSV.Export() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

// TestCSVRow tests the row functionality
func TestCSVRow(t *testing.T) {
	// Create a row with headers
	r := newRow(true)

	// Add a column
	testHeader := "test"
	testValue := "value"
	r.Col(testHeader, Source{data: newDynamicValue(testValue)})

	// Verify headers are tracked
	headers := r.getHeaders()
	if !slices.Contains(headers, testHeader) {
		t.Errorf("Col() didn't add header %s", testHeader)
	}

	// Verify value was added
	if val, exists := r.values[testHeader]; !exists {
		t.Errorf("Col() didn't add value for header %s", testHeader)
	} else {
		if str, err := val.strVal(); err != nil {
			t.Errorf("Col() value error: %v", err)
		} else if str != testValue {
			t.Errorf("Col() value = %s, want %s", str, testValue)
		}
	}

	// Create a row without headers
	r2 := newRow(false)
	r2.Col(testHeader, Source{data: newDynamicValue(testValue)})

	// Verify no headers are tracked
	headers = r2.getHeaders()
	if headers != nil {
		t.Errorf("Row without headers returned headers: %v", headers)
	}

	// Verify value was still added
	if val, exists := r2.values[testHeader]; !exists {
		t.Errorf("Col() didn't add value for header %s", testHeader)
	} else {
		if str, err := val.strVal(); err != nil {
			t.Errorf("Col() value error: %v", err)
		} else if str != testValue {
			t.Errorf("Col() value = %s, want %s", str, testValue)
		}
	}
}
