package api

import (
	"context"
	"fmt"
	"time"

	"flowngine/api/pb"
	"flowngine/service"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (api *Api) ExecuteTransfer(ctx context.Context, request *pb.ExecuteTransferRequest) (*pb.ExecuteTransferResponse, error) {
	const op = "api.Api.ExecuteTransfer"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":    op,
		"request": fmt.Sprintf("%+v", request),
	})

	logger.Info()

	// Initialize response
	response := &pb.ExecuteTransferResponse{}

	// Call service
	params := &service.ExecuteTransferParams{
		FromAccount: request.FromAccount,
		ToAccount:   request.ToAccount,
		Amount:      request.Amount,
		Currency:    request.Currency,
		Description: request.Description,
		ReferenceID: request.ReferenceId,
		RequestID:   request.RequestId,
	}

	results, err := api.service.ExecuteTransfer(ctx, params)
	if err != nil {
		logger.WithError(err).Error()

		return nil, err
	}

	// Set response
	response.TransactionId = results.TransactionID
	response.Status = pb.TransferStatus(pb.TransferStatus_value[results.Status])
	response.WorkflowId = results.WorkflowID
	response.RunId = results.RunID
	createdAt, err := time.Parse(time.RFC3339, results.CreatedAt)
	if err != nil {
		err = fmt.Errorf("failed to parse created_at timestamp: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}
	response.CreatedAt = timestamppb.New(createdAt)

	logger.WithField("request", fmt.Sprintf("%+v", request)).Info()

	return response, nil
}
