package service

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// createTestService creates a test service instance
func createTestService() *Service {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	return &Service{
		logger: logger,
		// store will be nil for unit tests that don't need database
	}
}

func TestGetSupportedCurrencies(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name            string
		includeInactive bool
		expectedCount   int
		checkInactive   bool
	}{
		{
			name:            "Get active currencies only",
			includeInactive: false,
			expectedCount:   10, // Should return only active currencies
			checkInactive:   false,
		},
		{
			name:            "Get all currencies including inactive",
			includeInactive: true,
			expectedCount:   12, // Should return active + inactive currencies
			checkInactive:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			params := GetSupportedCurrenciesParams{
				IncludeInactive: tt.includeInactive,
			}

			result, err := service.GetSupportedCurrencies(context.Background(), params)

			if err != nil {
				t.Errorf("GetSupportedCurrencies() error = %v", err)
				return
			}

			if result.Count != tt.expectedCount {
				t.Errorf("GetSupportedCurrencies() count = %d, want %d", result.Count, tt.expectedCount)
			}

			if len(result.Currencies) != tt.expectedCount {
				t.Errorf("GetSupportedCurrencies() currencies length = %d, want %d", len(result.Currencies), tt.expectedCount)
			}

			// Verify that all returned currencies are active if includeInactive is false
			if !tt.includeInactive {
				for _, currency := range result.Currencies {
					if !currency.IsActive {
						t.Errorf("GetSupportedCurrencies() returned inactive currency %s when includeInactive=false", currency.Code)
					}
				}
			}

			// Verify that inactive currencies are included if includeInactive is true
			if tt.checkInactive {
				hasInactive := false
				for _, currency := range result.Currencies {
					if !currency.IsActive {
						hasInactive = true
						break
					}
				}
				if !hasInactive {
					t.Error("GetSupportedCurrencies() should include inactive currencies when includeInactive=true")
				}
			}

			// Verify required fields are present
			for _, currency := range result.Currencies {
				if currency.Code == "" {
					t.Error("GetSupportedCurrencies() currency missing code")
				}
				if currency.Name == "" {
					t.Error("GetSupportedCurrencies() currency missing name")
				}
				if currency.Symbol == "" {
					t.Error("GetSupportedCurrencies() currency missing symbol")
				}
				if currency.ExchangeRate == nil {
					t.Error("GetSupportedCurrencies() currency missing exchange rate")
				}
			}
		})
	}
}

