package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type TransferParams struct {
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      int     `json:"amount"`
	Currency    string  `json:"currency"`
	Description *string `json:"description"`
	ReferenceID *string `json:"reference_id"`
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
}

func (service *Service) Transfer(ctx context.Context, params *TransferParams) (results *TransferResults, err error) {
	const op = "service.Service.Transfer"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Initialize results
	results = &TransferResults{}

	// TODO: Implement transfer

	// TODO: Set results

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info()

	return
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

	logger.Info()

	// Initialize results
	results = &GetTransferResults{}

	// TODO: Implement get transfer

	// TODO: Set results

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info()

	return
}
