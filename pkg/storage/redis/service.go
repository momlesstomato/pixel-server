package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	client := goredis.NewClient(opts)
	if debugLoggingEnabled() {
		client.AddHook(redisDebugHook{})
	}
	return &Service{client: client}, nil
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

// debugLoggingEnabled reports whether runtime debug logging is enabled.
func debugLoggingEnabled() bool {
	return strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug")
}

// redisDebugHook writes redis command diagnostics when debug logging is active.
type redisDebugHook struct{}

// DialHook returns the next dial hook unchanged.
func (h redisDebugHook) DialHook(next goredis.DialHook) goredis.DialHook {
	_ = h
	return next
}

// ProcessHook logs one redis command after execution.
func (h redisDebugHook) ProcessHook(next goredis.ProcessHook) goredis.ProcessHook {
	_ = h
	return func(ctx context.Context, cmd goredis.Cmder) error {
		err := next(ctx, cmd)
		log.Printf("level=debug component=redis cmd=%s args=%v err=%v", cmd.Name(), cmd.Args(), err)
		return err
	}
}

// ProcessPipelineHook logs pipeline command names after execution.
func (h redisDebugHook) ProcessPipelineHook(next goredis.ProcessPipelineHook) goredis.ProcessPipelineHook {
	_ = h
	return func(ctx context.Context, cmds []goredis.Cmder) error {
		err := next(ctx, cmds)
		names := make([]string, 0, len(cmds))
		for _, cmd := range cmds {
			names = append(names, cmd.Name())
		}
		log.Printf("level=debug component=redis pipeline=%v err=%v", names, err)
		return err
	}
}
