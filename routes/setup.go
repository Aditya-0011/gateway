package routes

import (
	"fmt"
	"gateway/db"
	"gateway/grpc/clients"
	"gateway/middlewares"
	"log/slog"
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

		if strings.Contains(host, "api.auth.") || strings.Contains(host, "api.manager.") || strings.Contains(host, "api.wallet.") || isDev {
			return c.Next()
		}

		slog.LogAttrs(c.Context(), slog.LevelWarn, "Top level middleware failed", slog.String("host", host))
		return notFoundHandler(c)
	})

	authMiddleware := middlewares.Authenticate(redis, clients.AuthClient)

	authRouter(app, middlewares.RequireHost("api.auth.", isDevEnv, notFoundHandler), redis, clients.AuthClient, authMiddleware, validator)

	managerRouter(app, middlewares.RequireHost("api.manager.", isDevEnv, notFoundHandler), authMiddleware, clients.ManagerUserClient, clients.ManagerPortfolioClient, validator)

	walletRouter(app, middlewares.RequireHost("api.wallet.", isDevEnv, notFoundHandler), authMiddleware, middlewares.RequireSession(), clients.WalletUserClient, clients.WalletFinanceClient, validator)

	app.Use(notFoundHandler)
}
