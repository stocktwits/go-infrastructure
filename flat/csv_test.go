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

// TestCSVExportSplit tests the CSV export with splitting functionality
func TestCSVExportSplit(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *CSV
		splits   func() []splitWriter
		wants    []string
		wantErrs []bool
	}{
		{
			name: "basic split by age",
			setup: func() *CSV {
				data := newDynamicValue([]map[string]any{
					{"name": "John", "age": float64(30)},
					{"name": "Jane", "age": float64(25)},
					{"name": "Bob", "age": float64(35)},
				})
				return newCsv(data, func(s Source, d Dest) {
					d.Col("name", s.Key("name"))
					d.Col("age", s.Key("age"))
				})
			},
			splits: func() []splitWriter {
				var buf1, buf2 bytes.Buffer
				return []splitWriter{
					Split(&buf1, "age", func(v float64) bool { return v >= 30 }),
					Split(&buf2, "age", func(v float64) bool { return v < 30 }),
				}
			},
			wants: []string{
				"name,age\nJohn,30\nBob,35\n",
				"name,age\nJane,25\n",
			},
			wantErrs: []bool{false, false},
		},
		{
			name: "split with no matches",
			setup: func() *CSV {
				data := newDynamicValue([]map[string]any{
					{"name": "John", "age": float64(30)},
				})
				return newCsv(data, func(s Source, d Dest) {
					d.Col("name", s.Key("name"))
					d.Col("age", s.Key("age"))
				})
			},
			splits: func() []splitWriter {
				var buf bytes.Buffer
				return []splitWriter{
					Split(&buf, "age", func(v float64) bool { return v > 100 }),
				}
			},
			wants:    []string{"name,age\n"},
			wantErrs: []bool{false},
		},
		{
			name: "split with multiple conditions",
			setup: func() *CSV {
				data := newDynamicValue([]map[string]any{
					{"name": "John", "age": float64(30), "city": "NYC"},
					{"name": "Jane", "age": float64(25), "city": "LA"},
					{"name": "Bob", "age": float64(35), "city": "NYC"},
				})
				return newCsv(data, func(s Source, d Dest) {
					d.Col("name", s.Key("name"))
					d.Col("age", s.Key("age"))
					d.Col("city", s.Key("city"))
				})
			},
			splits: func() []splitWriter {
				var buf1, buf2, buf3 bytes.Buffer
				return []splitWriter{
					Split(&buf1, "city", func(v string) bool { return v == "NYC" }),
					Split(&buf2, "age", func(v float64) bool { return v >= 30 }),
					NoSplit(&buf3), // Gets everything
				}
			},
			wants: []string{
				"name,age,city\nJohn,30,NYC\nBob,35,NYC\n",
				"name,age,city\nJohn,30,NYC\nBob,35,NYC\n",
				"name,age,city\nJohn,30,NYC\nJane,25,LA\nBob,35,NYC\n",
			},
			wantErrs: []bool{false, false, false},
		},
		{
			name: "error CSV",
			setup: func() *CSV {
				return newErrorCsv(fmt.Errorf("test error"))
			},
			splits: func() []splitWriter {
				var buf bytes.Buffer
				return []splitWriter{NoSplit(&buf)}
			},
			wants:    []string{""},
			wantErrs: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csv := tt.setup()
			splits := tt.splits()

			err := csv.ExportSplit(splits...)

			// Check if any of the splits produced an error
			if err != nil {
				for _, wantErr := range tt.wantErrs {
					if !wantErr {
						t.Errorf("CSV.ExportSplit() unexpected error = %v", err)
						return
					}
				}
			}

			// Verify output of each split writer
			for i, split := range splits {
				if sw, ok := split.(singleSplitWriter); ok {
					if buf, ok := sw.Writer.(*bytes.Buffer); ok {
						got := buf.String()
						if got != tt.wants[i] {
							t.Errorf("CSV.ExportSplit() split %d = %q, want %q", i, got, tt.wants[i])
						}
					}
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
	if val, exists := r.columns[testHeader]; !exists {
		t.Errorf("Col() didn't add column for header %s", testHeader)
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
	if val, exists := r2.columns[testHeader]; !exists {
		t.Errorf("Col() didn't add value for header %s", testHeader)
	} else {
		if str, err := val.strVal(); err != nil {
			t.Errorf("Col() value error: %v", err)
		} else if str != testValue {
			t.Errorf("Col() value = %s, want %s", str, testValue)
		}
	}
}
