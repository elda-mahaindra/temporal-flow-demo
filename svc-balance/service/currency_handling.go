package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// CurrencyInfo represents detailed information about a currency
type CurrencyInfo struct {
	Code         string           `json:"code"`
	Name         string           `json:"name"`
	Symbol       string           `json:"symbol"`
	DecimalPlace int              `json:"decimal_place"`
	IsActive     bool             `json:"is_active"`
	ExchangeRate *decimal.Decimal `json:"exchange_rate,omitempty"` // Rate to USD
}

// GetSupportedCurrenciesParams represents input parameters for getting supported currencies
type GetSupportedCurrenciesParams struct {
	IncludeInactive bool `json:"include_inactive"`
}

// GetSupportedCurrenciesResults represents the result of getting supported currencies
type GetSupportedCurrenciesResults struct {
	Currencies []CurrencyInfo `json:"currencies"`
	Count      int            `json:"count"`
}

// GetSupportedCurrencies returns a list of all supported currencies
func (service *Service) GetSupportedCurrencies(ctx context.Context, params GetSupportedCurrenciesParams) (*GetSupportedCurrenciesResults, error) {
	const op = "service.Service.GetSupportedCurrencies"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	currencies := service.getSupportedCurrencyList(params.IncludeInactive)

	result := &GetSupportedCurrenciesResults{
		Currencies: currencies,
		Count:      len(currencies),
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// ConvertCurrencyParams represents input parameters for currency conversion
type ConvertCurrencyParams struct {
	Amount       decimal.Decimal `json:"amount"`
	FromCurrency string          `json:"from_currency"`
	ToCurrency   string          `json:"to_currency"`
}

// ConvertCurrencyResults represents the result of currency conversion
type ConvertCurrencyResults struct {
	OriginalAmount    decimal.Decimal `json:"original_amount"`
	ConvertedAmount   decimal.Decimal `json:"converted_amount"`
	FromCurrency      string          `json:"from_currency"`
	ToCurrency        string          `json:"to_currency"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`
	ConversionApplied bool            `json:"conversion_applied"`
}

// ConvertCurrency converts an amount from one currency to another
func (service *Service) ConvertCurrency(ctx context.Context, params ConvertCurrencyParams) (*ConvertCurrencyResults, error) {
	const op = "service.Service.ConvertCurrency"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Validate input parameters
	if err := service.validateConvertCurrencyParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Get currency information
	fromCurrencyInfo, err := service.getCurrencyInfo(params.FromCurrency)
	if err != nil {
		err = fmt.Errorf("invalid from currency: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	toCurrencyInfo, err := service.getCurrencyInfo(params.ToCurrency)
	if err != nil {
		err = fmt.Errorf("invalid to currency: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Perform conversion
	convertedAmount, exchangeRate, conversionApplied := service.performCurrencyConversion(
		params.Amount, fromCurrencyInfo, toCurrencyInfo)

	result := &ConvertCurrencyResults{
		OriginalAmount:    params.Amount,
		ConvertedAmount:   convertedAmount,
		FromCurrency:      params.FromCurrency,
		ToCurrency:        params.ToCurrency,
		ExchangeRate:      exchangeRate,
		ConversionApplied: conversionApplied,
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// ValidateCurrencyParams represents input parameters for currency validation
type ValidateCurrencyParams struct {
	Currency          string   `json:"currency"`
	AllowedCurrencies []string `json:"allowed_currencies,omitempty"`
	RequireActive     bool     `json:"require_active"`
}

// ValidateCurrencyResults represents the result of currency validation
type ValidateCurrencyResults struct {
	Currency    string       `json:"currency"`
	IsValid     bool         `json:"is_valid"`
	IsSupported bool         `json:"is_supported"`
	IsActive    bool         `json:"is_active"`
	Info        CurrencyInfo `json:"info"`
	Message     string       `json:"message"`
}

// ValidateCurrencyEnhanced performs comprehensive currency validation
func (service *Service) ValidateCurrencyEnhanced(ctx context.Context, params ValidateCurrencyParams) (*ValidateCurrencyResults, error) {
	const op = "service.Service.ValidateCurrencyEnhanced"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Validate input parameters
	if params.Currency == "" {
		err := fmt.Errorf("currency code cannot be empty")

		logger.WithError(err).Error()

		return nil, err
	}

	// Normalize currency code
	normalizedCurrency := strings.ToUpper(strings.TrimSpace(params.Currency))

	// Get currency information
	currencyInfo, err := service.getCurrencyInfo(normalizedCurrency)
	isSupported := err == nil

	result := &ValidateCurrencyResults{
		Currency:    normalizedCurrency,
		IsSupported: isSupported,
		IsValid:     isSupported,
		IsActive:    isSupported && currencyInfo.IsActive,
	}

	if isSupported {
		result.Info = currencyInfo
		result.Message = fmt.Sprintf("Currency %s is supported", normalizedCurrency)

		// Check if currency is active if required
		if params.RequireActive && !currencyInfo.IsActive {
			result.IsValid = false
			result.Message = fmt.Sprintf("Currency %s is supported but not active", normalizedCurrency)
		}

		// Check against allowed currencies if provided
		if len(params.AllowedCurrencies) > 0 {
			allowed := false
			for _, allowedCurrency := range params.AllowedCurrencies {
				if strings.ToUpper(allowedCurrency) == normalizedCurrency {
					allowed = true
					break
				}
			}
			if !allowed {
				result.IsValid = false
				result.Message = fmt.Sprintf("Currency %s is not in the allowed list", normalizedCurrency)
			}
		}
	} else {
		result.Message = fmt.Sprintf("Currency %s is not supported", normalizedCurrency)
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// getSupportedCurrencyList returns the list of supported currencies with their information
func (service *Service) getSupportedCurrencyList(includeInactive bool) []CurrencyInfo {
	currencies := []CurrencyInfo{
		{Code: "USD", Name: "US Dollar", Symbol: "$", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(1.0))},
		{Code: "EUR", Name: "Euro", Symbol: "€", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(0.85))},
		{Code: "GBP", Name: "British Pound", Symbol: "£", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(0.73))},
		{Code: "JPY", Name: "Japanese Yen", Symbol: "¥", DecimalPlace: 0, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(110.0))},
		{Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(1.25))},
		{Code: "AUD", Name: "Australian Dollar", Symbol: "A$", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(1.35))},
		{Code: "CHF", Name: "Swiss Franc", Symbol: "Fr", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(0.92))},
		{Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(6.45))},
		{Code: "SGD", Name: "Singapore Dollar", Symbol: "S$", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(1.35))},
		{Code: "HKD", Name: "Hong Kong Dollar", Symbol: "HK$", DecimalPlace: 2, IsActive: true, ExchangeRate: decimalPtr(decimal.NewFromFloat(7.80))},
	}

	if includeInactive {
		// Add some inactive currencies for demonstration
		currencies = append(currencies, []CurrencyInfo{
			{Code: "BTC", Name: "Bitcoin", Symbol: "₿", DecimalPlace: 8, IsActive: false, ExchangeRate: decimalPtr(decimal.NewFromFloat(45000.0))},
			{Code: "ETH", Name: "Ethereum", Symbol: "Ξ", DecimalPlace: 8, IsActive: false, ExchangeRate: decimalPtr(decimal.NewFromFloat(3000.0))},
		}...)
	}

	return currencies
}

// getCurrencyInfo returns detailed information about a specific currency
func (service *Service) getCurrencyInfo(currencyCode string) (CurrencyInfo, error) {
	normalizedCode := strings.ToUpper(strings.TrimSpace(currencyCode))

	currencies := service.getSupportedCurrencyList(true) // Include inactive for info lookup

	for _, currency := range currencies {
		if currency.Code == normalizedCode {
			return currency, nil
		}
	}

	return CurrencyInfo{}, fmt.Errorf("currency %s not found", normalizedCode)
}

// validateConvertCurrencyParams validates currency conversion parameters
func (service *Service) validateConvertCurrencyParams(params ConvertCurrencyParams) error {
	if params.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}

	if params.FromCurrency == "" {
		return fmt.Errorf("from_currency cannot be empty")
	}

	if params.ToCurrency == "" {
		return fmt.Errorf("to_currency cannot be empty")
	}

	// Validate both currencies are supported
	if err := service.validateCurrency(params.FromCurrency); err != nil {
		return fmt.Errorf("invalid from_currency: %w", err)
	}

	if err := service.validateCurrency(params.ToCurrency); err != nil {
		return fmt.Errorf("invalid to_currency: %w", err)
	}

	return nil
}

