package api

import (
	"svc-transaction/service"

	"github.com/sirupsen/logrus"
)

type Api struct {
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
