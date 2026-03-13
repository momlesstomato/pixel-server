package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/momlesstomato/pixel-server/core/broadcast"
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

// DistributedCloseSignalBus defines broadcaster-backed close-signal behavior.
type DistributedCloseSignalBus struct {
	// broadcaster stores cross-instance publish and subscribe behavior.
	broadcaster broadcast.Broadcaster
	// prefix stores pub/sub channel namespace prefix.
	prefix string
}

// NewCloseSignalBus creates broadcaster-backed close-signal behavior.
func NewCloseSignalBus(broadcaster broadcast.Broadcaster, prefix string) (*DistributedCloseSignalBus, error) {
	if broadcaster == nil {
		return nil, fmt.Errorf("broadcaster is required")
	}
	value := strings.TrimSpace(prefix)
	if value == "" {
		value = "handshake:close"
	}
	return &DistributedCloseSignalBus{broadcaster: broadcaster, prefix: value}, nil
}

// NewRedisCloseSignalBus creates Redis-backed close-signal behavior.
func NewRedisCloseSignalBus(client *redislib.Client, prefix string) (*DistributedCloseSignalBus, error) {
	broadcaster, err := broadcast.NewRedisBroadcaster(client, "")
	if err != nil {
		return nil, err
	}
	return NewCloseSignalBus(broadcaster, prefix)
}

// Publish emits one close signal for one connection identifier.
func (bus *DistributedCloseSignalBus) Publish(ctx context.Context, connID string, signal CloseSignal) error {
	payload, err := json.Marshal(signal)
	if err != nil {
		return err
	}
	return bus.broadcaster.Publish(ctx, bus.channel(connID), payload)
}

// Subscribe listens for close signals targeting one connection identifier.
func (bus *DistributedCloseSignalBus) Subscribe(ctx context.Context, connID string) (<-chan CloseSignal, coreconnection.Disposable, error) {
	messages, disposable, err := bus.broadcaster.Subscribe(ctx, bus.channel(connID))
	if err != nil {
		return nil, nil, err
	}
	stream := make(chan CloseSignal, 1)
	go func() {
		defer close(stream)
		for {
			select {
			case <-ctx.Done():
				return
			case payload, open := <-messages:
				if !open {
					return
				}
				var signal CloseSignal
				if json.Unmarshal(payload, &signal) == nil {
					stream <- signal
				}
			}
		}
	}()
	return stream, disposable, nil
}

// channel returns pub/sub channel name for one connection.
func (bus *DistributedCloseSignalBus) channel(connID string) string {
	return bus.prefix + ":" + connID
}
