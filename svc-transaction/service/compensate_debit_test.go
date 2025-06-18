package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestValidateCompensateDebitParams(t *testing.T) {
	t.Parallel()

	service := &Service{logger: testLogger}

	testUUID := uuid.New()
	testAccountNumber := "ACC123456"
	emptyAccountNumber := ""
	testReferenceID := "REF123"
	emptyReferenceID := ""

	tests := []struct {
		name        string
		params      CompensateDebitParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid params with original transaction ID",
			params: CompensateDebitParams{
				OriginalTransactionID: &testUUID,
				Amount:                decimal.NewFromFloat(100.0),
				Currency:              "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with original reference ID",
			params: CompensateDebitParams{
				OriginalReferenceID: &testReferenceID,
				Amount:              decimal.NewFromFloat(100.0),
				Currency:            "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with account ID",
			params: CompensateDebitParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with account number",
			params: CompensateDebitParams{
				AccountNumber: &testAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: false,
		},
		{
			name: "Valid params with all fields",
			params: CompensateDebitParams{
				OriginalTransactionID: &testUUID,
				AccountID:             &testUUID,
				Amount:                decimal.NewFromFloat(100.0),
				Currency:              "USD",
				Description:           stringPtr("Test compensation"),
				ReferenceID:           stringPtr("REF123"),
				IdempotencyKey:        stringPtr("IDEM123"),
				CompensationReason:    stringPtr("Transaction rollback"),
				WorkflowID:            stringPtr("WF123"),
				RunID:                 stringPtr("RUN123"),
			},
			expectError: false,
		},
		{
			name: "Missing all identifiers",
			params: CompensateDebitParams{
				Amount:   decimal.NewFromFloat(100.0),
				Currency: "USD",
			},
			expectError: true,
			errorMsg:    "either original transaction information or account information must be provided",
		},
		{
			name: "Empty original transaction ID",
			params: CompensateDebitParams{
				OriginalTransactionID: &uuid.Nil,
				Amount:                decimal.NewFromFloat(100.0),
				Currency:              "USD",
			},
			expectError: true,
			errorMsg:    "original_transaction_id cannot be empty",
		},
		{
			name: "Empty original reference ID",
			params: CompensateDebitParams{
				OriginalReferenceID: &emptyReferenceID,
				Amount:              decimal.NewFromFloat(100.0),
				Currency:            "USD",
			},
			expectError: true,
			errorMsg:    "original_reference_id cannot be empty when provided",
		},
		{
			name: "Empty account ID",
			params: CompensateDebitParams{
				AccountID: &uuid.Nil,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "account_id cannot be empty",
		},
		{
			name: "Empty account number",
			params: CompensateDebitParams{
				AccountNumber: &emptyAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: true,
			errorMsg:    "account_number cannot be empty when provided",
		},
		{
			name: "Both account ID and account number provided",
			params: CompensateDebitParams{
				AccountID:     &testUUID,
				AccountNumber: &testAccountNumber,
				Amount:        decimal.NewFromFloat(100.0),
				Currency:      "USD",
			},
			expectError: true,
			errorMsg:    "only one of account_id or account_number should be provided",
		},
		{
			name: "Zero amount",
			params: CompensateDebitParams{
				AccountID: &testUUID,
				Amount:    decimal.Zero,
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "amount cannot be zero",
		},
		{
			name: "Negative amount",
			params: CompensateDebitParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(-100.0),
				Currency:  "USD",
			},
			expectError: true,
			errorMsg:    "amount cannot be negative",
		},
		{
			name: "Empty currency",
			params: CompensateDebitParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "",
			},
			expectError: true,
			errorMsg:    "currency cannot be empty",
		},
		{
			name: "Invalid currency format",
			params: CompensateDebitParams{
				AccountID: &testUUID,
				Amount:    decimal.NewFromFloat(100.0),
				Currency:  "US",
			},
			expectError: true,
			errorMsg:    "currency must be a 3-letter code",
		},
		{
			name: "Empty idempotency key when provided",
			params: CompensateDebitParams{
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
			t.Parallel()

			err := service.validateCompensateDebitParams(tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("validateCompensateDebitParams() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("validateCompensateDebitParams() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateCompensateDebitParams() error = %v", err)
				}
			}
		})
	}
}

func TestMapCurrencyToEnum(t *testing.T) {
	t.Parallel()

	service := &Service{logger: testLogger}

	tests := []struct {
		name     string
		currency string
		expected string
	}{
		{"USD", "USD", "USD"},
		{"EUR", "EUR", "EUR"},
		{"GBP", "GBP", "GBP"},
		{"JPY", "JPY", "JPY"},
		{"CAD", "CAD", "CAD"},
		{"AUD", "AUD", "AUD"},
		{"CHF", "CHF", "CHF"},
		{"CNY", "CNY", "CNY"},
		{"SGD", "SGD", "SGD"},
		{"HKD", "HKD", "HKD"},
		{"Unknown currency defaults to USD", "XYZ", "USD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := service.mapCurrencyToEnum(tt.currency)

			if string(result) != tt.expected {
				t.Errorf("mapCurrencyToEnum(%s) = %s, want %s", tt.currency, result, tt.expected)
			}
		})
	}
}

func TestResolveCompensationAccountID(t *testing.T) {
	t.Parallel()

	service := &Service{logger: testLogger}
	ctx := context.Background()

	testUUID := uuid.New()

	t.Run("AccountID provided", func(t *testing.T) {
		t.Parallel()

		params := CompensateDebitParams{
			AccountID: &testUUID,
		}

		result, err := service.resolveCompensationAccountID(ctx, params, nil)

		if err != nil {
			t.Errorf("resolveCompensationAccountID() error = %v", err)
			return
		}

		if result != testUUID {
			t.Errorf("resolveCompensationAccountID() = %v, want %v", result, testUUID)
		}
	})

	// Note: Testing with AccountNumber and OriginalTransaction would require mocking the store
	// This is left as a TODO for when we implement proper mocking
}
