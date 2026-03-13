package broadcast

import (
	"context"
	"testing"
	"time"
)

// TestLocalBroadcasterPublishSubscribe verifies in-process publish and subscribe behavior.
func TestLocalBroadcasterPublishSubscribe(t *testing.T) {
	broadcaster := NewLocalBroadcaster()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	messages, disposable, err := broadcaster.Subscribe(ctx, "broadcast:all")
	if err != nil {
		t.Fatalf("expected subscribe success, got %v", err)
	}
	defer disposable.Dispose()
	if err := broadcaster.Publish(ctx, "broadcast:all", []byte("payload")); err != nil {
		t.Fatalf("expected publish success, got %v", err)
	}
	select {
	case message := <-messages:
		if string(message) != "payload" {
			t.Fatalf("unexpected payload %q", string(message))
		}
	case <-time.After(time.Second):
		t.Fatalf("expected payload delivery")
	}
}

// TestLocalBroadcasterRejectsEmptyChannel verifies channel validation behavior.
func TestLocalBroadcasterRejectsEmptyChannel(t *testing.T) {
	broadcaster := NewLocalBroadcaster()
	if _, _, err := broadcaster.Subscribe(context.Background(), ""); err == nil {
		t.Fatalf("expected subscribe validation failure")
	}
	if err := broadcaster.Publish(context.Background(), "", []byte("x")); err == nil {
		t.Fatalf("expected publish validation failure")
	}
}
