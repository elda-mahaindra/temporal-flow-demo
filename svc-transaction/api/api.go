package api

import (
	"svc-transaction/middleware"
	"svc-transaction/service"

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

	// Enhanced Compensation Audit Routes
	compensationAudit := app.Group("/compensation-audit")
	compensationAudit.Get("/stats", api.GetCompensationStats)
	compensationAudit.Get("/workflow/:workflow_id", api.GetCompensationAuditByWorkflow)
	compensationAudit.Get("/pending", api.GetPendingCompensations)

	return app
}
