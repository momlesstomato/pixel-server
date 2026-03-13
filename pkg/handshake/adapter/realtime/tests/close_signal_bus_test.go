package tests

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	redislib "github.com/redis/go-redis/v9"
)

// TestNewRedisCloseSignalBusRejectsNilClient verifies constructor validation behavior.
func TestNewRedisCloseSignalBusRejectsNilClient(t *testing.T) {
	if _, err := handshakerealtime.NewRedisCloseSignalBus(nil, ""); err == nil {
		t.Fatalf("expected constructor error for nil client")
	}
}

// TestNewCloseSignalBusRejectsNilBroadcaster verifies constructor validation behavior.
func TestNewCloseSignalBusRejectsNilBroadcaster(t *testing.T) {
	if _, err := handshakerealtime.NewCloseSignalBus(nil, ""); err == nil {
		t.Fatalf("expected constructor error for nil broadcaster")
	}
}

// TestRedisCloseSignalBusPublishSubscribe verifies close signal delivery behavior.
func TestRedisCloseSignalBusPublishSubscribe(t *testing.T) {
	server := startMiniRedis(t)
	defer server.Close()
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	defer client.Close()
	bus, err := handshakerealtime.NewRedisCloseSignalBus(client, "handshake:test")
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
	if err := bus.Publish(ctx, "conn-1", handshakerealtime.CloseSignal{Code: 4002, Reason: "duplicate"}); err != nil {
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

// TestCloseSignalBusWithLocalBroadcaster verifies local broadcaster integration behavior.
func TestCloseSignalBusWithLocalBroadcaster(t *testing.T) {
	broadcaster := broadcast.NewLocalBroadcaster()
	bus, err := handshakerealtime.NewCloseSignalBus(broadcaster, "handshake:test")
	if err != nil {
		t.Fatalf("expected bus constructor success, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals, subscription, err := bus.Subscribe(ctx, "conn-2")
	if err != nil {
		t.Fatalf("expected subscribe success, got %v", err)
	}
	defer subscription.Dispose()
	if err := bus.Publish(ctx, "conn-2", handshakerealtime.CloseSignal{Code: 4001, Reason: "local"}); err != nil {
		t.Fatalf("expected publish success, got %v", err)
	}
	select {
	case signal := <-signals:
		if signal.Code != 4001 || signal.Reason != "local" {
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
