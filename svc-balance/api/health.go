package api

import (
	"github.com/gofiber/fiber/v2"
)

func (api *Api) Health(c *fiber.Ctx) error {
	return c.SendString("ok")
}