// performCurrencyConversion performs the actual currency conversion calculation
func (service *Service) performCurrencyConversion(amount decimal.Decimal, fromCurrency, toCurrency CurrencyInfo) (decimal.Decimal, decimal.Decimal, bool) {
	// If same currency, no conversion needed
	if fromCurrency.Code == toCurrency.Code {
		return amount, decimal.NewFromFloat(1.0), false
	}

	// Convert via USD as base currency
	// First convert from source currency to USD
	amountInUSD := amount.Div(*fromCurrency.ExchangeRate)

	// Then convert from USD to target currency
	convertedAmount := amountInUSD.Mul(*toCurrency.ExchangeRate)

	// Calculate the direct exchange rate
	exchangeRate := (*toCurrency.ExchangeRate).Div(*fromCurrency.ExchangeRate)

	// Round to appropriate decimal places for target currency
	convertedAmount = convertedAmount.Round(int32(toCurrency.DecimalPlace))

	return convertedAmount, exchangeRate, true
}

// NormalizeCurrencyAmount normalizes an amount to the appropriate decimal places for a currency
func (service *Service) NormalizeCurrencyAmount(amount decimal.Decimal, currencyCode string) (decimal.Decimal, error) {
	currencyInfo, err := service.getCurrencyInfo(currencyCode)
	if err != nil {
		return amount, fmt.Errorf("failed to get currency info: %w", err)
	}

	return amount.Round(int32(currencyInfo.DecimalPlace)), nil
}

// GetCurrencySymbol returns the symbol for a given currency code
func (service *Service) GetCurrencySymbol(currencyCode string) (string, error) {
	currencyInfo, err := service.getCurrencyInfo(currencyCode)
	if err != nil {
		return "", fmt.Errorf("failed to get currency info: %w", err)
	}

	return currencyInfo.Symbol, nil
}

// FormatCurrencyAmount formats an amount with the appropriate currency symbol and decimal places
func (service *Service) FormatCurrencyAmount(amount decimal.Decimal, currencyCode string) (string, error) {
	currencyInfo, err := service.getCurrencyInfo(currencyCode)
	if err != nil {
		return "", fmt.Errorf("failed to get currency info: %w", err)
	}

	normalizedAmount := amount.Round(int32(currencyInfo.DecimalPlace))

	return fmt.Sprintf("%s%s", currencyInfo.Symbol, normalizedAmount.String()), nil
}

// decimalPtr is a helper function to create a pointer to a decimal value
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
