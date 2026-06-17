package db

import (
	"context"
	"errors"
	"gateway/schema"
	"gateway/utils"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"

	fiberRedisStorage "github.com/gofiber/storage/redis/v3"
	driver "github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

var ErrInvalidSession = errors.New("invalid session")

const (
	sessionDuration = 15 * time.Minute
	timeoutDuration = 3 * time.Second
)

var (
	createSessionScript = driver.NewScript(`
		local mapping_key = KEYS[1]
		local new_session_key = KEYS[2]
		local new_session_data = ARGV[1]
		local duration = ARGV[2]

		local old_session_key = redis.call("GET", mapping_key)
		if old_session_key then
			redis.call("DEL", old_session_key)
		end

		redis.call("SET", mapping_key, new_session_key, "EX", duration)
		redis.call("SET", new_session_key, new_session_data, "EX", duration)

		return "OK"
	`)

	deleteSessionScript = driver.NewScript(`
		local mapping_key = KEYS[1]
		local session_id = ARGV[1]
		
		redis.call("DEL", session_id)
		
		if redis.call("GET", mapping_key) == session_id then
			redis.call("DEL", mapping_key)
		end
		
		return 1
	`)
)

type RedisParams struct {
	client          *driver.Client
	SessionDuration time.Duration
	Store           *fiberRedisStorage.Storage
}

func ConnectRedis(c context.Context, uri string) (*RedisParams, error) {
	opt, err := driver.ParseURL(uri)

	if err != nil {
		return nil, err
	}

	opt.PoolSize = 10
	opt.MinIdleConns = 2
	opt.MaxIdleConns = 5
	opt.ConnMaxIdleTime = 60 * time.Second
	opt.ConnMaxLifetime = 10 * time.Minute
	opt.PoolTimeout = 4 * time.Second
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	opt.MaintNotificationsConfig = &maintnotifications.Config{
		Mode: maintnotifications.ModeDisabled,
	}

	client := driver.NewClient(opt)

	ctx, cancel := context.WithTimeout(c, timeoutDuration)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	slog.Info("Connected to Redis!")

	store := fiberRedisStorage.NewFromConnection(client)

	return &RedisParams{
		client:          client,
		SessionDuration: sessionDuration,
		Store:           store,
	}, nil
}

func (r *RedisParams) CreateSession(c context.Context, userId int, userEmail string) (string, error) {
	ctx, cancel := context.WithTimeout(c, timeoutDuration)
	defer cancel()

	mappingKey := "active:" + userEmail

	sessionKey, err := utils.GenerateSessionKey()
	if err != nil {
		return "", err
	}

	tokenData := schema.Token{
		Id:    userId,
		Email: userEmail,
	}
	jsonData, err := sonic.Marshal(tokenData)
	if err != nil {
		return "", err
	}

	err = createSessionScript.Run(ctx, r.client,
		[]string{mappingKey, sessionKey},
		jsonData, int(sessionDuration.Seconds()),
	).Err()

	return sessionKey, err
}

func (r *RedisParams) CreateApiSession(c context.Context, key string, userId int, userEmail string) error {
	ctx, cancel := context.WithTimeout(c, timeoutDuration)
	defer cancel()

	tokenData := schema.Token{
		Id:    userId,
		Email: userEmail,
	}
	jsonData, err := sonic.Marshal(tokenData)
	if err != nil {
		return err
	}

	err = r.client.Set(ctx, key, jsonData, sessionDuration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisParams) GetSession(c context.Context, key string) (schema.Token, error) {
	ctx, cancel := context.WithTimeout(c, timeoutDuration)
	defer cancel()

	val, err := r.client.Get(ctx, key).Result()

	if err == driver.Nil {
		return schema.Token{}, ErrInvalidSession
	}
	if err != nil {
		return schema.Token{}, err
	}

	var tokenData schema.Token
	if err := sonic.UnmarshalString(val, &tokenData); err != nil {
		return schema.Token{}, err
	}
	return tokenData, nil
}

func (r *RedisParams) ExtendSession(c context.Context, sessionId, mappingKey string) error {
	ctx, cancel := context.WithTimeout(c, timeoutDuration)
	defer cancel()

	pipe := r.client.Pipeline()
	pipe.Expire(ctx, sessionId, sessionDuration)
	pipe.Expire(ctx, mappingKey, sessionDuration)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisParams) ExtendApiSession(c context.Context, key string) error {
	ctx, cancel := context.WithTimeout(c, timeoutDuration)
	defer cancel()

	return r.client.Expire(ctx, key, sessionDuration).Err()
}

func (r *RedisParams) DeleteSession(c context.Context, sessionId, mappingKey string) error {
	ctx, cancel := context.WithTimeout(c, timeoutDuration)
	defer cancel()

	return deleteSessionScript.Run(ctx, r.client, []string{mappingKey}, sessionId).Err()
}

func (r *RedisParams) Close() error {
	slog.Info("Disconnecting from Redis...")
	if r.Store != nil {
		slog.Info("Closing Redis store...")
		return r.Store.Close()
	}
	return r.client.Close()
}
