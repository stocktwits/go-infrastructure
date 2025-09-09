package pricefmt

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Supported currency codes.
// These constants represent the currency codes that can be used in price formatting.
const (
	CurrencyCodeUSD = "USD"
	CurrencyCodeEUR = "EUR"
	CurrencyCodeGBP = "GBP"
	CurrencyCodeINR = "INR"
	CurrencyCodeCAD = "CAD"
	CurrencyCodeAUD = "AUD"
	CurrencyCodePHP = "PHP"
	CurrencyCodeNZD = "NZD"
)

// defaultCurrencyCode is the default currency code used when formatting prices.
const defaultCurrencyCode = CurrencyCodeUSD

// defaultSubscriptLength is the minimum number of leading zeros required to use subscript formatting.
const defaultSubscriptLength = 5

// defaultValueLength is the maximum number of digits to include in the after zeros value.
const defaultValueLength = 4

// priceInput is a type constraint for price inputs that can be formatted.
type priceInput interface {
	~string | ~float64 | ~int | decimal.Decimal
}

// PriceFormatted holds the formatted price data, including currency and subscript information.
type PriceFormatted struct {
	UseSubscript      bool
	RawValue          string
	CurrencyCode      string
	CurrencyString    string
	IsNegative        bool
	ZerosAfterDecimal *int
	AfterZerosValue   *int64
}

// TryFormat attempts to format a price with the default currency code (USD).
// It returns nil if the formatting fails, which is useful for optional price fields.
func TryFormat[T priceInput](price T) *PriceFormatted {
	return TryFormatWithCurrency(price, defaultCurrencyCode)
}

// Format formats a price with the default currency code (USD).
func Format[T priceInput](price T) (*PriceFormatted, error) {
	return FormatWithCurrency(price, defaultCurrencyCode)
}

// TryFormatWithCurrency attempts to format a price with a specified currency code.
func TryFormatWithCurrency[T priceInput](price T, currencyCode string) *PriceFormatted {
	formattedPrice, err := FormatWithCurrency(price, currencyCode)
	if err != nil {
		return nil
	}
	return formattedPrice
}

// FormatWithCurrency gets formatting data for a price, primarily for handling small decimals.
func FormatWithCurrency[T priceInput](price T, currencyCode string) (*PriceFormatted, error) {
	return FormatWithOptions(price, currencyCode, defaultSubscriptLength, defaultValueLength)
}

// FormatWithOptions gets formatting data for a price with configurable subscript and value length parameters.
func FormatWithOptions[T priceInput](price T, currencyCode string, subscriptLength, valueLength int) (*PriceFormatted, error) {
	dPrice, err := getDecimalValue(price)
	if err != nil {
		return nil, fmt.Errorf("error converting price to decimal: %w", err)
	}

	priceData := &PriceFormatted{
		UseSubscript:   false,
		RawValue:       dPrice.String(),
		CurrencyCode:   currencyCode,
		CurrencyString: getCurrencySymbol(currencyCode),
		IsNegative:     dPrice.IsNegative(),
	}

	// If the price is not a small decimal, return the basic data.
	if dPrice.IsZero() || dPrice.Abs().GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return priceData, nil
	}

	strPrice := dPrice.Abs().String()

	// If the price does not contain a decimal point, it is not a small decimal.
	// We return the basic data without subscript formatting.
	if !strings.Contains(strPrice, ".") {
		return priceData, nil
	}

	parts := strings.SplitN(strPrice, ".", 2)
	wholePart := parts[0]
	decimalPart := parts[1]

	// If the whole part is not zero, we return the basic data without subscript formatting.
	// The subscript formatting only applies to small decimals (i.e. 0.0001)
	if wholePart != "0" {
		return priceData, nil
	}

	leadingZeroesCount := 0
	for _, r := range decimalPart {
		if r == '0' {
			leadingZeroesCount++
		} else {
			break
		}
	}

	if leadingZeroesCount == 0 {
		return priceData, nil
	}

	// Get the value after zeros, limited by valueLength
	afterZerosStr := decimalPart[leadingZeroesCount:]
	if len(afterZerosStr) > valueLength {
		afterZerosStr = afterZerosStr[:valueLength]
	}
	if afterZerosStr == "" {
		afterZerosStr = "0"
	}

	afterZerosValueDecimal, err := decimal.NewFromString(afterZerosStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing after zeros value: %w", err)
	}

	afterZerosValue := afterZerosValueDecimal.IntPart()

	priceData.UseSubscript = leadingZeroesCount >= subscriptLength
	priceData.ZerosAfterDecimal = &leadingZeroesCount
	priceData.AfterZerosValue = &afterZerosValue

	return priceData, nil
}

// getDecimalValue converts various types of price inputs to a decimal.Decimal.
func getDecimalValue(price any) (decimal.Decimal, error) {
	switch v := price.(type) {
	case string:
		return decimal.NewFromString(v)
	case float64:
		return decimal.NewFromFloat(v), nil
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case decimal.Decimal:
		return v, nil
	default:
		return decimal.Decimal{}, fmt.Errorf("unsupported price type: %T", v)
	}
}

// getCurrencySymbol returns the currency symbol for a given currency code.
func getCurrencySymbol(currencyCode string) string {
	switch currencyCode {
	case CurrencyCodeUSD:
		return "US$"
	case CurrencyCodeEUR:
		return "€"
	case CurrencyCodeGBP:
		return "£"
	case CurrencyCodeINR:
		return "₹"
	case CurrencyCodeCAD:
		return "CA$"
	case CurrencyCodeAUD:
		return "A$"
	case CurrencyCodePHP:
		return "₱"
	case CurrencyCodeNZD:
		return "NZ$"
	default:
		// Fallback to the code itself if unknown.
		return currencyCode
	}
}
