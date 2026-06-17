package routes

import (
	"fmt"
	"gateway/db"
	"gateway/grpc/clients"
	"gateway/middlewares"
	"os"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/gofiber/fiber/v3"
)

func Setup(app *fiber.App, redis *db.RedisParams, clients *clients.ClientParams) {
	isDevEnv := os.Getenv("DEVELOPMENT") == "true"

	validator, err := protovalidate.New()
	if err != nil {
		panic(fmt.Errorf("failed to initialize validator: %w", err))
	}

	app.Use(func(c fiber.Ctx) error {
		host := c.Hostname()
		isDev := isDevEnv && strings.Contains(host, "localhost")

		if strings.Contains(host, "api.auth.") || strings.Contains(host, "api.manager.") || isDev {
			return c.Next()
		}

		return notFoundHandler(c)
	})

	authMiddleware := middlewares.Authenticate(redis, clients.AuthClient)

	authGroup := app.Group("")
	authRouter(authGroup, redis, clients.AuthClient, authMiddleware, validator)

	managerGroup := app.Group("")
	managerRouter(managerGroup, authMiddleware, clients.ManagerUserClient, clients.ManagerPortfolioClient, validator)

	app.Use(notFoundHandler)
}
