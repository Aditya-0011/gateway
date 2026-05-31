package services

import (
	"fmt"
	"gateway/db"
	"gateway/grpc/calls"
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
		api.Post("/login", call.Login)

		authenticated := api.Group("/", middlewares.Authenticate(redis, client), middlewares.RequireSession())
		{
			authenticated.Get("/key", middlewares.Validate[auth.KeyRequest](validator), call.GetKey)
			authenticated.Post("/rotate-key", middlewares.Validate[auth.KeyRequest](validator), call.RotateKey)
			authenticated.Post("/logout", call.Logout)
		}
	}

}
