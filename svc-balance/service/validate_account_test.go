package service

import (
	"testing"

	"svc-balance/store/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func TestValidateAccountValidationParams(t *testing.T) {
	t.Parallel()

	service := createTestService()

	// Create test UUIDs
	testUUID := uuid.New()
	nilUUID := uuid.Nil
	testAccountNumber := "ACC123456"
	emptyAccountNumber := ""
	testTransactionType := "debit"
	invalidTransactionType := "invalid"
	testAmount := decimal.NewFromFloat(100.0)
	negativeAmount := decimal.NewFromFloat(-100.0)
	testCurrency := "USD"

	tests := []struct {
		name        string
		params      ValidateAccountParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid params with AccountID",
			params: ValidateAccountParams{
				AccountID: &testUUID,
			},
			expectError: false,
		},
		{
			name: "Valid params with AccountNumber",
			params: ValidateAccountParams{
				AccountNumber: &testAccountNumber,
			},
			expectError: false,
		},
		{
			name: "Valid params with transaction details",
			params: ValidateAccountParams{
				AccountID:         &testUUID,
				TransactionType:   &testTransactionType,
				TransactionAmount: &testAmount,
				ExpectedCurrency:  &testCurrency,
			},
			expectError: false,
		},
		{
			name: "Missing both AccountID and AccountNumber",
			params: ValidateAccountParams{
				TransactionType: &testTransactionType,
			},
			expectError: true,
			errorMsg:    "either account_id or account_number must be provided",
		},
		{
			name: "Both AccountID and AccountNumber provided",
			params: ValidateAccountParams{
				AccountID:     &testUUID,
				AccountNumber: &testAccountNumber,
			},
			expectError: true,
			errorMsg:    "only one of account_id or account_number should be provided",
		},
		{
			name: "Empty AccountID",
			params: ValidateAccountParams{
				AccountID: &nilUUID,
			},
			expectError: true,
			errorMsg:    "account_id cannot be empty",
		},
		{
			name: "Empty AccountNumber",
			params: ValidateAccountParams{
				AccountNumber: &emptyAccountNumber,
			},
			expectError: true,
			errorMsg:    "account_number cannot be empty",
		},
		{
			name: "Invalid transaction type",
			params: ValidateAccountParams{
				AccountID:       &testUUID,
				TransactionType: &invalidTransactionType,
			},
			expectError: true,
			errorMsg:    "invalid transaction_type: invalid",
		},
		{
			name: "Negative transaction amount",
			params: ValidateAccountParams{
				AccountID:         &testUUID,
				TransactionAmount: &negativeAmount,
			},
			expectError: true,
			errorMsg:    "transaction_amount cannot be negative",
		},
		{
			name: "Invalid expected currency",
			params: ValidateAccountParams{
				AccountID:        &testUUID,
				ExpectedCurrency: stringPtr("INVALID"),
			},
			expectError: true,
			errorMsg:    "invalid expected_currency: unsupported currency: INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := service.validateAccountValidationParams(tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("validateAccountValidationParams() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("validateAccountValidationParams() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateAccountValidationParams() error = %v", err)
				}
			}
		})
	}
}

