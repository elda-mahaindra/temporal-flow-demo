package activity

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDebitAccountActivityParams(t *testing.T) {
	params := DebitAccountActivityParams{
		AccountID:      "550e8400-e29b-41d4-a716-446655440000",
		Amount:         decimal.NewFromFloat(100.50),
		Currency:       "USD",
		Description:    "Test debit transaction",
		ReferenceID:    "ref-123",
		IdempotencyKey: "idemp-456",
		TransferID:     "transfer-789",
		WorkflowID:     "workflow-abc",
		RunID:          "run-def",
	}

	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", params.AccountID)
	assert.Equal(t, decimal.NewFromFloat(100.50), params.Amount)
	assert.Equal(t, "USD", params.Currency)
	assert.Equal(t, "Test debit transaction", params.Description)
	assert.Equal(t, "ref-123", params.ReferenceID)
	assert.Equal(t, "idemp-456", params.IdempotencyKey)
	assert.Equal(t, "transfer-789", params.TransferID)
	assert.Equal(t, "workflow-abc", params.WorkflowID)
	assert.Equal(t, "run-def", params.RunID)
}

func TestDebitAccountActivityResults(t *testing.T) {
	results := DebitAccountActivityResults{
		TransactionID:   "txn-123",
		AccountID:       "550e8400-e29b-41d4-a716-446655440000",
		AccountNumber:   "ACC-001",
		AccountName:     "Test Account",
		Amount:          decimal.NewFromFloat(100.50),
		Currency:        "USD",
		Description:     "Test debit transaction",
		ReferenceID:     "ref-123",
		IdempotencyKey:  "idemp-456",
		Status:          "completed",
		PreviousBalance: decimal.NewFromFloat(1000.00),
		NewBalance:      decimal.NewFromFloat(899.50),
		CreatedAt:       "2023-01-01T00:00:00Z",
		CompletedAt:     "2023-01-01T00:00:01Z",
	}

	assert.Equal(t, "txn-123", results.TransactionID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", results.AccountID)
	assert.Equal(t, "ACC-001", results.AccountNumber)
	assert.Equal(t, "Test Account", results.AccountName)
	assert.Equal(t, decimal.NewFromFloat(100.50), results.Amount)
	assert.Equal(t, "USD", results.Currency)
	assert.Equal(t, "Test debit transaction", results.Description)
	assert.Equal(t, "ref-123", results.ReferenceID)
	assert.Equal(t, "idemp-456", results.IdempotencyKey)
	assert.Equal(t, "completed", results.Status)
	assert.Equal(t, decimal.NewFromFloat(1000.00), results.PreviousBalance)
	assert.Equal(t, decimal.NewFromFloat(899.50), results.NewBalance)
	assert.Equal(t, "2023-01-01T00:00:00Z", results.CreatedAt)
	assert.Equal(t, "2023-01-01T00:00:01Z", results.CompletedAt)
}
