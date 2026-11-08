package db

import (
	"context"
	"errors"
	"time"

	"gateway/schema"

	"github.com/bytedance/sonic"
	driver "github.com/redis/go-redis/v9"
)

const (
	apiKeyCacheTTL     = 5 * time.Minute
	apiKeyPrefix       = "apikey:"
	keyOriginsKey      = "cors:key-origins"
	keyOriginsTTL      = 5 * time.Minute
	keyOriginsSentinel = "-"
)

var ErrOriginsNotLoaded = errors.New("key origins not loaded")

var originAllowedScript = driver.NewScript(`
	if redis.call("EXISTS", KEYS[1]) == 0 then
		return -1
	end
	return redis.call("SISMEMBER", KEYS[1], ARGV[1])
`)

func apiKeyPointer(userID int) string { return userTag(userID) + "apikey" }

func (s *Store) GetAPIKey(ctx context.Context, keyHash string) (schema.APIKeyIdentity, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	val, err := s.client.Get(ctx, apiKeyPrefix+keyHash).Result()
	if errors.Is(err, driver.Nil) {
		return schema.APIKeyIdentity{}, ErrInvalidSession
	}
	if err != nil {
		return schema.APIKeyIdentity{}, err
	}

	var identity schema.APIKeyIdentity
	if err := sonic.UnmarshalString(val, &identity); err != nil {
		return schema.APIKeyIdentity{}, err
	}
	return identity, nil
}

func (s *Store) CacheAPIKey(ctx context.Context, keyHash string, identity schema.APIKeyIdentity) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	data, err := sonic.Marshal(identity)
	if err != nil {
		return err
	}
	pipe := s.client.Pipeline()
	pipe.Set(ctx, apiKeyPrefix+keyHash, data, apiKeyCacheTTL)
	pipe.Set(ctx, apiKeyPointer(identity.UserID), keyHash, apiKeyCacheTTL)
	pipe.SAdd(ctx, authzUsersKey, identity.UserID)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) EvictAPIKey(ctx context.Context, keyHash string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	return s.client.Del(ctx, apiKeyPrefix+keyHash, keyOriginsKey).Err()
}

func (s *Store) KeyOriginAllowed(ctx context.Context, origin string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	res, err := originAllowedScript.Run(ctx, s.client, []string{keyOriginsKey}, origin).Int()
	if err != nil {
		return false, err
	}
	if res < 0 {
		return false, ErrOriginsNotLoaded
	}
	return res == 1, nil
}

func (s *Store) CacheKeyOrigins(ctx context.Context, origins []string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	_, err := s.client.TxPipelined(ctx, func(pipe driver.Pipeliner) error {
		pipe.Del(ctx, keyOriginsKey)
		pipe.SAdd(ctx, keyOriginsKey, keyOriginsSentinel, origins)
		pipe.Expire(ctx, keyOriginsKey, keyOriginsTTL)
		return nil
	})
	return err
}
