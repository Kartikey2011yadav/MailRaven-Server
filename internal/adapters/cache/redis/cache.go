package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache implements ports.Cache using Redis.
type Cache struct {
	client *Client
}

func NewCache(client *Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.rdb.Del(ctx, key).Err()
}

func (c *Cache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.rdb.IncrBy(ctx, key, delta).Result()
}
