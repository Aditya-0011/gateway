package middlewares

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		code := c.Response().StatusCode()

		var level slog.Level
		switch {
		case code >= 500:
			level = slog.LevelError
		case code >= 400:
			level = slog.LevelWarn
		default:
			level = slog.LevelInfo
		}

		attrs := []slog.Attr{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", code),
			slog.Duration("duration", duration),
			slog.String("ip", c.IP()),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		slog.LogAttrs(c.Context(), level, "HTTP Request", attrs...)

		return err
	}
}
