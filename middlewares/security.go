package middlewares

import "github.com/gofiber/fiber/v3"

func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
		c.Set(fiber.HeaderXFrameOptions, "DENY")
		c.Set(fiber.HeaderReferrerPolicy, "no-referrer")
		c.Set(fiber.HeaderContentSecurityPolicy, "default-src 'none'; frame-ancestors 'none'")
		c.Set(fiber.HeaderCacheControl, "no-store")
		return c.Next()
	}
}
