package activity

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCheckBalanceActivityParams(t *testing.T) {
	params := CheckBalanceActivityParams{
		AccountID:      "550e8400-e29b-41d4-a716-446655440000",
		RequiredAmount: decimal.NewFromFloat(100.50),
		Currency:       "USD",
		TransferID:     "transfer-123",
		WorkflowID:     "workflow-456",
		RunID:          "run-789",
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", params.AccountID)
	assert.Equal(t, decimal.NewFromFloat(100.50), params.RequiredAmount)
	assert.Equal(t, "USD", params.Currency)
	assert.Equal(t, "transfer-123", params.TransferID)
	assert.Equal(t, "workflow-456", params.WorkflowID)
	assert.Equal(t, "run-789", params.RunID)
}

func TestCheckBalanceActivityResults(t *testing.T) {
	results := CheckBalanceActivityResults{
		AccountID:       "550e8400-e29b-41d4-a716-446655440000",
		CurrentBalance:  decimal.NewFromFloat(200.75),
		RequiredAmount:  decimal.NewFromFloat(100.50),
		SufficientFunds: true,
		Currency:        "USD",
		CheckedAt:       "2023-12-01T10:30:00Z",
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", results.AccountID)
	assert.Equal(t, decimal.NewFromFloat(200.75), results.CurrentBalance)
	assert.Equal(t, decimal.NewFromFloat(100.50), results.RequiredAmount)
	assert.True(t, results.SufficientFunds)
	assert.Equal(t, "USD", results.Currency)
	assert.Equal(t, "2023-12-01T10:30:00Z", results.CheckedAt)
}
