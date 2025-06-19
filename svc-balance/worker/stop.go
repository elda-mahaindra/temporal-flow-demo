package worker

import (
	"github.com/sirupsen/logrus"
)

// Stop gracefully stops the Temporal worker
func (worker *Worker) Stop() {
	const op = "worker.Worker.Stop"

	logger := worker.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.WithField("message", "Stopping Temporal worker").Info()

	worker.worker.Stop()
	worker.client.Close()
}
