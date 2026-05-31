package main

import (
	"context"
	"fmt"
	"gateway/grpc/clients"
	"gateway/db"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	setupCtx, setupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer setupCancel()

	database, err := db.Setup(setupCtx)
	if err != nil {
		slog.Error("Failed to setup databases", "error", err)
		os.Exit(1)
	}
	defer database.Cleanup()

	serviceClients := clients.Setup()
	defer serviceClients.Close()

	configs := fiber.Config{AppName: "Gateway"}

	if os.Getenv("DEVELOPMENT") == "" {
		configs.ProxyHeader = fiber.HeaderXForwardedFor
		configs.TrustProxy = true
		configs.TrustProxyConfig = fiber.TrustProxyConfig{
			Proxies: []string{"127.0.0.1"},
		}
	}

	app := fiber.New(configs)

	app.Use(helmet.New())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:4173,http://127.0.0.1:4173"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(allowedOrigins, ","),
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:               10,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
		Storage:           database.Redis.Store,
		KeyGenerator: func(c fiber.Ctx) string {
			return "ratelimit" + c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return fiber.NewError(fiber.ErrTooManyRequests.Code, "Limit reached")
		},
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		slog.Info("Gateway server listening", "address", port)
		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-quit:
		slog.Info("Interrupt received. Starting graceful shutdown...")
	case err := <-errChan:
		slog.Error("Gateway server failed", "error", err)
		slog.Info("Starting graceful shutdown due to server error...")
	}

	stopped := make(chan struct{})
	go func() {
		app.Shutdown()
		close(stopped)
	}()

	select {
	case <-time.After(5 * time.Second):
		slog.Info("Timeout reached (5s). Forcing server shutdown...")
		app.Shutdown()
	case <-stopped:
		slog.Info("Server gracefully stopped within timeout.")
	}

	slog.Info("Shutdown complete. Exiting main...")

}
