package routes

import (
	"gateway/db"
	calls "gateway/grpc/calls/auth"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func authRouter(router fiber.Router, redis *db.RedisParams, client auth.AuthServiceClient, authMiddleware fiber.Handler, validator protovalidate.Validator) {
	authCall := calls.AuthCalls(redis, client, validator)

	router.Post("/login", middlewares.Validate[auth.LoginRequest](validator), authCall.Login)

	router.Get("/key", authMiddleware, middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), authCall.GetKey)
	router.Post("/rotate-key", authMiddleware, middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), authCall.RotateKey)
	router.Post("/logout", authMiddleware, middlewares.RequireSession(), authCall.Logout)

}
