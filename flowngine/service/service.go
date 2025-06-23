package service

import (
	"flowngine/util/config"
	"time"

	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Service struct {
	logger *logrus.Logger
	config config.Config

	temporalClient client.Client
}

func NewService(
	logger *logrus.Logger,
	config config.Config,
	temporalClient client.Client,
) *Service {
	service := &Service{
		logger: logger,
		config: config,

		temporalClient: temporalClient,
	}

	// Initialize the global ActivityOptionsProvider for workflows
	ActivityOptionsProvider = service.GetActivityOptions

	return service
}

// GetActivityOptions returns banking-optimized activity options from configuration
func (s *Service) GetActivityOptions() workflow.ActivityOptions {
	activityConfig := s.config.Temporal.ActivityOptions

	return workflow.ActivityOptions{
		StartToCloseTimeout:    time.Duration(activityConfig.StartToCloseTimeoutSeconds) * time.Second,
		HeartbeatTimeout:       time.Duration(activityConfig.HeartbeatTimeoutSeconds) * time.Second,
		ScheduleToCloseTimeout: time.Duration(activityConfig.ScheduleToCloseTimeoutSeconds) * time.Second,
		ScheduleToStartTimeout: time.Duration(activityConfig.ScheduleToStartTimeoutSeconds) * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Duration(activityConfig.RetryPolicy.InitialIntervalMs) * time.Millisecond,
			BackoffCoefficient:     activityConfig.RetryPolicy.BackoffCoefficient,
			MaximumInterval:        time.Duration(activityConfig.RetryPolicy.MaximumIntervalSeconds) * time.Second,
			MaximumAttempts:        int32(activityConfig.RetryPolicy.MaximumAttempts),
			NonRetryableErrorTypes: activityConfig.RetryPolicy.NonRetryableErrorTypes,
		},
	}
}

// SetTemporalClient updates the Temporal client (used for delayed connection)
func (s *Service) SetTemporalClient(temporalClient client.Client) {
	s.temporalClient = temporalClient
}
