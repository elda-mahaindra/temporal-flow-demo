package activity

import (
	"context"
	"fmt"

	"svc-balance/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
)

// CheckBalanceActivityParams defines parameters for the CheckBalance activity
// This matches the structure expected by the workflow
type CheckBalanceActivityParams struct {
	AccountID      string          `json:"account_id"`
	RequiredAmount decimal.Decimal `json:"required_amount"`
	Currency       string          `json:"currency"`
	TransferID     string          `json:"transfer_id"`
	WorkflowID     string          `json:"workflow_id"`
	RunID          string          `json:"run_id"`
}

// CheckBalanceActivityResults defines results from the CheckBalance activity
// This matches the structure expected by the workflow
type CheckBalanceActivityResults struct {
	AccountID       string          `json:"account_id"`
	CurrentBalance  decimal.Decimal `json:"current_balance"`
	RequiredAmount  decimal.Decimal `json:"required_amount"`
	SufficientFunds bool            `json:"sufficient_funds"`
	Currency        string          `json:"currency"`
	CheckedAt       string          `json:"checked_at"`
}

// CheckBalance is the Temporal activity that handles CheckBalance requests
func (api *Activity) CheckBalance(ctx context.Context, params CheckBalanceActivityParams) (*CheckBalanceActivityResults, error) {
	const op = "activity.Activity.CheckBalance"

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

	logger.WithField("message", "Starting CheckBalance activity").Info()

	// FAILURE SIMULATION: Check if we should inject a failure
	if err := api.service.SimulateFailure(ctx, "CheckBalance", params.AccountID); err != nil {
		logger.WithError(err).Warn("Failure simulation triggered")
		return nil, err
	}

	// Parse account ID
	accountID, err := uuid.Parse(params.AccountID)
	if err != nil {
		err = fmt.Errorf("invalid account_id format: %w", err)

		logger.WithError(err).Error("Failed to parse account ID")

		return nil, err
	}

	// Convert to service parameters
	serviceParams := service.CheckBalanceParams{
		AccountID:        &accountID,
		RequiredAmount:   &params.RequiredAmount,
		ExpectedCurrency: &params.Currency,
		IncludeDetails:   false, // Keep it simple for workflow activities
	}

	// Call the service method
	result, err := api.service.CheckBalance(ctx, serviceParams)
	if err != nil {
		err = fmt.Errorf("balance check failed: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}

	// Convert to activity result format
	activityResult := &CheckBalanceActivityResults{
		AccountID:       result.AccountID.String(),
		CurrentBalance:  result.CurrentBalance,
		RequiredAmount:  params.RequiredAmount,
		SufficientFunds: result.SufficientFunds,
		Currency:        result.Currency,
		CheckedAt:       activityInfo.StartedTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	logger.WithField("result", fmt.Sprintf("%+v", activityResult)).Info()

	return activityResult, nil
}