func TestConvertCurrency(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name                  string
		params                ConvertCurrencyParams
		expectedFromCurrency  string
		expectedToCurrency    string
		expectedConversion    bool
		expectError           bool
		validateExchangeRate  bool
		expectedErrorContains string
	}{
		{
			name: "Convert USD to EUR",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("100.00"),
				FromCurrency: "USD",
				ToCurrency:   "EUR",
			},
			expectedFromCurrency: "USD",
			expectedToCurrency:   "EUR",
			expectedConversion:   true,
			expectError:          false,
			validateExchangeRate: true,
		},
		{
			name: "Convert EUR to USD",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("85.00"),
				FromCurrency: "EUR",
				ToCurrency:   "USD",
			},
			expectedFromCurrency: "EUR",
			expectedToCurrency:   "USD",
			expectedConversion:   true,
			expectError:          false,
			validateExchangeRate: true,
		},
		{
			name: "Same currency conversion",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("100.00"),
				FromCurrency: "USD",
				ToCurrency:   "USD",
			},
			expectedFromCurrency: "USD",
			expectedToCurrency:   "USD",
			expectedConversion:   false,
			expectError:          false,
			validateExchangeRate: false,
		},
		{
			name: "Convert to JPY (0 decimal places)",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("100.00"),
				FromCurrency: "USD",
				ToCurrency:   "JPY",
			},
			expectedFromCurrency: "USD",
			expectedToCurrency:   "JPY",
			expectedConversion:   true,
			expectError:          false,
			validateExchangeRate: true,
		},
		{
			name: "Invalid from currency",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("100.00"),
				FromCurrency: "INVALID",
				ToCurrency:   "USD",
			},
			expectError:           true,
			expectedErrorContains: "invalid from_currency",
		},
		{
			name: "Invalid to currency",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("100.00"),
				FromCurrency: "USD",
				ToCurrency:   "INVALID",
			},
			expectError:           true,
			expectedErrorContains: "invalid to_currency",
		},
		{
			name: "Negative amount",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("-100.00"),
				FromCurrency: "USD",
				ToCurrency:   "EUR",
			},
			expectError:           true,
			expectedErrorContains: "amount cannot be negative",
		},
		{
			name: "Empty from currency",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("100.00"),
				FromCurrency: "",
				ToCurrency:   "EUR",
			},
			expectError:           true,
			expectedErrorContains: "from_currency cannot be empty",
		},
		{
			name: "Empty to currency",
			params: ConvertCurrencyParams{
				Amount:       decimal.RequireFromString("100.00"),
				FromCurrency: "USD",
				ToCurrency:   "",
			},
			expectError:           true,
			expectedErrorContains: "to_currency cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			result, err := service.ConvertCurrency(context.Background(), tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("ConvertCurrency() expected error but got none")
					return
				}
				if tt.expectedErrorContains != "" && !strings.Contains(err.Error(), tt.expectedErrorContains) {
					t.Errorf("ConvertCurrency() error = %v, should contain %v", err.Error(), tt.expectedErrorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertCurrency() error = %v", err)
				return
			}

			// Validate basic fields
			if result.FromCurrency != tt.expectedFromCurrency {
				t.Errorf("ConvertCurrency() FromCurrency = %v, want %v", result.FromCurrency, tt.expectedFromCurrency)
			}

			if result.ToCurrency != tt.expectedToCurrency {
				t.Errorf("ConvertCurrency() ToCurrency = %v, want %v", result.ToCurrency, tt.expectedToCurrency)
			}

			if result.ConversionApplied != tt.expectedConversion {
				t.Errorf("ConvertCurrency() ConversionApplied = %v, want %v", result.ConversionApplied, tt.expectedConversion)
			}

			// Validate original amount
			if !result.OriginalAmount.Equal(tt.params.Amount) {
				t.Errorf("ConvertCurrency() OriginalAmount = %v, want %v", result.OriginalAmount, tt.params.Amount)
			}

			// Validate exchange rate logic
			if tt.validateExchangeRate {
				if result.ExchangeRate.IsZero() {
					t.Error("ConvertCurrency() ExchangeRate should not be zero for different currencies")
				}
				if result.ConvertedAmount.IsZero() {
					t.Error("ConvertCurrency() ConvertedAmount should not be zero")
				}
			}

			// For same currency, converted amount should equal original
			if !tt.expectedConversion {
				if !result.ConvertedAmount.Equal(result.OriginalAmount) {
					t.Errorf("ConvertCurrency() same currency conversion: ConvertedAmount = %v, want %v", result.ConvertedAmount, result.OriginalAmount)
				}
				if !result.ExchangeRate.Equal(decimal.NewFromFloat(1.0)) {
					t.Errorf("ConvertCurrency() same currency conversion: ExchangeRate = %v, want 1.0", result.ExchangeRate)
				}
			}
		})
	}
}

