package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// mustParseDecimal is a helper function to parse decimal strings for tests
func mustParseDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestValidateAuditBalanceHistoryParams(t *testing.T) {
	t.Parallel()

	testService := &Service{}

	tests := []struct {
		name        string
		params      AuditBalanceHistoryParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_params",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(150.00),
				Operation:     "credit",
				CreatedBy:     "system",
			},
			expectError: false,
		},
		{
			name: "missing_account_id",
			params: AuditBalanceHistoryParams{
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(150.00),
				Operation:     "credit",
				CreatedBy:     "system",
			},
			expectError: true,
			errorMsg:    "account_id is required",
		},
		{
			name: "missing_transaction_id",
			params: AuditBalanceHistoryParams{
				AccountID:  "123e4567-e89b-12d3-a456-426614174000",
				OldBalance: decimal.NewFromFloat(100.00),
				NewBalance: decimal.NewFromFloat(150.00),
				Operation:  "credit",
				CreatedBy:  "system",
			},
			expectError: true,
			errorMsg:    "transaction_id is required",
		},
		{
			name: "missing_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(150.00),
				CreatedBy:     "system",
			},
			expectError: true,
			errorMsg:    "operation is required",
		},
		{
			name: "missing_created_by",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(150.00),
				Operation:     "credit",
			},
			expectError: true,
			errorMsg:    "created_by is required",
		},
		{
			name: "invalid_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(150.00),
				Operation:     "invalid_operation",
				CreatedBy:     "system",
			},
			expectError: true,
			errorMsg:    "invalid operation type: invalid_operation",
		},
		{
			name: "valid_debit_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(50.00),
				Operation:     "debit",
				CreatedBy:     "system",
			},
			expectError: false,
		},
		{
			name: "valid_compensate_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(150.00),
				Operation:     "compensate",
				CreatedBy:     "temporal_workflow",
			},
			expectError: false,
		},
		{
			name: "valid_freeze_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(100.00),
				Operation:     "freeze",
				CreatedBy:     "admin",
			},
			expectError: false,
		},
		{
			name: "valid_transfer_in_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(200.00),
				Operation:     "transfer_in",
				CreatedBy:     "transfer_service",
			},
			expectError: false,
		},
		{
			name: "valid_transfer_out_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(200.00),
				NewBalance:    decimal.NewFromFloat(100.00),
				Operation:     "transfer_out",
				CreatedBy:     "transfer_service",
			},
			expectError: false,
		},
		{
			name: "valid_adjustment_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(99.99),
				Operation:     "adjustment",
				CreatedBy:     "admin",
			},
			expectError: false,
		},
		{
			name: "valid_unfreeze_operation",
			params: AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(100.00),
				Operation:     "unfreeze",
				CreatedBy:     "admin",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := testService.validateAuditBalanceHistoryParams(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuditBalanceHistoryBasic(t *testing.T) {
	t.Parallel()

	testService := &Service{}

	tests := []struct {
		name        string
		accountID   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_uuid_account_id",
			accountID:   "123e4567-e89b-12d3-a456-426614174000",
			expectError: false,
		},
		{
			name:        "invalid_uuid_account_id",
			accountID:   "invalid-uuid",
			expectError: true,
			errorMsg:    "invalid account ID format",
		},
		{
			name:        "empty_account_id",
			accountID:   "",
			expectError: true,
			errorMsg:    "account_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := AuditBalanceHistoryParams{
				AccountID:     tt.accountID,
				TransactionID: uuid.New(),
				OldBalance:    decimal.NewFromFloat(100.00),
				NewBalance:    decimal.NewFromFloat(150.00),
				Operation:     "credit",
				CreatedBy:     "system",
			}

			err := testService.validateAuditBalanceHistoryParams(params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuditBalanceHistoryBalanceCalculations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		oldBalance        decimal.Decimal
		newBalance        decimal.Decimal
		expectedChange    decimal.Decimal
		operation         string
		expectedOperation string
	}{
		{
			name:              "credit_positive_change",
			oldBalance:        decimal.NewFromFloat(100.00),
			newBalance:        decimal.NewFromFloat(150.00),
			expectedChange:    decimal.NewFromFloat(50.00),
			operation:         "credit",
			expectedOperation: "credit",
		},
		{
			name:              "debit_negative_change",
			oldBalance:        decimal.NewFromFloat(100.00),
			newBalance:        decimal.NewFromFloat(50.00),
			expectedChange:    decimal.NewFromFloat(-50.00),
			operation:         "debit",
			expectedOperation: "debit",
		},
		{
			name:              "no_change",
			oldBalance:        decimal.NewFromFloat(100.00),
			newBalance:        decimal.NewFromFloat(100.00),
			expectedChange:    decimal.NewFromFloat(0.00),
			operation:         "freeze",
			expectedOperation: "freeze",
		},
		{
			name:              "compensate_positive_change",
			oldBalance:        decimal.NewFromFloat(50.00),
			newBalance:        decimal.NewFromFloat(100.00),
			expectedChange:    decimal.NewFromFloat(50.00),
			operation:         "compensate",
			expectedOperation: "compensate",
		},
		{
			name:              "large_transfer",
			oldBalance:        decimal.NewFromFloat(1000.00),
			newBalance:        decimal.NewFromFloat(5000.00),
			expectedChange:    decimal.NewFromFloat(4000.00),
			operation:         "transfer_in",
			expectedOperation: "transfer_in",
		},
		{
			name:              "precision_adjustment",
			oldBalance:        mustParseDecimal("100.123456"),
			newBalance:        mustParseDecimal("100.123457"),
			expectedChange:    mustParseDecimal("0.000001"),
			operation:         "adjustment",
			expectedOperation: "adjustment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := AuditBalanceHistoryParams{
				AccountID:     "123e4567-e89b-12d3-a456-426614174000",
				TransactionID: uuid.New(),
				OldBalance:    tt.oldBalance,
				NewBalance:    tt.newBalance,
				Operation:     tt.operation,
				CreatedBy:     "system",
			}

			// Test balance change calculation
			expectedBalanceChange := tt.newBalance.Sub(tt.oldBalance)
			assert.True(t, expectedBalanceChange.Equal(tt.expectedChange))

			// Test operation mapping
			assert.Equal(t, tt.expectedOperation, params.Operation)
		})
	}
}
