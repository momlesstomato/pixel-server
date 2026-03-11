package redis

import (
	"fmt"

	redislib "github.com/redis/go-redis/v9"
)

// Stage defines Redis startup behavior.
type Stage interface {
	// Name returns a stable startup unit identifier.
	Name() string
	// InitializeRedis creates a Redis client from loaded configuration.
	InitializeRedis(Config) (*redislib.Client, error)
}

// Initializer provides default Redis startup behavior.
type Initializer struct{}

// Name returns the stable initializer name.
func (initializer Initializer) Name() string {
	return "redis"
}

// InitializeRedis builds and returns a configured Redis client.
func (initializer Initializer) InitializeRedis(loaded Config) (*redislib.Client, error) {
	if loaded.Address == "" {
		return nil, fmt.Errorf("redis address is required")
	}
	return NewClient(loaded)
}
