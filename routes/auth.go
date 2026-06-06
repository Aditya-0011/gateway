package routes

import (
	"fmt"
	"gateway/db"
	calls "gateway/grpc/calls/auth"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func authRouter(router fiber.Router, redis *db.RedisParams, client auth.AuthServiceClient) {
	validator, err := protovalidate.New()

	if err != nil {
		panic(fmt.Errorf("failed to initialize validator: %w", err))
	}

	authCall := calls.AuthCalls(redis, client, validator)

	router.Post("/login", middlewares.Validate[auth.LoginRequest](validator), authCall.Login)

	router.Get("/key", middlewares.Authenticate(redis, client), middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), authCall.GetKey)
	router.Post("/rotate-key", middlewares.Authenticate(redis, client), middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), authCall.RotateKey)
	router.Post("/logout", middlewares.Authenticate(redis, client), middlewares.RequireSession(), authCall.Logout)

}
