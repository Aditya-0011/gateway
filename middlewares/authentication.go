package middlewares

import (
	"context"
	"fmt"
	"gateway/db"
	"gateway/utils"
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
		slog.Error("DOMAIN environment variable is not set")
		os.Exit(1)
	}

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
				if err.Error() != "invalid session" {
					slog.Error("Redis Error", "error", err)
					c.ClearCookie("session")
					return fiber.NewError(fiber.ErrInternalServerError.Code, "Logged out due to security reasons")
				}
				c.ClearCookie("session")
				return fiber.NewError(fiber.ErrUnauthorized.Code, "Invalid Session")
			}

			go func() {
				mappingKey := fmt.Sprintf("active:%s", userData.Email)
				if err := redis.ExtendSession(context.WithoutCancel(c.Context()), sessionKey, mappingKey); err != nil {
					slog.Error("Redis Error extending session", "error", err)
				}
			}()

			c.Cookie(&fiber.Cookie{
				Name:     "session",
				Value:    sessionKey,
				Domain:   domain,
				HTTPOnly: true,
				Secure:   !isDev,
				SameSite: "Strict",
				MaxAge:   int(redis.SessionDuration.Seconds()),
			})

			c.Locals("userId", userData.Id)
			c.Locals("userEmail", userData.Email)
			c.Locals("authType", "session")

		} else {
			key := fmt.Sprintf("api:%s", utils.HashSHA256(apiKey))

			userData, err := redis.GetSession(c.Context(), key)

			if err != nil {
				if err.Error() != "invalid session" {
					slog.Error("Redis Error retrieving api key", "error", err)
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
					slog.Error("Redis Error creating api session", "error", err)
					return fiber.NewError(fiber.ErrInternalServerError.Code, "Internal server error")
				}

				c.Locals("userId", int(res.GetUserId()))
				c.Locals("userEmail", res.GetUserEmail())
				c.Locals("authType", "api_key")
			} else {
				c.Locals("userId", userData.Id)
				c.Locals("userEmail", userData.Email)
				c.Locals("authType", "api_key")

				go func() {
					if err := redis.ExtendApiSession(context.WithoutCancel(c.Context()), key); err != nil {
						slog.Error("Redis Error extending api session", "error", err)
					}
				}()
			}
		}

		return c.Next()
	}
}
