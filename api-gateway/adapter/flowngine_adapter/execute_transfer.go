package flowngine_adapter

import (
	"context"
	"fmt"

	"api-gateway/adapter/flowngine_adapter/pb"

	"github.com/sirupsen/logrus"
)

func (adapter *Adapter) ExecuteTransfer(ctx context.Context, request *pb.ExecuteTransferRequest) (response *pb.ExecuteTransferResponse, err error) {
	const op = "flowngine_adapter.Adapter.ExecuteTransfer"

	logger := adapter.logger.WithFields(logrus.Fields{
		"[op]":    op,
		"request": fmt.Sprintf("%+v", request),
		"type":    fmt.Sprintf("%T", request),
	})

	logger.Info()

	// Call service
	response, err = adapter.serviceBClient.ExecuteTransfer(ctx, request)
	if err != nil {
		logger.WithError(err).Error()

		return nil, err
	}

	logger.WithField("response", fmt.Sprintf("%+v", response)).Info()

	return response, nil
}
