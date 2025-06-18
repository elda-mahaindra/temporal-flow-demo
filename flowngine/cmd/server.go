package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"flowngine/api"
	"flowngine/api/pb"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func runGrpcServer(port int, server *api.Api) *grpc.Server {
	// Create new gRPC server
	opts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithPropagators(propagation.TraceContext{}),
		)),
	}
	grpcServer := grpc.NewServer(opts...)

	// Register gRPC services
	pb.RegisterFlowEngineServer(grpcServer, server)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	// Listen at specified port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("failed to listen at port: %v!", port)

		os.Exit(1)
	}

	log.Printf("listening at port: %d", port)

	// Serve the gRPC server
	go func() {
		log.Printf("gRPC server started successfully 🚀")

		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("failed to serve: %v", err)
		}
	}()

	return grpcServer
}
