package flowngine_adapter

import (
	"context"
	"fmt"

	"api-gateway/adapter/flowngine_adapter/pb"

	"github.com/sirupsen/logrus"
)

func (adapter *Adapter) GetTransferStatus(ctx context.Context, request *pb.GetTransferStatusRequest) (response *pb.GetTransferStatusResponse, err error) {
	const op = "flowngine_adapter.Adapter.GetTransferStatus"

	logger := adapter.logger.WithFields(logrus.Fields{
		"[op]":    op,
		"request": fmt.Sprintf("%+v", request),
		"type":    fmt.Sprintf("%T", request),
	})

	logger.Info()

	// Call service
	response, err = adapter.serviceBClient.GetTransferStatus(ctx, request)
	if err != nil {
		logger.WithError(err).Error()

		return nil, err
	}

	logger.WithField("response", fmt.Sprintf("%+v", response)).Info()

	return response, nil
}
