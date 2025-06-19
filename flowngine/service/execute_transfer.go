package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
)

type ExecuteTransferParams struct {
	FromAccount       string `json:"from_account"`
	ToAccount         string `json:"to_account"`
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
	Description       string `json:"description"`
	ReferenceID       string `json:"reference_id"`
	RequestID         string `json:"request_id"`
	WaitForCompletion bool   `json:"wait_for_completion"`
}

type ExecuteTransferResults struct {
	TransactionID       string           `json:"transaction_id"`
	Status              string           `json:"status"`
	WorkflowID          string           `json:"workflow_id"`
	RunID               string           `json:"run_id"`
	CreatedAt           string           `json:"created_at"`
	CompletedAt         *string          `json:"completed_at,omitempty"`
	FinalAmount         *decimal.Decimal `json:"final_amount,omitempty"`
	ErrorMessage        string           `json:"error_message,omitempty"`
	CompensationApplied *bool            `json:"compensation_applied,omitempty"`
}

func (svc *Service) ExecuteTransfer(ctx context.Context, params *ExecuteTransferParams) (*ExecuteTransferResults, error) {
	const op = "service.Service.ExecuteTransfer"

	logger := svc.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info("Starting transfer execution", "sync_mode", params.WaitForCompletion)

	// Validate input parameters
	if err := validateExecuteTransferParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Generate transaction and workflow IDs
	transactionID := uuid.New().String()
	workflowID := fmt.Sprintf("transfer_workflow_%s", transactionID)
	idempotencyKey := fmt.Sprintf("%s_%s", params.RequestID, transactionID)

	// Convert amount to decimal (from minor units to major units for internal processing)
	amountDecimal := decimal.NewFromInt(params.Amount).Div(decimal.NewFromInt(100))

	// Prepare workflow parameters
	workflowParams := TransferWorkflowParams{
		TransferID:     transactionID,
		FromAccount:    params.FromAccount,
		ToAccount:      params.ToAccount,
		Amount:         amountDecimal,
		Currency:       params.Currency,
		Description:    params.Description,
		IdempotencyKey: idempotencyKey,
		RequestedBy:    params.RequestID,
	}

	// Configure workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                "transfer-task-queue",
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowRunTimeout:       time.Minute * 5,
	}

	logger.Info("Starting Temporal workflow", "workflow_id", workflowID, "transaction_id", transactionID)

	// 🎯 THIS IS THE TEMPORAL MAGIC!
	// Just start the workflow - Temporal handles ALL the complexity:
	// - State persistence
	// - Activity coordination
	// - Error handling
	// - Retries and timeouts
	// - Compensation logic
	workflowRun, err := svc.temporalClient.ExecuteWorkflow(ctx, workflowOptions, transferWorkflow, workflowParams)
	if err != nil {
		err = fmt.Errorf("failed to start workflow: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	runID := workflowRun.GetRunID()

	logger.Info("🎉 Temporal workflow started!", "workflow_id", workflowID, "run_id", runID, "transaction_id", transactionID)

	// Initialize base results
	results := &ExecuteTransferResults{
		TransactionID: transactionID,
		WorkflowID:    workflowID,
		RunID:         runID,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	if params.WaitForCompletion {
		// 🔄 SYNC MODE: Wait for workflow completion
		logger.Info("⏳ Waiting for workflow completion (SYNC mode)")

		var workflowResult TransferWorkflowResults
		err = workflowRun.Get(ctx, &workflowResult)

		if err != nil {
			err = fmt.Errorf("workflow execution failed: %w", err)

			logger.WithError(err).Error()

			results.Status = "TRANSFER_STATUS_FAILED"
			results.ErrorMessage = fmt.Sprintf("workflow failed: %v", err)

			return results, nil // Don't return error - client gets the failure result
		}

		// 🎉 Workflow completed successfully!
		logger.Info("✅ Workflow completed successfully (SYNC mode)", "status", workflowResult.Status)

		results.Status = workflowResult.Status
		if workflowResult.CompletedAt != nil {
			completedAt := workflowResult.CompletedAt.Format(time.RFC3339)
			results.CompletedAt = &completedAt
		}
		results.FinalAmount = &workflowResult.Amount
		results.ErrorMessage = workflowResult.ErrorMessage
		results.CompensationApplied = &workflowResult.CompensationApplied

		logger.WithField("results", fmt.Sprintf("%+v", results)).Info("Transfer execution completed synchronously")

	} else {
		// 🚀 ASYNC MODE: Return immediately (Current Implementation)
		logger.Info("🚀 Returning immediately (ASYNC mode) - client must poll for status")
		results.Status = "TRANSFER_STATUS_PENDING"

		logger.WithField("results", fmt.Sprintf("%+v", results)).Info("Transfer execution initiated - client should poll GetTransferStatus")
	}

	return results, nil
}

// validateExecuteTransferParams validates the input parameters for transfer execution
func validateExecuteTransferParams(params *ExecuteTransferParams) error {
	if params.FromAccount == "" {
		return fmt.Errorf("from_account is required")
	}

	if params.ToAccount == "" {
		return fmt.Errorf("to_account is required")
	}

	if params.FromAccount == params.ToAccount {
		return fmt.Errorf("from_account and to_account cannot be the same")
	}

	if params.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if params.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if params.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}

	// Note: WaitForCompletion is optional and defaults to false (async mode)
	// - false: Async mode - returns immediately, client polls GetTransferStatus
	// - true: Sync mode - waits for workflow completion, returns final result

	// Validate currency format (basic validation)
	validCurrencies := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true,
		"CAD": true, "AUD": true, "CHF": true, "CNY": true,
		"SGD": true, "HKD": true, "IDR": true,
	}

	if !validCurrencies[params.Currency] {
		return fmt.Errorf("unsupported currency: %s", params.Currency)
	}

	return nil
}
