package flat

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"slices"
)

const bufferSize = 100

// rootDataTypes defines the types of data that can be used as root data for CSV generation.
var rootDataTypes = []DataType{
	DataTypeObject,
	DataTypeArray,
	DataTypeArrayOfObjects,
	DataTypeStreamOfObjects,
}

// CSV represets data that can be exported to CSV format.
type CSV struct {
	rootData  *DynamicValue
	err       error
	flattener flattener
}

// newCsv creates a new CSV instance from the provided rootDynamicValue and flattener function.
// It checks if the rootDynamicValue contains an error or if its data type is supported for CSV generation.
// If the data type is not supported, it returns an error CSV instance.
func newCsv(rootDynamicValue *DynamicValue, f flattener) *CSV {
	if rootDynamicValue.Error() != nil {
		return newErrorCsv(rootDynamicValue.Error())
	}

	if !slices.Contains(rootDataTypes, rootDynamicValue.dataType) {
		return newErrorCsv(fmt.Errorf("data type is not supported for CSV generation"))
	}

	return &CSV{
		rootData:  rootDynamicValue,
		flattener: f,
	}
}

// newErrorCsv creates a new CSV instance that represents an error.
func newErrorCsv(err error) *CSV {
	return &CSV{
		err: err,
	}
}

// Export writes the CSV data to the provided writers.
// It writes the headers first, then the data rows.
// If an error has occurred during the process, it returns an error.
func (t *CSV) Export(w io.Writer) error {
	return t.ExportSplit(NoSplit(w))
}

// ExportSplit writes the CSV data to multiple writers based on the provided Splits.
// A Split contains a writer and an optional split function.
// The split function is used to determine whether a row should be written to that writer.
func (t *CSV) ExportSplit(splitters ...splitWriter) error {
	if t.err != nil {
		return fmt.Errorf("cannot export CSV due to previous error: %w", t.err)
	}

	csvWriters := make([]*csv.Writer, len(splitters))
	for i, s := range splitters {
		csvWriters[i] = csv.NewWriter(s)
	}

	rows := make(chan *row, bufferSize)
	go t.streamRows(rows)

	var headers []string
	for row := range rows {
		if row.hasHeaders() {
			headers = row.getHeaders()

			for _, csvWriter := range csvWriters {
				if err := csvWriter.Write(headers); err != nil {
					return fmt.Errorf("failed to write CSV headers: %w", err)
				}
			}
		}

		for i, csvWriter := range csvWriters {
			columnValues := make([]string, len(headers))
			includeLine := true
			for j, header := range headers {
				if column, exists := row.columns[header]; exists {
					// Check if the split function should include this line
					shouldInclude, err := splitters[i].shouldInclude(header, column.data)
					if err != nil {
						return fmt.Errorf("error checking split condition for header %s: %w", header, err)
					}

					if !shouldInclude {
						includeLine = false
						break // Skip writing this line for this writer
					}

					val, err := column.strVal()
					if err != nil {
						return fmt.Errorf("failed to get value for header %s: %w", header, err)
					}

					columnValues[j] = val
				}
			}

			if !includeLine {
				continue // Skip writing this line for this writer
			}

			if err := csvWriter.Write(columnValues); err != nil {
				return fmt.Errorf("failed to write CSV data: %w", err)
			}
		}
	}

	for _, csvWriter := range csvWriters {
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return fmt.Errorf("failed to flush CSV writer: %w", err)
		}
	}

	return nil
}

// Dest is an interface for writing data to a CSV.
// Add more detailed documentation for interfaces
type Dest interface {
	// Col adds a column to the CSV.
	// Parameters:
	//   name: The column header name
	//   value: The source value to add
	Col(name string, value Source)

	// ColFormatted adds a formatted column to the CSV.
	// Parameters:
	//   name: The column header name
	//   value: The source value to add
	//   formatter: A function to format the value before adding
	ColFormatted(name string, value Source, formatter Formatter)
}

// Source represents a source of data for CSV generation.
// It provides methods to access data by index or key.
type Source struct {
	data *DynamicValue
}

// FixValue creates a new Source instance with a fixed value.
// This is useful when you want to include a constant value in the CSV output.
func FixValue[T any](value T) Source {
	// Create a new DynamicValue from the provided value
	dv := newDynamicValue(value)
	return Source{data: dv}
}

