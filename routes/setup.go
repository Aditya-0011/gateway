package routes

import (
	"fmt"
	"time"

	"gateway/db"
	"gateway/grpc/clients"
	"gateway/internal/config"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/gofiber/fiber/v3"
)

func Setup(app *fiber.App, cfg config.Config, store *db.Store, cl *clients.Clients) error {
	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("creating validator: %w", err)
	}

	cookie := middlewares.NewSessionCookie(cfg.APIDomain, !cfg.Dev)
	authn := middlewares.NewAuthenticator(store, cl.Auth, cookie, cfg.AllowedOrigins)

	host := func(service string) fiber.Router {
		if cfg.Dev {
			return app
		}
		return app.Domain(cfg.Host(service))
	}

	authRoutes(host("auth"), store, cl, cookie, authn, credentialLimiter(store), validator)
	managerRoutes(host("manager"), store, cl, authn.Required(), middlewares.RequireProject("manager"), validator)
	walletRoutes(host("wallet"), cl, authn.Required(), middlewares.RequireProject("wallet"), validator)

	app.Use(notFoundHandler)
	return nil
}

func GlobalLimiter(store *db.Store) fiber.Handler {
	return middlewares.RateLimit{
		Name:    "global",
		Limit:   60,
		Window:  time.Minute,
		Message: "Too many requests",
	}.Handler(store)
}

func credentialLimiter(store *db.Store) fiber.Handler {
	return middlewares.RateLimit{
		Name:    "credentials",
		Limit:   5,
		Window:  time.Minute,
		Message: "Too many attempts, try again in a minute",
	}.Handler(store)
}
