package middlewares

import (
	"log/slog"
	"math"
	"strconv"
	"time"

	"gateway/db"

	"github.com/gofiber/fiber/v3"
)

type RateLimit struct {
	Name    string
	Limit   int
	Window  time.Duration
	Message string
	By      func(fiber.Ctx) string
}

func (r RateLimit) Handler(store *db.Store) fiber.Handler {
	limit := strconv.Itoa(r.Limit)

	return func(c fiber.Ctx) error {
		subject := c.IP()
		if r.By != nil {
			subject = r.By(c)
		}
		res, err := store.Allow(c.Context(), r.Name+":"+subject, r.Limit, r.Window)
		if err != nil {
			slog.LogAttrs(c.Context(), slog.LevelError, "rate limiter unavailable", slog.Any("error", err))
			return fiber.NewError(fiber.StatusServiceUnavailable, "Service unavailable")
		}

		c.Set("X-RateLimit-Limit", limit)
		c.Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))

		if !res.Allowed {
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(int(math.Ceil(res.RetryAfter.Seconds()))))
			return fiber.NewError(fiber.StatusTooManyRequests, r.Message)
		}
		return c.Next()
	}
}
