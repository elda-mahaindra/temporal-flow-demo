package service

import (
	"context"
	"fmt"
	"time"

	"api-gateway/adapter/flowngine_adapter/pb"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TransferParams struct {
	FromAccount       string  `json:"from_account"`
	ToAccount         string  `json:"to_account"`
	Amount            int     `json:"amount"`
	Currency          string  `json:"currency"`
	Description       *string `json:"description"`
	ReferenceID       *string `json:"reference_id"`
	WaitForCompletion bool    `json:"wait_for_completion"` // Sync vs async mode
}

type TransferResults struct {
	TransactionID       string `json:"transaction_id"`
	Status              string `json:"status"`
	FromAccount         string `json:"from_account"`
	ToAccount           string `json:"to_account"`
	Amount              int    `json:"amount"`
	Currency            string `json:"currency"`
	Description         string `json:"description"`
	ReferenceID         string `json:"reference_id"`
	CreatedAt           string `json:"created_at"`
	EstimatedCompletion string `json:"estimated_completion"`
	// Fields for sync mode (when WaitForCompletion=true)
	CompletedAt         *string `json:"completed_at,omitempty"`
	ErrorMessage        string  `json:"error_message,omitempty"`
	CompensationApplied *bool   `json:"compensation_applied,omitempty"`
	WorkflowID          string  `json:"workflow_id"`
	RunID               string  `json:"run_id"`
}

func (service *Service) Transfer(ctx context.Context, params *TransferParams) (results *TransferResults, err error) {
	const op = "service.Service.Transfer"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info("Initiating transfer through FlowEngine")

	// Generate request ID for idempotency
	requestID := uuid.New().String()

	// Set default values for optional fields
	description := ""
	if params.Description != nil {
		description = *params.Description
	}

	referenceID := ""
	if params.ReferenceID != nil {
		referenceID = *params.ReferenceID
	}

	// Create FlowEngine request
	flowEngineRequest := &pb.ExecuteTransferRequest{
		FromAccount: params.FromAccount,
		ToAccount:   params.ToAccount,
		Amount:      int64(params.Amount),
		Currency:    params.Currency,
		Description: description,
		ReferenceId: referenceID,
		RequestId:   requestID,
	}

	// Call FlowEngine adapter
	flowEngineResponse, err := service.flowngineAdapter.ExecuteTransfer(ctx, flowEngineRequest)
	if err != nil {
		err = fmt.Errorf("failed to execute transfer via FlowEngine: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Convert status enum to string
	statusString := flowEngineResponse.Status.String()

	// Convert timestamp to string
	createdAtString := flowEngineResponse.CreatedAt.AsTime().Format(time.RFC3339)

	// Estimate completion time (adding 2 minutes as default)
	estimatedCompletion := flowEngineResponse.CreatedAt.AsTime().Add(2 * time.Minute).Format(time.RFC3339)

	// Initialize results
	results = &TransferResults{
		TransactionID:       flowEngineResponse.TransactionId,
		Status:              statusString,
		FromAccount:         params.FromAccount,
		ToAccount:           params.ToAccount,
		Amount:              params.Amount,
		Currency:            params.Currency,
		Description:         description,
		ReferenceID:         referenceID,
		CreatedAt:           createdAtString,
		EstimatedCompletion: estimatedCompletion,
		WorkflowID:          flowEngineResponse.WorkflowId,
		RunID:               flowEngineResponse.RunId,
	}

	// For sync mode, wait for completion
	if params.WaitForCompletion {
		logger.Info("Waiting for transfer completion (sync mode)")

		// Poll for status updates until completion or timeout
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.After(5 * time.Minute) // 5 minute timeout for sync mode

		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-timeout:
				logger.Warn("Transfer completion timeout reached")
				results.ErrorMessage = "Transfer completion timeout reached"
				return results, nil
			case <-ticker.C:
				// Check status
				statusResponse, err := service.checkTransferStatus(ctx, flowEngineResponse.TransactionId)
				if err != nil {
					logger.WithError(err).Warn("Failed to check transfer status")
					continue
				}

				// Update status
				results.Status = statusResponse.Status.String()

				// Check if completed (success or failure)
				if statusResponse.Status == pb.TransferStatus_TRANSFER_STATUS_COMPLETED ||
					statusResponse.Status == pb.TransferStatus_TRANSFER_STATUS_FAILED ||
					statusResponse.Status == pb.TransferStatus_TRANSFER_STATUS_COMPENSATED ||
					statusResponse.Status == pb.TransferStatus_TRANSFER_STATUS_CANCELLED {

					// Set completion time
					if statusResponse.CompletedAt != nil {
						completedAt := statusResponse.CompletedAt.AsTime().Format(time.RFC3339)
						results.CompletedAt = &completedAt
					}

					// Set error message if failed
					if statusResponse.ErrorMessage != "" {
						results.ErrorMessage = statusResponse.ErrorMessage
					}

					// Set compensation flag
					if statusResponse.Status == pb.TransferStatus_TRANSFER_STATUS_COMPENSATED {
						compensated := true
						results.CompensationApplied = &compensated
					}

					logger.WithField("final_status", statusResponse.Status.String()).Info("Transfer completed")
					return results, nil
				}
			}
		}
	}

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info("Transfer initiated successfully")

	return results, nil
}

// checkTransferStatus is a helper method to poll transfer status
func (service *Service) checkTransferStatus(ctx context.Context, transactionID string) (*pb.GetTransferStatusResponse, error) {
	statusRequest := &pb.GetTransferStatusRequest{
		TransactionId: transactionID,
	}

	return service.flowngineAdapter.GetTransferStatus(ctx, statusRequest)
}

type GetTransferParams struct {
	TransactionID string `json:"transaction_id"`
}

type GetTransferResults struct {
	TransactionID     string `json:"transaction_id"`
	Status            string `json:"status"`
	FromAccount       string `json:"from_account"`
	ToAccount         string `json:"to_account"`
	Amount            int    `json:"amount"`
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
}

func (service *Service) GetTransfer(ctx context.Context, params *GetTransferParams) (results *GetTransferResults, err error) {
	const op = "service.Service.GetTransfer"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info("Getting transfer status from FlowEngine")

	// Create FlowEngine request
	statusRequest := &pb.GetTransferStatusRequest{
		TransactionId: params.TransactionID,
	}

	// Call FlowEngine adapter
	statusResponse, err := service.flowngineAdapter.GetTransferStatus(ctx, statusRequest)
	if err != nil {
		err = fmt.Errorf("failed to get transfer status from FlowEngine: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Convert timestamps to strings
	createdAt := statusResponse.CreatedAt.AsTime().Format(time.RFC3339)
	completedAt := ""
	if statusResponse.CompletedAt != nil {
		completedAt = statusResponse.CompletedAt.AsTime().Format(time.RFC3339)
	}

	// Initialize results
	results = &GetTransferResults{
		TransactionID: statusResponse.TransactionId,
		Status:        statusResponse.Status.String(),
		FromAccount:   statusResponse.FromAccount,
		ToAccount:     statusResponse.ToAccount,
		Amount:        int(statusResponse.Amount),
		Currency:      statusResponse.Currency,
		Description:   statusResponse.Description,
		ReferenceID:   statusResponse.ReferenceId,
		CreatedAt:     createdAt,
		CompletedAt:   completedAt,
	}

	// Set workflow execution details
	if statusResponse.WorkflowExecution != nil {
		results.WorkflowExecution.WorkflowID = statusResponse.WorkflowExecution.WorkflowId
		results.WorkflowExecution.RunID = statusResponse.WorkflowExecution.RunId
		results.WorkflowExecution.Status = statusResponse.WorkflowExecution.Status
	}

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info("Transfer status retrieved successfully")

	return results, nil
}