func TestValidateAccountStatus(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name              string
		accountStatus     string
		expectValid       bool
		expectCanTransact bool
	}{
		{
			name:              "Active account",
			accountStatus:     "active",
			expectValid:       true,
			expectCanTransact: true,
		},
		{
			name:              "Inactive account",
			accountStatus:     "inactive",
			expectValid:       false,
			expectCanTransact: false,
		},
		{
			name:              "Suspended account",
			accountStatus:     "suspended",
			expectValid:       false,
			expectCanTransact: false,
		},
		{
			name:              "Closed account",
			accountStatus:     "closed",
			expectValid:       false,
			expectCanTransact: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := &ValidateAccountResults{
				IsValid:     true,
				CanTransact: true,
				Validations: []ValidationResult{},
			}

			account := sqlc.CoreAccount{
				Status: sqlc.CoreAccountStatus(tt.accountStatus),
			}

			service.validateAccountStatus(result, account)

			if result.IsValid != tt.expectValid {
				t.Errorf("validateAccountStatus() IsValid = %v, want %v", result.IsValid, tt.expectValid)
			}

			if result.CanTransact != tt.expectCanTransact {
				t.Errorf("validateAccountStatus() CanTransact = %v, want %v", result.CanTransact, tt.expectCanTransact)
			}

			// Check that validation result was added
			if len(result.Validations) != 1 {
				t.Errorf("validateAccountStatus() validations count = %d, want 1", len(result.Validations))
			}

			validation := result.Validations[0]
			if validation.Type != "status" {
				t.Errorf("validateAccountStatus() validation type = %s, want status", validation.Type)
			}

			if validation.Rule != "account_active" {
				t.Errorf("validateAccountStatus() validation rule = %s, want account_active", validation.Rule)
			}

			if validation.Passed != tt.expectValid {
				t.Errorf("validateAccountStatus() validation passed = %v, want %v", validation.Passed, tt.expectValid)
			}
		})
	}
}

func TestValidateAccountBalance(t *testing.T) {
	t.Parallel()

	service := createTestService()

	// Create test account with balance of 1000.00
	account := sqlc.CoreAccount{
		Balance: createPgNumeric("1000.00"),
	}

	tests := []struct {
		name              string
		transactionType   *string
		transactionAmount *decimal.Decimal
		expectValid       bool
		expectCanTransact bool
	}{
		{
			name:              "Sufficient funds for debit",
			transactionType:   stringPtr("debit"),
			transactionAmount: decimalPtr(decimal.NewFromFloat(500.0)),
			expectValid:       true,
			expectCanTransact: true,
		},
		{
			name:              "Insufficient funds for debit",
			transactionType:   stringPtr("debit"),
			transactionAmount: decimalPtr(decimal.NewFromFloat(1500.0)),
			expectValid:       false,
			expectCanTransact: false,
		},
		{
			name:              "Credit transaction (no balance check)",
			transactionType:   stringPtr("credit"),
			transactionAmount: decimalPtr(decimal.NewFromFloat(500.0)),
			expectValid:       true,
			expectCanTransact: true,
		},
		{
			name:              "No transaction type (balance validation only)",
			transactionType:   nil,
			transactionAmount: nil,
			expectValid:       true,
			expectCanTransact: true,
		},
		{
			name:              "Exact balance debit",
			transactionType:   stringPtr("debit"),
			transactionAmount: decimalPtr(decimal.NewFromFloat(1000.0)),
			expectValid:       true,
			expectCanTransact: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := &ValidateAccountResults{
				IsValid:     true,
				CanTransact: true,
				Validations: []ValidationResult{},
			}

			params := ValidateAccountParams{
				TransactionType:   tt.transactionType,
				TransactionAmount: tt.transactionAmount,
			}

			err := service.validateAccountBalance(result, account, params)
			if err != nil {
				t.Errorf("validateAccountBalance() error = %v", err)
				return
			}

			if result.IsValid != tt.expectValid {
				t.Errorf("validateAccountBalance() IsValid = %v, want %v", result.IsValid, tt.expectValid)
			}

			if result.CanTransact != tt.expectCanTransact {
				t.Errorf("validateAccountBalance() CanTransact = %v, want %v", result.CanTransact, tt.expectCanTransact)
			}

			// Check that validation result was added
			if len(result.Validations) != 1 {
				t.Errorf("validateAccountBalance() validations count = %d, want 1", len(result.Validations))
			}

			validation := result.Validations[0]
			if validation.Type != "balance" {
				t.Errorf("validateAccountBalance() validation type = %s, want balance", validation.Type)
			}

			if validation.Passed != tt.expectValid {
				t.Errorf("validateAccountBalance() validation passed = %v, want %v", validation.Passed, tt.expectValid)
			}
		})
	}
}

