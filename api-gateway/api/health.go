package api

import (
	"fmt"

	"api-gateway/service"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func (api *Api) CheckHealth(c *fiber.Ctx) error {
	const op = "api.Api.CheckHealth"

	// Create params
	params := &service.CheckHealthParams{}

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Call service
	results, err := api.service.CheckHealth(c.Context(), params)
	if err != nil {
		logger.WithError(err).Error()

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(results)
}
