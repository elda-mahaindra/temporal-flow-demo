package activity

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// TemporalWorkerConfig holds configuration for the Temporal worker
type TemporalWorkerConfig struct {
	HostPort  string `json:"host_port" mapstructure:"host_port"`
	Namespace string `json:"namespace" mapstructure:"namespace"`
	TaskQueue string `json:"task_queue" mapstructure:"task_queue"`
}

// TemporalWorker wraps the Temporal worker and client
type TemporalWorker struct {
	client client.Client
	worker worker.Worker
	config TemporalWorkerConfig
	api    *Activity
}

// NewTemporalWorker creates a new Temporal worker instance
func NewTemporalWorker(config TemporalWorkerConfig, api *Activity) (*TemporalWorker, error) {
	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	// Create worker with default options
	temporalWorker := worker.New(temporalClient, config.TaskQueue, worker.Options{})

	return &TemporalWorker{
		client: temporalClient,
		worker: temporalWorker,
		config: config,
		api:    api,
	}, nil
}

// registerActivities registers all activities with the worker
func (tw *TemporalWorker) registerActivities() {
	activities := tw.api.GetActivities()
	for _, activity := range activities {
		tw.worker.RegisterActivity(activity)
	}

	tw.api.logger.WithFields(logrus.Fields{
		"task_queue":         tw.config.TaskQueue,
		"activity_count":     len(activities),
		"temporal_host":      tw.config.HostPort,
		"temporal_namespace": tw.config.Namespace,
	}).Info("Temporal activities registered successfully")
}

// Run starts the Temporal worker following the established API pattern
func (tw *TemporalWorker) Run(ctx context.Context) error {
	const op = "api.TemporalWorker.Run"

	logger := tw.api.logger.WithFields(logrus.Fields{
		"[op]":               op,
		"task_queue":         tw.config.TaskQueue,
		"temporal_host":      tw.config.HostPort,
		"temporal_namespace": tw.config.Namespace,
	})

	logger.Info("Starting Temporal worker")

	// Register activities before starting
	tw.registerActivities()

	// Start the worker
	err := tw.worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start Temporal worker: %w", err)
	}

	logger.Info("Temporal worker started successfully")

	// Wait for context cancellation (following the established pattern)
	<-ctx.Done()

	logger.Info("Shutting down Temporal worker")
	tw.worker.Stop()
	tw.client.Close()

	return nil
}

// Stop gracefully stops the Temporal worker
func (tw *TemporalWorker) Stop() {
	tw.api.logger.Info("Stopping Temporal worker")
	tw.worker.Stop()
	tw.client.Close()
}

// HealthCheck checks if the worker is healthy
func (tw *TemporalWorker) HealthCheck(ctx context.Context) error {
	// Create a simple context with timeout for health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get workflow service to verify connectivity
	_, err := tw.client.WorkflowService().GetSystemInfo(healthCtx, &workflowservice.GetSystemInfoRequest{})
	if err != nil {
		err := fmt.Errorf("temporal health check failed: %w", err)

		tw.api.logger.Error(err)

		return err
	}

	return nil
}
