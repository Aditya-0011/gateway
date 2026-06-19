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
		slog.LogAttrs(context.Background(), slog.LevelError, "REDIS_URL environment variable is not set")
		os.Exit(1)
	}

	redis, err := ConnectRedis(ctx, redisURI)
	if err != nil {
		return nil, err
	}

	slog.LogAttrs(context.Background(), slog.LevelInfo, "Redis initialized")

	return &DatabaseParams{
		Redis: redis,
	}, nil
}

func (s *DatabaseParams) Cleanup() error {
	slog.LogAttrs(context.Background(), slog.LevelInfo, "Closing database connections")

	if err := s.Redis.Close(); err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "Error closing Redis", slog.String("error", err.Error()))
	}

	return nil
}
