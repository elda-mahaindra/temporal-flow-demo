package service

import (
	"context"
	"encoding/json"
	"fmt"

	"svc-transaction/store/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// ValidationResult represents a single validation result
type ValidationResult struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Level   string `json:"level"` // "error", "warning", "info"
	Passed  bool   `json:"passed"`
}

// DebitAccountParams represents the input parameters for debiting an account
type DebitAccountParams struct {
	AccountID      *uuid.UUID      `json:"account_id,omitempty"`
	AccountNumber  *string         `json:"account_number,omitempty"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    *string         `json:"description,omitempty"`
	ReferenceID    *string         `json:"reference_id,omitempty"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	Metadata       map[string]any  `json:"metadata,omitempty"`
}

// DebitAccountResults represents the output of a debit operation
type DebitAccountResults struct {
	TransactionID     uuid.UUID          `json:"transaction_id"`
	AccountID         uuid.UUID          `json:"account_id"`
	AccountNumber     string             `json:"account_number"`
	AccountName       string             `json:"account_name"`
	Amount            decimal.Decimal    `json:"amount"`
	Currency          string             `json:"currency"`
	Description       *string            `json:"description,omitempty"`
	ReferenceID       *string            `json:"reference_id,omitempty"`
	IdempotencyKey    *string            `json:"idempotency_key,omitempty"`
	Status            string             `json:"status"`
	PreviousBalance   decimal.Decimal    `json:"previous_balance"`
	NewBalance        decimal.Decimal    `json:"new_balance"`
	CreatedAt         string             `json:"created_at"`
	CompletedAt       *string            `json:"completed_at,omitempty"`
	ValidationResults []ValidationResult `json:"validation_results,omitempty"`
	Metadata          map[string]any     `json:"metadata,omitempty"`
}

