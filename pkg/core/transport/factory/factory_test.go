package factory

import (
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"pixelsv/pkg/core/transport/local"
	natsbus "pixelsv/pkg/core/transport/nats"
)

// TestNewLocalWhenNoNATS validates local adapter selection by default.
func TestNewLocalWhenNoNATS(t *testing.T) {
	bus, err := New(Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer bus.Close()
	if _, ok := bus.(*local.Bus); !ok {
		t.Fatalf("expected local bus")
	}
}

// TestNewForceLocal validates local selection when ForceLocal is enabled.
func TestNewForceLocal(t *testing.T) {
	bus, err := New(Config{NATSURL: "nats://127.0.0.1:4222", ForceLocal: true})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer bus.Close()
	if _, ok := bus.(*local.Bus); !ok {
		t.Fatalf("expected local bus")
	}
}

// TestNewNATS validates NATS adapter selection when URL is provided.
func TestNewNATS(t *testing.T) {
	srv := runNATSServer(t)
	bus, err := New(Config{NATSURL: srv.ClientURL()})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer bus.Close()
	if _, ok := bus.(*natsbus.Bus); !ok {
		t.Fatalf("expected nats bus")
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
