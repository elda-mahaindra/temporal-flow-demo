package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestValidateTransferWorkflowParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		params      TransferWorkflowParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_params",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "USD",
				Description:    "Test transfer",
				IdempotencyKey: "idempotency-123",
				RequestedBy:    "user-123",
			},
			expectError: false,
		},
		{
			name: "missing_transfer_id",
			params: TransferWorkflowParams{
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "transfer_id is required",
		},
		{
			name: "missing_from_account",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "from_account is required",
		},
		{
			name: "missing_to_account",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "to_account is required",
		},
		{
			name: "same_from_and_to_account",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "same-account",
				ToAccount:      "same-account",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "from_account and to_account cannot be the same",
		},
		{
			name: "zero_amount",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.Zero,
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "amount must be positive",
		},
		{
			name: "negative_amount",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(-50.00),
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "amount must be positive",
		},
		{
			name: "missing_currency",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "currency is required",
		},
		{
			name: "invalid_currency",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "INVALID",
				IdempotencyKey: "idempotency-123",
			},
			expectError: true,
			errorMsg:    "unsupported currency: INVALID",
		},
		{
			name: "missing_idempotency_key",
			params: TransferWorkflowParams{
				TransferID:  "transfer-123",
				FromAccount: "account-from",
				ToAccount:   "account-to",
				Amount:      decimal.NewFromFloat(100.00),
				Currency:    "USD",
			},
			expectError: true,
			errorMsg:    "idempotency_key is required",
		},
		{
			name: "valid_eur_currency",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "EUR",
				IdempotencyKey: "idempotency-123",
			},
			expectError: false,
		},
		{
			name: "valid_gbp_currency",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "GBP",
				IdempotencyKey: "idempotency-123",
			},
			expectError: false,
		},
		{
			name: "valid_jpy_currency",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(10000),
				Currency:       "JPY",
				IdempotencyKey: "idempotency-123",
			},
			expectError: false,
		},
		{
			name: "valid_large_amount",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(1000000.50),
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: false,
		},
		{
			name: "valid_small_amount",
			params: TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(0.01),
				Currency:       "USD",
				IdempotencyKey: "idempotency-123",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateTransferWorkflowParams(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransferWorkflowStructures(t *testing.T) {
	t.Parallel()

	t.Run("transfer_workflow_params_structure", func(t *testing.T) {
		params := TransferWorkflowParams{
			TransferID:     "test-transfer-123",
			FromAccount:    "from-account-456",
			ToAccount:      "to-account-789",
			Amount:         decimal.NewFromFloat(250.75),
			Currency:       "USD",
			Description:    "Test payment",
			IdempotencyKey: "idempotency-key-abc",
			RequestedBy:    "user-xyz",
		}

		assert.Equal(t, "test-transfer-123", params.TransferID)
		assert.Equal(t, "from-account-456", params.FromAccount)
		assert.Equal(t, "to-account-789", params.ToAccount)
		assert.True(t, params.Amount.Equal(decimal.NewFromFloat(250.75)))
		assert.Equal(t, "USD", params.Currency)
		assert.Equal(t, "Test payment", params.Description)
		assert.Equal(t, "idempotency-key-abc", params.IdempotencyKey)
		assert.Equal(t, "user-xyz", params.RequestedBy)
	})

	t.Run("transfer_workflow_results_structure", func(t *testing.T) {
		results := TransferWorkflowResults{
			TransferID:          "test-transfer-123",
			Status:              "completed",
			FromAccount:         "from-account-456",
			ToAccount:           "to-account-789",
			Amount:              decimal.NewFromFloat(250.75),
			Currency:            "USD",
			Description:         "Test payment",
			CompensationApplied: false,
			WorkflowID:          "workflow-id-123",
			RunID:               "run-id-456",
		}

		assert.Equal(t, "test-transfer-123", results.TransferID)
		assert.Equal(t, "completed", results.Status)
		assert.Equal(t, "from-account-456", results.FromAccount)
		assert.Equal(t, "to-account-789", results.ToAccount)
		assert.True(t, results.Amount.Equal(decimal.NewFromFloat(250.75)))
		assert.Equal(t, "USD", results.Currency)
		assert.Equal(t, "Test payment", results.Description)
		assert.False(t, results.CompensationApplied)
		assert.Equal(t, "workflow-id-123", results.WorkflowID)
		assert.Equal(t, "run-id-456", results.RunID)
	})
}

func TestCurrencyValidation(t *testing.T) {
	t.Parallel()

	validCurrencies := []string{"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "CNY", "SGD", "HKD"}
	invalidCurrencies := []string{"", "INVALID", "usd", "Bitcoin", "XYZ", "123"}

	for _, currency := range validCurrencies {
		t.Run("valid_currency_"+currency, func(t *testing.T) {
			params := TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       currency,
				IdempotencyKey: "idempotency-123",
			}

			err := validateTransferWorkflowParams(params)
			assert.NoError(t, err)
		})
	}

	for _, currency := range invalidCurrencies {
		t.Run("invalid_currency_"+currency, func(t *testing.T) {
			params := TransferWorkflowParams{
				TransferID:     "transfer-123",
				FromAccount:    "account-from",
				ToAccount:      "account-to",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       currency,
				IdempotencyKey: "idempotency-123",
			}

			err := validateTransferWorkflowParams(params)
			assert.Error(t, err)
			if currency == "" {
				assert.Contains(t, err.Error(), "currency is required")
			} else {
				assert.Contains(t, err.Error(), "unsupported currency")
			}
		})
	}
}

func TestWorkflowParameterEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("very_large_amount", func(t *testing.T) {
		amount, _ := decimal.NewFromString("999999999999.99")
		params := TransferWorkflowParams{
			TransferID:     "transfer-123",
			FromAccount:    "account-from",
			ToAccount:      "account-to",
			Amount:         amount,
			Currency:       "USD",
			IdempotencyKey: "idempotency-123",
		}

		err := validateTransferWorkflowParams(params)
		assert.NoError(t, err)
	})

	t.Run("very_small_amount", func(t *testing.T) {
		amount, _ := decimal.NewFromString("0.000001")
		params := TransferWorkflowParams{
			TransferID:     "transfer-123",
			FromAccount:    "account-from",
			ToAccount:      "account-to",
			Amount:         amount,
			Currency:       "USD",
			IdempotencyKey: "idempotency-123",
		}

		err := validateTransferWorkflowParams(params)
		assert.NoError(t, err)
	})

	t.Run("long_account_names", func(t *testing.T) {
		longAccountName := "very-long-account-name-that-might-be-used-in-some-systems-with-detailed-naming-conventions"
		params := TransferWorkflowParams{
			TransferID:     "transfer-123",
			FromAccount:    longAccountName + "-from",
			ToAccount:      longAccountName + "-to",
			Amount:         decimal.NewFromFloat(100.00),
			Currency:       "USD",
			IdempotencyKey: "idempotency-123",
		}

		err := validateTransferWorkflowParams(params)
		assert.NoError(t, err)
	})

	t.Run("unicode_description", func(t *testing.T) {
		params := TransferWorkflowParams{
			TransferID:     "transfer-123",
			FromAccount:    "account-from",
			ToAccount:      "account-to",
			Amount:         decimal.NewFromFloat(100.00),
			Currency:       "USD",
			Description:    "Payment for 商品 with émojis 🏦💰",
			IdempotencyKey: "idempotency-123",
		}

		err := validateTransferWorkflowParams(params)
		assert.NoError(t, err)
	})
}
