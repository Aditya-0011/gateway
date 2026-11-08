package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	driver "github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

var ErrInvalidSession = errors.New("invalid session")

const opTimeout = 3 * time.Second

type Store struct {
	client *driver.Client
}

func Connect(ctx context.Context, uri string) (*Store, error) {
	opt, err := driver.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("parsing REDIS_URL: %w", err)
	}

	opt.PoolSize = 10
	opt.Protocol = 3
	opt.MinIdleConns = 2
	opt.MaxIdleConns = 5
	opt.ConnMaxIdleTime = 60 * time.Second
	opt.ConnMaxLifetime = 10 * time.Minute
	opt.PoolTimeout = 4 * time.Second
	opt.DialTimeout = 2 * time.Second
	opt.ReadTimeout = opTimeout
	opt.WriteTimeout = opTimeout
	opt.MaintNotificationsConfig = &maintnotifications.Config{Mode: maintnotifications.ModeDisabled}

	client := driver.NewClient(opt)

	pingCtx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return &Store{client: client}, nil
}

func (s *Store) Close() error {
	return s.client.Close()
}
