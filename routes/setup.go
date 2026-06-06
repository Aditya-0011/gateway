package routes

import (
	"gateway/db"
	"gateway/grpc/clients"
	"gateway/middlewares"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func Setup(app *fiber.App, redis *db.RedisParams, clients *clients.ClientParams) {
	isDevEnv := os.Getenv("DEVELOPMENT") == "true"

	app.Use(func(c fiber.Ctx) error {
		host := c.Hostname()
		isDev := isDevEnv && strings.Contains(host, "localhost")

		if strings.Contains(host, "api.auth.") || strings.Contains(host, "api.manager.") || isDev {
			return c.Next()
		}

		return notFoundHandler(c)
	})

	authGroup := app.Group("")
	authRouter(authGroup, redis, clients.AuthClient)

	authMiddleware := middlewares.Authenticate(redis, clients.AuthClient)
	managerGroup := app.Group("")
	managerRouter(managerGroup, authMiddleware, clients.ManagerUserClient, clients.ManagerPortfolioClient)

	app.Use(notFoundHandler)
}
