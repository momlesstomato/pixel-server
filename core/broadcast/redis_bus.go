package broadcast

import (
	"context"
	"fmt"
	"strings"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	redislib "github.com/redis/go-redis/v9"
)

// RedisBroadcaster defines Redis Pub/Sub broadcast behavior.
type RedisBroadcaster struct {
	// client stores Redis connectivity.
	client *redislib.Client
	// prefix stores optional Redis channel namespace prefix.
	prefix string
}

// NewRedisBroadcaster creates one Redis-backed broadcaster.
func NewRedisBroadcaster(client *redislib.Client, prefix string) (*RedisBroadcaster, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	return &RedisBroadcaster{client: client, prefix: strings.TrimSpace(prefix)}, nil
}

// Publish sends one payload to one Redis Pub/Sub channel.
func (broadcaster *RedisBroadcaster) Publish(ctx context.Context, channel string, payload []byte) error {
	resolved, err := broadcaster.resolveChannel(channel)
	if err != nil {
		return err
	}
	return broadcaster.client.Publish(ctx, resolved, payload).Err()
}

// Subscribe listens for payloads in one Redis Pub/Sub channel.
func (broadcaster *RedisBroadcaster) Subscribe(ctx context.Context, channel string) (<-chan []byte, coreconnection.Disposable, error) {
	resolved, err := broadcaster.resolveChannel(channel)
	if err != nil {
		return nil, nil, err
	}
	pubSub := broadcaster.client.Subscribe(ctx, resolved)
	if _, receiveErr := pubSub.Receive(ctx); receiveErr != nil {
		_ = pubSub.Close()
		return nil, nil, receiveErr
	}
	stream := make(chan []byte, 8)
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
				payload := append([]byte(nil), message.Payload...)
				stream <- payload
			}
		}
	}()
	return stream, coreconnection.DisposeFunc(pubSub.Close), nil
}

// resolveChannel builds one namespaced channel.
func (broadcaster *RedisBroadcaster) resolveChannel(channel string) (string, error) {
	trimmed := strings.TrimSpace(channel)
	if trimmed == "" {
		return "", fmt.Errorf("channel is required")
	}
	if broadcaster.prefix == "" {
		return trimmed, nil
	}
	return broadcaster.prefix + ":" + trimmed, nil
}
