package routes

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"gateway/db"
	"gateway/internal/config"
	"gateway/middlewares"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

const originLookupTimeout = 3 * time.Second

// CORS serves two lanes. First-party UIs (CORS_ALLOWED_ORIGINS) get credentialed
// CORS everywhere. Any other origin is only considered on Manager's portfolio
// routes, without credentials, and only if some API key lists it; the key's own
// origin list is enforced again by the authenticator on the actual request.
func CORS(cfg config.Config, store *db.Store, authClient auth.AuthServiceClient) fiber.Handler {
	firstParty := cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodQuery},
		AllowHeaders:     []string{fiber.HeaderContentType},
		ExposeHeaders:    []string{fiber.HeaderContentDisposition},
		AllowCredentials: true,
		MaxAge:           3600,
	})

	portfolio := cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return keyOriginAllowed(store, authClient, origin)
		},
		AllowMethods: []string{fiber.MethodGet, fiber.MethodPost},
		AllowHeaders: []string{fiber.HeaderContentType, middlewares.APIKeyHeader},
		MaxAge:       600,
	})

	managerHost := cfg.Host("manager")
	return func(c fiber.Ctx) error {
		origin := c.Get(fiber.HeaderOrigin)
		if origin != "" && !slices.Contains(cfg.AllowedOrigins, origin) &&
			(cfg.Dev || c.Hostname() == managerHost) && slices.Contains(portfolioPaths, c.Path()) {
			return portfolio(c)
		}
		return firstParty(c)
	}
}

func keyOriginAllowed(store *db.Store, authClient auth.AuthServiceClient, origin string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), originLookupTimeout)
	defer cancel()

	allowed, err := store.KeyOriginAllowed(ctx, origin)
	if err == nil {
		return allowed
	}
	if !errors.Is(err, db.ErrOriginsNotLoaded) {
		slog.LogAttrs(ctx, slog.LevelError, "reading key origins", slog.Any("error", err))
		return false
	}

	res, err := authClient.ListKeyOrigins(ctx, &auth.ListKeyOriginsRequest{})
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "loading key origins", slog.Any("error", err))
		return false
	}
	if err := store.CacheKeyOrigins(ctx, res.GetOrigins()); err != nil {
		slog.LogAttrs(ctx, slog.LevelWarn, "caching key origins", slog.Any("error", err))
	}
	return slices.Contains(res.GetOrigins(), origin)
}
