package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"svc-balance/store/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// MockStore implements the store interface for testing
type MockStore struct {
	checkAccountBalanceFunc           func(ctx context.Context, arg sqlc.CheckAccountBalanceParams) (sqlc.CheckAccountBalanceRow, error)
	getAccountByIDFunc                func(ctx context.Context, id pgtype.UUID) (sqlc.CoreAccount, error)
	getAccountByNumberFunc            func(ctx context.Context, accountNumber string) (sqlc.CoreAccount, error)
	getAccountBalanceHistoryFunc      func(ctx context.Context, arg sqlc.GetAccountBalanceHistoryParams) ([]sqlc.CoreAccountBalanceHistory, error)
	getAccountSummaryFunc             func(ctx context.Context, id pgtype.UUID) (sqlc.GetAccountSummaryRow, error)
	getAccountsByBalanceRangeFunc     func(ctx context.Context, arg sqlc.GetAccountsByBalanceRangeParams) ([]sqlc.GetAccountsByBalanceRangeRow, error)
	getAccountsByCurrencyFunc         func(ctx context.Context, arg sqlc.GetAccountsByCurrencyParams) ([]sqlc.CoreAccount, error)
	getAccountsByStatusFunc           func(ctx context.Context, arg sqlc.GetAccountsByStatusParams) ([]sqlc.CoreAccount, error)
	getAccountsWithLowBalanceFunc     func(ctx context.Context, arg sqlc.GetAccountsWithLowBalanceParams) ([]sqlc.GetAccountsWithLowBalanceRow, error)
	validateAccountForTransactionFunc func(ctx context.Context, arg sqlc.ValidateAccountForTransactionParams) (sqlc.ValidateAccountForTransactionRow, error)
}

func (m *MockStore) CheckAccountBalance(ctx context.Context, arg sqlc.CheckAccountBalanceParams) (sqlc.CheckAccountBalanceRow, error) {
	if m.checkAccountBalanceFunc != nil {
		return m.checkAccountBalanceFunc(ctx, arg)
	}
	return sqlc.CheckAccountBalanceRow{}, errors.New("not implemented")
}

func (m *MockStore) GetAccountByID(ctx context.Context, id pgtype.UUID) (sqlc.CoreAccount, error) {
	if m.getAccountByIDFunc != nil {
		return m.getAccountByIDFunc(ctx, id)
	}
	return sqlc.CoreAccount{}, errors.New("not implemented")
}

func (m *MockStore) GetAccountByNumber(ctx context.Context, accountNumber string) (sqlc.CoreAccount, error) {
	if m.getAccountByNumberFunc != nil {
		return m.getAccountByNumberFunc(ctx, accountNumber)
	}
	return sqlc.CoreAccount{}, errors.New("not implemented")
}

