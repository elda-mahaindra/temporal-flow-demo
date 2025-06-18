package service

import (
	"api-gateway/adapter/flowngine_adapter"

	"github.com/sirupsen/logrus"
)

type Service struct {
	logger *logrus.Logger

	flowngineAdapter *flowngine_adapter.Adapter
}

func NewService(
	logger *logrus.Logger,
	flowngineAdapter *flowngine_adapter.Adapter,
) *Service {
	return &Service{
		logger: logger,

		flowngineAdapter: flowngineAdapter,
	}
}
