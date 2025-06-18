package api

import (
	"flowngine/api/pb"
	"flowngine/service"

	"github.com/sirupsen/logrus"
)

type Api struct {
	pb.UnimplementedFlowEngineServer

	logger *logrus.Logger

	service *service.Service
}

func NewApi(
	logger *logrus.Logger,
	service *service.Service,
) *Api {
	return &Api{
		logger: logger,

		service: service,
	}
}
