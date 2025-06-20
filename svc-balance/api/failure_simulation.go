package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// GetFailureSimulationStats returns statistics about the failure simulation
func (api *Api) GetFailureSimulationStats(c *fiber.Ctx) error {
	const op = "api.Api.GetFailureSimulationStats"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info("Getting failure simulation statistics")

	// Get stats from service
	stats := api.service.GetFailureSimulationStats()

	return c.JSON(fiber.Map{
		"status":             "success",
		"message":            "Failure simulation statistics retrieved successfully",
		"failure_simulation": stats,
	})
}

// ResetFailureSimulation resets the failure simulation state
func (api *Api) ResetFailureSimulation(c *fiber.Ctx) error {
	const op = "api.Api.ResetFailureSimulation"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info("Resetting failure simulation state")

	// Reset simulation state
	api.service.ResetFailureSimulation()

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Failure simulation state reset successfully",
	})
}

// GetLearningScenarios returns available failure scenarios for learning
func (api *Api) GetLearningScenarios(c *fiber.Ctx) error {
	const op = "api.Api.GetLearningScenarios"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info("Getting learning scenarios")

	// Get scenarios from service
	scenarios := api.service.GetLearningScenarios()

	return c.JSON(fiber.Map{
		"status":      "success",
		"message":     "Learning scenarios retrieved successfully",
		"scenarios":   scenarios,
		"description": "These are hardcoded failure scenarios designed to demonstrate Temporal's fault tolerance capabilities",
	})
}
