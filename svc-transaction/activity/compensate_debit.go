package activity

import (
	"context"
	"fmt"

	"svc-transaction/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
)

// CompensateDebitActivityParams defines parameters for the CompensateDebit activity
// This matches the structure expected by the workflow
type CompensateDebitActivityParams struct {
	OriginalTransactionID string          `json:"original_transaction_id"`
	AccountID             string          `json:"account_id"`
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	CompensationReason    string          `json:"compensation_reason"`
	ReferenceID           string          `json:"reference_id"`
	IdempotencyKey        string          `json:"idempotency_key"`
	TransferID            string          `json:"transfer_id"`
	WorkflowID            string          `json:"workflow_id"`
	RunID                 string          `json:"run_id"`
}

// CompensateDebitActivityResults defines results from the CompensateDebit activity
// This matches the structure expected by the workflow
type CompensateDebitActivityResults struct {
	TransactionID         string          `json:"transaction_id"`
	AccountID             string          `json:"account_id"`
	AccountNumber         string          `json:"account_number"`
	AccountName           string          `json:"account_name"`
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	Description           string          `json:"description"`
	ReferenceID           string          `json:"reference_id"`
	IdempotencyKey        string          `json:"idempotency_key"`
	Status                string          `json:"status"`
	PreviousBalance       decimal.Decimal `json:"previous_balance"`
	NewBalance            decimal.Decimal `json:"new_balance"`
	CreatedAt             string          `json:"created_at"`
	CompletedAt           string          `json:"completed_at"`
	OriginalTransactionID string          `json:"original_transaction_id"`
	CompensationReason    string          `json:"compensation_reason"`
}

// CompensateDebit is the Temporal activity that handles CompensateDebit requests
func (api *Activity) CompensateDebit(ctx context.Context, params CompensateDebitActivityParams) (*CompensateDebitActivityResults, error) {
	const op = "activity.Activity.CompensateDebit"

	// Get activity info for logging
	activityInfo := activity.GetInfo(ctx)

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":                    op,
		"activity_id":             activityInfo.ActivityID,
		"activity_type":           activityInfo.ActivityType.Name,
		"workflow_id":             params.WorkflowID,
		"run_id":                  params.RunID,
		"transfer_id":             params.TransferID,
		"account_id":              params.AccountID,
		"original_transaction_id": params.OriginalTransactionID,
	})

	logger.WithField("message", "Starting CompensateDebit activity").Info()

	// PERFORMANCE OPTIMIZATION: Record heartbeat for long-running activity monitoring
	activity.RecordHeartbeat(ctx, "CompensateDebit_started")

	// FAILURE SIMULATION: Check if we should inject a failure
	if err := api.service.SimulateFailure(ctx, "CompensateDebit", params.AccountID); err != nil {
		logger.WithError(err).Warn("🚨 Transaction failure simulation triggered")
		return nil, err
	}

	// PERFORMANCE OPTIMIZATION: Record heartbeat before parsing
	activity.RecordHeartbeat(ctx, "CompensateDebit_parsing")

	// Parse account ID
	accountID, err := uuid.Parse(params.AccountID)
	if err != nil {
		err = fmt.Errorf("invalid account_id format: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Parse original transaction ID
	originalTransactionID, err := uuid.Parse(params.OriginalTransactionID)
	if err != nil {
		err = fmt.Errorf("invalid original_transaction_id format: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Convert to service parameters
	serviceParams := service.CompensateDebitParams{
		OriginalTransactionID: &originalTransactionID,
		AccountID:             &accountID,
		Amount:                params.Amount,
		Currency:              params.Currency,
		Description:           &params.CompensationReason,
		ReferenceID:           &params.ReferenceID,
		IdempotencyKey:        &params.IdempotencyKey,
		CompensationReason:    &params.CompensationReason,
		WorkflowID:            &params.WorkflowID,
		RunID:                 &params.RunID,
		Metadata: map[string]any{
			"transfer_id": params.TransferID,
			"workflow_id": params.WorkflowID,
			"run_id":      params.RunID,
			"activity_id": activityInfo.ActivityID,
		},
	}

	// PERFORMANCE OPTIMIZATION: Record heartbeat before service call
	activity.RecordHeartbeat(ctx, "CompensateDebit_service_call")

	// Call the service method
	result, err := api.service.CompensateDebit(ctx, serviceParams)
	if err != nil {
		err = fmt.Errorf("compensate debit failed: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// PERFORMANCE OPTIMIZATION: Record heartbeat after service completion
	activity.RecordHeartbeat(ctx, "CompensateDebit_completed")

	// Convert to activity result format
	activityResult := &CompensateDebitActivityResults{
		TransactionID:         result.TransactionID.String(),
		AccountID:             result.AccountID.String(),
		AccountNumber:         result.AccountNumber,
		AccountName:           result.AccountName,
		Amount:                result.Amount,
		Currency:              result.Currency,
		Description:           *result.Description,
		ReferenceID:           *result.ReferenceID,
		IdempotencyKey:        *result.IdempotencyKey,
		Status:                result.Status,
		PreviousBalance:       result.PreviousBalance,
		NewBalance:            result.NewBalance,
		CreatedAt:             result.CreatedAt,
		CompletedAt:           *result.CompletedAt,
		OriginalTransactionID: result.OriginalTransactionID.String(),
		CompensationReason:    *result.CompensationReason,
	}

	logger.WithField("result", fmt.Sprintf("%+v", activityResult)).Info()

	return activityResult, nil
}