// DebitAccount processes a debit transaction against an account
// This is the main entry point for debit operations with full validation and transaction handling
func (service *Service) DebitAccount(ctx context.Context, params DebitAccountParams) (*DebitAccountResults, error) {
	const op = "service.Service.DebitAccount"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Step 1: Validate input parameters
	if err := service.validateDebitAccountParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 2: Check for existing transaction with same idempotency key
	if params.IdempotencyKey != nil {
		existingResult, err := service.checkExistingDebitTransaction(ctx, *params.IdempotencyKey)
		if err != nil {
			err = fmt.Errorf("failed to check existing transaction: %w", err)

			logger.WithError(err).Error()

			return nil, err
		}
		if existingResult != nil {
			service.logger.WithFields(logrus.Fields{
				"transaction_id": existingResult.TransactionID,
				"message":        "Returning existing debit transaction",
			}).Info()

			return existingResult, nil
		}
	}

	// Step 3: Resolve account ID if account number is provided
	accountID, err := service.resolveAccountID(ctx, params)
	if err != nil {
		err = fmt.Errorf("failed to resolve account: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 4: Perform account validation (balance, status, currency)
	validationResults, err := service.validateAccountForDebit(ctx, accountID, params)
	if err != nil {
		err = fmt.Errorf("account validation failed: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 5: Check if validation passed (no errors)
	hasErrors := false
	for _, result := range validationResults {
		if !result.Passed && result.Level == "error" {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		err = fmt.Errorf("account validation failed")

		logger.WithError(err).Error()

		return &DebitAccountResults{
			AccountID:         accountID,
			Amount:            params.Amount,
			Currency:          params.Currency,
			Status:            "validation_failed",
			ValidationResults: validationResults,
		}, err
	}

	// Step 6: Execute the debit transaction within a database transaction
	result, err := service.executeDebitTransaction(ctx, accountID, params, validationResults)
	if err != nil {
		err = fmt.Errorf("failed to execute debit transaction: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// validateDebitAccountParams validates the input parameters for debit operation
func (service *Service) validateDebitAccountParams(params DebitAccountParams) error {
	// Must have either account ID or account number
	if params.AccountID == nil && params.AccountNumber == nil {
		return fmt.Errorf("either account_id or account_number must be provided")
	}

	if params.AccountID != nil && params.AccountNumber != nil {
		return fmt.Errorf("only one of account_id or account_number should be provided")
	}

	// Validate account ID if provided
	if params.AccountID != nil && *params.AccountID == uuid.Nil {
		return fmt.Errorf("account_id cannot be empty")
	}

	// Validate account number if provided
	if params.AccountNumber != nil && *params.AccountNumber == "" {
		return fmt.Errorf("account_number cannot be empty")
	}

	// Validate amount
	if params.Amount.IsZero() {
		return fmt.Errorf("amount cannot be zero")
	}

	if params.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}

	// Validate currency
	if params.Currency == "" {
		return fmt.Errorf("currency cannot be empty")
	}

	// Validate currency format (basic check)
	if len(params.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter code")
	}

	// Validate idempotency key if provided
	if params.IdempotencyKey != nil && *params.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key cannot be empty when provided")
	}

	return nil
}

// checkExistingDebitTransaction checks if a transaction with the same idempotency key already exists
func (service *Service) checkExistingDebitTransaction(ctx context.Context, idempotencyKey string) (*DebitAccountResults, error) {
	pgIdempotencyKey := pgtype.Text{String: idempotencyKey, Valid: true}

	transaction, err := service.store.GetTransactionByIdempotencyKey(ctx, pgIdempotencyKey)
	if err != nil {
		// If not found, that's okay - we can proceed with new transaction
		return nil, nil
	}

	// Convert existing transaction to result format
	result, err := service.convertTransactionToDebitResult(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to convert existing transaction: %w", err)
	}

	return result, nil
}

// resolveAccountID resolves the account ID from either AccountID or AccountNumber
func (service *Service) resolveAccountID(ctx context.Context, params DebitAccountParams) (uuid.UUID, error) {
	if params.AccountID != nil {
		return *params.AccountID, nil
	}

	// Get account by account number
	account, err := service.store.GetAccountByAccountNumber(ctx, *params.AccountNumber)
	if err != nil {
		return uuid.Nil, fmt.Errorf("account not found with account number %s: %w", *params.AccountNumber, err)
	}

	return uuid.UUID(account.ID.Bytes), nil
}

// validateAccountForDebit performs comprehensive account validation before debit
func (service *Service) validateAccountForDebit(ctx context.Context, accountID uuid.UUID, params DebitAccountParams) ([]ValidationResult, error) {
	var results []ValidationResult

	// Convert account ID to pgtype.UUID
	pgAccountID := pgtype.UUID{Bytes: accountID, Valid: true}

	// Convert amount to pgtype.Numeric for balance check
	pgAmount, err := service.decimalToPgNumeric(params.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount for validation: %w", err)
	}

	// Check account balance and status
	balanceCheck, err := service.store.CheckAccountBalance(ctx, sqlc.CheckAccountBalanceParams{
		ID:      pgAccountID,
		Column2: pgAmount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check account balance: %w", err)
	}

	// Validate account status
	if balanceCheck.Status != sqlc.CoreAccountStatusActive {
		results = append(results, ValidationResult{
			Field:   "account_status",
			Message: fmt.Sprintf("Account is not active. Current status: %s", balanceCheck.Status),
			Level:   "error",
			Passed:  false,
		})
	} else {
		results = append(results, ValidationResult{
			Field:   "account_status",
			Message: "Account is active",
			Level:   "info",
			Passed:  true,
		})
	}

	// Validate sufficient funds
	if !balanceCheck.SufficientFunds {
		currentBalance, _ := service.pgNumericToDecimal(balanceCheck.Balance)
		results = append(results, ValidationResult{
			Field: "account_balance",
			Message: fmt.Sprintf("Insufficient funds. Current balance: %s, Required: %s",
				currentBalance.String(), params.Amount.String()),
			Level:  "error",
			Passed: false,
		})
	} else {
		results = append(results, ValidationResult{
			Field:   "account_balance",
			Message: "Sufficient funds available",
			Level:   "info",
			Passed:  true,
		})
	}

	// Validate currency match
	if string(balanceCheck.Currency) != params.Currency {
		results = append(results, ValidationResult{
			Field: "currency",
			Message: fmt.Sprintf("Currency mismatch. Account currency: %s, Transaction currency: %s",
				balanceCheck.Currency, params.Currency),
			Level:  "error",
			Passed: false,
		})
	} else {
		results = append(results, ValidationResult{
			Field:   "currency",
			Message: "Currency matches account currency",
			Level:   "info",
			Passed:  true,
		})
	}

	return results, nil
}

// executeDebitTransaction executes the debit transaction within a database transaction
func (service *Service) executeDebitTransaction(ctx context.Context, accountID uuid.UUID, params DebitAccountParams, validationResults []ValidationResult) (*DebitAccountResults, error) {
	// Get account details before transaction
	pgAccountID := pgtype.UUID{Bytes: accountID, Valid: true}
	account, err := service.store.GetAccountByID(ctx, pgAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account details: %w", err)
	}

	previousBalance, err := service.pgNumericToDecimal(account.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to convert previous balance: %w", err)
	}

	// Convert parameters to database types
	pgAmount, err := service.decimalToPgNumeric(params.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: %w", err)
	}

	pgCurrency := sqlc.CoreCurrencyCode(params.Currency)

	var pgDescription pgtype.Text
	if params.Description != nil {
		pgDescription = pgtype.Text{String: *params.Description, Valid: true}
	}

	var pgReferenceID pgtype.Text
	if params.ReferenceID != nil {
		pgReferenceID = pgtype.Text{String: *params.ReferenceID, Valid: true}
	}

	var pgIdempotencyKey pgtype.Text
	if params.IdempotencyKey != nil {
		pgIdempotencyKey = pgtype.Text{String: *params.IdempotencyKey, Valid: true}
	}

	var pgMetadata []byte
	if params.Metadata != nil {
		// Convert metadata map to JSON bytes
		metadataBytes, err := json.Marshal(params.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		pgMetadata = metadataBytes
	}

	// Create the transaction record
	createParams := sqlc.CreateTransactionParams{
		AccountID:       pgAccountID,
		TransactionType: sqlc.CoreTransactionTypeDebit,
		Amount:          pgAmount,
		Currency:        pgCurrency,
		Description:     pgDescription,
		ReferenceID:     pgReferenceID,
		IdempotencyKey:  pgIdempotencyKey,
		Metadata:        pgMetadata,
	}

	transaction, err := service.store.CreateTransaction(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Complete the transaction (this will update the account balance)
	completedTransaction, err := service.store.CompleteTransaction(ctx, transaction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete transaction: %w", err)
	}

	// Calculate new balance (previous balance minus debit amount)
	newBalance := previousBalance.Sub(params.Amount)

	// Build the result
	result := &DebitAccountResults{
		TransactionID:     uuid.UUID(transaction.ID.Bytes),
		AccountID:         accountID,
		AccountNumber:     account.AccountNumber,
		AccountName:       account.AccountName,
		Amount:            params.Amount,
		Currency:          params.Currency,
		Description:       params.Description,
		ReferenceID:       params.ReferenceID,
		IdempotencyKey:    params.IdempotencyKey,
		Status:            string(completedTransaction.Status),
		PreviousBalance:   previousBalance,
		NewBalance:        newBalance,
		CreatedAt:         transaction.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		ValidationResults: validationResults,
		Metadata:          params.Metadata,
	}

	if completedTransaction.CompletedAt.Valid {
		completedAtStr := completedTransaction.CompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		result.CompletedAt = &completedAtStr
	}

	return result, nil
}

// convertTransactionToDebitResult converts a database transaction to DebitAccountResults
func (service *Service) convertTransactionToDebitResult(ctx context.Context, transaction sqlc.GetTransactionByIdempotencyKeyRow) (*DebitAccountResults, error) {
	// Convert database types to Go types
	transactionID := uuid.UUID(transaction.ID.Bytes)
	accountID := uuid.UUID(transaction.AccountID.Bytes)

	amount, err := service.pgNumericToDecimal(transaction.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: %w", err)
	}

	// Get account details
	pgAccountID := pgtype.UUID{Bytes: accountID, Valid: true}
	account, err := service.store.GetAccountByID(ctx, pgAccountID)
	if err != nil {
		// If we can't get account details, continue with basic info
		service.logger.WithError(err).Warn("Failed to get account details for existing transaction")
	}

	result := &DebitAccountResults{
		TransactionID: transactionID,
		AccountID:     accountID,
		Amount:        amount,
		Currency:      string(transaction.Currency),
		Status:        string(transaction.Status),
		CreatedAt:     transaction.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add account details if available
	if err == nil {
		result.AccountNumber = account.AccountNumber
		result.AccountName = account.AccountName

		currentBalance, balanceErr := service.pgNumericToDecimal(account.Balance)
		if balanceErr == nil {
			result.NewBalance = currentBalance
			// Calculate previous balance (current + debit amount since it was subtracted)
			result.PreviousBalance = currentBalance.Add(amount)
		}
	}

	if transaction.Description.Valid {
		result.Description = &transaction.Description.String
	}

	if transaction.ReferenceID.Valid {
		result.ReferenceID = &transaction.ReferenceID.String
	}

	if transaction.IdempotencyKey.Valid {
		result.IdempotencyKey = &transaction.IdempotencyKey.String
	}

	if transaction.CompletedAt.Valid {
		completedAtStr := transaction.CompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		result.CompletedAt = &completedAtStr
	}

	// Convert metadata from JSON bytes to map
	if len(transaction.Metadata) > 0 {
		var metadata map[string]any
		if err := json.Unmarshal(transaction.Metadata, &metadata); err == nil {
			result.Metadata = metadata
		}
	}

	return result, nil
}
