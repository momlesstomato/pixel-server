package realtime

import (
	"context"
	"encoding/json"
	"fmt"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	redislib "github.com/redis/go-redis/v9"
)

// CloseSignal defines one remote close instruction payload.
type CloseSignal struct {
	// Code stores websocket close status code.
	Code int `json:"code"`
	// Reason stores websocket close reason phrase.
	Reason string `json:"reason"`
}

// CloseSignalBus defines close-signal publish/subscribe behavior.
type CloseSignalBus interface {
	// Publish emits one close signal for one connection identifier.
	Publish(context.Context, string, CloseSignal) error
	// Subscribe listens for close signals targeting one connection identifier.
	Subscribe(context.Context, string) (<-chan CloseSignal, coreconnection.Disposable, error)
}

// RedisCloseSignalBus defines Redis pub/sub close-signal behavior.
type RedisCloseSignalBus struct {
	// client stores Redis connectivity for pub/sub operations.
	client *redislib.Client
	// prefix stores pub/sub channel namespace prefix.
	prefix string
}

// NewRedisCloseSignalBus creates Redis close-signal behavior.
func NewRedisCloseSignalBus(client *redislib.Client, prefix string) (*RedisCloseSignalBus, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	value := prefix
	if value == "" {
		value = "handshake:close"
	}
	return &RedisCloseSignalBus{client: client, prefix: value}, nil
}

// Publish emits one close signal for one connection identifier.
func (bus *RedisCloseSignalBus) Publish(ctx context.Context, connID string, signal CloseSignal) error {
	payload, err := json.Marshal(signal)
	if err != nil {
		return err
	}
	return bus.client.Publish(ctx, bus.channel(connID), payload).Err()
}

// Subscribe listens for close signals targeting one connection identifier.
func (bus *RedisCloseSignalBus) Subscribe(ctx context.Context, connID string) (<-chan CloseSignal, coreconnection.Disposable, error) {
	pubSub := bus.client.Subscribe(ctx, bus.channel(connID))
	if _, err := pubSub.Receive(ctx); err != nil {
		_ = pubSub.Close()
		return nil, nil, err
	}
	stream := make(chan CloseSignal, 1)
	go func() {
		defer close(stream)
		for {
			select {
			case <-ctx.Done():
				return
			case message, open := <-pubSub.Channel():
				if !open {
					return
				}
				var signal CloseSignal
				if err := json.Unmarshal([]byte(message.Payload), &signal); err == nil {
					stream <- signal
				}
			}
		}
	}()
	return stream, coreconnection.DisposeFunc(pubSub.Close), nil
}

// channel returns Redis pub/sub channel name for one connection.
func (bus *RedisCloseSignalBus) channel(connID string) string {
	return bus.prefix + ":" + connID
}
