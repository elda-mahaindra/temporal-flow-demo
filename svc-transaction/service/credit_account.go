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

// CreditAccountParams represents the input parameters for crediting an account
type CreditAccountParams struct {
	AccountID      *uuid.UUID      `json:"account_id,omitempty"`
	AccountNumber  *string         `json:"account_number,omitempty"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    *string         `json:"description,omitempty"`
	ReferenceID    *string         `json:"reference_id,omitempty"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	Metadata       map[string]any  `json:"metadata,omitempty"`
}

// CreditAccountResults represents the output of a credit operation
type CreditAccountResults struct {
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

// CreditAccount processes a credit transaction against an account
// This is the main entry point for credit operations with full validation and transaction handling
func (service *Service) CreditAccount(ctx context.Context, params CreditAccountParams) (*CreditAccountResults, error) {
	const op = "service.Service.CreditAccount"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Step 1: Validate input parameters
	if err := service.validateCreditAccountParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 2: Check for existing transaction with same idempotency key
	if params.IdempotencyKey != nil {
		existingResult, err := service.checkExistingCreditTransaction(ctx, *params.IdempotencyKey)
		if err != nil {
			err = fmt.Errorf("failed to check existing transaction: %w", err)

			logger.WithError(err).Error()

			return nil, err
		}
		if existingResult != nil {
			service.logger.WithFields(logrus.Fields{
				"transaction_id": existingResult.TransactionID,
				"message":        "Returning existing credit transaction",
			}).Info()

			return existingResult, nil
		}
	}

	// Step 3: Resolve account ID if account number is provided
	accountID, err := service.resolveCreditAccountID(ctx, params)
	if err != nil {
		err = fmt.Errorf("failed to resolve account: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 4: Perform account validation (status, currency)
	validationResults, err := service.validateAccountForCredit(ctx, accountID, params)
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

		return &CreditAccountResults{
			AccountID:         accountID,
			Amount:            params.Amount,
			Currency:          params.Currency,
			Status:            "validation_failed",
			ValidationResults: validationResults,
		}, err
	}

	// Step 6: Execute the credit transaction
	result, err := service.executeCreditTransaction(ctx, accountID, params, validationResults)
	if err != nil {
		err = fmt.Errorf("failed to execute credit transaction: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// validateCreditAccountParams validates the input parameters for credit operation
func (service *Service) validateCreditAccountParams(params CreditAccountParams) error {
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

// checkExistingCreditTransaction checks if a transaction already exists with the given idempotency key
func (service *Service) checkExistingCreditTransaction(ctx context.Context, idempotencyKey string) (*CreditAccountResults, error) {
	pgIdempotencyKey := pgtype.Text{String: idempotencyKey, Valid: true}

	transaction, err := service.store.GetTransactionByIdempotencyKey(ctx, pgIdempotencyKey)
	if err != nil {
		// If not found, that's okay - we can proceed with new transaction
		return nil, nil
	}

	// Convert existing transaction to result format
	result, err := service.convertTransactionToCreditResult(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to convert existing transaction: %w", err)
	}

	return result, nil
}

// resolveCreditAccountID resolves account ID from account number if needed
func (service *Service) resolveCreditAccountID(ctx context.Context, params CreditAccountParams) (uuid.UUID, error) {
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

// validateAccountForCredit performs account validation specific to credit operations
func (service *Service) validateAccountForCredit(ctx context.Context, accountID uuid.UUID, params CreditAccountParams) ([]ValidationResult, error) {
	var results []ValidationResult

	// Convert account ID to pgtype.UUID
	pgAccountID := pgtype.UUID{Bytes: accountID, Valid: true}

	// Get account details
	account, err := service.store.GetAccountByID(ctx, pgAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Validate account status - only active accounts can receive credits
	if account.Status != sqlc.CoreAccountStatusActive {
		results = append(results, ValidationResult{
			Field:   "account_status",
			Message: fmt.Sprintf("Account status is %s, must be active to receive credits", account.Status),
			Level:   "error",
			Passed:  false,
		})
	} else {
		results = append(results, ValidationResult{
			Field:   "account_status",
			Message: "Account status is valid",
			Level:   "info",
			Passed:  true,
		})
	}

	// Validate currency match
	if string(account.Currency) != params.Currency {
		results = append(results, ValidationResult{
			Field:   "currency",
			Message: fmt.Sprintf("Currency mismatch: account has %s, transaction has %s", account.Currency, params.Currency),
			Level:   "error",
			Passed:  false,
		})
	} else {
		results = append(results, ValidationResult{
			Field:   "currency",
			Message: "Currency matches account currency",
			Level:   "info",
			Passed:  true,
		})
	}

	// Credit amount validation (informational - generally no maximum for credits)
	results = append(results, ValidationResult{
		Field:   "amount",
		Message: fmt.Sprintf("Credit amount: %s %s", params.Amount.String(), params.Currency),
		Level:   "info",
		Passed:  true,
	})

	return results, nil
}

// executeCreditTransaction executes the credit transaction
func (service *Service) executeCreditTransaction(ctx context.Context, accountID uuid.UUID, params CreditAccountParams, validationResults []ValidationResult) (*CreditAccountResults, error) {
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
		TransactionType: sqlc.CoreTransactionTypeCredit,
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

	// Calculate new balance (previous balance plus credit amount)
	newBalance := previousBalance.Add(params.Amount)

	// Build the result
	result := &CreditAccountResults{
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

// convertTransactionToCreditResult converts a database transaction to CreditAccountResults
func (service *Service) convertTransactionToCreditResult(ctx context.Context, transaction sqlc.GetTransactionByIdempotencyKeyRow) (*CreditAccountResults, error) {
	// Verify this is a credit transaction
	if transaction.TransactionType != sqlc.CoreTransactionTypeCredit {
		return nil, fmt.Errorf("transaction is not a credit transaction: %s", transaction.TransactionType)
	}

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

	result := &CreditAccountResults{
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
			// Calculate previous balance (current minus credit amount since it was added)
			result.PreviousBalance = currentBalance.Sub(amount)
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
