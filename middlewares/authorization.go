package middlewares

import (
	"github.com/gofiber/fiber/v3"
)

func RequireSession() fiber.Handler {
	return func(c fiber.Ctx) error {
		authType := c.Locals("authType")

		if authType != "session" {
			return fiber.ErrForbidden
		}

		return c.Next()
	}
}
