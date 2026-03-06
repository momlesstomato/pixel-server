package natsbus

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"pixelsv/pkg/core/transport"
)

// TestNewInvalidURL validates NATS URL validation.
func TestNewInvalidURL(t *testing.T) {
	if _, err := New("not-a-url"); err == nil {
		t.Fatalf("expected nats connection error")
	}
}

// TestBusPublishSubscribe validates NATS publish/subscribe transport flow.
func TestBusPublishSubscribe(t *testing.T) {
	srv := runNATSServer(t)
	bus, err := New(srv.ClientURL())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer bus.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got := make(chan string, 1)
	_, err = bus.Subscribe(ctx, "session.output.*", func(_ context.Context, message transport.Message) error {
		got <- string(message.Payload)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := bus.Publish(ctx, "session.output.7", []byte("ok")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case value := <-got:
		if value != "ok" {
			t.Fatalf("unexpected payload: %s", value)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected message")
	}
}

func runNATSServer(t *testing.T) *server.Server {
	t.Helper()
	opts := &server.Options{Host: "127.0.0.1", Port: -1, NoLog: true, NoSigs: true}
	srv, err := server.NewServer(opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	go srv.Start()
	if !srv.ReadyForConnections(5 * time.Second) {
		t.Fatalf("nats server not ready")
	}
	t.Cleanup(srv.Shutdown)
	return srv
}
