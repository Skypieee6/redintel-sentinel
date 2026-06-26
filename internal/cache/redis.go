// Package cache manages the Redis client.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Skypieee6/redintel-sentinel/internal/config"
)

// Redis wraps a go-redis client.
type Redis struct {
	Client *redis.Client
}

// New creates a Redis client and verifies connectivity.
func New(ctx context.Context, cfg config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{Client: client}, nil
}

// Ping verifies Redis is reachable.
func (r *Redis) Ping(ctx context.Context) error {
	if r == nil || r.Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return r.Client.Ping(ctx).Err()
}

// Close closes the underlying Redis connection.
func (r *Redis) Close() error {
	if r != nil && r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
