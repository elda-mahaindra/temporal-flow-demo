package flowngine_adapter

import (
	"api-gateway/adapter/flowngine_adapter/pb"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Adapter is a wrapper around the grpc client
type Adapter struct {
	serviceName string

	logger *logrus.Logger

	serviceBClient pb.FlowEngineClient
}

// NewAdapter creates a new grpc adapter
func NewAdapter(
	serviceName string,
	logger *logrus.Logger,
	cc *grpc.ClientConn,
) *Adapter {
	serviceBClient := pb.NewFlowEngineClient(cc)

	return &Adapter{
		serviceName: serviceName,

		logger: logger,

		serviceBClient: serviceBClient,
	}
}
