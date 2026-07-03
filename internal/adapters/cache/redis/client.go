package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
)

// Client wraps a Redis connection and provides all distributed adapters.
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client from config.
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     50,
		MinIdleConns: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close shuts down the Redis client.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping checks if Redis is reachable.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
