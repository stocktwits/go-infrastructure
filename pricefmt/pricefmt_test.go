package pricefmt

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newPtr[T any](v T) *T {
	return &v
}

// formatTestCase is a struct to hold test cases for price formatting functions.
type formatTestCase[T priceInput] struct {
	name         string
	price        T // Now T, which must satisfy priceInput
	currencyCode string
	expected     *PriceFormatted
	expectedErr  bool
}

// runFormatTests is a helper to reduce duplication for FormatWithCurrency and TryFormatWithCurrency tests.
func runFormatTests[T priceInput](t *testing.T, tests []formatTestCase[T]) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test FormatWithCurrency
			formatted, err := FormatWithCurrency(tt.price, tt.currencyCode)
			if tt.expectedErr {
				assert.Error(t, err, "FormatWithCurrency should return an error for invalid input")
				assert.Nil(t, formatted, "Formatted price should be nil on error")
			} else {
				assert.NoError(t, err, "FormatWithCurrency should not return an error for valid input")
				assert.NotNil(t, formatted, "Formatted price should not be nil")
				assert.Equal(t, tt.expected.UseSubscript, formatted.UseSubscript, "UseSubscript mismatch")
				assert.Equal(t, tt.expected.RawValue, formatted.RawValue, "RawValue mismatch")
				assert.Equal(t, tt.expected.CurrencyCode, formatted.CurrencyCode, "CurrencyCode mismatch")
				assert.Equal(t, tt.expected.CurrencyString, formatted.CurrencyString, "CurrencyString mismatch")
				assert.Equal(t, tt.expected.ZerosAfterDecimal, formatted.ZerosAfterDecimal, "ZerosAfterDecimal mismatch")
				assert.Equal(t, tt.expected.AfterZerosValue, formatted.AfterZerosValue, "AfterZerosValue mismatch")
			}

			// Test TryFormatWithCurrency
			tryFormatted := TryFormatWithCurrency(tt.price, tt.currencyCode)
			if tt.expectedErr {
				assert.Nil(t, tryFormatted, "TryFormatWithCurrency should return nil for invalid input")
			} else {
				assert.NotNil(t, tryFormatted, "TryFormatWithCurrency should not return nil")
				assert.Equal(t, tt.expected.UseSubscript, tryFormatted.UseSubscript, "TryFormatWithCurrency UseSubscript mismatch")
				assert.Equal(t, tt.expected.RawValue, tryFormatted.RawValue, "TryFormatWithCurrency RawValue mismatch")
				assert.Equal(t, tt.expected.CurrencyCode, tryFormatted.CurrencyCode, "TryFormatWithCurrency CurrencyCode mismatch")
				assert.Equal(t, tt.expected.CurrencyString, tryFormatted.CurrencyString, "TryFormatWithCurrency CurrencyString mismatch")
				assert.Equal(t, tt.expected.ZerosAfterDecimal, tryFormatted.ZerosAfterDecimal, "TryFormatWithCurrency ZerosAfterDecimal mismatch")
				assert.Equal(t, tt.expected.AfterZerosValue, tryFormatted.AfterZerosValue, "TryFormatWithCurrency AfterZerosValue mismatch")
			}
		})
	}
}

// runFormatDefaultCurrencyTests is a helper for Format and TryFormat tests.
func runFormatDefaultCurrencyTests[T priceInput](t *testing.T, tests []formatTestCase[T]) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Format (default currency)
			formatted, err := Format(tt.price)
			if tt.expectedErr {
				assert.Error(t, err, "Format should return an error for invalid input")
				assert.Nil(t, formatted, "Formatted price should be nil on error")
			} else {
				assert.NoError(t, err, "Format should not return an error for valid input")
				assert.NotNil(t, formatted, "Formatted price should not be nil")
				assert.Equal(t, tt.expected.UseSubscript, formatted.UseSubscript, "UseSubscript mismatch")
				assert.Equal(t, tt.expected.RawValue, formatted.RawValue, "RawValue mismatch")
				assert.Equal(t, tt.expected.CurrencyCode, formatted.CurrencyCode, "CurrencyCode mismatch")
				assert.Equal(t, tt.expected.CurrencyString, formatted.CurrencyString, "CurrencyString mismatch")
				assert.Equal(t, tt.expected.ZerosAfterDecimal, formatted.ZerosAfterDecimal, "ZerosAfterDecimal mismatch")
				assert.Equal(t, tt.expected.AfterZerosValue, formatted.AfterZerosValue, "AfterZerosValue mismatch")
			}

			// Test TryFormat (default currency)
			tryFormatted := TryFormat(tt.price)
			if tt.expectedErr {
				assert.Nil(t, tryFormatted, "TryFormat should return nil for invalid input")
			} else {
				assert.NotNil(t, tryFormatted, "TryFormat should not return nil")
				assert.Equal(t, tt.expected.UseSubscript, tryFormatted.UseSubscript, "TryFormat UseSubscript mismatch")
				assert.Equal(t, tt.expected.RawValue, tryFormatted.RawValue, "TryFormat RawValue mismatch")
				assert.Equal(t, tt.expected.CurrencyCode, tryFormatted.CurrencyCode, "TryFormat CurrencyCode mismatch")
				assert.Equal(t, tt.expected.CurrencyString, tryFormatted.CurrencyString, "TryFormat CurrencyString mismatch")
				assert.Equal(t, tt.expected.ZerosAfterDecimal, tryFormatted.ZerosAfterDecimal, "TryFormat ZerosAfterDecimal mismatch")
				assert.Equal(t, tt.expected.AfterZerosValue, tryFormatted.AfterZerosValue, "TryFormat AfterZerosValue mismatch")
			}
		})
	}
}

