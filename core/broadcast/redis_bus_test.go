package broadcast

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
)

// TestNewRedisBroadcasterRejectsNilClient verifies constructor validation behavior.
func TestNewRedisBroadcasterRejectsNilClient(t *testing.T) {
	if _, err := NewRedisBroadcaster(nil, ""); err == nil {
		t.Fatalf("expected constructor validation failure")
	}
}

// TestRedisBroadcasterRoundTrip verifies Redis pub/sub round-trip behavior.
func TestRedisBroadcasterRoundTrip(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer server.Close()
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	defer client.Close()
	broadcaster, err := NewRedisBroadcaster(client, "")
	if err != nil {
		t.Fatalf("expected broadcaster creation success, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	messages, disposable, err := broadcaster.Subscribe(ctx, "broadcast:all")
	if err != nil {
		t.Fatalf("expected subscribe success, got %v", err)
	}
	defer disposable.Dispose()
	if err := broadcaster.Publish(ctx, "broadcast:all", []byte("redis-payload")); err != nil {
		t.Fatalf("expected publish success, got %v", err)
	}
	select {
	case message := <-messages:
		if string(message) != "redis-payload" {
			t.Fatalf("unexpected payload %q", string(message))
		}
	case <-time.After(time.Second):
		t.Fatalf("expected payload delivery")
	}
}
