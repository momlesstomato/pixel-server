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

// TestPublishDeliversBurstMessages verifies burst messages are not dropped.
func TestPublishDeliversBurstMessages(t *testing.T) {
	broadcaster := NewLocalBroadcaster()
	stream, disposable, err := broadcaster.Subscribe(context.Background(), "burst")
	if err != nil {
		t.Fatalf("expected subscribe success, got %v", err)
	}
	defer func() { _ = disposable.Dispose() }()
	received := make([]byte, 0, 16)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 16; i++ {
			payload := <-stream
			if len(payload) == 1 {
				received = append(received, payload[0])
			}
		}
	}()
	for i := range 16 {
		if err = broadcaster.Publish(context.Background(), "burst", []byte{byte(i)}); err != nil {
			t.Fatalf("publish burst message %d failed: %v", i, err)
		}
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected consumer to receive all burst messages")
	}
	if len(received) != 16 {
		t.Fatalf("expected 16 burst payloads, got %d", len(received))
	}
	for i := range 16 {
		if received[i] != byte(i) {
			t.Fatalf("unexpected payload at index %d: %d", i, received[i])
		}
	}
}
