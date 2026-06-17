package calls

import (
	"context"
	"gateway/db"
	rpc "gateway/grpc/calls"
	"gateway/schema"
	"log/slog"
	"os"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

type AuthCallsParams struct {
	redis     *db.RedisParams
	client    auth.AuthServiceClient
	validator protovalidate.Validator
	isDev     bool
	domain    string
}

func AuthCalls(redis *db.RedisParams, client auth.AuthServiceClient, validator protovalidate.Validator) *AuthCallsParams {
	isDev := os.Getenv("DEVELOPMENT") == "true"
	domain := os.Getenv("DOMAIN")

	if domain == "" && isDev {
		domain = ""
	} else if domain == "" {
		slog.Error("DOMAIN environment variable is not set")
		os.Exit(1)
	}

	return &AuthCallsParams{
		redis:     redis,
		client:    client,
		validator: validator,
		isDev:     isDev,
		domain:    domain,
	}
}

func (ac *AuthCallsParams) Login(c fiber.Ctx) error {

	req := c.Locals("req").(*auth.LoginRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.LoginResponse, error) {
		return ac.client.Login(ctx, req)
	})
	if err != nil {
		return err
	}

	sessionKey, err := ac.redis.CreateSession(c.Context(), int(res.GetUserId()), res.GetUserEmail())
	if err != nil {
		slog.Error("Redis error during CreateSession", "error", err)
		return fiber.ErrInternalServerError
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    sessionKey,
		Domain:   ac.domain,
		HTTPOnly: true,
		Secure:   !ac.isDev,
		SameSite: "Strict",
		MaxAge:   int(ac.redis.SessionDuration.Seconds()),
	})

	return c.JSON(&auth.SimpleResponse{
		Message: "Login successful",
	})
}

func (ac *AuthCallsParams) GetKey(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.KeyRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.GetKeyResponse, error) {
		return ac.client.GetKey(ctx, req)
	})
}

func (ac *AuthCallsParams) RotateKey(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.KeyRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return ac.client.RotateKey(ctx, req)
	})
}

func (ac *AuthCallsParams) Logout(c fiber.Ctx) error {
	cookie := c.Cookies("session")

	authInfo, ok := c.Locals("auth").(*schema.AuthInfo)

	if !ok || authInfo.UserEmail == "" {
		c.ClearCookie("session")
		return c.JSON(&auth.SimpleResponse{
			Message: "No active session",
		})
	}

	email := authInfo.UserEmail

	go func() {
		mappingKey := "active:" + email
		ac.redis.DeleteSession(context.WithoutCancel(c.Context()), cookie, mappingKey)
	}()

	c.ClearCookie("session")

	return c.JSON(&auth.SimpleResponse{
		Message: "Successfully logged out",
	})
}
