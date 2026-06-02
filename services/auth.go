package services

import (
	"fmt"
	"gateway/db"
	calls "gateway/grpc/calls/auth"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func authService(router fiber.Router, redis *db.RedisParams, client auth.AuthServiceClient) {
	validator, err := protovalidate.New()

	if err != nil {
		panic(fmt.Errorf("failed to initialize validator: %w", err))
	}

	call := calls.AuthCalls(redis, client, validator)

	router.Post("/login", middlewares.Validate[auth.LoginRequest](validator), call.Login)

	router.Get("/key", middlewares.Authenticate(redis, client), middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), call.GetKey)
	router.Post("/rotate-key", middlewares.Authenticate(redis, client), middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), call.RotateKey)
	router.Post("/logout", middlewares.Authenticate(redis, client), middlewares.RequireSession(), call.Logout)

}
