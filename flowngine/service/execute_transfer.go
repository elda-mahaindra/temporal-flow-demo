package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type ExecuteTransferParams struct {
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	ReferenceID string `json:"reference_id"`
	RequestID   string `json:"request_id"`
}

type ExecuteTransferResults struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	WorkflowID    string `json:"workflow_id"`
	RunID         string `json:"run_id"`
	CreatedAt     string `json:"created_at"`
}

func (service *Service) ExecuteTransfer(ctx context.Context, params *ExecuteTransferParams) (*ExecuteTransferResults, error) {
	const op = "service.Service.ExecuteTransfer"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Initialize results
	results := &ExecuteTransferResults{}

	// TODO: Implement cancel transfer

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info()

	return results, nil
}
