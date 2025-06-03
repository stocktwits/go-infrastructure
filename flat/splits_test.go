package flat

import (
	"bytes"
	"strings"
	"testing"
)

func TestSingleSplitter(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		includeFunc func(string) bool
		inputHeader string
		inputValue  any
		wantInclude bool
		wantErr     bool
	}{
		{
			name:        "match header and include",
			header:      "test",
			includeFunc: func(v string) bool { return true },
			inputHeader: "test",
			inputValue:  "value",
			wantInclude: true,
			wantErr:     false,
		},
		{
			name:        "match header and exclude",
			header:      "test",
			includeFunc: func(v string) bool { return false },
			inputHeader: "test",
			inputValue:  "value",
			wantInclude: false,
			wantErr:     false,
		},
		{
			name:        "different header always includes",
			header:      "test",
			includeFunc: func(v string) bool { return false },
			inputHeader: "other",
			inputValue:  "value",
			wantInclude: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSplitter(tt.header, tt.includeFunc)
			include, err := s.shouldInclude(tt.inputHeader, newDynamicValue(tt.inputValue))

			if (err != nil) != tt.wantErr {
				t.Errorf("shouldInclude() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if include != tt.wantInclude {
				t.Errorf("shouldInclude() = %v, want %v", include, tt.wantInclude)
			}
		})
	}

	t.Run("type mismatch causes error", func(t *testing.T) {
		s := NewSplitter("test", func(v string) bool {
			return v == "value"
		})
		_, err := s.shouldInclude("test", newDynamicValue(42)) // int when string expected
		if err == nil {
			t.Error("shouldInclude() expected error for type mismatch, got nil")
		} else {
			if !strings.Contains(err.Error(), "split function type mismatch with data type") {
				t.Errorf("shouldInclude() unexpected error = %v", err)
			}
		}
	})
}

func TestNoSplit(t *testing.T) {
	var buf bytes.Buffer
	sw := NoSplit(&buf)

	// Test that NoSplit always includes data
	include, err := sw.shouldInclude("any_header", newDynamicValue("any_value"))
	if err != nil {
		t.Errorf("NoSplit.shouldInclude() unexpected error = %v", err)
	}
	if !include {
		t.Error("NoSplit.shouldInclude() = false, want true")
	}

	// Test that it correctly implements io.Writer
	testData := "test data"
	n, err := sw.Write([]byte(testData))
	if err != nil {
		t.Errorf("NoSplit.Write() unexpected error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("NoSplit.Write() wrote %d bytes, want %d bytes", n, len(testData))
	}
	if buf.String() != testData {
		t.Errorf("NoSplit.Write() content = %q, want %q", buf.String(), testData)
	}
}

func TestSplit(t *testing.T) {
	// Test type-safe splitting with a string value
	var buf bytes.Buffer
	sw := Split(&buf, "name", func(s string) bool {
		return s == "John"
	})

	include, err := sw.shouldInclude("name", newDynamicValue("John"))
	if err != nil {
		t.Errorf("Split.shouldInclude() unexpected error = %v", err)
	}
	if !include {
		t.Error("Split.shouldInclude() = false, want true for matching value")
	}

	include, err = sw.shouldInclude("name", newDynamicValue("Jane"))
	if err != nil {
		t.Errorf("Split.shouldInclude() unexpected error = %v", err)
	}
	if include {
		t.Error("Split.shouldInclude() = true, want false for non-matching value")
	}

	// Test that non-matching headers are always included
	include, err = sw.shouldInclude("age", newDynamicValue(30))
	if err != nil {
		t.Errorf("Split.shouldInclude() unexpected error = %v", err)
	}
	if !include {
		t.Error("Split.shouldInclude() = false, want true for non-matching header")
	}
}

func TestSplitAnd(t *testing.T) {
	tests := []struct {
		name      string
		splitters []splitter
		header    string
		value     any
		want      bool
		wantErr   bool
	}{
		{
			name:      "no splitters",
			splitters: []splitter{},
			header:    "test",
			value:     "value",
			want:      true,
			wantErr:   false,
		},
		{
			name: "all splitters return true",
			splitters: []splitter{
				NewSplitter("test", func(v string) bool { return true }),
				NewSplitter("test", func(v string) bool { return true }),
			},
			header:  "test",
			value:   "value",
			want:    true,
			wantErr: false,
		},
		{
			name: "one splitter returns false",
			splitters: []splitter{
				NewSplitter("test", func(v string) bool { return true }),
				NewSplitter("test", func(v string) bool { return false }),
			},
			header:  "test",
			value:   "value",
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			sw := SplitAnd(&buf, tt.splitters...)
			got, err := sw.shouldInclude(tt.header, newDynamicValue(tt.value))

			if (err != nil) != tt.wantErr {
				t.Errorf("SplitAnd.shouldInclude() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SplitAnd.shouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitOr(t *testing.T) {
	tests := []struct {
		name      string
		splitters []splitter
		header    string
		value     any
		want      bool
		wantErr   bool
	}{
		{
			name:      "no splitters",
			splitters: []splitter{},
			header:    "test",
			value:     "value",
			want:      true,
			wantErr:   false,
		},
		{
			name: "any splitter returns true",
			splitters: []splitter{
				NewSplitter("test", func(v string) bool { return false }),
				NewSplitter("test", func(v string) bool { return true }),
			},
			header:  "test",
			value:   "value",
			want:    true,
			wantErr: false,
		},
		{
			name: "all splitters return false",
			splitters: []splitter{
				NewSplitter("test", func(v string) bool { return false }),
				NewSplitter("test", func(v string) bool { return false }),
			},
			header:  "test",
			value:   "value",
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			sw := SplitOr(&buf, tt.splitters...)
			got, err := sw.shouldInclude(tt.header, newDynamicValue(tt.value))

			if (err != nil) != tt.wantErr {
				t.Errorf("SplitOr.shouldInclude() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SplitOr.shouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntegration_CSVSplitting(t *testing.T) {
	// Create test data
	data := []map[string]any{
		{"name": "John", "age": 30, "city": "New York"},
		{"name": "Jane", "age": 25, "city": "Boston"},
		{"name": "Bob", "age": 35, "city": "Chicago"},
	}

	// Create CSV instance
	csv := newCsv(newDynamicValue(data), func(s Source, d Dest) {
		d.Col("name", s.Key("name"))
		d.Col("age", s.Key("age"))
		d.Col("city", s.Key("city"))
	})

	// Test splitting by age
	var youngBuf, oldBuf bytes.Buffer
	youngFilter := Split(&youngBuf, "age", func(age int) bool {
		return age < 30
	})
	oldFilter := Split(&oldBuf, "age", func(age int) bool {
		return age >= 30
	})

	// Export with splits
	err := csv.ExportSplit(youngFilter, oldFilter)
	if err != nil {
		t.Fatalf("ExportSplit() error = %v", err)
	}

	// Verify young people CSV
	youngCSV := youngBuf.String()
	if !strings.Contains(youngCSV, "Jane,25,Boston") {
		t.Errorf("Young CSV missing expected data: %s", youngCSV)
	}
	if strings.Contains(youngCSV, "John,30,New York") {
		t.Errorf("Young CSV contains unexpected data: %s", youngCSV)
	}

	// Verify old people CSV
	oldCSV := oldBuf.String()
	if !strings.Contains(oldCSV, "John,30,New York") {
		t.Errorf("Old CSV missing expected data: %s", oldCSV)
	}
	if strings.Contains(oldCSV, "Jane,25,Boston") {
		t.Errorf("Old CSV contains unexpected data: %s", oldCSV)
	}
}
