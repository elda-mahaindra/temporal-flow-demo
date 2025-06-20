package service

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ActivityOptionsProvider holds the configured activity options for workflows
var ActivityOptionsProvider func() workflow.ActivityOptions

// TransferWorkflowParams defines the input parameters for the transfer workflow
type TransferWorkflowParams struct {
	TransferID     string          `json:"transfer_id"`
	FromAccount    string          `json:"from_account"`
	ToAccount      string          `json:"to_account"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    string          `json:"description"`
	IdempotencyKey string          `json:"idempotency_key"`
	RequestedBy    string          `json:"requested_by"`
}

// TransferWorkflowResults defines the output results from the transfer workflow
type TransferWorkflowResults struct {
	TransferID          string          `json:"transfer_id"`
	Status              string          `json:"status"`
	FromAccount         string          `json:"from_account"`
	ToAccount           string          `json:"to_account"`
	Amount              decimal.Decimal `json:"amount"`
	Currency            string          `json:"currency"`
	Description         string          `json:"description"`
	StartedAt           time.Time       `json:"started_at"`
	CompletedAt         *time.Time      `json:"completed_at,omitempty"`
	ErrorMessage        string          `json:"error_message,omitempty"`
	CompensationApplied bool            `json:"compensation_applied"`
	WorkflowID          string          `json:"workflow_id"`
	RunID               string          `json:"run_id"`
}

// transferWorkflow orchestrates the money transfer process using the orchestration-based saga pattern
func transferWorkflow(ctx workflow.Context, params TransferWorkflowParams) (*TransferWorkflowResults, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting TransferWorkflow", "transfer_id", params.TransferID, "from_account", params.FromAccount, "to_account", params.ToAccount, "amount", params.Amount)

	// Initialize workflow results
	workflowInfo := workflow.GetInfo(ctx)
	results := &TransferWorkflowResults{
		TransferID:          params.TransferID,
		Status:              "processing",
		FromAccount:         params.FromAccount,
		ToAccount:           params.ToAccount,
		Amount:              params.Amount,
		Currency:            params.Currency,
		Description:         params.Description,
		StartedAt:           workflow.Now(ctx),
		CompensationApplied: false,
		WorkflowID:          workflowInfo.WorkflowExecution.ID,
		RunID:               workflowInfo.WorkflowExecution.RunID,
	}

	// Validate workflow parameters
	if err := validateTransferWorkflowParams(params); err != nil {
		logger.Error("Invalid workflow parameters", "error", err)
		results.Status = "failed"
		results.ErrorMessage = fmt.Sprintf("validation failed: %v", err)
		completedAt := workflow.Now(ctx)
		results.CompletedAt = &completedAt
		return results, err
	}

	// PERFORMANCE OPTIMIZATION: Configure optimized activity options for banking operations from configuration
	var activityOptions workflow.ActivityOptions
	if ActivityOptionsProvider != nil {
		activityOptions = ActivityOptionsProvider()
	} else {
		// Fallback to default banking-optimized options if configuration is not available
		activityOptions = workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute * 2,  // Banking operations should complete within 2 minutes
			HeartbeatTimeout:    time.Second * 30, // Heartbeat every 30 seconds for monitoring
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    time.Millisecond * 500, // Start with 500ms retry interval (faster for banking)
				BackoffCoefficient: 1.5,                    // Moderate backoff to prevent thundering herd
				MaximumInterval:    time.Second * 15,       // Max 15 seconds between retries (banking needs quick response)
				MaximumAttempts:    3,                      // Fail fast for banking operations
				NonRetryableErrorTypes: []string{ // Don't retry these banking-specific errors
					"INSUFFICIENT_FUNDS",
					"ACCOUNT_NOT_FOUND",
					"INVALID_CURRENCY",
					"ACCOUNT_BLOCKED",
				},
			},
			ScheduleToCloseTimeout: time.Minute * 3,  // Total time including queuing
			ScheduleToStartTimeout: time.Second * 30, // Max time in queue before starting
		}
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Check Balance
	logger.Info("Step 1: Checking balance", "account_id", params.FromAccount)
	balanceCheckParams := map[string]interface{}{
		"account_id":      params.FromAccount,
		"required_amount": params.Amount,
		"currency":        params.Currency,
		"transfer_id":     params.TransferID,
		"workflow_id":     workflowInfo.WorkflowExecution.ID,
		"run_id":          workflowInfo.WorkflowExecution.RunID,
	}

	var balanceResult map[string]interface{}
	err := workflow.ExecuteActivity(ctx, "CheckBalance", balanceCheckParams).Get(ctx, &balanceResult)
	if err != nil {
		logger.Error("Balance check failed", "error", err)
		results.Status = "failed"
		results.ErrorMessage = fmt.Sprintf("balance check failed: %v", err)
		completedAt := workflow.Now(ctx)
		results.CompletedAt = &completedAt
		return results, err
	}

	sufficientFunds, ok := balanceResult["sufficient_funds"].(bool)
	if !ok || !sufficientFunds {
		logger.Error("Insufficient funds", "balance_result", balanceResult)
		results.Status = "failed"
		results.ErrorMessage = "insufficient funds"
		completedAt := workflow.Now(ctx)
		results.CompletedAt = &completedAt
		return results, fmt.Errorf("insufficient funds")
	}

	logger.Info("Balance check successful", "balance_result", balanceResult)

	// Step 2: Debit Account
	logger.Info("Step 2: Debiting account", "account_id", params.FromAccount, "amount", params.Amount)
	debitParams := map[string]interface{}{
		"account_id":      params.FromAccount,
		"amount":          params.Amount,
		"currency":        params.Currency,
		"description":     fmt.Sprintf("Transfer to %s: %s", params.ToAccount, params.Description),
		"reference_id":    params.TransferID,
		"idempotency_key": fmt.Sprintf("%s-debit", params.IdempotencyKey),
		"transfer_id":     params.TransferID,
		"workflow_id":     workflowInfo.WorkflowExecution.ID,
		"run_id":          workflowInfo.WorkflowExecution.RunID,
	}

	var debitResult map[string]interface{}
	err = workflow.ExecuteActivity(ctx, "DebitAccount", debitParams).Get(ctx, &debitResult)
	if err != nil {
		logger.Error("Debit account failed", "error", err)
		results.Status = "failed"
		results.ErrorMessage = fmt.Sprintf("debit account failed: %v", err)
		completedAt := workflow.Now(ctx)
		results.CompletedAt = &completedAt
		return results, err
	}

	logger.Info("Debit account successful", "debit_result", debitResult)

	// Step 3: Credit Account (with compensation logic if it fails)
	logger.Info("Step 3: Crediting account", "account_id", params.ToAccount, "amount", params.Amount)
	creditParams := map[string]interface{}{
		"account_id":      params.ToAccount,
		"amount":          params.Amount,
		"currency":        params.Currency,
		"description":     fmt.Sprintf("Transfer from %s: %s", params.FromAccount, params.Description),
		"reference_id":    params.TransferID,
		"idempotency_key": fmt.Sprintf("%s-credit", params.IdempotencyKey),
		"transfer_id":     params.TransferID,
		"workflow_id":     workflowInfo.WorkflowExecution.ID,
		"run_id":          workflowInfo.WorkflowExecution.RunID,
	}

	var creditResult map[string]interface{}
	err = workflow.ExecuteActivity(ctx, "CreditAccount", creditParams).Get(ctx, &creditResult)
	if err != nil {
		logger.Error("Credit account failed, executing compensation", "error", err)

		// Execute compensation: reverse the debit
		compensationParams := map[string]interface{}{
			"original_transaction_id": debitResult["transaction_id"],
			"account_id":              params.FromAccount,
			"amount":                  params.Amount,
			"currency":                params.Currency,
			"compensation_reason":     fmt.Sprintf("Credit to %s failed: %v", params.ToAccount, err),
			"reference_id":            params.TransferID,
			"idempotency_key":         fmt.Sprintf("%s-compensate", params.IdempotencyKey),
			"transfer_id":             params.TransferID,
			"workflow_id":             workflowInfo.WorkflowExecution.ID,
			"run_id":                  workflowInfo.WorkflowExecution.RunID,
		}

		var compensationResult map[string]interface{}
		compensationErr := workflow.ExecuteActivity(ctx, "CompensateDebit", compensationParams).Get(ctx, &compensationResult)
		if compensationErr != nil {
			logger.Error("Compensation failed", "error", compensationErr)
			results.Status = "failed"
			results.ErrorMessage = fmt.Sprintf("credit failed and compensation failed: credit_error=%v, compensation_error=%v", err, compensationErr)
		} else {
			logger.Info("Compensation successful", "compensation_result", compensationResult)
			results.Status = "failed"
			results.ErrorMessage = fmt.Sprintf("credit account failed: %v", err)
			results.CompensationApplied = true
		}

		completedAt := workflow.Now(ctx)
		results.CompletedAt = &completedAt
		return results, err
	}

	logger.Info("Credit account successful", "credit_result", creditResult)

	// Step 4: Confirm Transfer (finalization)
	logger.Info("Step 4: Transfer completed successfully")
	results.Status = "completed"
	completedAt := workflow.Now(ctx)
	results.CompletedAt = &completedAt

	logger.Info("TransferWorkflow completed successfully",
		"transfer_id", params.TransferID,
		"debit_result", debitResult,
		"credit_result", creditResult,
		"duration", workflow.Now(ctx).Sub(results.StartedAt))

	return results, nil
}

// validateTransferWorkflowParams validates the input parameters for the transfer workflow
func validateTransferWorkflowParams(params TransferWorkflowParams) error {
	if params.TransferID == "" {
		return fmt.Errorf("transfer_id is required")
	}

	if params.FromAccount == "" {
		return fmt.Errorf("from_account is required")
	}

	if params.ToAccount == "" {
		return fmt.Errorf("to_account is required")
	}

	if params.FromAccount == params.ToAccount {
		return fmt.Errorf("from_account and to_account cannot be the same")
	}

	if params.Amount.IsZero() || params.Amount.IsNegative() {
		return fmt.Errorf("amount must be positive")
	}

	if params.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if params.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}

	// Validate currency format (basic validation)
	validCurrencies := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true,
		"CAD": true, "AUD": true, "CHF": true, "CNY": true,
		"SGD": true, "HKD": true,
	}

	if !validCurrencies[params.Currency] {
		return fmt.Errorf("unsupported currency: %s", params.Currency)
	}

	return nil
}
