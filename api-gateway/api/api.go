package api

import (
	"api-gateway/middleware"
	"api-gateway/service"

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

	// Transfer Routes
	transfer := app.Group("/transfer")
	transfer.Post("/", api.Transfer)
	transfer.Get("/:id", api.GetTransfer)

	// Health Check Routes
	health := app.Group("/health")
	health.Get("/", api.CheckHealth)

	// Metrics Routes
	metrics := app.Group("/metrics")
	metrics.Get("/", api.GetMetrics)

	return app
}
