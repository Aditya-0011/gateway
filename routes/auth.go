package routes

import (
	"gateway/db"
	calls "gateway/grpc/calls/auth"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func authRouter(router fiber.Router, requireHost fiber.Handler, redis *db.RedisParams, client auth.AuthServiceClient, authMiddleware fiber.Handler, validator protovalidate.Validator) {
	authCall := calls.AuthCalls(redis, client, validator)

	router.Post("/login", requireHost, middlewares.Validate[auth.LoginRequest](validator), authCall.Login)

	router.Get("/key", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), authCall.GetKey)
	router.Post("/rotate-key", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[auth.KeyRequest](validator), authCall.RotateKey)
	router.Post("/logout", requireHost, authMiddleware, middlewares.RequireSession(), authCall.Logout)
}