func TestValidateAccountCurrency(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name              string
		accountCurrency   string
		expectedCurrency  *string
		expectValid       bool
		expectCanTransact bool
	}{
		{
			name:              "Valid supported currency",
			accountCurrency:   "USD",
			expectedCurrency:  nil,
			expectValid:       true,
			expectCanTransact: true,
		},
		{
			name:              "Currency matches expected",
			accountCurrency:   "USD",
			expectedCurrency:  stringPtr("USD"),
			expectValid:       true,
			expectCanTransact: true,
		},
		{
			name:              "Currency doesn't match expected",
			accountCurrency:   "USD",
			expectedCurrency:  stringPtr("EUR"),
			expectValid:       false,
			expectCanTransact: false,
		},
		{
			name:              "Unsupported currency",
			accountCurrency:   "INVALID",
			expectedCurrency:  nil,
			expectValid:       false,
			expectCanTransact: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := &ValidateAccountResults{
				IsValid:     true,
				CanTransact: true,
				Validations: []ValidationResult{},
			}

			account := sqlc.CoreAccount{
				Currency: sqlc.CoreCurrencyCode(tt.accountCurrency),
			}

			params := ValidateAccountParams{
				ExpectedCurrency: tt.expectedCurrency,
			}

			service.validateAccountCurrency(result, account, params)

			if result.IsValid != tt.expectValid {
				t.Errorf("validateAccountCurrency() IsValid = %v, want %v", result.IsValid, tt.expectValid)
			}

			if result.CanTransact != tt.expectCanTransact {
				t.Errorf("validateAccountCurrency() CanTransact = %v, want %v", result.CanTransact, tt.expectCanTransact)
			}

			// Check that validation result was added
			if len(result.Validations) != 1 {
				t.Errorf("validateAccountCurrency() validations count = %d, want 1", len(result.Validations))
			}

			validation := result.Validations[0]
			if validation.Type != "currency" {
				t.Errorf("validateAccountCurrency() validation type = %s, want currency", validation.Type)
			}

			if validation.Passed != tt.expectValid {
				t.Errorf("validateAccountCurrency() validation passed = %v, want %v", validation.Passed, tt.expectValid)
			}
		})
	}
}

func TestValidateBusinessRules(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name              string
		accountVersion    int32
		accountName       string
		transactionAmount *decimal.Decimal
		expectValid       bool // Overall expectation for critical validations
	}{
		{
			name:              "Valid account with normal transaction",
			accountVersion:    1,
			accountName:       "John Doe",
			transactionAmount: decimalPtr(decimal.NewFromFloat(1000.0)),
			expectValid:       true,
		},
		{
			name:              "Valid account with large transaction",
			accountVersion:    1,
			accountName:       "John Doe",
			transactionAmount: decimalPtr(decimal.NewFromFloat(150000.0)),
			expectValid:       true, // Large transaction gets warning but doesn't fail validation
		},
		{
			name:              "Invalid version",
			accountVersion:    0,
			accountName:       "John Doe",
			transactionAmount: nil,
			expectValid:       true, // Version issue is warning, not error
		},
		{
			name:              "Short account name",
			accountVersion:    1,
			accountName:       "A",
			transactionAmount: nil,
			expectValid:       true, // Name issue is warning, not error
		},
		{
			name:              "Empty account name",
			accountVersion:    1,
			accountName:       "",
			transactionAmount: nil,
			expectValid:       true, // Name issue is warning, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := &ValidateAccountResults{
				IsValid:     true,
				CanTransact: true,
				Validations: []ValidationResult{},
			}

			account := sqlc.CoreAccount{
				Version:     tt.accountVersion,
				AccountName: tt.accountName,
			}

			params := ValidateAccountParams{
				TransactionAmount: tt.transactionAmount,
			}

			err := service.validateBusinessRules(result, account, params)
			if err != nil {
				t.Errorf("validateBusinessRules() error = %v", err)
				return
			}

			// Business rules should not fail validation (they generate warnings)
			if !result.IsValid {
				t.Errorf("validateBusinessRules() should not fail validation, IsValid = %v", result.IsValid)
			}

			// Check that validation results were added
			expectedValidations := 2 // version + name
			if tt.transactionAmount != nil {
				expectedValidations = 3 // + transaction limits
			}

			if len(result.Validations) != expectedValidations {
				t.Errorf("validateBusinessRules() validations count = %d, want %d", len(result.Validations), expectedValidations)
			}

			// Check validation types
			validationTypes := make(map[string]bool)
			for _, validation := range result.Validations {
				if validation.Type != "business_rule" {
					t.Errorf("validateBusinessRules() validation type = %s, want business_rule", validation.Type)
				}
				validationTypes[validation.Rule] = true
			}

			if !validationTypes["version_consistency"] {
				t.Error("validateBusinessRules() missing version_consistency validation")
			}

			if !validationTypes["name_format"] {
				t.Error("validateBusinessRules() missing name_format validation")
			}

			if tt.transactionAmount != nil && !validationTypes["transaction_limits"] {
				t.Error("validateBusinessRules() missing transaction_limits validation")
			}
		})
	}
}

