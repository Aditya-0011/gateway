package services

import (
	"gateway/grpc/clients"
	"gateway/db"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func Setup(app *fiber.App, redis *db.RedisParams, clients clients.ClientParams) {
	authGroup := app.Group("", func(c fiber.Ctx) error {
		if strings.Contains(c.Hostname(), "auth.") || strings.Contains(c.Hostname(), "localhost") {
			return c.Next()
		}
		return c.Next()
	})
	authService(authGroup, redis, clients.AuthClient)
}