// Test cases for int inputs
func TestFormatWithCurrency_Int(t *testing.T) {
	// Tests for FormatWithCurrency and TryFormatWithCurrency (explicit currency)
	explicitCurrencyTests := []formatTestCase[int]{
		{
			name:         "USD integer",
			price:        123,
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "123",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "GBP large number",
			price:        1000000,
			currencyCode: CurrencyCodeGBP,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "1000000",
				CurrencyCode:      CurrencyCodeGBP,
				CurrencyString:    "£",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "AUD with no decimal",
			price:        77,
			currencyCode: CurrencyCodeAUD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "77",
				CurrencyCode:      CurrencyCodeAUD,
				CurrencyString:    "A$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "Unsupported currency code",
			price:        100,
			currencyCode: "XYZ", // Unsupported
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "100",
				CurrencyCode:      "XYZ",
				CurrencyString:    "XYZ", // Should fall back to code itself
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
	}
	runFormatTests(t, explicitCurrencyTests)

	// Tests for Format and TryFormat (default USD currency)
	defaultCurrencyTests := []formatTestCase[int]{
		{
			name:  "Default USD integer",
			price: 456,
			// For default currency functions, the expected currency is always USD
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "456",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
	}
	runFormatDefaultCurrencyTests(t, defaultCurrencyTests)
}

// Test cases for float64 inputs
func TestFormatWithCurrency_Float64(t *testing.T) {
	// Tests for FormatWithCurrency and TryFormatWithCurrency (explicit currency)
	explicitCurrencyTests := []formatTestCase[float64]{
		{
			name:         "USD float",
			price:        123.45,
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "123.45",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "USD zero",
			price:        0.0,
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "0",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "USD small decimal with one zero",
			price:        0.0123,
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.0123",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: newPtr(1),
				AfterZerosValue:   newPtr[int64](123),
			},
			expectedErr: false,
		},
		{
			name:         "USD small decimal with multiple zeros",
			price:        0.00000456,
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.00000456",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: newPtr(5),
				AfterZerosValue:   newPtr[int64](456),
			},
			expectedErr: false,
		},
		{
			name:         "USD small decimal with trailing zeros in raw value (decimal handles internally)",
			price:        0.00100, // decimal.NewFromFloat(0.00100) will be "0.001"
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.001",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: newPtr(2),
				AfterZerosValue:   newPtr[int64](1),
			},
			expectedErr: false,
		},
		{
			name:         "USD small decimal ending in zero",
			price:        0.00010, // decimal.NewFromFloat(0.00010) will be "0.0001"
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.0001",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: newPtr(3),
				AfterZerosValue:   newPtr[int64](1),
			},
			expectedErr: false,
		},
		{
			name:         "USD non-zero integer with decimal part, no leading zeros",
			price:        1.0000001,
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "1.0000001",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "EUR small decimal",
			price:        0.003,
			currencyCode: CurrencyCodeEUR,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.003",
				CurrencyCode:      CurrencyCodeEUR,
				CurrencyString:    "€",
				ZerosAfterDecimal: newPtr(2),
				AfterZerosValue:   newPtr[int64](3),
			},
			expectedErr: false,
		},
		{
			name:         "PHP small decimal from float",
			price:        0.0000000001,
			currencyCode: CurrencyCodePHP,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.0000000001",
				CurrencyCode:      CurrencyCodePHP,
				CurrencyString:    "₱",
				ZerosAfterDecimal: newPtr(9),
				AfterZerosValue:   newPtr[int64](1),
			},
			expectedErr: false,
		},
		{
			name:         "NZD just over zero",
			price:        0.5,
			currencyCode: CurrencyCodeNZD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "0.5",
				CurrencyCode:      CurrencyCodeNZD,
				CurrencyString:    "NZ$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
	}
	runFormatTests(t, explicitCurrencyTests)

	// Tests for Format and TryFormat (default USD currency)
	defaultCurrencyTests := []formatTestCase[float64]{
		{
			name:  "Default USD small decimal",
			price: 0.0005,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.0005",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: newPtr(3),
				AfterZerosValue:   newPtr[int64](5),
			},
			expectedErr: false,
		},
	}
	runFormatDefaultCurrencyTests(t, defaultCurrencyTests)
}

