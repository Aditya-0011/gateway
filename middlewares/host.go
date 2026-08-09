package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func RequireHost(domain string, isDevEnv bool, notFoundHandler fiber.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		host := c.Hostname()

		if (isDevEnv && strings.Contains(host, "localhost")) || strings.Contains(host, domain) {
			return c.Next()
		}

		if notFoundHandler != nil {
			return notFoundHandler(c)
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	}
}
