package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type CancelTransferParams struct {
	TransactionID string `json:"transaction_id"`
	Reason        string `json:"reason"`
}

type CancelTransferResults struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Cancelled bool   `json:"cancelled"`
}

func (svc *Service) CancelTransfer(ctx context.Context, params *CancelTransferParams) (*CancelTransferResults, error) {
	const op = "service.Service.CancelTransfer"

	logger := svc.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info("Cancelling transfer")

	// Validate input parameters
	if err := validateCancelTransferParams(params); err != nil {
		err = fmt.Errorf("invalid parameters: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Generate workflow ID from transaction ID (following the same pattern as ExecuteTransfer)
	workflowID := fmt.Sprintf("transfer_workflow_%s", params.TransactionID)

	logger.Info("Cancelling Temporal workflow", "workflow_id", workflowID, "transaction_id", params.TransactionID, "reason", params.Reason)

	// 🎯 THIS IS THE TEMPORAL MAGIC!
	// Just cancel the workflow - Temporal handles ALL the complexity:
	// - Graceful workflow termination
	// - Activity cancellation
	// - Automatic compensation if needed
	// - State cleanup
	// - No custom database state management required!
	err := svc.temporalClient.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		err = fmt.Errorf("failed to cancel workflow: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	logger.Info("🎉 Temporal workflow cancelled! All cleanup and compensation handled automatically",
		"workflow_id", workflowID,
		"transaction_id", params.TransactionID,
		"reason", params.Reason)

	// Return success response
	message := fmt.Sprintf("Transfer %s cancelled successfully: %s", params.TransactionID, params.Reason)

	results := &CancelTransferResults{
		Success:   true,
		Message:   message,
		Cancelled: true,
	}

	logger.Info("Transfer cancellation completed - Temporal handled all the complexity!",
		"transaction_id", params.TransactionID,
		"workflow_id", workflowID)

	return results, nil
}

// validateCancelTransferParams validates the input parameters for transfer cancellation
func validateCancelTransferParams(params *CancelTransferParams) error {
	if params.TransactionID == "" {
		return fmt.Errorf("transaction_id is required")
	}

	if params.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	return nil
}
