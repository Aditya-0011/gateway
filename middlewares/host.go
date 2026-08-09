package middlewares

import (
	"log/slog"
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
			slog.LogAttrs(c.Context(), slog.LevelWarn, "RequireHost failed, calling notFoundHandler", slog.String("domain", domain), slog.String("host", host))
			return notFoundHandler(c)
		}

		slog.LogAttrs(c.Context(), slog.LevelWarn, "RequireHost failed, returning 404 manually", slog.String("domain", domain), slog.String("host", host))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	}
}
