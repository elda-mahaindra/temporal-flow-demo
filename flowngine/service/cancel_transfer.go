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
}

func (service *Service) CancelTransfer(ctx context.Context, params *CancelTransferParams) (*CancelTransferResults, error) {
	const op = "service.Service.CancelTransfer"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Initialize results
	results := &CancelTransferResults{}

	// TODO: Implement cancel transfer

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info()

	return results, nil
}
