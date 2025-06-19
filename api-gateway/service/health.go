package service

import (
	"context"
	"fmt"
	"time"

	"api-gateway/adapter/flowngine_adapter/pb"

	"github.com/sirupsen/logrus"
)

type CheckHealthParams struct {
}

type CheckHealthResults struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

func (service *Service) CheckHealth(ctx context.Context, params *CheckHealthParams) (results *CheckHealthResults, err error) {
	const op = "service.Service.CheckHealth"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info("Performing health check")

	// Initialize results with basic information
	results = &CheckHealthResults{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  make(map[string]string),
	}

	// Check FlowEngine connectivity
	flowEngineStatus := "healthy"

	// Create a simple status request with a short timeout
	statusCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Try to call FlowEngine with a dummy transaction ID to test connectivity
	statusRequest := &pb.GetTransferStatusRequest{
		TransactionId: "health-check-dummy-id",
	}

	_, err = service.flowngineAdapter.GetTransferStatus(statusCtx, statusRequest)
	if err != nil {
		// Expected to fail for dummy ID, but if we get a response, gRPC connection is working
		logger.WithError(err).Debug("FlowEngine health check call (expected to fail for dummy ID)")

		// Check if it's a connection error vs application error
		if statusCtx.Err() == context.DeadlineExceeded {
			flowEngineStatus = "timeout"
			results.Status = "degraded"
		} else {
			// If we get any response (even error), the connection is working
			// This is expected since we're using a dummy transaction ID
			flowEngineStatus = "healthy"
		}
	}

	// Set service statuses
	results.Services["api-gateway"] = "healthy"
	results.Services["flowngine"] = flowEngineStatus

	// Overall status determination
	if flowEngineStatus != "healthy" {
		results.Status = "degraded"
	}

	logger.WithField("results", fmt.Sprintf("%+v", results)).Info("Health check completed")

	return results, nil
}
