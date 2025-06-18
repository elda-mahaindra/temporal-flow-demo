package main

import (
	"fmt"

	"api-gateway/adapter/flowngine_adapter"
	"api-gateway/util/config"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func createFlowngineAdapter(config config.Flowngine, logger *logrus.Logger) (*flowngine_adapter.Adapter, error) {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithPropagators(propagation.TraceContext{}),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s grpc server: %w", config.Name, err)
	}

	grpcAdapter := flowngine_adapter.NewAdapter(config.Name, logger, conn)

	return grpcAdapter, nil
}
