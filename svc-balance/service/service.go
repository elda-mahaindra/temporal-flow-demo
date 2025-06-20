package service

import (
	"svc-balance/store"
	"svc-balance/util/failure"

	"github.com/sirupsen/logrus"
)

type Service struct {
	logger *logrus.Logger

	store store.IStore

	failureSimulator *failure.Simulator
}

func NewService(
	logger *logrus.Logger,
	store store.IStore,
) *Service {
	return &Service{
		logger: logger,

		store: store,

		failureSimulator: failure.NewSimulator(logger),
	}
}
