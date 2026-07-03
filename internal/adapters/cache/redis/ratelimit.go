package redis

import (
	"context"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

// RateLimiter implements ports.DistributedRateLimiter using Redis sliding window.
type RateLimiter struct {
	client *Client
	prefix string
}

func NewRateLimiter(client *Client) *RateLimiter {
	return &RateLimiter{client: client, prefix: "rl:"}
}

// Allow checks if the request is within the rate limit using a sliding window counter.
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	redisKey := r.prefix + key
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	pipe := r.client.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZCard(ctx, redisKey)
	pipe.ZAdd(ctx, redisKey, redislib.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, redisKey, window)

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := cmds[1].(*redislib.IntCmd).Val()
	if count >= int64(limit) {
		// Remove the entry we just added since we're rejecting
		pipe2 := r.client.rdb.Pipeline()
		pipe2.ZRem(ctx, redisKey, now)
		_, _ = pipe2.Exec(ctx)
		return false, nil
	}

	return true, nil
}
