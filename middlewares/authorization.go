package middlewares

import (
	"gateway/schema"

	"github.com/gofiber/fiber/v3"
)

func RequireSession() fiber.Handler {
	return func(c fiber.Ctx) error {
		authInfo, ok := c.Locals("auth").(*schema.AuthInfo)

		if !ok || !authInfo.IsSession {
			return fiber.ErrForbidden
		}

		return c.Next()
	}
}
