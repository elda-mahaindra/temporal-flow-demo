package api

import (
	"context"
	"fmt"

	"flowngine/api/pb"
	"flowngine/service"

	"github.com/sirupsen/logrus"
)

func (api *Api) CancelTransfer(ctx context.Context, request *pb.CancelTransferRequest) (*pb.CancelTransferResponse, error) {
	const op = "api.Api.CancelTransfer"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":    op,
		"request": fmt.Sprintf("%+v", request),
	})

	logger.Info()

	// Initialize response
	response := &pb.CancelTransferResponse{}

	// Call service
	params := &service.CancelTransferParams{
		TransactionID: request.TransactionId,
		Reason:        request.Reason,
	}

	_, err := api.service.CancelTransfer(ctx, params)
	if err != nil {
		logger.WithError(err).Error()

		return nil, err
	}

	// Set response
	response.Success = true
	response.Message = "Transfer cancelled successfully"

	logger.WithField("request", fmt.Sprintf("%+v", request)).Info()

	return response, nil
}
