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

// CompensateDebitParams represents the input parameters for compensating a debit transaction
type CompensateDebitParams struct {
	// Original transaction information
	OriginalTransactionID *uuid.UUID `json:"original_transaction_id,omitempty"`
	OriginalReferenceID   *string    `json:"original_reference_id,omitempty"`

	// Account information (alternative to looking up via original transaction)
	AccountID     *uuid.UUID `json:"account_id,omitempty"`
	AccountNumber *string    `json:"account_number,omitempty"`

	// Compensation transaction details
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    *string         `json:"description,omitempty"`
	ReferenceID    *string         `json:"reference_id,omitempty"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	Metadata       map[string]any  `json:"metadata,omitempty"`

	// Compensation specific fields
	CompensationReason *string `json:"compensation_reason,omitempty"`
	WorkflowID         *string `json:"workflow_id,omitempty"`
	RunID              *string `json:"run_id,omitempty"`
}

// CompensateDebitResults represents the output of a compensation operation
type CompensateDebitResults struct {
	// Compensation transaction details
	TransactionID   uuid.UUID       `json:"transaction_id"`
	AccountID       uuid.UUID       `json:"account_id"`
	AccountNumber   string          `json:"account_number"`
	AccountName     string          `json:"account_name"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	Description     *string         `json:"description,omitempty"`
	ReferenceID     *string         `json:"reference_id,omitempty"`
	IdempotencyKey  *string         `json:"idempotency_key,omitempty"`
	Status          string          `json:"status"`
	PreviousBalance decimal.Decimal `json:"previous_balance"`
	NewBalance      decimal.Decimal `json:"new_balance"`
	CreatedAt       string          `json:"created_at"`
	CompletedAt     *string         `json:"completed_at,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`

	// Original transaction details
	OriginalTransactionID *uuid.UUID `json:"original_transaction_id,omitempty"`
	OriginalReferenceID   *string    `json:"original_reference_id,omitempty"`

	// Compensation specific fields
	CompensationReason *string            `json:"compensation_reason,omitempty"`
	WorkflowID         *string            `json:"workflow_id,omitempty"`
	RunID              *string            `json:"run_id,omitempty"`
	ValidationResults  []ValidationResult `json:"validation_results,omitempty"`
}

