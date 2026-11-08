package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"gateway/db"
	"gateway/grpc/clients"
	"gateway/internal/config"
	"gateway/middlewares"
	"gateway/routes"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

const (
	startupTimeout  = 10 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("gateway failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	slog.Info("runtime config",
		slog.Int("GOMAXPROCS", runtime.GOMAXPROCS(0)),
		slog.Int64("GOMEMLIMIT_MiB", debug.SetMemoryLimit(-1)/1024/1024),
	)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	startCtx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	store, err := db.Connect(startCtx, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer store.Close()
	slog.Info("connected to redis")

	serviceClients, err := clients.New(cfg)
	if err != nil {
		return err
	}
	defer serviceClients.Close()

	app := fiber.New(fiberConfig(cfg))

	app.Use(middlewares.Logger())
	app.Use(middlewares.SecurityHeaders())
	app.Use(routes.CORS(cfg, store, serviceClients.Auth))
	app.Use(routes.GlobalLimiter(store))

	if err := routes.Setup(app, cfg, store, serviceClients); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- app.Listen(net.JoinHostPort("", cfg.Port), fiber.ListenConfig{DisableStartupMessage: true})
	}()
	slog.Info("gateway listening", slog.String("port", cfg.Port))

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-listenErr:
		return err
	}

	if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		slog.Warn("shutdown did not complete cleanly", slog.Any("error", err))
	}
	return nil
}

func fiberConfig(cfg config.Config) fiber.Config {
	fc := fiber.Config{
		AppName:           "Gateway",
		JSONEncoder:       sonic.Marshal,
		JSONDecoder:       sonic.Unmarshal,
		ReduceMemoryUsage: true,
		BodyLimit:         1 << 20,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if !cfg.Dev {
		fc.ProxyHeader = fiber.HeaderXForwardedFor
		fc.TrustProxy = true
		fc.EnableIPValidation = true
		fc.TrustProxyConfig = fiber.TrustProxyConfig{Loopback: true, Private: true}
	}
	return fc
}
