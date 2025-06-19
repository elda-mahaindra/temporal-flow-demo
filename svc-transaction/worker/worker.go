package worker

import (
	"context"
	"fmt"
	"time"

	"svc-transaction/activity"

	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Worker wraps the Temporal worker and client
type Worker struct {
	client    client.Client
	worker    worker.Worker
	taskQueue string
	activity  *activity.Activity
	logger    *logrus.Logger
}

// NewWorker creates a new Temporal worker instance
func NewWorker(
	logger *logrus.Logger,
	temporalClient client.Client,
	taskQueue string,
	activity *activity.Activity,
) (*Worker, error) {
	// Create worker with default options
	temporalWorker := worker.New(temporalClient, taskQueue, worker.Options{})

	return &Worker{
		client:    temporalClient,
		worker:    temporalWorker,
		taskQueue: taskQueue,
		activity:  activity,
		logger:    logger,
	}, nil
}

// registerActivities registers all activities with the worker
func (w *Worker) registerActivities() {
	activities := w.activity.GetActivities()
	for _, activity := range activities {
		w.worker.RegisterActivity(activity)
	}

	w.logger.WithFields(logrus.Fields{
		"task_queue":     w.taskQueue,
		"activity_count": len(activities),
	}).Info("Temporal activities registered successfully")
}

// Run starts the Temporal worker following the established API pattern
func (w *Worker) Run(ctx context.Context) error {
	const op = "worker.Worker.Run"

	logger := w.logger.WithFields(logrus.Fields{
		"[op]":       op,
		"task_queue": w.taskQueue,
	})

	logger.Info("Starting Temporal worker")

	// Register activities before starting
	w.registerActivities()

	// Start the worker
	err := w.worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start Temporal worker: %w", err)
	}

	logger.Info("Temporal worker started successfully")

	// Wait for context cancellation (following the established pattern)
	<-ctx.Done()

	logger.Info("Shutting down Temporal worker")
	w.worker.Stop()

	return nil
}

// Stop gracefully stops the Temporal worker
func (w *Worker) Stop() {
	w.logger.Info("Stopping Temporal worker")
	w.worker.Stop()
}

// HealthCheck checks if the worker is healthy
func (w *Worker) HealthCheck(ctx context.Context) error {
	// Create a simple context with timeout for health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get workflow service to verify connectivity
	_, err := w.client.WorkflowService().GetSystemInfo(healthCtx, nil)
	if err != nil {
		return fmt.Errorf("Temporal health check failed: %w", err)
	}

	return nil
}
