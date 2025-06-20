package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// GetFailureSimulationStats returns statistics about the transaction failure simulation
func (api *Api) GetFailureSimulationStats(c *fiber.Ctx) error {
	const op = "api.Api.GetFailureSimulationStats"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info("Getting transaction failure simulation statistics")

	// Get stats from service
	stats := api.service.GetFailureSimulationStats()

	return c.JSON(fiber.Map{
		"status":             "success",
		"message":            "Transaction failure simulation statistics retrieved successfully",
		"failure_simulation": stats,
	})
}

// ResetFailureSimulation resets the transaction failure simulation state
func (api *Api) ResetFailureSimulation(c *fiber.Ctx) error {
	const op = "api.Api.ResetFailureSimulation"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info("Resetting transaction failure simulation state")

	// Reset simulation state
	api.service.ResetFailureSimulation()

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Transaction failure simulation state reset successfully",
	})
}

// GetLearningScenarios returns available transaction failure scenarios for learning
func (api *Api) GetLearningScenarios(c *fiber.Ctx) error {
	const op = "api.Api.GetLearningScenarios"

	logger := api.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info("Getting transaction learning scenarios")

	// Get scenarios from service
	scenarios := api.service.GetLearningScenarios()

	return c.JSON(fiber.Map{
		"status":      "success",
		"message":     "Transaction learning scenarios retrieved successfully",
		"scenarios":   scenarios,
		"description": "These are hardcoded failure scenarios designed to demonstrate Temporal's transaction and compensation capabilities",
	})
}