// Idx retrieves an element from the Source instance that holds an array or an array of objects.
// If the index is out of bounds or the data type is not an array, it returns a new Source with NullData.
// If the data is not an array, it returns a new Source with NullData.
// If the index is valid, it returns a new Source with the data at that index.
func (s Source) Idx(index int) Source {
	return Source{
		data: s.data.Idx(index),
	}
}

// Key retrieves a value from the Source instance using a sequence of keys.
// If no keys are provided, it returns a new Source with NullData.
// If the data is not an object or the key does not exist, it returns a new Source with NullData.
// If the keys are valid, it returns a new Source with the data for those keys.
func (s Source) Key(keys ...string) Source {
	return Source{
		data: s.data.Key(keys...),
	}
}

// format applies a formatting function to the data in the Source instance.
// The formatter function is used to transform the data before it is written to the CSV.
// If multiple formatters are applied, the last one will take precedence.
func (s Source) format(formater Formatter) Source {
	if formater == nil {
		return s
	}

	newData, err := formater(s.data)
	if err != nil {
		return Source{
			data: errorDynamicValue(fmt.Errorf("error formatting data: %w", err)),
		}
	}

	return Source{
		data: newData,
	}
}

// strVal retrieves the string representation of the data in the Source instance.
func (s Source) strVal() (string, error) {
	return s.data.strVal()
}

// flattener is a function type that takes a Source and a Dest as arguments.
// It is used to map data from the Source to the Dest during CSV generation.
type flattener func(s Source, b Dest)

// row represents a single row of data in the CSV.
// It contains a map of column names to their corresponding Source values,
// a slice of headers (if applicable), and a flag indicating whether headers are included.
type row struct {
	columns     map[string]Source
	headers     []string
	withHeaders bool
}

// newRow creates a new row instance.
// If withHeaders is true, it initializes the headers slice to track column names.
func newRow(withHeaders bool) *row {
	r := &row{
		columns:     make(map[string]Source),
		withHeaders: withHeaders,
	}

	if withHeaders {
		r.headers = make([]string, 0)
	}

	return r
}

// Col adds a column to the row with the specified name and value.
func (r *row) Col(name string, value Source) {
	r.ColFormatted(name, value, nil)
}

// Col adds a column to the row with the specified name and value.
func (r *row) ColFormatted(name string, value Source, formatter Formatter) {
	if r.withHeaders {
		if !slices.Contains(r.headers, name) {
			r.headers = append(r.headers, name)
		}
	}

	if formatter != nil {
		value = value.format(formatter)
	}

	r.columns[name] = value
}

// hasHeaders checks if the row has headers.
func (r *row) hasHeaders() bool {
	return r.withHeaders
}

// getHeaders returns the headers of the row if they are included.
func (r *row) getHeaders() []string {
	if r.withHeaders {
		return r.headers
	}
	return nil
}

// streamRows streams the rows from the rootData based on its data type.
func (t *CSV) streamRows(rows chan *row) {
	switch t.rootData.DataType() {
	case DataTypeObject:
		s := Source{data: t.rootData}
		d := newRow(true)
		t.flattener(s, d)
		rows <- d
	case DataTypeArray:
		arr := t.rootData.value.([]any)
		for i, item := range arr {
			s := Source{data: newDynamicValue(item)}
			d := newRow(i == 0) // Only write headers for the first item
			t.flattener(s, d)
			rows <- d
		}
	case DataTypeArrayOfObjects:
		arr := t.rootData.value.([]map[string]any)
		for i, item := range arr {
			s := Source{data: newDynamicValue(item)}
			d := newRow(i == 0) // Only write headers for the first item
			t.flattener(s, d)
			rows <- d
		}
	case DataTypeStreamOfObjects:
		reader := t.rootData.value.(io.Reader)
		decoder := json.NewDecoder(reader)

		withHeaders := true
		for {
			var item map[string]any
			if err := decoder.Decode(&item); err == io.EOF {
				break // End of stream
			} else if err != nil {
				t.err = fmt.Errorf("error decoding JSON stream: %w", err)
				return
			}
			s := Source{data: newDynamicValue(item)}
			d := newRow(withHeaders)
			t.flattener(s, d)
			rows <- d

			withHeaders = false // Only write headers for the first item
		}
	}

	close(rows)
}
