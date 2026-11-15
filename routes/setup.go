package routes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/db"
	"gateway/grpc/clients"
	"gateway/internal/config"
	"gateway/internal/crypto"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/gofiber/fiber/v3"
)

const healthTimeout = time.Second

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
		Skip: func(c fiber.Ctx) bool {
			return c.Get(middlewares.APIKeyHeader) != ""
		},
	}.Handler(store)
}

func Health(store *db.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		if string(c.Request().URI().PathOriginal()) != "/healthz" {
			return c.Next()
		}
		ctx, cancel := context.WithTimeout(c.Context(), healthTimeout)
		defer cancel()
		if err := store.Ping(ctx); err != nil {
			return c.SendStatus(fiber.StatusServiceUnavailable)
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func accountLimiter[M interface{ GetEmail() string }](store *db.Store) fiber.Handler {
	return middlewares.RateLimit{
		Name:    "account",
		Limit:   10,
		Window:  15 * time.Minute,
		Message: "Too many attempts for this account, try again later",
		By: func(c fiber.Ctx) string {
			return crypto.HashSHA256(strings.ToLower(c.Locals("req").(M).GetEmail()))
		},
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