func TestValidateCurrencyEnhanced(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name           string
		params         ValidateCurrencyParams
		expectedValid  bool
		expectedActive bool
		expectError    bool
	}{
		{
			name: "Valid active currency",
			params: ValidateCurrencyParams{
				Currency:      "USD",
				RequireActive: false,
			},
			expectedValid:  true,
			expectedActive: true,
			expectError:    false,
		},
		{
			name: "Valid inactive currency without requiring active",
			params: ValidateCurrencyParams{
				Currency:      "BTC",
				RequireActive: false,
			},
			expectedValid:  true,
			expectedActive: false,
			expectError:    false,
		},
		{
			name: "Valid inactive currency but requiring active",
			params: ValidateCurrencyParams{
				Currency:      "BTC",
				RequireActive: true,
			},
			expectedValid:  false,
			expectedActive: false,
			expectError:    false,
		},
		{
			name: "Invalid currency",
			params: ValidateCurrencyParams{
				Currency:      "INVALID",
				RequireActive: false,
			},
			expectedValid:  false,
			expectedActive: false,
			expectError:    false,
		},
		{
			name: "Empty currency",
			params: ValidateCurrencyParams{
				Currency:      "",
				RequireActive: false,
			},
			expectError: true,
		},
		{
			name: "Currency in allowed list",
			params: ValidateCurrencyParams{
				Currency:          "USD",
				AllowedCurrencies: []string{"USD", "EUR"},
				RequireActive:     false,
			},
			expectedValid:  true,
			expectedActive: true,
			expectError:    false,
		},
		{
			name: "Currency not in allowed list",
			params: ValidateCurrencyParams{
				Currency:          "JPY",
				AllowedCurrencies: []string{"USD", "EUR"},
				RequireActive:     false,
			},
			expectedValid:  false,
			expectedActive: true,
			expectError:    false,
		},
		{
			name: "Case insensitive currency",
			params: ValidateCurrencyParams{
				Currency:      "usd",
				RequireActive: false,
			},
			expectedValid:  true,
			expectedActive: true,
			expectError:    false,
		},
		{
			name: "Currency with whitespace",
			params: ValidateCurrencyParams{
				Currency:      " EUR ",
				RequireActive: false,
			},
			expectedValid:  true,
			expectedActive: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			result, err := service.ValidateCurrencyEnhanced(context.Background(), tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("ValidateCurrencyEnhanced() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateCurrencyEnhanced() error = %v", err)
				return
			}

			if result.IsValid != tt.expectedValid {
				t.Errorf("ValidateCurrencyEnhanced() IsValid = %v, want %v", result.IsValid, tt.expectedValid)
			}

			if result.IsActive != tt.expectedActive {
				t.Errorf("ValidateCurrencyEnhanced() IsActive = %v, want %v", result.IsActive, tt.expectedActive)
			}

			// Verify currency normalization
			expectedCurrency := strings.ToUpper(strings.TrimSpace(tt.params.Currency))
			if result.Currency != expectedCurrency {
				t.Errorf("ValidateCurrencyEnhanced() Currency = %v, want %v", result.Currency, expectedCurrency)
			}

			// Verify message is not empty
			if result.Message == "" {
				t.Error("ValidateCurrencyEnhanced() Message should not be empty")
			}
		})
	}
}

func TestNormalizeCurrencyAmount(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name           string
		amount         decimal.Decimal
		currencyCode   string
		expectedAmount string
		expectError    bool
	}{
		{
			name:           "Normalize USD amount",
			amount:         decimal.RequireFromString("123.456"),
			currencyCode:   "USD",
			expectedAmount: "123.46", // Rounded to 2 decimal places
			expectError:    false,
		},
		{
			name:           "Normalize JPY amount",
			amount:         decimal.RequireFromString("123.456"),
			currencyCode:   "JPY",
			expectedAmount: "123", // Rounded to 0 decimal places
			expectError:    false,
		},
		{
			name:           "Normalize EUR amount",
			amount:         decimal.RequireFromString("99.999"),
			currencyCode:   "EUR",
			expectedAmount: "100", // Rounded to 2 decimal places but String() doesn't show trailing zeros
			expectError:    false,
		},
		{
			name:         "Invalid currency",
			amount:       decimal.RequireFromString("123.456"),
			currencyCode: "INVALID",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			result, err := service.NormalizeCurrencyAmount(tt.amount, tt.currencyCode)

			if tt.expectError {
				if err == nil {
					t.Error("NormalizeCurrencyAmount() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NormalizeCurrencyAmount() error = %v", err)
				return
			}

			if result.String() != tt.expectedAmount {
				t.Errorf("NormalizeCurrencyAmount() = %s, want %s", result.String(), tt.expectedAmount)
			}
		})
	}
}

