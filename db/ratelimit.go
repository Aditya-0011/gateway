package db

import (
	"context"
	"strconv"
	"time"

	driver "github.com/redis/go-redis/v9"
)

const rateLimitPrefix = "ratelimit:"

var slidingWindowScript = driver.NewScript(`
	local current = tonumber(redis.call("GET", KEYS[1]) or "0")
	local previous = tonumber(redis.call("GET", KEYS[2]) or "0")
	local limit = tonumber(ARGV[1])
	local window_ms = tonumber(ARGV[2])
	local elapsed_ms = tonumber(ARGV[3])

	local estimated = previous * (window_ms - elapsed_ms) / window_ms + current
	if estimated >= limit then
		return {0, 0}
	end

	current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("PEXPIRE", KEYS[1], window_ms * 2)
	end

	return {1, math.max(limit - math.ceil(estimated) - 1, 0)}
`)

type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

func (s *Store) Allow(ctx context.Context, bucket string, limit int, window time.Duration) (RateLimitResult, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	now := time.Now()
	windowMs := window.Milliseconds()
	index := now.UnixMilli() / windowMs
	elapsed := now.UnixMilli() - index*windowMs

	tag := rateLimitPrefix + "{" + bucket + "}:"
	keys := []string{tag + strconv.FormatInt(index, 10), tag + strconv.FormatInt(index-1, 10)}

	res, err := slidingWindowScript.Run(ctx, s.client, keys, limit, windowMs, elapsed).Int64Slice()
	if err != nil {
		return RateLimitResult{}, err
	}

	result := RateLimitResult{Allowed: res[0] == 1, Remaining: int(res[1])}
	if !result.Allowed {
		result.RetryAfter = time.Duration(windowMs-elapsed) * time.Millisecond
	}
	return result, nil
}
