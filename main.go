package main

import (
	"context"
	"fmt"
	"gateway/db"
	"gateway/grpc/clients"
	"gateway/routes"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
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

	configs := fiber.Config{
		AppName:           "Gateway",
		JSONEncoder:       sonic.Marshal,
		JSONDecoder:       sonic.Unmarshal,
		ReduceMemoryUsage: true,
	}

	if os.Getenv("DEVELOPMENT") == "" {
		configs.ProxyHeader = fiber.HeaderXForwardedFor
		configs.TrustProxy = true
		configs.TrustProxyConfig = fiber.TrustProxyConfig{
			Proxies: []string{"127.0.0.1"},
		}
	}

	app := fiber.New(configs)

	isDev := os.Getenv("DEVELOPMENT") == "true"

	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" && isDev {
		allowedOrigins = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:5174,http://127.0.0.1:5174"
	} else if allowedOrigins == "" {
		slog.Error("CORS_ALLOWED_ORIGINS environment variable is not set")
		os.Exit(1)
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(allowedOrigins, ","),
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:               10,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.FixedWindow{},
		Storage:           database.Redis.Store,
		KeyGenerator: func(c fiber.Ctx) string {
			return "ratelimit" + c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return fiber.NewError(fiber.ErrTooManyRequests.Code, "Limit reached")
		},
	}))

	routes.Setup(app, database.Redis, serviceClients)

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
