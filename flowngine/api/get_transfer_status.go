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

func (api *Api) GetTransferStatus(ctx context.Context, request *pb.GetTransferStatusRequest) (*pb.GetTransferStatusResponse, error) {
	const op = "api.Api.GetTransferStatus"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":    op,
		"request": fmt.Sprintf("%+v", request),
	})

	logger.Info()

	// Initialize response
	response := &pb.GetTransferStatusResponse{}

	// Call service
	params := &service.GetTransferStatusParams{
		TransactionID: request.TransactionId,
	}

	results, err := api.service.GetTransferStatus(ctx, params)
	if err != nil {
		logger.WithError(err).Error()

		return nil, err
	}

	// Set response
	response.TransactionId = results.TransactionID
	response.Status = pb.TransferStatus(pb.TransferStatus_value[results.Status])
	response.FromAccount = results.FromAccount
	response.ToAccount = results.ToAccount
	response.Amount = results.Amount
	response.Currency = results.Currency
	response.Description = results.Description
	response.ReferenceId = results.ReferenceID
	createdAt, err := time.Parse(time.RFC3339, results.CreatedAt)
	if err != nil {
		err = fmt.Errorf("failed to parse created_at timestamp: %w", err)

		logger.WithError(err).Error()
		logger.WithError(err).Error()

		return nil, err
	}
	response.CreatedAt = timestamppb.New(createdAt)
	completedAt, err := time.Parse(time.RFC3339, results.CompletedAt)
	if err != nil {
		err = fmt.Errorf("failed to parse completed_at timestamp: %w", err)

		logger.WithError(err).Error()

		return nil, err
	}
	response.CompletedAt = timestamppb.New(completedAt)
	response.WorkflowExecution = &pb.WorkflowExecution{
		WorkflowId: results.WorkflowExecution.WorkflowID,
		RunId:      results.WorkflowExecution.RunID,
		Status:     results.WorkflowExecution.Status,
	}
	response.ErrorMessage = results.ErrorMessage

	logger.WithField("request", fmt.Sprintf("%+v", request)).Info()

	return response, nil
}
