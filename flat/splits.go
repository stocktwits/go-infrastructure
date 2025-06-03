package flat

import (
	"fmt"
	"io"
)

// splitter defines an interface for splitting data based on a header and a DynamicValue.
type splitter interface {
	// shouldInclude checks if the data should be skipped based on the header and the DynamicValue.
	shouldInclude(header string, dv *DynamicValue) (bool, error)
}

// splitWriter defines an interface that combines io.Writer and splitter.
// It allows writing data while also determining if the data should be split based on a header and a DynamicValue.
type splitWriter interface {
	io.Writer
	splitter
}

// singleSplitter implements the splitter interface
// splitting the data based on a single header and a split function.
type singleSplitter struct {
	header      string
	includeFunc func(*DynamicValue) (bool, error)
}

// NewSplitter creates a new splitter instance that uses the provided header and rawIncludeFunc.
// The rawIncludeFunc should return false if the line should be skipped.
func NewSplitter[T any](header string, rawIncludeFunc func(T) bool) splitter {
	return &singleSplitter{
		header:      header,
		includeFunc: getSplitFunc(rawIncludeFunc),
	}
}

// getSplitFunc returns a function that checks if a DynamicValue should be split based on the provided rawSplitFunc.
func getSplitFunc[T any](rawSplitFunc func(T) bool) func(*DynamicValue) (bool, error) {
	return func(dv *DynamicValue) (bool, error) {
		expectedType := getDataTypeFromType[T]()
		dvType := dv.DataType()

		// Handle numeric type conversions
		if expectedType == DataTypeInt && dvType == DataTypeFloat {
			if fval, ok := dv.value.(float64); ok {
				if ival := int(fval); float64(ival) == fval { // Only convert if no precision is lost
					return rawSplitFunc(any(ival).(T)), nil
				}
			}
		}

		if dvType != expectedType {
			return false, fmt.Errorf("split function type mismatch with data type")
		}

		rawValue := dv.value.(T)
		return rawSplitFunc(rawValue), nil
	}
}

// singleSplitWriter implements the splitWriter interface.
// It combines an io.Writer with a singleSplitter.
type singleSplitWriter struct {
	io.Writer
	splitter
}

// NoSplit returns a splitWriter that does not perform any splitting.
func NoSplit(w io.Writer) splitWriter {
	return singleSplitWriter{
		Writer: w,
		splitter: NewSplitter("", func(_ any) bool {
			return true // Always include data when no split is defined
		}),
	}
}

// Split creates a split instance that writes to the provided writer and uses the specified header
// and rawIncludeFunc to determine if the data should be inluded in the split.
// the rawIncludeFunc should return false if the line should be skipped.
func Split[T any](w io.Writer, header string, rawIncludeFunc func(T) bool) splitWriter {
	return singleSplitWriter{
		Writer:   w,
		splitter: NewSplitter(header, rawIncludeFunc),
	}
}

// shouldInclude checks if the data should be included based on the header and the DynamicValue.
func (s *singleSplitter) shouldInclude(header string, dv *DynamicValue) (bool, error) {
	// If the header does not match the split header, do not skip
	if s.header != header {
		return true, nil
	}

	// If no split function is defined, do not skip
	if s.includeFunc == nil {
		return true, nil
	}

	return s.includeFunc(dv)
}

// splitOperation defines the type of logical operation to be performed
// when combining multiple splitters
type splitOperation int

const (
	splitAndOperation splitOperation = iota
	splitOrOperation
)

// splitWriterOperation implements the splitWriter interface by combining multiple splitters
// with a logical operation (AND/OR) to determine if data should be included
type splitWriterOperation struct {
	io.Writer
	splitters []splitter
	operation splitOperation
}

// SplitAnd creates a splitWriter that writes to the provided writer and applies an AND operation
// across the provided splitters. If any splitter returns false, the data is skipped.
// If all splitters return true, the data is included.
func SplitAnd(w io.Writer, spliters ...splitter) splitWriter {
	return &splitWriterOperation{
		Writer:    w,
		splitters: spliters,
		operation: splitAndOperation,
	}
}

// SplitOr creates a splitWriter that writes to the provided writer and applies an OR operation
// across the provided splitters. If any splitter returns true, the data is included.
// If all splitters return false, the data is skipped.
func SplitOr(w io.Writer, spliters ...splitter) splitWriter {
	return &splitWriterOperation{
		Writer:    w,
		splitters: spliters,
		operation: splitOrOperation,
	}
}

func (s *splitWriterOperation) shouldInclude(header string, dv *DynamicValue) (bool, error) {
	// If no splitters are defined, include all data
	if len(s.splitters) == 0 {
		return true, nil
	}

	// Iterate through all splitters and apply the operation
	for _, splitter := range s.splitters {
		include, err := splitter.shouldInclude(header, dv)
		if err != nil {
			return false, fmt.Errorf("error checking split condition: %w", err)
		}

		if s.operation == splitAndOperation && !include {
			return false, nil // If any splitter returns false in AND operation, skip
		}
		if s.operation == splitOrOperation && include {
			return true, nil // If any splitter returns true in OR operation, include
		}
	}

	// If all splitters returned true in AND operation, include; otherwise skip
	return s.operation == splitAndOperation, nil
}
