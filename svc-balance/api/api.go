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

	// Failure Simulation Routes (for testing and monitoring)
	failureSimulation := app.Group("/failure-simulation")
	failureSimulation.Get("/stats", api.GetFailureSimulationStats)
	failureSimulation.Post("/reset", api.ResetFailureSimulation)
	failureSimulation.Get("/scenarios", api.GetLearningScenarios)

	return app
}
