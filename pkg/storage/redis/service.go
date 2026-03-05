package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// Service owns a Redis client.
type Service struct {
	client *goredis.Client
}

// New creates a Service from Config.
func New(cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &Service{client: goredis.NewClient(opts)}, nil
}

// Ping checks Redis connectivity.
func (s *Service) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Close closes the Redis client.
func (s *Service) Close() error {
	return s.client.Close()
}

// Client returns the underlying redis client.
func (s *Service) Client() *goredis.Client {
	return s.client
}
