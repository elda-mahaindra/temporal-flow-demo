package service

import (
	"context"
	"fmt"
	"math/big"

	"svc-balance/store/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// CheckBalanceParams represents the input parameters for balance checking
type CheckBalanceParams struct {
	// Account identifier - either AccountID or AccountNumber must be provided
	AccountID     *uuid.UUID `json:"account_id,omitempty"`
	AccountNumber *string    `json:"account_number,omitempty"`

	// Required amount to check against (optional - if not provided, just returns current balance)
	RequiredAmount *decimal.Decimal `json:"required_amount,omitempty"`

	// Expected currency for validation (optional)
	ExpectedCurrency *string `json:"expected_currency,omitempty"`

	// Whether to include detailed account information
	IncludeDetails bool `json:"include_details"`
}

// CheckBalanceDetails provides additional account information
type CheckBalanceDetails struct {
	CreatedAt        string               `json:"created_at"`
	UpdatedAt        string               `json:"updated_at"`
	TransactionCount int64                `json:"transaction_count,omitempty"`
	TotalDebits      *decimal.Decimal     `json:"total_debits,omitempty"`
	TotalCredits     *decimal.Decimal     `json:"total_credits,omitempty"`
	BalanceHistory   []BalanceHistoryItem `json:"balance_history,omitempty"`
}

// BalanceHistoryItem represents a balance change record
type BalanceHistoryItem struct {
	TransactionID *uuid.UUID      `json:"transaction_id,omitempty"`
	OldBalance    decimal.Decimal `json:"old_balance"`
	NewBalance    decimal.Decimal `json:"new_balance"`
	BalanceChange decimal.Decimal `json:"balance_change"`
	Operation     string          `json:"operation"`
	CreatedAt     string          `json:"created_at"`
	CreatedBy     string          `json:"created_by,omitempty"`
}

// CheckBalanceResults represents the result of balance checking
type CheckBalanceResults struct {
	// Account information
	AccountID     uuid.UUID `json:"account_id"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`

	// Balance information
	CurrentBalance  decimal.Decimal `json:"current_balance"`
	Currency        string          `json:"currency"`
	SufficientFunds bool            `json:"sufficient_funds"`

	// Account status
	Status   string `json:"status"`
	IsActive bool   `json:"is_active"`

	// Additional details (if requested)
	Details *CheckBalanceDetails `json:"details,omitempty"`

	// Validation messages
	ValidationMessages []string `json:"validation_messages,omitempty"`
}

// CheckBalance implements the core balance checking business logic
func (service *Service) CheckBalance(ctx context.Context, params CheckBalanceParams) (*CheckBalanceResults, error) {
	const op = "service.Service.CheckBalance"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Validate input parameters
	if err := service.validateCheckBalanceParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Get account information
	account, err := service.getAccountForBalance(ctx, params)
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Convert account data to result format
	result, err := service.buildCheckBalanceResult(account)
	if err != nil {
		err = fmt.Errorf("failed to build result: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Enhance result with currency formatting
	if formattedBalance, err := service.FormatCurrencyAmount(result.CurrentBalance, result.Currency); err == nil {
		// Add formatted balance as additional information (not replacing the decimal value)
		logger.WithField("formatted_balance", formattedBalance).Debug("Currency formatting applied")
	}

	// Perform business validations
	service.performBalanceValidations(result, params)

	// Perform comprehensive account validation if requested
	if params.IncludeDetails {
		validationParams := ValidateAccountParams{
			AccountID:             params.AccountID,
			AccountNumber:         params.AccountNumber,
			ExpectedCurrency:      params.ExpectedCurrency,
			ValidateStatus:        true,
			ValidateBusinessRules: true,
		}

		// If we have a required amount, also validate balance
		if params.RequiredAmount != nil {
			transactionType := "debit"
			validationParams.TransactionType = &transactionType
			validationParams.TransactionAmount = params.RequiredAmount
			validationParams.ValidateBalance = true
		}

		validationResult, err := service.ValidateAccount(ctx, validationParams)
		if err != nil {
			logger.WithError(err).Warn("Failed to perform comprehensive account validation")
		} else {
			// Add validation messages to the result
			for _, validation := range validationResult.Validations {
				if !validation.Passed && validation.Severity == "error" {
					result.ValidationMessages = append(result.ValidationMessages, validation.Message)
				}
			}
		}
	}

	// Include additional details if requested
	if params.IncludeDetails {
		details, err := service.getBalanceDetails(ctx, account.ID)
		if err != nil {
			err = fmt.Errorf("failed to get balance details: %w", err)

			logger.WithError(err).Error()

			// Don't fail the entire operation for details
		} else {
			result.Details = details
		}
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// validateCheckBalanceParams validates the input parameters
func (service *Service) validateCheckBalanceParams(params CheckBalanceParams) error {
	// Either AccountID or AccountNumber must be provided
	if params.AccountID == nil && params.AccountNumber == nil {
		return fmt.Errorf("either account_id or account_number must be provided")
	}

	// Both AccountID and AccountNumber cannot be provided
	if params.AccountID != nil && params.AccountNumber != nil {
		return fmt.Errorf("only one of account_id or account_number should be provided")
	}

	// Validate AccountID format if provided
	if params.AccountID != nil && *params.AccountID == uuid.Nil {
		return fmt.Errorf("account_id cannot be empty")
	}

	// Validate AccountNumber format if provided
	if params.AccountNumber != nil && *params.AccountNumber == "" {
		return fmt.Errorf("account_number cannot be empty")
	}

	// Validate RequiredAmount if provided
	if params.RequiredAmount != nil && params.RequiredAmount.IsNegative() {
		return fmt.Errorf("required_amount cannot be negative")
	}

	// Validate ExpectedCurrency if provided
	if params.ExpectedCurrency != nil {
		if err := service.validateCurrency(*params.ExpectedCurrency); err != nil {
			return fmt.Errorf("invalid expected_currency: %w", err)
		}
	}

	return nil
}

// getAccountForBalance retrieves account information for balance checking
func (service *Service) getAccountForBalance(ctx context.Context, params CheckBalanceParams) (sqlc.CheckAccountBalanceRow, error) {
	var requiredAmount pgtype.Numeric

	// Convert required amount to pgtype.Numeric if provided
	if params.RequiredAmount != nil {
		bigFloat, _ := params.RequiredAmount.Float64()
		requiredAmount = pgtype.Numeric{
			Int:   big.NewInt(0).SetBytes(params.RequiredAmount.BigInt().Bytes()),
			Exp:   int32(-params.RequiredAmount.Exponent()),
			Valid: true,
		}
		_ = bigFloat // Use the variable to avoid unused variable error
	}

	if params.AccountID != nil {
		// Query by Account ID
		accountUUID := pgtype.UUID{
			Bytes: *params.AccountID,
			Valid: true,
		}

		return service.store.CheckAccountBalance(ctx, sqlc.CheckAccountBalanceParams{
			ID:      accountUUID,
			Column2: requiredAmount,
		})
	} else {
		// Query by Account Number - first get the account, then check balance
		account, err := service.store.GetAccountByNumber(ctx, *params.AccountNumber)
		if err != nil {
			return sqlc.CheckAccountBalanceRow{}, err
		}

		return service.store.CheckAccountBalance(ctx, sqlc.CheckAccountBalanceParams{
			ID:      account.ID,
			Column2: requiredAmount,
		})
	}
}

// buildCheckBalanceResult converts database result to service result format
func (service *Service) buildCheckBalanceResult(account sqlc.CheckAccountBalanceRow) (*CheckBalanceResults, error) {
	// Convert UUID
	accountID, err := uuid.FromBytes(account.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}

	// Convert balance
	balance, err := service.pgNumericToDecimal(account.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid balance format: %w", err)
	}

	result := &CheckBalanceResults{
		AccountID:       accountID,
		AccountNumber:   account.AccountNumber,
		AccountName:     account.AccountName,
		CurrentBalance:  balance,
		Currency:        string(account.Currency),
		SufficientFunds: account.SufficientFunds,
		Status:          string(account.Status),
		IsActive:        account.Status == sqlc.CoreAccountStatusActive,
	}

	return result, nil
}

// performBalanceValidations performs business logic validations
func (service *Service) performBalanceValidations(result *CheckBalanceResults, params CheckBalanceParams) {
	var messages []string

	// Check account status
	if !result.IsActive {
		messages = append(messages, fmt.Sprintf("Account is not active (status: %s)", result.Status))
	}

	// Check currency match if expected currency is provided
	if params.ExpectedCurrency != nil && result.Currency != *params.ExpectedCurrency {
		messages = append(messages, fmt.Sprintf("Currency mismatch: expected %s, got %s", *params.ExpectedCurrency, result.Currency))
	}

	// Check sufficient funds if required amount is provided
	if params.RequiredAmount != nil && !result.SufficientFunds {
		messages = append(messages, fmt.Sprintf("Insufficient funds: required %s, available %s",
			params.RequiredAmount.String(), result.CurrentBalance.String()))
	}

	// Check for zero balance warning
	if result.CurrentBalance.IsZero() {
		messages = append(messages, "Account has zero balance")
	}

	result.ValidationMessages = messages
}

// getBalanceDetails retrieves additional account details
func (service *Service) getBalanceDetails(ctx context.Context, accountID pgtype.UUID) (*CheckBalanceDetails, error) {
	// Get account summary with transaction statistics
	summary, err := service.store.GetAccountSummary(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account summary: %w", err)
	}

	// Get recent balance history
	historyParams := sqlc.GetAccountBalanceHistoryParams{
		AccountID: accountID,
		Limit:     10, // Last 10 balance changes
		Offset:    0,
	}

	historyRecords, err := service.store.GetAccountBalanceHistory(ctx, historyParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance history: %w", err)
	}

	// Convert history records
	var history []BalanceHistoryItem
	for _, record := range historyRecords {
		item := BalanceHistoryItem{
			Operation: record.Operation,
			CreatedAt: record.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Convert transaction ID if valid
		if record.TransactionID.Valid {
			txID, err := uuid.FromBytes(record.TransactionID.Bytes[:])
			if err == nil {
				item.TransactionID = &txID
			}
		}

		// Convert balances
		if oldBalance, err := service.pgNumericToDecimal(record.OldBalance); err == nil {
			item.OldBalance = oldBalance
		}
		if newBalance, err := service.pgNumericToDecimal(record.NewBalance); err == nil {
			item.NewBalance = newBalance
		}
		if balanceChange, err := service.pgNumericToDecimal(record.BalanceChange); err == nil {
			item.BalanceChange = balanceChange
		}

		// Convert created by
		if record.CreatedBy.Valid {
			item.CreatedBy = record.CreatedBy.String
		}

		history = append(history, item)
	}

	// Convert total debits and credits
	var totalDebits, totalCredits *decimal.Decimal
	if summary.TotalDebits != nil {
		if debitStr, ok := summary.TotalDebits.(string); ok {
			if d, err := decimal.NewFromString(debitStr); err == nil {
				totalDebits = &d
			}
		}
	}
	if summary.TotalCredits != nil {
		if creditStr, ok := summary.TotalCredits.(string); ok {
			if d, err := decimal.NewFromString(creditStr); err == nil {
				totalCredits = &d
			}
		}
	}

	details := &CheckBalanceDetails{
		CreatedAt:        summary.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        summary.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		TransactionCount: summary.TransactionCount,
		TotalDebits:      totalDebits,
		TotalCredits:     totalCredits,
		BalanceHistory:   history,
	}

	return details, nil
}
