package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// createTestLogger creates a test logger instance
var testLogger = func() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	return logger
}()

func TestValidateCreditAccountParams(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := &Service{logger: testLogger}

	testUUID := uuid.New()
	testAccountNumber := "ACC123456"
	emptyAccountNumber := ""

	tests := []struct {
		name        string
		params      CreditAccountParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid params with AccountID",
			params: CreditAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with AccountNumber",
			params: CreditAccountParams{
				AccountNumber: &testAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with all optional fields",
			params: CreditAccountParams{
				AccountID:      &testUUID,
				Amount:         decimal.NewFromFloat(100.0),
				Currency:       "USD",
				Description:    stringPtr("Test credit"),
				ReferenceID:    stringPtr("REF123"),
				IdempotencyKey: stringPtr("IDEM123"),
				Metadata:       map[string]any{"key": "value"},
			},
			expectError: false,
		},
		{
			name: "Missing both AccountID and AccountNumber",
			params: CreditAccountParams{
				Amount:   decimal.NewFromFloat(100.0),
				Currency: "USD",
			},
			expectError: true,
			errorMsg:    "either account_id or account_number must be provided",
		},
		{
			name: "Both AccountID and AccountNumber provided",
			params: CreditAccountParams{
				AccountID:     &testUUID,
				AccountNumber: &testAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: true,
			errorMsg:    "only one of account_id or account_number should be provided",
		},
		{
			name: "Empty AccountID",
			params: CreditAccountParams{
				AccountID: &uuid.Nil,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "account_id cannot be empty",
		},
		{
			name: "Empty AccountNumber",
			params: CreditAccountParams{
				AccountNumber: &emptyAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: true,
			errorMsg:    "account_number cannot be empty",
		},
		{
			name: "Zero amount",
			params: CreditAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.Zero,
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "amount cannot be zero",
		},
		{
			name: "Negative amount",
			params: CreditAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(-100.0),
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
		{
			name: "Empty currency",
			params: CreditAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "",
			},
			expectError: true,
			errorMsg:    "currency cannot be empty",
		},
		{
			name: "Invalid currency format",
			params: CreditAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "US",
			},
			expectError: true,
			errorMsg:    "currency must be a 3-letter code",
		},
		{
			name: "Empty idempotency key when provided",
			params: CreditAccountParams{
				AccountID:      &testUUID,
				Amount:         decimal.NewFromFloat(100.0),
				Currency:       "USD",
				IdempotencyKey: stringPtr(""),
			},
			expectError: true,
			errorMsg:    "idempotency_key cannot be empty when provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			err := service.validateCreditAccountParams(tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("validateCreditAccountParams() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("validateCreditAccountParams() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateCreditAccountParams() error = %v", err)
				}
			}
		})
	}
}

func TestCreditAccountUtilityFunctions(t *testing.T) {
	t.Parallel()

	service := &Service{logger: testLogger}

	t.Run("decimalToPgNumeric", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			input decimal.Decimal
		}{
			{
				name:  "Positive decimal",
				input: decimal.NewFromFloat(123.45),
			},
			{
				name:  "Zero",
				input: decimal.Zero,
			},
			{
				name:  "Large number",
				input: decimal.NewFromFloat(999999.9999),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				result, err := service.decimalToPgNumeric(tt.input)

				if err != nil {
					t.Errorf("decimalToPgNumeric() error = %v", err)
					return
				}

				if !result.Valid {
					t.Error("decimalToPgNumeric() result should be valid")
				}

				// Convert back to verify
				converted, err := service.pgNumericToDecimal(result)
				if err != nil {
					t.Errorf("pgNumericToDecimal() error = %v", err)
					return
				}

				if !converted.Equal(tt.input) {
					t.Errorf("Round trip conversion failed: got %v, want %v", converted, tt.input)
				}
			})
		}
	})

	t.Run("pgNumericToDecimal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			input       pgtype.Numeric
			expectError bool
			expected    string
		}{
			{
				name: "Valid positive decimal",
				input: pgtype.Numeric{
					Int:   decimal.NewFromFloat(123.45).Coefficient(),
					Exp:   int32(decimal.NewFromFloat(123.45).Exponent()),
					Valid: true,
				},
				expectError: false,
				expected:    "123.45",
			},
			{
				name: "Valid zero",
				input: pgtype.Numeric{
					Int:   decimal.Zero.Coefficient(),
					Exp:   int32(decimal.Zero.Exponent()),
					Valid: true,
				},
				expectError: false,
				expected:    "0",
			},
			{
				name: "Invalid numeric",
				input: pgtype.Numeric{
					Valid: false,
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				result, err := service.pgNumericToDecimal(tt.input)

				if tt.expectError {
					if err == nil {
						t.Error("pgNumericToDecimal() expected error but got none")
					}
					return
				}

				if err != nil {
					t.Errorf("pgNumericToDecimal() error = %v", err)
					return
				}

				if result.String() != tt.expected {
					t.Errorf("pgNumericToDecimal() = %s, want %s", result.String(), tt.expected)
				}
			})
		}
	})
}

func TestResolveCreditAccountID(t *testing.T) {
	t.Parallel()

	service := &Service{logger: testLogger}
	ctx := context.Background()

	testUUID := uuid.New()

	t.Run("AccountID provided", func(t *testing.T) {
		t.Parallel()

		params := CreditAccountParams{
			AccountID: &testUUID,
		}

		result, err := service.resolveCreditAccountID(ctx, params)

		if err != nil {
			t.Errorf("resolveCreditAccountID() error = %v", err)
			return
		}

		if result != testUUID {
			t.Errorf("resolveCreditAccountID() = %v, want %v", result, testUUID)
		}
	})

	// Note: Testing with AccountNumber would require mocking the store
	// This is left as a TODO for when we implement proper mocking
}

// Helper functions for tests

func creditDecimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func createCreditPgNumeric(value string) pgtype.Numeric {
	d, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}

	return pgtype.Numeric{
		Int:   d.Coefficient(),
		Exp:   int32(d.Exponent()),
		Valid: true,
	}
}
