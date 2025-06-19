package worker

import (
	"svc-balance/activity"

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

// NewWorker creates a new Temporal worker instance
func NewWorker(
	logger *logrus.Logger,
	client client.Client,
	taskQueue string,
	activity *activity.Activity,
) (*Worker, error) {
	// Create worker with default options
	temporalWorker := worker.New(client, taskQueue, worker.Options{})

	return &Worker{
		logger: logger,

		client: client,
		worker: temporalWorker,

		taskQueue: taskQueue,
		activity:  activity,
	}, nil
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
