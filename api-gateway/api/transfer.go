package api

import (
	"fmt"

	"api-gateway/service"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func (api *Api) Transfer(c *fiber.Ctx) error {
	const op = "api.Api.Transfer"

	var req struct {
		FromAccount       string  `json:"from_account" validate:"required,min=12,max=12"`
		ToAccount         string  `json:"to_account" validate:"required,min=12,max=12"`
		Amount            int     `json:"amount" validate:"required,min=1,max=1000000000"`
		Currency          string  `json:"currency" validate:"required,min=3,max=3"`
		Description       *string `json:"description" validate:"max=100"`
		ReferenceID       *string `json:"reference_id" validate:"max=50"`
		WaitForCompletion *bool   `json:"wait_for_completion"`
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// Default to async mode if not specified
	waitForCompletion := false
	if req.WaitForCompletion != nil {
		waitForCompletion = *req.WaitForCompletion
	}

	// Create params
	params := &service.TransferParams{
		FromAccount:       req.FromAccount,
		ToAccount:         req.ToAccount,
		Amount:            req.Amount,
		Currency:          req.Currency,
		Description:       req.Description,
		ReferenceID:       req.ReferenceID,
		WaitForCompletion: waitForCompletion,
	}

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Call service
	results, err := api.service.Transfer(c.Context(), params)
	if err != nil {
		logger.WithError(err).Error()

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(results)
}

func (api *Api) GetTransfer(c *fiber.Ctx) error {
	const op = "api.Api.GetTransfer"

	id := c.Params("id")

	params := &service.GetTransferParams{
		TransactionID: id,
	}

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	// Call service
	results, err := api.service.GetTransfer(c.Context(), params)
	if err != nil {
		logger.WithError(err).Error()

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(results)
}
