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

// DebitAccountActivityParams defines parameters for the DebitAccount activity
// This matches the structure expected by the workflow
type DebitAccountActivityParams struct {
	AccountID      string          `json:"account_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Description    string          `json:"description"`
	ReferenceID    string          `json:"reference_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	TransferID     string          `json:"transfer_id"`
	WorkflowID     string          `json:"workflow_id"`
	RunID          string          `json:"run_id"`
}

// DebitAccountActivityResults defines results from the DebitAccount activity
// This matches the structure expected by the workflow
type DebitAccountActivityResults struct {
	TransactionID   string          `json:"transaction_id"`
	AccountID       string          `json:"account_id"`
	AccountNumber   string          `json:"account_number"`
	AccountName     string          `json:"account_name"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	Description     string          `json:"description"`
	ReferenceID     string          `json:"reference_id"`
	IdempotencyKey  string          `json:"idempotency_key"`
	Status          string          `json:"status"`
	PreviousBalance decimal.Decimal `json:"previous_balance"`
	NewBalance      decimal.Decimal `json:"new_balance"`
	CreatedAt       string          `json:"created_at"`
	CompletedAt     string          `json:"completed_at"`
}

// DebitAccount is the Temporal activity that handles DebitAccount requests
func (api *Activity) DebitAccount(ctx context.Context, params DebitAccountActivityParams) (*DebitAccountActivityResults, error) {
	const op = "activity.Activity.DebitAccount"

	// Get activity info for logging
	activityInfo := activity.GetInfo(ctx)

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":          op,
		"activity_id":   activityInfo.ActivityID,
		"activity_type": activityInfo.ActivityType.Name,
		"workflow_id":   params.WorkflowID,
		"run_id":        params.RunID,
		"transfer_id":   params.TransferID,
		"account_id":    params.AccountID,
	})

	logger.WithField("message", "Starting DebitAccount activity").Info()

	// PERFORMANCE OPTIMIZATION: Record heartbeat for long-running activity monitoring
	activity.RecordHeartbeat(ctx, "DebitAccount_started")

	// FAILURE SIMULATION: Check if we should inject a failure
	if err := api.service.SimulateFailure(ctx, "DebitAccount", params.AccountID); err != nil {
		logger.WithError(err).Warn("🚨 Transaction failure simulation triggered")
		return nil, err
	}

	// PERFORMANCE OPTIMIZATION: Record heartbeat before parsing
	activity.RecordHeartbeat(ctx, "DebitAccount_parsing")

	// Parse account ID
	accountID, err := uuid.Parse(params.AccountID)
	if err != nil {
		err = fmt.Errorf("invalid account_id format: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Convert to service parameters
	serviceParams := service.DebitAccountParams{
		AccountID:      &accountID,
		Amount:         params.Amount,
		Currency:       params.Currency,
		Description:    &params.Description,
		ReferenceID:    &params.ReferenceID,
		IdempotencyKey: &params.IdempotencyKey,
		Metadata: map[string]any{
			"transfer_id": params.TransferID,
			"workflow_id": params.WorkflowID,
			"run_id":      params.RunID,
			"activity_id": activityInfo.ActivityID,
		},
	}

	// PERFORMANCE OPTIMIZATION: Record heartbeat before service call
	activity.RecordHeartbeat(ctx, "DebitAccount_service_call")

	// Call the service method
	result, err := api.service.DebitAccount(ctx, serviceParams)
	if err != nil {
		err = fmt.Errorf("debit account failed: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// PERFORMANCE OPTIMIZATION: Record heartbeat after service completion
	activity.RecordHeartbeat(ctx, "DebitAccount_completed")

	// Convert to activity result format
	activityResult := &DebitAccountActivityResults{
		TransactionID:   result.TransactionID.String(),
		AccountID:       result.AccountID.String(),
		AccountNumber:   result.AccountNumber,
		AccountName:     result.AccountName,
		Amount:          result.Amount,
		Currency:        result.Currency,
		Description:     *result.Description,
		ReferenceID:     *result.ReferenceID,
		IdempotencyKey:  *result.IdempotencyKey,
		Status:          result.Status,
		PreviousBalance: result.PreviousBalance,
		NewBalance:      result.NewBalance,
		CreatedAt:       result.CreatedAt,
		CompletedAt:     *result.CompletedAt,
	}

	logger.WithField("result", fmt.Sprintf("%+v", activityResult)).Info()

	return activityResult, nil
}
