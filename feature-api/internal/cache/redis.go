package cache

import (
	"context"

	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/featureflags/feature-api/internal/config"
)

// Connect creates a Redis client and verifies connectivity with PING.
// The caller owns the client lifecycle and must call Close.
func Connect(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPass,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}

// Close shuts down the Redis client.
func Close(client *redis.Client) {
	if client != nil {
		_ = client.Close()
	}
}
