package local

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"pixelsv/pkg/core/transport"
)

// TestBusPublishSubscribe validates publish/subscribe round-trip behavior.
func TestBusPublishSubscribe(t *testing.T) {
	bus := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan string, 1)
	_, err := bus.Subscribe(ctx, "session.output.*", func(_ context.Context, message transport.Message) error {
		ch <- string(message.Payload)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := bus.Publish(ctx, "session.output.1", []byte("ok")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case value := <-ch:
		if value != "ok" {
			t.Fatalf("unexpected payload: %s", value)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected message")
	}
}

// TestBusUnsubscribe validates explicit unsubscribe behavior.
func TestBusUnsubscribe(t *testing.T) {
	bus := New()
	ctx := context.Background()
	var calls atomic.Int64
	sub, err := bus.Subscribe(ctx, "room.input.*", func(context.Context, transport.Message) error {
		calls.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := sub.Unsubscribe(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := bus.Publish(ctx, "room.input.1", []byte("x")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("expected zero calls, got %d", calls.Load())
	}
}

// TestBusClose validates close behavior.
func TestBusClose(t *testing.T) {
	bus := New()
	if err := bus.Close(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := bus.Publish(context.Background(), "a.b", []byte("x")); err == nil {
		t.Fatalf("expected closed error")
	}
	if _, err := bus.Subscribe(context.Background(), "a.b", func(context.Context, transport.Message) error { return nil }); err == nil {
		t.Fatalf("expected closed error")
	}
}
