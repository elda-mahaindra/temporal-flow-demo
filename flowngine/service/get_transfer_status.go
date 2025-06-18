package service

import (
	"context"
	"fmt"

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

func (service *Service) GetTransferStatus(ctx context.Context, params *GetTransferStatusParams) (*GetTransferStatusResults, error) {
	const op = "service.Service.GetTransferStatus"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Initialize results
	results := &GetTransferStatusResults{}

	// TODO: Implement get transfer status

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info()

	return results, nil
}
