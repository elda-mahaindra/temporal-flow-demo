package worker

import (
	"context"
	"fmt"
	"time"

	"svc-transaction/activity"
	"svc-transaction/util/config"

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

// NewWorker creates a new Temporal worker instance with performance optimizations
func NewWorker(
	logger *logrus.Logger,
	temporalClient client.Client,
	taskQueue string,
	activity *activity.Activity,
	temporalConfig config.Temporal,
) (*Worker, error) {
	// Create worker with performance-optimized options
	workerOptions := worker.Options{
		// Transaction service optimization: fewer but heavier database operations
		MaxConcurrentActivityExecutionSize:      getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentActivityExecutions, 80),
		MaxConcurrentWorkflowTaskExecutionSize:  getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentWorkflowExecutions, 40),
		MaxConcurrentLocalActivityExecutionSize: getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentLocalActivities, 160),
		MaxConcurrentActivityTaskPollers:        getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentActivityTaskPollers, 5),
		MaxConcurrentWorkflowTaskPollers:        getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentWorkflowTaskPollers, 5),

		// Performance optimization: Enable session worker for resource management
		EnableSessionWorker: temporalConfig.WorkerOptions.EnableSessionWorker,

		// Task timeout optimizations
		MaxHeartbeatThrottleInterval:     60 * time.Second,
		DefaultHeartbeatThrottleInterval: 30 * time.Second,
	}

	temporalWorker := worker.New(temporalClient, taskQueue, workerOptions)

	logger.WithFields(logrus.Fields{
		"max_concurrent_activities": workerOptions.MaxConcurrentActivityExecutionSize,
		"max_concurrent_workflows":  workerOptions.MaxConcurrentWorkflowTaskExecutionSize,
		"max_local_activities":      workerOptions.MaxConcurrentLocalActivityExecutionSize,
		"activity_task_pollers":     workerOptions.MaxConcurrentActivityTaskPollers,
		"workflow_task_pollers":     workerOptions.MaxConcurrentWorkflowTaskPollers,
		"session_worker_enabled":    workerOptions.EnableSessionWorker,
	}).Info("🚀 Transaction service worker created with performance optimizations")

	return &Worker{
		client:    temporalClient,
		worker:    temporalWorker,
		taskQueue: taskQueue,
		activity:  activity,
		logger:    logger,
	}, nil
}

// getOptimizedValue returns the configured value or a default if not set
func getOptimizedValue(configValue, defaultValue int) int {
	if configValue > 0 {
		return configValue
	}
	return defaultValue
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
