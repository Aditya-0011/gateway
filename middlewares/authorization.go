package middlewares

import (
	"gateway/schema"

	"github.com/gofiber/fiber/v3"
)

func authInfo(c fiber.Ctx) (*schema.AuthInfo, bool) {
	info, ok := c.Locals("auth").(*schema.AuthInfo)
	return info, ok
}

func RequireSession() fiber.Handler {
	return func(c fiber.Ctx) error {
		if info, ok := authInfo(c); !ok || !info.IsSession() {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		if info, ok := authInfo(c); !ok || !info.IsAdmin() {
			return fiber.NewError(fiber.StatusForbidden, "Administrator access required")
		}
		return c.Next()
	}
}

func RequireProject(slug string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if info, ok := authInfo(c); !ok || !info.CanAccess(slug) {
			return fiber.NewError(fiber.StatusForbidden, "No access to this project")
		}
		return c.Next()
	}
}
