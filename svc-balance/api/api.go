package api

import (
	"svc-balance/middleware"
	"svc-balance/service"

	"github.com/gofiber/fiber/v2"
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

func (api *Api) SetupRoutes(app *fiber.App) *fiber.App {
	// Error handler middleware
	app.Use(middleware.ErrorHandler())

	// Health Routes
	health := app.Group("/health")
	health.Get("/", api.Health)

	return app
}
