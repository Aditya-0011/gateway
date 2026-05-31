package db

import (
	"context"
	"log/slog"
	"os"
)

type DatabaseParams struct {
	Redis *RedisParams
}

func Setup(ctx context.Context) (*DatabaseParams, error) {
	redisURI := os.Getenv("REDIS_URL")
	if redisURI == "" {
		slog.Error("REDIS_URL environment variable is not set")
		os.Exit(1)
	}

	redis, err := ConnectRedis(ctx, redisURI)
	if err != nil {
		return nil, err
	}

	slog.Info("Redis initialized")

	return &DatabaseParams{
		Redis: redis,
	}, nil
}

func (s *DatabaseParams) Cleanup() error {
	slog.Info("Closing database connections")

	if err := s.Redis.Close(); err != nil {
		slog.Error("Error closing Redis", "error", err)
	}

	return nil
}
