package worker

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Run starts the Temporal worker
func (worker *Worker) Run(ctx context.Context) error {
	const op = "worker.Worker.Run"

	logger := worker.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	// Register activities before starting
	worker.registerActivities()

	// Start the worker
	err := worker.worker.Start()
	if err != nil {
		err = fmt.Errorf("failed to start Temporal worker: %w", err)

		logger.WithError(err).Error()

		return err
	}

	// Wait for context cancellation (following the established pattern)
	<-ctx.Done()

	logger.Info("Shutting down Temporal worker")
	worker.client.Close()

	return nil
}
