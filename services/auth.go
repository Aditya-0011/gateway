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

	api := router.Group("/")
	{
		api.Post("/login", middlewares.Validate[auth.LoginRequest](validator), call.Login)

		authenticated := api.Group("/", middlewares.RequireSession(), middlewares.Authenticate(redis, client))
		{
			authenticated.Get("/key", call.GetKey)
			authenticated.Post("/rotate-key", call.RotateKey)
			authenticated.Post("/logout", call.Logout)
		}
	}

}
