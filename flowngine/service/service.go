package service

import (
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
)

type Service struct {
	logger *logrus.Logger

	temporalClient client.Client
}

func NewService(
	logger *logrus.Logger,
	temporalClient client.Client,
) *Service {
	return &Service{
		logger: logger,

		temporalClient: temporalClient,
	}
}