// CompensateDebit processes a compensation credit transaction that reverses a previous debit
// This is used in distributed transaction scenarios when a debit operation needs to be rolled back
func (service *Service) CompensateDebit(ctx context.Context, params CompensateDebitParams) (*CompensateDebitResults, error) {
	const op = "service.Service.CompensateDebit"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Step 1: Validate input parameters
	if err := service.validateCompensateDebitParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 2: Check for existing compensation transaction with same idempotency key
	if params.IdempotencyKey != nil {
		existingResult, err := service.checkExistingCompensationTransaction(ctx, *params.IdempotencyKey)
		if err != nil {
			err = fmt.Errorf("failed to check existing compensation transaction: %w", err)

			logger.WithError(err).Error()

			return nil, err
		}
		if existingResult != nil {
			service.logger.WithFields(logrus.Fields{
				"transaction_id": existingResult.TransactionID,
				"message":        "Returning existing compensation transaction",
			}).Info()

			return existingResult, nil
		}
	}

	// Step 3: Resolve and validate original transaction if provided
	originalTransaction, err := service.resolveOriginalTransaction(ctx, params)
	if err != nil {
		err = fmt.Errorf("failed to resolve original transaction: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 4: Resolve account ID (from original transaction or direct params)
	accountID, err := service.resolveCompensationAccountID(ctx, params, originalTransaction)
	if err != nil {
		err = fmt.Errorf("failed to resolve account: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 5: Perform account validation
	validationResults, err := service.validateAccountForCompensation(ctx, accountID, params, originalTransaction)
	if err != nil {
		err = fmt.Errorf("account validation failed: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Step 6: Check if validation passed (no errors)
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

		return &CompensateDebitResults{
			AccountID:         accountID,
			Amount:            params.Amount,
			Currency:          params.Currency,
			Status:            "validation_failed",
			ValidationResults: validationResults,
		}, err
	}

	// Step 7: Execute the compensation transaction
	result, err := service.executeCompensationTransaction(ctx, accountID, params, originalTransaction, validationResults)
	if err != nil {
		err = fmt.Errorf("failed to execute compensation transaction: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// validateCompensateDebitParams validates the input parameters for compensation operation
func (service *Service) validateCompensateDebitParams(params CompensateDebitParams) error {
	// Must have either original transaction info or account info
	hasOriginalTxnInfo := params.OriginalTransactionID != nil || params.OriginalReferenceID != nil
	hasAccountInfo := params.AccountID != nil || params.AccountNumber != nil

	if !hasOriginalTxnInfo && !hasAccountInfo {
		return fmt.Errorf("either original transaction information or account information must be provided")
	}

	// Validate original transaction ID if provided
	if params.OriginalTransactionID != nil && *params.OriginalTransactionID == uuid.Nil {
		return fmt.Errorf("original_transaction_id cannot be empty")
	}

	// Validate original reference ID if provided
	if params.OriginalReferenceID != nil && *params.OriginalReferenceID == "" {
		return fmt.Errorf("original_reference_id cannot be empty when provided")
	}

	// Validate account ID if provided
	if params.AccountID != nil && *params.AccountID == uuid.Nil {
		return fmt.Errorf("account_id cannot be empty")
	}

	// Validate account number if provided
	if params.AccountNumber != nil && *params.AccountNumber == "" {
		return fmt.Errorf("account_number cannot be empty when provided")
	}

	// Don't allow both account ID and account number
	if params.AccountID != nil && params.AccountNumber != nil {
		return fmt.Errorf("only one of account_id or account_number should be provided")
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

// checkExistingCompensationTransaction checks if a compensation transaction already exists with the given idempotency key
func (service *Service) checkExistingCompensationTransaction(ctx context.Context, idempotencyKey string) (*CompensateDebitResults, error) {
	pgIdempotencyKey := pgtype.Text{String: idempotencyKey, Valid: true}

	transaction, err := service.store.GetTransactionByIdempotencyKey(ctx, pgIdempotencyKey)
	if err != nil {
		// If not found, that's okay - we can proceed with new transaction
		return nil, nil
	}

	// Convert the existing transaction to compensation result
	return service.convertTransactionToCompensationResult(ctx, transaction)
}

// resolveOriginalTransaction resolves the original transaction details if provided
func (service *Service) resolveOriginalTransaction(ctx context.Context, params CompensateDebitParams) (*sqlc.GetTransactionByIDRow, error) {
	if params.OriginalTransactionID != nil {
		// Get transaction by ID
		pgTransactionID := pgtype.UUID{Bytes: *params.OriginalTransactionID, Valid: true}
		transaction, err := service.store.GetTransactionByID(ctx, pgTransactionID)
		if err != nil {
			return nil, fmt.Errorf("original transaction not found: %w", err)
		}

		// Validate it's a debit transaction
		if transaction.TransactionType != sqlc.CoreTransactionTypeDebit {
			return nil, fmt.Errorf("original transaction must be a debit transaction")
		}

		return &transaction, nil
	}

	if params.OriginalReferenceID != nil {
		// Get transactions by reference ID and find the debit transaction
		pgReferenceID := pgtype.Text{String: *params.OriginalReferenceID, Valid: true}
		transactions, err := service.store.GetTransactionsByReference(ctx, pgReferenceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get transactions by reference: %w", err)
		}

		// Find the debit transaction
		for _, txn := range transactions {
			if txn.TransactionType == sqlc.CoreTransactionTypeDebit {
				// Convert to GetTransactionByIDRow format
				fullTxn, err := service.store.GetTransactionByID(ctx, txn.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to get full transaction details: %w", err)
				}
				return &fullTxn, nil
			}
		}

		return nil, fmt.Errorf("no debit transaction found with reference ID: %s", *params.OriginalReferenceID)
	}

	// No original transaction info provided, that's okay
	return nil, nil
}

// resolveCompensationAccountID resolves the account ID for the compensation transaction
func (service *Service) resolveCompensationAccountID(ctx context.Context, params CompensateDebitParams, originalTransaction *sqlc.GetTransactionByIDRow) (uuid.UUID, error) {
	// If account ID is provided directly, use it
	if params.AccountID != nil {
		return *params.AccountID, nil
	}

	// If account number is provided, resolve it
	if params.AccountNumber != nil {
		account, err := service.store.GetAccountByAccountNumber(ctx, *params.AccountNumber)
		if err != nil {
			return uuid.Nil, fmt.Errorf("account not found: %w", err)
		}
		return account.ID.Bytes, nil
	}

	// If original transaction is available, use its account ID
	if originalTransaction != nil {
		return originalTransaction.AccountID.Bytes, nil
	}

	return uuid.Nil, fmt.Errorf("could not resolve account ID")
}

// validateAccountForCompensation validates the account for compensation operation
func (service *Service) validateAccountForCompensation(ctx context.Context, accountID uuid.UUID, params CompensateDebitParams, originalTransaction *sqlc.GetTransactionByIDRow) ([]ValidationResult, error) {
	var results []ValidationResult

	// Get account details
	pgAccountID := pgtype.UUID{Bytes: accountID, Valid: true}
	account, err := service.store.GetAccountByID(ctx, pgAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account details: %w", err)
	}

	// Validate account status
	if account.Status != sqlc.CoreAccountStatusActive {
		results = append(results, ValidationResult{
			Field:   "account_status",
			Message: fmt.Sprintf("Account status is %s, expected active", account.Status),
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

	// Validate currency match
	expectedCurrency := service.mapCurrencyToEnum(params.Currency)
	if account.Currency != expectedCurrency {
		results = append(results, ValidationResult{
			Field:   "currency",
			Message: fmt.Sprintf("Currency mismatch: account has %s, compensation uses %s", account.Currency, params.Currency),
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

	// Validate amount matches original transaction if available
	if originalTransaction != nil {
		originalAmount, err := service.pgNumericToDecimal(originalTransaction.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to convert original transaction amount: %w", err)
		}

		if !params.Amount.Equal(originalAmount) {
			results = append(results, ValidationResult{
				Field:   "amount",
				Message: fmt.Sprintf("Compensation amount %s does not match original debit amount %s", params.Amount.String(), originalAmount.String()),
				Level:   "warning",
				Passed:  true, // Warning, not error
			})
		} else {
			results = append(results, ValidationResult{
				Field:   "amount",
				Message: "Compensation amount matches original debit amount",
				Level:   "info",
				Passed:  true,
			})
		}
	}

	return results, nil
}

// executeCompensationTransaction executes the compensation transaction within a database transaction
func (service *Service) executeCompensationTransaction(ctx context.Context, accountID uuid.UUID, params CompensateDebitParams, originalTransaction *sqlc.GetTransactionByIDRow, validationResults []ValidationResult) (*CompensateDebitResults, error) {
	// Get account details for balance information
	pgAccountID := pgtype.UUID{Bytes: accountID, Valid: true}
	account, err := service.store.GetAccountByID(ctx, pgAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account details: %w", err)
	}

	previousBalance, err := service.pgNumericToDecimal(account.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to convert previous balance: %w", err)
	}

	// Prepare transaction metadata with compensation information
	metadata := make(map[string]any)
	if params.Metadata != nil {
		for k, v := range params.Metadata {
			metadata[k] = v
		}
	}

	// Add compensation-specific metadata
	metadata["compensation"] = true
	metadata["compensation_type"] = "debit_reversal"
	if params.CompensationReason != nil {
		metadata["compensation_reason"] = *params.CompensationReason
	}
	if originalTransaction != nil {
		originalID := uuid.UUID(originalTransaction.ID.Bytes)
		metadata["original_transaction_id"] = originalID.String()
	}
	if params.WorkflowID != nil {
		metadata["workflow_id"] = *params.WorkflowID
	}
	if params.RunID != nil {
		metadata["run_id"] = *params.RunID
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Prepare transaction parameters
	pgAmount, err := service.decimalToPgNumeric(params.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: %w", err)
	}

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

	createParams := sqlc.CreateTransactionParams{
		AccountID:       pgAccountID,
		TransactionType: sqlc.CoreTransactionTypeCredit, // Compensation is a credit
		Amount:          pgAmount,
		Currency:        service.mapCurrencyToEnum(params.Currency),
		Description:     pgDescription,
		ReferenceID:     pgReferenceID,
		IdempotencyKey:  pgIdempotencyKey,
		Metadata:        metadataJSON,
	}

	// Create the transaction
	createResult, err := service.store.CreateTransaction(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create compensation transaction: %w", err)
	}

	// Complete the transaction immediately (compensation is typically atomic)
	pgTransactionID := pgtype.UUID{Bytes: createResult.ID.Bytes, Valid: true}
	completeResult, err := service.store.CompleteTransaction(ctx, pgTransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to complete compensation transaction: %w", err)
	}

	// Get updated account balance
	updatedAccount, err := service.store.GetAccountByID(ctx, pgAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated account details: %w", err)
	}

	newBalance, err := service.pgNumericToDecimal(updatedAccount.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to convert new balance: %w", err)
	}

	// Build the result
	result := &CompensateDebitResults{
		TransactionID:      createResult.ID.Bytes,
		AccountID:          accountID,
		AccountNumber:      account.AccountNumber,
		AccountName:        account.AccountName,
		Amount:             params.Amount,
		Currency:           params.Currency,
		Description:        params.Description,
		ReferenceID:        params.ReferenceID,
		IdempotencyKey:     params.IdempotencyKey,
		Status:             string(completeResult.Status),
		PreviousBalance:    previousBalance,
		NewBalance:         newBalance,
		CreatedAt:          createResult.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		ValidationResults:  validationResults,
		Metadata:           metadata,
		CompensationReason: params.CompensationReason,
		WorkflowID:         params.WorkflowID,
		RunID:              params.RunID,
	}

	// Add completion time if available
	if completeResult.CompletedAt.Valid {
		completedAt := completeResult.CompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		result.CompletedAt = &completedAt
	}

	// Add original transaction info if available
	if originalTransaction != nil {
		originalID := uuid.UUID(originalTransaction.ID.Bytes)
		result.OriginalTransactionID = &originalID
		if originalTransaction.ReferenceID.Valid {
			result.OriginalReferenceID = &originalTransaction.ReferenceID.String
		}
	} else if params.OriginalReferenceID != nil {
		result.OriginalReferenceID = params.OriginalReferenceID
	}

	return result, nil
}

// convertTransactionToCompensationResult converts an existing transaction to compensation result
func (service *Service) convertTransactionToCompensationResult(ctx context.Context, transaction sqlc.GetTransactionByIdempotencyKeyRow) (*CompensateDebitResults, error) {
	// Get account details
	account, err := service.store.GetAccountByID(ctx, transaction.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account details: %w", err)
	}

	// Convert amount
	amount, err := service.pgNumericToDecimal(transaction.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: %w", err)
	}

	// Get current balance as "new balance"
	currentBalance, err := service.pgNumericToDecimal(account.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to convert current balance: %w", err)
	}

	// Calculate previous balance (current - amount since this was a credit)
	previousBalance := currentBalance.Sub(amount)

	// Parse metadata
	var metadata map[string]any
	if len(transaction.Metadata) > 0 {
		if err := json.Unmarshal(transaction.Metadata, &metadata); err != nil {
			metadata = make(map[string]any)
		}
	} else {
		metadata = make(map[string]any)
	}

	result := &CompensateDebitResults{
		TransactionID:   transaction.ID.Bytes,
		AccountID:       transaction.AccountID.Bytes,
		AccountNumber:   account.AccountNumber,
		AccountName:     account.AccountName,
		Amount:          amount,
		Currency:        string(transaction.Currency),
		Status:          string(transaction.Status),
		PreviousBalance: previousBalance,
		NewBalance:      currentBalance,
		CreatedAt:       transaction.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		Metadata:        metadata,
	}

	// Add optional fields
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
		completedAt := transaction.CompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		result.CompletedAt = &completedAt
	}

	// Extract compensation-specific metadata
	if originalTxnID, ok := metadata["original_transaction_id"].(string); ok {
		if parsedID, err := uuid.Parse(originalTxnID); err == nil {
			result.OriginalTransactionID = &parsedID
		}
	}

	if reason, ok := metadata["compensation_reason"].(string); ok {
		result.CompensationReason = &reason
	}

	if workflowID, ok := metadata["workflow_id"].(string); ok {
		result.WorkflowID = &workflowID
	}

	if runID, ok := metadata["run_id"].(string); ok {
		result.RunID = &runID
	}

	return result, nil
}
