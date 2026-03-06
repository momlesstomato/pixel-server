package e2e_test

import (
	"context"
	"testing"
	"time"

	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/factory"
)

// Test04LocalTransportComposition validates local transport factory and message flow.
func Test04LocalTransportComposition(t *testing.T) {
	bus, err := factory.New(factory.Config{NATSURL: "", ForceLocal: true})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer bus.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan string, 1)
	_, err = bus.Subscribe(ctx, sessionmessaging.OutputTopic("1"), func(_ context.Context, message coretransport.Message) error {
		out <- string(message.Payload)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := bus.Publish(ctx, sessionmessaging.OutputTopic("1"), []byte("ok")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case value := <-out:
		if value != "ok" {
			t.Fatalf("unexpected payload: %s", value)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected message")
	}
}
