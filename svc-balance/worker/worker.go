package worker

import (
	"svc-balance/activity"
	"svc-balance/util/config"
	"time"

	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Worker wraps the Temporal worker and client
type Worker struct {
	logger *logrus.Logger

	client client.Client
	worker worker.Worker

	taskQueue string
	activity  *activity.Activity
}

// NewWorker creates a new Temporal worker instance with performance optimizations
func NewWorker(
	logger *logrus.Logger,
	client client.Client,
	taskQueue string,
	activity *activity.Activity,
	temporalConfig config.Temporal,
) (*Worker, error) {
	// Create worker with performance-optimized options
	workerOptions := worker.Options{
		// Balance service optimization: handle many quick balance checks
		MaxConcurrentActivityExecutionSize:      getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentActivityExecutions, 100),
		MaxConcurrentWorkflowTaskExecutionSize:  getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentWorkflowExecutions, 50),
		MaxConcurrentLocalActivityExecutionSize: getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentLocalActivities, 200),
		MaxConcurrentActivityTaskPollers:        getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentActivityTaskPollers, 5),
		MaxConcurrentWorkflowTaskPollers:        getOptimizedValue(temporalConfig.WorkerOptions.MaxConcurrentWorkflowTaskPollers, 5),

		// Performance optimization: Enable session worker for resource management
		EnableSessionWorker: temporalConfig.WorkerOptions.EnableSessionWorker,

		// Task timeout optimizations
		MaxHeartbeatThrottleInterval:     60 * time.Second,
		DefaultHeartbeatThrottleInterval: 30 * time.Second,
	}

	temporalWorker := worker.New(client, taskQueue, workerOptions)

	logger.WithFields(logrus.Fields{
		"max_concurrent_activities": workerOptions.MaxConcurrentActivityExecutionSize,
		"max_concurrent_workflows":  workerOptions.MaxConcurrentWorkflowTaskExecutionSize,
		"max_local_activities":      workerOptions.MaxConcurrentLocalActivityExecutionSize,
		"activity_task_pollers":     workerOptions.MaxConcurrentActivityTaskPollers,
		"workflow_task_pollers":     workerOptions.MaxConcurrentWorkflowTaskPollers,
		"session_worker_enabled":    workerOptions.EnableSessionWorker,
	}).Info("🚀 Balance service worker created with performance optimizations")

	return &Worker{
		logger: logger,

		client: client,
		worker: temporalWorker,

		taskQueue: taskQueue,
		activity:  activity,
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
func (worker *Worker) registerActivities() {
	const op = "worker.Worker.registerActivities"

	logger := worker.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	activities := worker.activity.GetActivities()
	for _, activity := range activities {
		worker.worker.RegisterActivity(activity)
	}

	logger.WithFields(logrus.Fields{
		"task_queue":     worker.taskQueue,
		"activity_count": len(activities),
		"message":        "Temporal activities registered successfully",
	}).Info()
}