// Test cases for string inputs
func TestFormatWithCurrency_String(t *testing.T) {
	// Tests for FormatWithCurrency and TryFormatWithCurrency (explicit currency)
	explicitCurrencyTests := []formatTestCase[string]{
		{
			name:         "USD string",
			price:        "123.456",
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "123.456",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "USD small decimal with string input ending in zero",
			price:        "0.00010", // shopspring/decimal will normalize "0.00010" to "0.0001"
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.0001",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: newPtr(3),
				AfterZerosValue:   newPtr[int64](1),
			},
			expectedErr: false,
		},
		{
			name:         "CAD small decimal from string",
			price:        "0.00005",
			currencyCode: CurrencyCodeCAD,
			expected: &PriceFormatted{
				UseSubscript:      true,
				RawValue:          "0.00005",
				CurrencyCode:      CurrencyCodeCAD,
				CurrencyString:    "CA$",
				ZerosAfterDecimal: newPtr(4),
				AfterZerosValue:   newPtr[int64](5),
			},
			expectedErr: false,
		},
		{
			name:         "Invalid string price input",
			price:        "invalid-price",
			currencyCode: CurrencyCodeUSD,
			expected:     nil,
			expectedErr:  true,
		},
	}
	runFormatTests(t, explicitCurrencyTests)

	// Tests for Format and TryFormat (default USD currency)
	defaultCurrencyTests := []formatTestCase[string]{
		{
			name:        "Default USD invalid string input",
			price:       "not-a-number",
			expected:    nil,
			expectedErr: true,
		},
	}
	runFormatDefaultCurrencyTests(t, defaultCurrencyTests)
}

// Test cases for decimal.Decimal inputs
func TestFormatWithCurrency_Decimal(t *testing.T) {
	// Tests for FormatWithCurrency and TryFormatWithCurrency (explicit currency)
	explicitCurrencyTests := []formatTestCase[decimal.Decimal]{
		{
			name:         "USD decimal",
			price:        decimal.NewFromFloat(987.65),
			currencyCode: CurrencyCodeUSD,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "987.65",
				CurrencyCode:      CurrencyCodeUSD,
				CurrencyString:    "US$",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
		{
			name:         "INR decimal with no fractional part",
			price:        decimal.NewFromInt(500),
			currencyCode: CurrencyCodeINR,
			expected: &PriceFormatted{
				UseSubscript:      false,
				RawValue:          "500",
				CurrencyCode:      CurrencyCodeINR,
				CurrencyString:    "₹",
				ZerosAfterDecimal: nil,
				AfterZerosValue:   nil,
			},
			expectedErr: false,
		},
	}
	runFormatTests(t, explicitCurrencyTests)

	// Tests for Format and TryFormat (default USD currency)
	// No specific additional decimal cases needed beyond what's already covered by explicit,
	// but keeping the structure for consistency.
	defaultCurrencyTests := []formatTestCase[decimal.Decimal]{
		// You might add specific default USD decimal tests here if different behavior is expected
	}
	runFormatDefaultCurrencyTests(t, defaultCurrencyTests)
}

// Test getDecimalValue function
func TestGetDecimalValue(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    decimal.Decimal
		expectedErr bool
	}{
		{
			name:        "string to decimal",
			input:       "123.45",
			expected:    decimal.NewFromFloat(123.45),
			expectedErr: false,
		},
		{
			name:        "float64 to decimal",
			input:       98.76,
			expected:    decimal.NewFromFloat(98.76),
			expectedErr: false,
		},
		{
			name:        "int to decimal",
			input:       500,
			expected:    decimal.NewFromInt(500),
			expectedErr: false,
		},
		{
			name:        "decimal.Decimal (no conversion)",
			input:       decimal.NewFromFloat(1.23),
			expected:    decimal.NewFromFloat(1.23),
			expectedErr: false,
		},
		{
			name:        "invalid string to decimal",
			input:       "abc",
			expected:    decimal.Decimal{},
			expectedErr: true,
		},
		{
			name:        "unsupported type (bool)",
			input:       true,
			expected:    decimal.Decimal{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := getDecimalValue(tt.input)
			if tt.expectedErr {
				assert.Error(t, err, "getDecimalValue should return an error")
			} else {
				assert.NoError(t, err, "getDecimalValue should not return an error")
				assert.True(t, d.Equal(tt.expected), fmt.Sprintf("Expected %v, got %v", tt.expected, d))
			}
		})
	}
}

// Test getCurrencySymbol function
func TestGetCurrencySymbol(t *testing.T) {
	tests := []struct {
		currencyCode string
		expected     string
	}{
		{CurrencyCodeUSD, "US$"},
		{CurrencyCodeEUR, "€"},
		{CurrencyCodeGBP, "£"},
		{CurrencyCodeINR, "₹"},
		{CurrencyCodeCAD, "CA$"},
		{CurrencyCodeAUD, "A$"},
		{CurrencyCodePHP, "₱"},
		{CurrencyCodeNZD, "NZ$"},
		{"UNKNOWN", "UNKNOWN"}, // Test for unsupported code
		{"XYZ", "XYZ"},         // Another unsupported code
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Currency: %s", tt.currencyCode), func(t *testing.T) {
			symbol := getCurrencySymbol(tt.currencyCode)
			assert.Equal(t, tt.expected, symbol, "Currency symbol mismatch")
		})
	}
}
