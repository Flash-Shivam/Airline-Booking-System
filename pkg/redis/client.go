package redis

import (
	"context"
	"fmt"
	"time"

	"airline-booking-system/internal/config"

	"github.com/go-redis/redis/v8"
)

// Client represents Redis client wrapper
type Client struct {
	*redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Client{rdb}
}

// SetJSON sets a JSON value in Redis with TTL
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, ttl).Err()
}

// Get gets a value from Redis
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Exists checks if a key exists in Redis
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.Client.Exists(ctx, key).Result()
	return count > 0, err
}

// Delete deletes a key from Redis
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// AcquireLock acquires a distributed lock
func (c *Client) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, "locked", ttl).Result()
}

// ReleaseLock releases a distributed lock
func (c *Client) ReleaseLock(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// IncrBy increments a key by the specified amount
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.Client.IncrBy(ctx, key, value).Result()
}

// GetInt gets an integer value from Redis
func (c *Client) GetInt(ctx context.Context, key string) (int64, error) {
	return c.Client.Get(ctx, key).Int64()
}

// Ping checks Redis connectivity
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
