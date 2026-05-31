package calls

import (
	"context"
	"fmt"
	"gateway/db"
	"gateway/utils"
	"log/slog"
	"os"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthCallsParams struct {
	redis     *db.RedisParams
	client    auth.AuthServiceClient
	validator protovalidate.Validator
}

func AuthCalls(redis *db.RedisParams, client auth.AuthServiceClient, validator protovalidate.Validator) *AuthCallsParams {
	return &AuthCallsParams{
		redis:     redis,
		client:    client,
		validator: validator,
	}
}

func (ac *AuthCallsParams) Login(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.LoginRequest)

	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	res, err := ac.client.Login(ctx, req)

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.Unauthenticated {
				return fiber.NewError(fiber.StatusUnauthorized, st.Message())
			}
			return fiber.NewError(fiber.StatusInternalServerError, st.Message())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	sessionKey, err := ac.redis.CreateSession(c.Context(), int(res.GetUserId()), res.GetUserEmail())
	if err != nil {
		slog.Error("Redis error during CreateSession", "error", err)
		return fiber.ErrInternalServerError
	}

	development := os.Getenv("DEVELOPMENT")
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    sessionKey,
		HTTPOnly: true,
		Secure:   development != "true",
		SameSite: "Strict",
		MaxAge:   int(ac.redis.SessionDuration.Seconds()),
	})

	return c.JSON(res)
}

func (ac *AuthCallsParams) GetKey(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.KeyRequest)

	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	res, err := ac.client.GetKey(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.NotFound {
				return fiber.NewError(fiber.StatusNotFound, st.Message())
			}
			return fiber.NewError(fiber.StatusInternalServerError, st.Message())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(res)
}

func (ac *AuthCallsParams) RotateKey(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.KeyRequest)

	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	res, err := ac.client.RotateKey(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.NotFound {
				return fiber.NewError(fiber.StatusNotFound, st.Message())
			}
			return fiber.NewError(fiber.StatusInternalServerError, st.Message())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(res)
}

func (ac *AuthCallsParams) Logout(c fiber.Ctx) error {
	cookie := c.Cookies("session")
	email, ok := c.Locals("userEmail").(string)

	if !ok || email == "" {
		c.ClearCookie("session")
		return c.JSON(&manager.SimpleResponse{
			Message: "No active session",
		})
	}

	go func() {
		mappingKey := fmt.Sprintf("active:%s", email)
		ac.redis.DeleteSession(context.WithoutCancel(c.Context()), cookie, mappingKey)
	}()

	c.ClearCookie("session")

	return c.JSON(&manager.SimpleResponse{
		Message: "Successfully logged out",
	})
}