func TestGetCurrencySymbol(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name           string
		currencyCode   string
		expectedSymbol string
		expectError    bool
	}{
		{
			name:           "Get USD symbol",
			currencyCode:   "USD",
			expectedSymbol: "$",
			expectError:    false,
		},
		{
			name:           "Get EUR symbol",
			currencyCode:   "EUR",
			expectedSymbol: "€",
			expectError:    false,
		},
		{
			name:           "Get JPY symbol",
			currencyCode:   "JPY",
			expectedSymbol: "¥",
			expectError:    false,
		},
		{
			name:         "Invalid currency",
			currencyCode: "INVALID",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			result, err := service.GetCurrencySymbol(tt.currencyCode)

			if tt.expectError {
				if err == nil {
					t.Error("GetCurrencySymbol() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetCurrencySymbol() error = %v", err)
				return
			}

			if result != tt.expectedSymbol {
				t.Errorf("GetCurrencySymbol() = %s, want %s", result, tt.expectedSymbol)
			}
		})
	}
}

func TestFormatCurrencyAmount(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name           string
		amount         decimal.Decimal
		currencyCode   string
		expectedFormat string
		expectError    bool
	}{
		{
			name:           "Format USD amount",
			amount:         decimal.RequireFromString("123.456"),
			currencyCode:   "USD",
			expectedFormat: "$123.46",
			expectError:    false,
		},
		{
			name:           "Format JPY amount",
			amount:         decimal.RequireFromString("123.456"),
			currencyCode:   "JPY",
			expectedFormat: "¥123",
			expectError:    false,
		},
		{
			name:           "Format EUR amount",
			amount:         decimal.RequireFromString("99.999"),
			currencyCode:   "EUR",
			expectedFormat: "€100",
			expectError:    false,
		},
		{
			name:         "Invalid currency",
			amount:       decimal.RequireFromString("123.456"),
			currencyCode: "INVALID",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			result, err := service.FormatCurrencyAmount(tt.amount, tt.currencyCode)

			if tt.expectError {
				if err == nil {
					t.Error("FormatCurrencyAmount() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FormatCurrencyAmount() error = %v", err)
				return
			}

			if result != tt.expectedFormat {
				t.Errorf("FormatCurrencyAmount() = %s, want %s", result, tt.expectedFormat)
			}
		})
	}
}

func TestGetCurrencyInfo(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name         string
		currencyCode string
		expectError  bool
		expectedCode string
		expectedName string
	}{
		{
			name:         "Get USD info",
			currencyCode: "USD",
			expectError:  false,
			expectedCode: "USD",
			expectedName: "US Dollar",
		},
		{
			name:         "Get EUR info",
			currencyCode: "EUR",
			expectError:  false,
			expectedCode: "EUR",
			expectedName: "Euro",
		},
		{
			name:         "Get BTC info (inactive)",
			currencyCode: "BTC",
			expectError:  false,
			expectedCode: "BTC",
			expectedName: "Bitcoin",
		},
		{
			name:         "Invalid currency",
			currencyCode: "INVALID",
			expectError:  true,
		},
		{
			name:         "Case insensitive",
			currencyCode: "usd",
			expectError:  false,
			expectedCode: "USD",
			expectedName: "US Dollar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			result, err := service.getCurrencyInfo(tt.currencyCode)

			if tt.expectError {
				if err == nil {
					t.Error("getCurrencyInfo() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("getCurrencyInfo() error = %v", err)
				return
			}

			if result.Code != tt.expectedCode {
				t.Errorf("getCurrencyInfo() code = %s, want %s", result.Code, tt.expectedCode)
			}

			if result.Name != tt.expectedName {
				t.Errorf("getCurrencyInfo() name = %s, want %s", result.Name, tt.expectedName)
			}
		})
	}
}
