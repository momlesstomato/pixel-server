package realtime

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
)

// TestNewRedisCloseSignalBusRejectsNilClient verifies constructor validation.
func TestNewRedisCloseSignalBusRejectsNilClient(t *testing.T) {
	if _, err := NewRedisCloseSignalBus(nil, ""); err == nil {
		t.Fatalf("expected constructor error for nil client")
	}
}

// TestRedisCloseSignalBusPublishSubscribe verifies signal delivery behavior.
func TestRedisCloseSignalBusPublishSubscribe(t *testing.T) {
	server := startMiniRedis(t)
	defer server.Close()
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	defer client.Close()
	bus, err := NewRedisCloseSignalBus(client, "handshake:test")
	if err != nil {
		t.Fatalf("expected bus constructor success, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals, subscription, err := bus.Subscribe(ctx, "conn-1")
	if err != nil {
		t.Fatalf("expected subscribe success, got %v", err)
	}
	defer subscription.Dispose()
	if err := bus.Publish(ctx, "conn-1", CloseSignal{Code: 4002, Reason: "duplicate"}); err != nil {
		t.Fatalf("expected publish success, got %v", err)
	}
	select {
	case signal := <-signals:
		if signal.Code != 4002 || signal.Reason != "duplicate" {
			t.Fatalf("unexpected signal payload: %#v", signal)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected close signal delivery")
	}
}

// startMiniRedis creates one isolated Redis test server.
func startMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup, got %v", err)
	}
	return server
}
