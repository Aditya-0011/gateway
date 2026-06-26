package main

import (
	"context"
	"fmt"
	"gateway/db"
	"gateway/grpc/clients"
	"gateway/middlewares"
	"gateway/routes"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
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

	slog.LogAttrs(context.Background(), slog.LevelInfo, "runtime config",
		slog.Int("GOMAXPROCS", runtime.GOMAXPROCS(0)),
		slog.Int("NumCPU", runtime.NumCPU()),
		slog.Int64("GOMEMLIMIT_MiB", debug.SetMemoryLimit(-1)/1024/1024),
	)

	setupCtx, setupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer setupCancel()

	database, err := db.Setup(setupCtx)
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "Failed to setup databases", slog.String("error", err.Error()))
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
		configs.ProxyHeader = "X-Real-IP"
		configs.TrustProxy = true
		configs.TrustProxyConfig = fiber.TrustProxyConfig{
			Loopback: true,
			Private:  true,
		}
	}

	app := fiber.New(configs)

	isDev := os.Getenv("DEVELOPMENT") == "true"

	allowedOrigins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if allowedOrigins == "" && isDev {
		allowedOrigins = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:5174,http://127.0.0.1:5174"
	} else if allowedOrigins == "" {
		slog.LogAttrs(context.Background(), slog.LevelError, "CORS_ALLOWED_ORIGINS environment variable is not set")
		os.Exit(1)
	}

	app.Use(middlewares.Logger())

	origins := strings.Split(allowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:               15,
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

	routes.Setup(app, database.Redis, serviceClients)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		slog.LogAttrs(context.Background(), slog.LevelInfo, "Gateway server listening", slog.String("port", port))
		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-quit:
		slog.LogAttrs(context.Background(), slog.LevelInfo, "Interrupt received. Starting graceful shutdown...")
	case err := <-errChan:
		slog.LogAttrs(context.Background(), slog.LevelError, "Gateway server failed", slog.String("error", err.Error()))
		slog.LogAttrs(context.Background(), slog.LevelInfo, "Starting graceful shutdown due to server error...")
	}

	stopped := make(chan struct{})
	go func() {
		app.Shutdown()
		close(stopped)
	}()

	shutdownTimer := time.NewTimer(5 * time.Second)
	defer shutdownTimer.Stop()

	select {
	case <-shutdownTimer.C:
		slog.LogAttrs(context.Background(), slog.LevelInfo, "Timeout reached (5s). Forcing server shutdown...")
		app.Shutdown()
	case <-stopped:
		slog.LogAttrs(context.Background(), slog.LevelInfo, "Server gracefully stopped within timeout.")
	}

	slog.LogAttrs(context.Background(), slog.LevelInfo, "Shutdown complete. Exiting main...")

}