func (m *MockStore) GetAccountBalanceHistory(ctx context.Context, arg sqlc.GetAccountBalanceHistoryParams) ([]sqlc.CoreAccountBalanceHistory, error) {
	if m.getAccountBalanceHistoryFunc != nil {
		return m.getAccountBalanceHistoryFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) GetAccountSummary(ctx context.Context, id pgtype.UUID) (sqlc.GetAccountSummaryRow, error) {
	if m.getAccountSummaryFunc != nil {
		return m.getAccountSummaryFunc(ctx, id)
	}
	return sqlc.GetAccountSummaryRow{}, errors.New("not implemented")
}

func (m *MockStore) GetAccountsByBalanceRange(ctx context.Context, arg sqlc.GetAccountsByBalanceRangeParams) ([]sqlc.GetAccountsByBalanceRangeRow, error) {
	if m.getAccountsByBalanceRangeFunc != nil {
		return m.getAccountsByBalanceRangeFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) GetAccountsByCurrency(ctx context.Context, arg sqlc.GetAccountsByCurrencyParams) ([]sqlc.CoreAccount, error) {
	if m.getAccountsByCurrencyFunc != nil {
		return m.getAccountsByCurrencyFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) GetAccountsByStatus(ctx context.Context, arg sqlc.GetAccountsByStatusParams) ([]sqlc.CoreAccount, error) {
	if m.getAccountsByStatusFunc != nil {
		return m.getAccountsByStatusFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) GetAccountsWithLowBalance(ctx context.Context, arg sqlc.GetAccountsWithLowBalanceParams) ([]sqlc.GetAccountsWithLowBalanceRow, error) {
	if m.getAccountsWithLowBalanceFunc != nil {
		return m.getAccountsWithLowBalanceFunc(ctx, arg)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) ValidateAccountForTransaction(ctx context.Context, arg sqlc.ValidateAccountForTransactionParams) (sqlc.ValidateAccountForTransactionRow, error) {
	if m.validateAccountForTransactionFunc != nil {
		return m.validateAccountForTransactionFunc(ctx, arg)
	}
	return sqlc.ValidateAccountForTransactionRow{}, errors.New("not implemented")
}

func TestCheckBalanceValidateParams(t *testing.T) {
	t.Parallel()

	service := createTestService()

	testUUID := uuid.New()
	nilUUID := uuid.UUID{}
	testAccountNumber := "ACC123456"
	emptyAccountNumber := ""
	negativeAmount := decimal.NewFromFloat(-100.0)

	tests := []struct {
		name        string
		params      CheckBalanceParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid params with AccountID",
			params: CheckBalanceParams{
				AccountID: &testUUID,
			},
			expectError: false,
		},
		{
			name: "Valid params with AccountNumber",
			params: CheckBalanceParams{
				AccountNumber: &testAccountNumber,
			},
			expectError: false,
		},
		{
			name: "Valid params with required amount",
			params: CheckBalanceParams{
				AccountID:      &testUUID,
				RequiredAmount: decimalPtr(decimal.NewFromFloat(100.0)),
			},
			expectError: false,
		},
		{
			name: "Missing both AccountID and AccountNumber",
			params: CheckBalanceParams{
				RequiredAmount: decimalPtr(decimal.NewFromFloat(100.0)),
			},
			expectError: true,
			errorMsg:    "either account_id or account_number must be provided",
		},
		{
			name: "Both AccountID and AccountNumber provided",
			params: CheckBalanceParams{
				AccountID:     &testUUID,
				AccountNumber: &testAccountNumber,
			},
			expectError: true,
			errorMsg:    "only one of account_id or account_number should be provided",
		},
		{
			name: "Empty AccountID",
			params: CheckBalanceParams{
				AccountID: &nilUUID,
			},
			expectError: true,
			errorMsg:    "account_id cannot be empty",
		},
		{
			name: "Empty AccountNumber",
			params: CheckBalanceParams{
				AccountNumber: &emptyAccountNumber,
			},
			expectError: true,
			errorMsg:    "account_number cannot be empty",
		},
		{
			name: "Negative required amount",
			params: CheckBalanceParams{
				AccountID:      &testUUID,
				RequiredAmount: &negativeAmount,
			},
			expectError: true,
			errorMsg:    "required_amount cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := service.validateCheckBalanceParams(tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("validateCheckBalanceParams() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("validateCheckBalanceParams() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateCheckBalanceParams() error = %v", err)
				}
			}
		})
	}
}

func TestPerformBalanceValidations(t *testing.T) {
	t.Parallel()

	service := createTestService()

	tests := []struct {
		name                    string
		result                  *CheckBalanceResults
		params                  CheckBalanceParams
		expectedMessageCount    int
		expectedMessageContains string
	}{
		{
			name: "Sufficient funds",
			result: &CheckBalanceResults{
				CurrentBalance:  decimal.NewFromFloat(1000.0),
				SufficientFunds: true,
				IsActive:        true,
				Status:          "active",
				Currency:        "USD",
			},
			params: CheckBalanceParams{
				RequiredAmount: decimalPtr(decimal.NewFromFloat(500.0)),
			},
			expectedMessageCount: 0, // No validation messages expected
		},
		{
			name: "Insufficient funds",
			result: &CheckBalanceResults{
				CurrentBalance:  decimal.NewFromFloat(100.0),
				SufficientFunds: false,
				IsActive:        true,
				Status:          "active",
				Currency:        "USD",
			},
			params: CheckBalanceParams{
				RequiredAmount: decimalPtr(decimal.NewFromFloat(500.0)),
			},
			expectedMessageCount:    1,
			expectedMessageContains: "Insufficient funds",
		},
		{
			name: "Exact amount",
			result: &CheckBalanceResults{
				CurrentBalance:  decimal.NewFromFloat(500.0),
				SufficientFunds: true,
				IsActive:        true,
				Status:          "active",
				Currency:        "USD",
			},
			params: CheckBalanceParams{
				RequiredAmount: decimalPtr(decimal.NewFromFloat(500.0)),
			},
			expectedMessageCount: 0, // No validation messages expected
		},
		{
			name: "No required amount",
			result: &CheckBalanceResults{
				CurrentBalance:  decimal.NewFromFloat(100.0),
				SufficientFunds: true,
				IsActive:        true,
				Status:          "active",
				Currency:        "USD",
			},
			params: CheckBalanceParams{
				RequiredAmount: nil,
			},
			expectedMessageCount: 0, // No validation messages expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clear any existing validation messages
			tt.result.ValidationMessages = nil

			service.performBalanceValidations(tt.result, tt.params)

			if len(tt.result.ValidationMessages) != tt.expectedMessageCount {
				t.Errorf("performBalanceValidations() ValidationMessages count = %d, want %d", len(tt.result.ValidationMessages), tt.expectedMessageCount)
			}

			if tt.expectedMessageContains != "" {
				found := false
				for _, msg := range tt.result.ValidationMessages {
					if strings.Contains(msg, tt.expectedMessageContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("performBalanceValidations() ValidationMessages should contain '%s', got %v", tt.expectedMessageContains, tt.result.ValidationMessages)
				}
			}
		})
	}
}

func TestBuildCheckBalanceResult(t *testing.T) {
	t.Parallel()

	service := createTestService()

	testUUID := uuid.New()
	testPgUUID := pgtype.UUID{Bytes: testUUID, Valid: true}

	testAccount := sqlc.CheckAccountBalanceRow{
		ID:              testPgUUID,
		AccountNumber:   "ACC123456",
		AccountName:     "John Doe",
		Balance:         createPgNumeric("1500.75"),
		Currency:        sqlc.CoreCurrencyCodeUSD,
		Status:          sqlc.CoreAccountStatusActive,
		SufficientFunds: true, // This should come from the database result
	}

	result, err := service.buildCheckBalanceResult(testAccount)

	if err != nil {
		t.Errorf("buildCheckBalanceResult() error = %v", err)
		return
	}

	// Verify account ID conversion
	if result.AccountID != testUUID {
		t.Errorf("buildCheckBalanceResult() AccountID = %v, want %v", result.AccountID, testUUID)
	}

	// Verify basic fields
	if result.AccountNumber != testAccount.AccountNumber {
		t.Errorf("buildCheckBalanceResult() AccountNumber = %v, want %v", result.AccountNumber, testAccount.AccountNumber)
	}

	if result.AccountName != testAccount.AccountName {
		t.Errorf("buildCheckBalanceResult() AccountName = %v, want %v", result.AccountName, testAccount.AccountName)
	}

	// Verify balance conversion
	expectedBalance, _ := service.pgNumericToDecimal(testAccount.Balance)
	if !result.CurrentBalance.Equal(expectedBalance) {
		t.Errorf("buildCheckBalanceResult() CurrentBalance = %v, want %v", result.CurrentBalance, expectedBalance)
	}

	// Verify currency
	if result.Currency != string(testAccount.Currency) {
		t.Errorf("buildCheckBalanceResult() Currency = %v, want %v", result.Currency, string(testAccount.Currency))
	}

	// Verify status
	if result.Status != string(testAccount.Status) {
		t.Errorf("buildCheckBalanceResult() Status = %v, want %v", result.Status, string(testAccount.Status))
	}

	if result.IsActive != (testAccount.Status == sqlc.CoreAccountStatusActive) {
		t.Errorf("buildCheckBalanceResult() IsActive = %v, want %v", result.IsActive, testAccount.Status == sqlc.CoreAccountStatusActive)
	}

	// Verify SufficientFunds comes from database result
	if result.SufficientFunds != testAccount.SufficientFunds {
		t.Errorf("buildCheckBalanceResult() SufficientFunds = %v, want %v", result.SufficientFunds, testAccount.SufficientFunds)
	}
}
