package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// createTestService creates a test service instance
func createTestService() *Service {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	return &Service{
		logger: logger,
		// store will be mocked for unit tests
	}
}

func TestValidateDebitAccountParams(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	testUUID := uuid.New()
	testAccountNumber := "ACC123456"
	emptyAccountNumber := ""

	tests := []struct {
		name        string
		params      DebitAccountParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid params with AccountID",
			params: DebitAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with AccountNumber",
			params: DebitAccountParams{
				AccountNumber: &testAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with all optional fields",
			params: DebitAccountParams{
				AccountID:      &testUUID,
				Amount:         decimal.NewFromFloat(100.0),
				Currency:       "USD",
				Description:    stringPtr("Test debit"),
				ReferenceID:    stringPtr("REF123"),
				IdempotencyKey: stringPtr("IDEM123"),
				Metadata:       map[string]any{"key": "value"},
			},
			expectError: false,
		},
		{
			name: "Missing both AccountID and AccountNumber",
			params: DebitAccountParams{
				Amount:   decimal.NewFromFloat(100.0),
				Currency: "USD",
			},
			expectError: true,
			errorMsg:    "either account_id or account_number must be provided",
		},
		{
			name: "Both AccountID and AccountNumber provided",
			params: DebitAccountParams{
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
			params: DebitAccountParams{
				AccountID: &uuid.Nil,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "account_id cannot be empty",
		},
		{
			name: "Empty AccountNumber",
			params: DebitAccountParams{
				AccountNumber: &emptyAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: true,
			errorMsg:    "account_number cannot be empty",
		},
		{
			name: "Zero amount",
			params: DebitAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.Zero,
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "amount cannot be zero",
		},
		{
			name: "Negative amount",
			params: DebitAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(-100.0),
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
		{
			name: "Empty currency",
			params: DebitAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "",
			},
			expectError: true,
			errorMsg:    "currency cannot be empty",
		},
		{
			name: "Invalid currency format",
			params: DebitAccountParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "US",
			},
			expectError: true,
			errorMsg:    "currency must be a 3-letter code",
		},
		{
			name: "Empty idempotency key when provided",
			params: DebitAccountParams{
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

			err := service.validateDebitAccountParams(tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("validateDebitAccountParams() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("validateDebitAccountParams() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateDebitAccountParams() error = %v", err)
				}
			}
		})
	}
}

func TestDecimalToPgNumeric(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

	tests := []struct {
		name     string
		input    decimal.Decimal
		expected pgtype.Numeric
	}{
		{
			name:  "Positive decimal",
			input: decimal.NewFromFloat(123.45),
			expected: pgtype.Numeric{
				Int:   decimal.NewFromFloat(123.45).Coefficient(),
				Exp:   int32(decimal.NewFromFloat(123.45).Exponent()),
				Valid: true,
			},
		},
		{
			name:  "Zero",
			input: decimal.Zero,
			expected: pgtype.Numeric{
				Int:   decimal.Zero.Coefficient(),
				Exp:   int32(decimal.Zero.Exponent()),
				Valid: true,
			},
		},
		{
			name:  "Large number",
			input: decimal.NewFromFloat(999999.9999),
			expected: pgtype.Numeric{
				Int:   decimal.NewFromFloat(999999.9999).Coefficient(),
				Exp:   int32(decimal.NewFromFloat(999999.9999).Exponent()),
				Valid: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Allow subtests to run parallel with each other

			result, err := service.decimalToPgNumeric(tt.input)

			if err != nil {
				t.Errorf("decimalToPgNumeric() error = %v", err)
				return
			}

			if !result.Valid {
				t.Error("decimalToPgNumeric() result should be valid")
			}

			if result.Int.Cmp(tt.expected.Int) != 0 {
				t.Errorf("decimalToPgNumeric() Int = %v, want %v", result.Int, tt.expected.Int)
			}

			if result.Exp != tt.expected.Exp {
				t.Errorf("decimalToPgNumeric() Exp = %v, want %v", result.Exp, tt.expected.Exp)
			}
		})
	}
}

func TestPgNumericToDecimal(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	service := createTestService()

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
			t.Parallel() // Allow subtests to run parallel with each other

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
}

func TestResolveAccountID(t *testing.T) {
	t.Parallel() // Allow this test function to run parallel with others

	// Note: This test would require mocking the store
	// For now, we'll test the basic logic without database calls

	service := createTestService()
	ctx := context.Background()

	testUUID := uuid.New()

	t.Run("AccountID provided", func(t *testing.T) {
		t.Parallel()

		params := DebitAccountParams{
			AccountID: &testUUID,
		}

		result, err := service.resolveAccountID(ctx, params)

		if err != nil {
			t.Errorf("resolveAccountID() error = %v", err)
			return
		}

		if result != testUUID {
			t.Errorf("resolveAccountID() = %v, want %v", result, testUUID)
		}
	})

	// Note: Testing with AccountNumber would require mocking the store
	// This is left as a TODO for when we implement proper mocking
}

// Helper functions for tests

func stringPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func createPgNumeric(value string) pgtype.Numeric {
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