func TestUpdateValidationSummary(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name                    string
		validations             []ValidationResult
		expectValid             bool
		expectCanTransact       bool
		expectedSummaryContains string
	}{
		{
			name: "All validations passed",
			validations: []ValidationResult{
				{Type: "status", Passed: true, Severity: "info"},
				{Type: "balance", Passed: true, Severity: "info"},
			},
			expectValid:             true,
			expectCanTransact:       true,
			expectedSummaryContains: "All validations passed",
		},
		{
			name: "Some validations failed with errors",
			validations: []ValidationResult{
				{Type: "status", Passed: false, Severity: "error"},
				{Type: "balance", Passed: true, Severity: "info"},
			},
			expectValid:             false,
			expectCanTransact:       false,
			expectedSummaryContains: "failed with 1 errors",
		},
		{
			name: "Some validations failed with warnings only",
			validations: []ValidationResult{
				{Type: "status", Passed: true, Severity: "info"},
				{Type: "business_rule", Passed: false, Severity: "warning"},
			},
			expectValid:             true,
			expectCanTransact:       true,
			expectedSummaryContains: "passed with 1 warnings",
		},
		{
			name: "Mixed errors and warnings",
			validations: []ValidationResult{
				{Type: "status", Passed: false, Severity: "error"},
				{Type: "business_rule", Passed: false, Severity: "warning"},
				{Type: "balance", Passed: true, Severity: "info"},
			},
			expectValid:             false,
			expectCanTransact:       false,
			expectedSummaryContains: "failed with 1 errors, 1 warnings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := &ValidateAccountResults{
				IsValid:     true,
				CanTransact: true,
				Validations: tt.validations,
			}

			service.updateValidationSummary(result)

			if result.IsValid != tt.expectValid {
				t.Errorf("updateValidationSummary() IsValid = %v, want %v", result.IsValid, tt.expectValid)
			}

			if result.CanTransact != tt.expectCanTransact {
				t.Errorf("updateValidationSummary() CanTransact = %v, want %v", result.CanTransact, tt.expectCanTransact)
			}

			if result.ValidationSummary == "" {
				t.Error("updateValidationSummary() ValidationSummary should not be empty")
			}

			if tt.expectedSummaryContains != "" {
				if !contains(result.ValidationSummary, tt.expectedSummaryContains) {
					t.Errorf("updateValidationSummary() ValidationSummary = %s, should contain %s", result.ValidationSummary, tt.expectedSummaryContains)
				}
			}
		})
	}
}

// Helper functions for tests

func stringPtr(s string) *string {
	return &s
}

func createPgNumeric(value string) pgtype.Numeric {
	d, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}

	// Convert decimal to pgtype.Numeric
	// For PostgreSQL numeric, we need to store the unscaled value and the scale
	// For example, 1000.00 should be stored as Int=100000, Exp=-2
	exp := int32(d.Exponent())
	coefficient := d.Coefficient()

	return pgtype.Numeric{
		Int:   coefficient,
		Exp:   exp,
		Valid: true,
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
