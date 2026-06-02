package middlewares

import "github.com/gofiber/fiber/v3"

func RequireSession() fiber.Handler {
	return func(c fiber.Ctx) error {
		authType := c.Locals("authType")

		if authType != "session" {
			return fiber.NewError(fiber.StatusForbidden, "API Keys are restricted to read-only endpoints")
		}

		return c.Next()
	}
}
