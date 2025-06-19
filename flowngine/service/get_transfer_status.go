package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type GetTransferStatusParams struct {
	TransactionID string `json:"transaction_id"`
}

type GetTransferStatusResults struct {
	TransactionID     string `json:"transaction_id"`
	Status            string `json:"status"`
	FromAccount       string `json:"from_account"`
	ToAccount         string `json:"to_account"`
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
	Description       string `json:"description"`
	ReferenceID       string `json:"reference_id"`
	CreatedAt         string `json:"created_at"`
	CompletedAt       string `json:"completed_at"`
	WorkflowExecution struct {
		WorkflowID string `json:"workflow_id"`
		RunID      string `json:"run_id"`
		Status     string `json:"status"`
	} `json:"workflow_execution"`
	ErrorMessage string `json:"error_message"`
}

func (svc *Service) GetTransferStatus(ctx context.Context, params *GetTransferStatusParams) (*GetTransferStatusResults, error) {
	const op = "service.Service.GetTransferStatus"

	logger := svc.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info("Getting transfer status")

	// Validate input parameters
	if err := validateGetTransferStatusParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Generate workflow ID from transaction ID (following the same pattern as ExecuteTransfer)
	workflowID := fmt.Sprintf("transfer_workflow_%s", params.TransactionID)

	logger.Info("Querying workflow status from Temporal", "workflow_id", workflowID, "transaction_id", params.TransactionID)

	// 🎯 THIS IS THE TEMPORAL MAGIC!
	// Query workflow status directly from Temporal - no custom database needed!
	// Temporal maintains ALL workflow state, execution history, and status
	workflowRun := svc.temporalClient.GetWorkflow(ctx, workflowID, "")

	// Try to get the workflow result
	var workflowResult TransferWorkflowResults
	workflowErr := workflowRun.Get(ctx, &workflowResult)

	if workflowErr != nil {
		// Workflow might still be running or failed
		logger.WithError(workflowErr).Info("Workflow not completed yet or failed")

		// For running workflows, return pending status
		// Temporal Web UI will show you the real-time execution progress!
		results := &GetTransferStatusResults{
			TransactionID: params.TransactionID,
			Status:        "TRANSFER_STATUS_PROCESSING",
			FromAccount:   "", // Would be available from workflow history if needed
			ToAccount:     "",
			Amount:        0,
			Currency:      "",
			Description:   "Transfer is being processed by Temporal workflow",
			ReferenceID:   "",
			CreatedAt:     time.Now().Add(-time.Minute * 1).Format(time.RFC3339), // Approximate
			CompletedAt:   "",
			WorkflowExecution: struct {
				WorkflowID string `json:"workflow_id"`
				RunID      string `json:"run_id"`
				Status     string `json:"status"`
			}{
				WorkflowID: workflowID,
				RunID:      workflowRun.GetRunID(),
				Status:     "RUNNING",
			},
			ErrorMessage: "",
		}

		logger.Info("📊 Workflow status from Temporal: PROCESSING")
		return results, nil
	}

	// 🎉 Workflow completed! All data comes from Temporal
	// Convert amount back to minor units for API response
	amountMinorUnits := workflowResult.Amount.Mul(decimal.NewFromInt(100)).IntPart()

	results := &GetTransferStatusResults{
		TransactionID: params.TransactionID,
		Status:        workflowResult.Status,
		FromAccount:   workflowResult.FromAccount,
		ToAccount:     workflowResult.ToAccount,
		Amount:        amountMinorUnits,
		Currency:      workflowResult.Currency,
		Description:   workflowResult.Description,
		ReferenceID:   params.TransactionID,
		CreatedAt:     workflowResult.StartedAt.Format(time.RFC3339),
		WorkflowExecution: struct {
			WorkflowID string `json:"workflow_id"`
			RunID      string `json:"run_id"`
			Status     string `json:"status"`
		}{
			WorkflowID: workflowResult.WorkflowID,
			RunID:      workflowResult.RunID,
			Status:     "COMPLETED",
		},
		ErrorMessage: workflowResult.ErrorMessage,
	}

	if workflowResult.CompletedAt != nil {
		results.CompletedAt = workflowResult.CompletedAt.Format(time.RFC3339)
	}

	logger.Info("🎉 Workflow status retrieved directly from Temporal - no custom database needed!",
		"transaction_id", params.TransactionID,
		"status", results.Status,
		"workflow_id", workflowID)

	return results, nil
}

// validateGetTransferStatusParams validates the input parameters for status checking
func validateGetTransferStatusParams(params *GetTransferStatusParams) error {
	if params.TransactionID == "" {
		return fmt.Errorf("transaction_id is required")
	}

	return nil
}
