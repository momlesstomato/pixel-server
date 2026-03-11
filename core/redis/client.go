package redis

import (
	"fmt"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

// NewClient creates a Redis client from application configuration.
func NewClient(redisConfig Config) (*redislib.Client, error) {
	address := strings.TrimSpace(redisConfig.Address)
	if address == "" {
		return nil, fmt.Errorf("redis address is required")
	}
	options := &redislib.Options{
		Addr:         address,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		DialTimeout:  3 * time.Second,
	}
	return redislib.NewClient(options), nil
}
