package service

import (
	"context"
	"fmt"

	"svc-balance/store/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// ValidationResult represents a single validation check result
type ValidationResult struct {
	Type     string `json:"type"`     // "status", "balance", "currency", "business_rule"
	Rule     string `json:"rule"`     // Specific rule name
	Passed   bool   `json:"passed"`   // Whether validation passed
	Message  string `json:"message"`  // Validation message
	Severity string `json:"severity"` // "error", "warning", "info"
}

// ValidateAccountParams represents the input parameters for account validation
type ValidateAccountParams struct {
	// Account identifier - either AccountID or AccountNumber must be provided
	AccountID     *uuid.UUID `json:"account_id,omitempty"`
	AccountNumber *string    `json:"account_number,omitempty"`

	// Transaction details for validation (optional)
	TransactionType   *string          `json:"transaction_type,omitempty"` // "debit", "credit", "check"
	TransactionAmount *decimal.Decimal `json:"transaction_amount,omitempty"`

	// Expected currency for validation (optional)
	ExpectedCurrency *string `json:"expected_currency,omitempty"`

	// Validation options
	ValidateStatus        bool `json:"validate_status"`         // Default: true
	ValidateBalance       bool `json:"validate_balance"`        // Default: false unless transaction details provided
	ValidateCurrency      bool `json:"validate_currency"`       // Default: false unless expected_currency provided
	ValidateBusinessRules bool `json:"validate_business_rules"` // Default: true
}

// ValidateAccountResults represents the result of account validation
type ValidateAccountResults struct {
	// Account information
	AccountID     uuid.UUID `json:"account_id"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	Status        string    `json:"status"`
	Currency      string    `json:"currency"`

	// Validation results
	IsValid     bool               `json:"is_valid"`
	CanTransact bool               `json:"can_transact"`
	Validations []ValidationResult `json:"validations"`

	// Summary
	ValidationSummary string `json:"validation_summary"`
}

// ValidateAccount implements comprehensive account validation logic
func (service *Service) ValidateAccount(ctx context.Context, params ValidateAccountParams) (*ValidateAccountResults, error) {
	const op = "service.Service.ValidateAccount"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Validate input parameters
	if err := service.validateAccountValidationParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Get account information
	account, err := service.getAccountForValidation(ctx, params)
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Perform validations
	result, err := service.performAccountValidations(account, params)
	if err != nil {
		err = fmt.Errorf("failed to perform validations: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	logger.WithField("results", fmt.Sprintf("%+v", result)).Info()

	return result, nil
}

// validateAccountValidationParams validates the input parameters
func (service *Service) validateAccountValidationParams(params ValidateAccountParams) error {
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

	// Validate transaction type if provided
	if params.TransactionType != nil {
		validTypes := map[string]bool{"debit": true, "credit": true, "check": true}
		if !validTypes[*params.TransactionType] {
			return fmt.Errorf("invalid transaction_type: %s", *params.TransactionType)
		}
	}

	// Validate transaction amount if provided
	if params.TransactionAmount != nil && params.TransactionAmount.IsNegative() {
		return fmt.Errorf("transaction_amount cannot be negative")
	}

	// Validate expected currency if provided
	if params.ExpectedCurrency != nil {
		if err := service.validateCurrency(*params.ExpectedCurrency); err != nil {
			return fmt.Errorf("invalid expected_currency: %w", err)
		}
	}

	return nil
}

// getAccountForValidation retrieves account information for validation
func (service *Service) getAccountForValidation(ctx context.Context, params ValidateAccountParams) (sqlc.CoreAccount, error) {
	if params.AccountID != nil {
		// Query by Account ID
		accountUUID := pgtype.UUID{
			Bytes: *params.AccountID,
			Valid: true,
		}

		return service.store.GetAccountByID(ctx, accountUUID)
	} else {
		// Query by Account Number
		return service.store.GetAccountByNumber(ctx, *params.AccountNumber)
	}
}

// performAccountValidations performs all requested validation checks
func (service *Service) performAccountValidations(account sqlc.CoreAccount, params ValidateAccountParams) (*ValidateAccountResults, error) {
	// Convert UUID
	accountID, err := uuid.FromBytes(account.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("invalid account ID format: %w", err)
	}

	// Initialize result
	result := &ValidateAccountResults{
		AccountID:     accountID,
		AccountNumber: account.AccountNumber,
		AccountName:   account.AccountName,
		Status:        string(account.Status),
		Currency:      string(account.Currency),
		IsValid:       true,
		CanTransact:   true,
		Validations:   []ValidationResult{},
	}

	// Perform status validation (default: enabled)
	if params.ValidateStatus {
		service.validateAccountStatus(result, account)
	}

	// Perform balance validation if transaction details provided or explicitly requested
	if params.ValidateBalance || (params.TransactionType != nil && params.TransactionAmount != nil) {
		if err := service.validateAccountBalance(result, account, params); err != nil {
			return nil, err
		}
	}

	// Perform currency validation if expected currency provided or explicitly requested
	if params.ValidateCurrency || params.ExpectedCurrency != nil {
		service.validateAccountCurrency(result, account, params)
	}

	// Perform business rules validation (default: enabled)
	if params.ValidateBusinessRules {
		if err := service.validateBusinessRules(result, account, params); err != nil {
			return nil, err
		}
	}

	// Update overall validation status
	service.updateValidationSummary(result)

	return result, nil
}

// validateAccountStatus validates account status
func (service *Service) validateAccountStatus(result *ValidateAccountResults, account sqlc.CoreAccount) {
	validation := ValidationResult{
		Type: "status",
		Rule: "account_active",
	}

	if string(account.Status) == "active" {
		validation.Passed = true
		validation.Message = "Account is active"
		validation.Severity = "info"
	} else {
		validation.Passed = false
		validation.Message = fmt.Sprintf("Account status is '%s', expected 'active'", string(account.Status))
		validation.Severity = "error"
		result.IsValid = false
		result.CanTransact = false
	}

	result.Validations = append(result.Validations, validation)
}

// validateAccountBalance validates account balance for transactions
func (service *Service) validateAccountBalance(result *ValidateAccountResults, account sqlc.CoreAccount, params ValidateAccountParams) error {
	validation := ValidationResult{
		Type: "balance",
		Rule: "sufficient_funds",
	}

	// Convert balance
	balance, err := service.pgNumericToDecimal(account.Balance)
	if err != nil {
		return fmt.Errorf("invalid balance format: %w", err)
	}

	// Only validate balance for debit transactions or when transaction amount is provided
	if params.TransactionType != nil && *params.TransactionType == "debit" && params.TransactionAmount != nil {
		if balance.GreaterThanOrEqual(*params.TransactionAmount) {
			validation.Passed = true
			validation.Message = fmt.Sprintf("Sufficient funds available (Balance: %s, Required: %s)", balance.String(), params.TransactionAmount.String())
			validation.Severity = "info"
		} else {
			validation.Passed = false
			validation.Message = fmt.Sprintf("Insufficient funds (Balance: %s, Required: %s)", balance.String(), params.TransactionAmount.String())
			validation.Severity = "error"
			result.IsValid = false
			result.CanTransact = false
		}
	} else {
		// Just validate that balance is valid and non-negative
		if balance.GreaterThanOrEqual(decimal.Zero) {
			validation.Passed = true
			validation.Message = fmt.Sprintf("Balance is valid (%s)", balance.String())
			validation.Severity = "info"
		} else {
			validation.Passed = false
			validation.Message = fmt.Sprintf("Negative balance detected (%s)", balance.String())
			validation.Severity = "warning"
			// Don't fail validation for negative balance unless it's a debit transaction
		}
	}

	result.Validations = append(result.Validations, validation)
	return nil
}

// validateAccountCurrency validates account currency
func (service *Service) validateAccountCurrency(result *ValidateAccountResults, account sqlc.CoreAccount, params ValidateAccountParams) {
	validation := ValidationResult{
		Type: "currency",
		Rule: "currency_match",
	}

	// First validate that the account currency is supported
	accountCurrency := string(account.Currency)
	if err := service.validateCurrency(accountCurrency); err != nil {
		validation.Passed = false
		validation.Message = fmt.Sprintf("Unsupported account currency: %s", accountCurrency)
		validation.Severity = "error"
		result.IsValid = false
		result.CanTransact = false
	} else if params.ExpectedCurrency != nil {
		// Validate currency match if expected currency is provided
		if accountCurrency == *params.ExpectedCurrency {
			validation.Passed = true
			validation.Message = fmt.Sprintf("Currency matches expected (%s)", accountCurrency)
			validation.Severity = "info"
		} else {
			validation.Passed = false
			validation.Message = fmt.Sprintf("Currency mismatch (Account: %s, Expected: %s)", accountCurrency, *params.ExpectedCurrency)
			validation.Severity = "error"
			result.IsValid = false
			result.CanTransact = false
		}
	} else {
		// Just validate that currency is supported
		validation.Passed = true
		validation.Message = fmt.Sprintf("Account currency is supported (%s)", accountCurrency)
		validation.Severity = "info"
	}

	result.Validations = append(result.Validations, validation)
}

// validateBusinessRules validates business-specific rules
func (service *Service) validateBusinessRules(result *ValidateAccountResults, account sqlc.CoreAccount, params ValidateAccountParams) error {
	// Business Rule 1: Account version consistency
	versionValidation := ValidationResult{
		Type:     "business_rule",
		Rule:     "version_consistency",
		Passed:   true,
		Message:  fmt.Sprintf("Account version is valid (%d)", account.Version),
		Severity: "info",
	}

	if account.Version < 1 {
		versionValidation.Passed = false
		versionValidation.Message = fmt.Sprintf("Invalid account version (%d)", account.Version)
		versionValidation.Severity = "warning"
	}

	result.Validations = append(result.Validations, versionValidation)

	// Business Rule 2: Account name consistency
	nameValidation := ValidationResult{
		Type: "business_rule",
		Rule: "name_format",
	}

	if account.AccountName != "" && len(account.AccountName) >= 2 {
		nameValidation.Passed = true
		nameValidation.Message = "Account name format is valid"
		nameValidation.Severity = "info"
	} else {
		nameValidation.Passed = false
		nameValidation.Message = "Account name is too short or empty"
		nameValidation.Severity = "warning"
		// Don't fail validation for name format issues
	}

	result.Validations = append(result.Validations, nameValidation)

	// Business Rule 3: Transaction limits for large amounts
	if params.TransactionAmount != nil {
		limitValidation := ValidationResult{
			Type: "business_rule",
			Rule: "transaction_limits",
		}

		// Define transaction limits (these could be configurable)
		maxSingleTransaction := decimal.NewFromFloat(100000.00) // $100,000

		if params.TransactionAmount.LessThanOrEqual(maxSingleTransaction) {
			limitValidation.Passed = true
			limitValidation.Message = fmt.Sprintf("Transaction amount within limits (%s)", params.TransactionAmount.String())
			limitValidation.Severity = "info"
		} else {
			limitValidation.Passed = false
			limitValidation.Message = fmt.Sprintf("Transaction amount exceeds limit (Amount: %s, Limit: %s)", params.TransactionAmount.String(), maxSingleTransaction.String())
			limitValidation.Severity = "warning"
			// Don't fail validation for limit checks, but flag for review
		}

		result.Validations = append(result.Validations, limitValidation)
	}

	return nil
}

// updateValidationSummary updates the overall validation summary
func (service *Service) updateValidationSummary(result *ValidateAccountResults) {
	errorCount := 0
	warningCount := 0
	passedCount := 0

	for _, validation := range result.Validations {
		switch validation.Severity {
		case "error":
			if !validation.Passed {
				errorCount++
			}
		case "warning":
			if !validation.Passed {
				warningCount++
			}
		case "info":
			if validation.Passed {
				passedCount++
			}
		}
	}

	if errorCount > 0 {
		result.ValidationSummary = fmt.Sprintf("Validation failed with %d errors, %d warnings (%d checks passed)", errorCount, warningCount, passedCount)
		result.IsValid = false
		result.CanTransact = false
	} else if warningCount > 0 {
		result.ValidationSummary = fmt.Sprintf("Validation passed with %d warnings (%d checks passed)", warningCount, passedCount)
		// Keep IsValid and CanTransact as determined by individual validations
	} else {
		result.ValidationSummary = fmt.Sprintf("All validations passed (%d checks)", passedCount)
		// Keep IsValid and CanTransact as true (unless changed by specific validations)
	}
}
