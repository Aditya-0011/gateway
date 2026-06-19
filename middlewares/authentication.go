package middlewares

import (
	"context"
	"errors"
	"gateway/db"
	"gateway/internal/crypto"
	"gateway/schema"
	"log/slog"
	"os"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func Authenticate(redis *db.RedisParams, authClient auth.AuthServiceClient) fiber.Handler {
	isDev := os.Getenv("DEVELOPMENT") == "true"
	domain := os.Getenv("DOMAIN")

	if domain == "" && isDev {
		domain = ""
	} else if domain == "" {
		slog.LogAttrs(context.Background(), slog.LevelError, "DOMAIN environment variable is not set")
		os.Exit(1)
	}

	maxAge := int(redis.SessionDuration.Seconds())

	return func(c fiber.Ctx) error {
		sessionKey := c.Cookies("session")
		var apiKey string

		if sessionKey == "" {
			apiKey = c.Get("X-API-KEY")
			if apiKey == "" {
				return fiber.ErrUnauthorized
			}
		}

		if sessionKey != "" {
			userData, err := redis.GetSession(c.Context(), sessionKey)
			if err != nil {
				if !errors.Is(err, db.ErrInvalidSession) {
					slog.LogAttrs(c.Context(), slog.LevelError, "Redis Error", slog.String("error", err.Error()))
					c.ClearCookie("session")
					return fiber.NewError(fiber.ErrInternalServerError.Code, "Logged out due to security reasons")
				}
				c.ClearCookie("session")
				return fiber.NewError(fiber.ErrUnauthorized.Code, "Invalid Session")
			}

			go func() {
				mappingKey := "active:" + userData.Email
				if err := redis.ExtendSession(context.WithoutCancel(c.Context()), sessionKey, mappingKey); err != nil {
					slog.LogAttrs(context.Background(), slog.LevelError, "Redis Error extending session", slog.String("error", err.Error()))
				}
			}()

			c.Cookie(&fiber.Cookie{
				Name:     "session",
				Value:    sessionKey,
				Domain:   domain,
				HTTPOnly: true,
				Secure:   !isDev,
				SameSite: "Strict",
				MaxAge:   maxAge,
			})

			c.Locals("auth", &schema.AuthInfo{
				UserId:    userData.Id,
				UserEmail: userData.Email,
				IsSession: true,
			})

		} else {
			key := "api:" + crypto.HashSHA256(apiKey)

			userData, err := redis.GetSession(c.Context(), key)

			if err != nil {
				if !errors.Is(err, db.ErrInvalidSession) {
					slog.LogAttrs(c.Context(), slog.LevelError, "Redis Error retrieving api key", slog.String("error", err.Error()))
					return fiber.NewError(fiber.ErrInternalServerError.Code, "Internal server error")
				}

				res, err := authClient.ValidateKey(c.Context(), &auth.ValidateKeyRequest{
					Key: apiKey,
				})
				if err != nil {
					return fiber.NewError(fiber.ErrUnauthorized.Code, "Invalid API Key")
				}

				err = redis.CreateApiSession(c.Context(), key, int(res.GetUserId()), res.GetUserEmail())
				if err != nil {
					slog.LogAttrs(c.Context(), slog.LevelError, "Redis Error creating api session", slog.String("error", err.Error()))
					return fiber.NewError(fiber.ErrInternalServerError.Code, "Internal server error")
				}

				c.Locals("auth", &schema.AuthInfo{
					UserId:    int(res.GetUserId()),
					UserEmail: res.GetUserEmail(),
					IsSession: false,
				})
			} else {
				c.Locals("auth", &schema.AuthInfo{
					UserId:    userData.Id,
					UserEmail: userData.Email,
					IsSession: false,
				})

				go func() {
					if err := redis.ExtendApiSession(context.WithoutCancel(c.Context()), key); err != nil {
						slog.LogAttrs(context.Background(), slog.LevelError, "Redis Error extending api session", slog.String("error", err.Error()))
					}
				}()
			}
		}

		return c.Next()
	}
}
