package app

import (
	"testing"
	"time"

	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/plugin"
	"pixelsv/pkg/plugin/eventbus"
)

// TestServiceLifecycle validates connect/auth/disconnect behavior.
func TestServiceLifecycle(t *testing.T) {
	service := NewService(nil, time.Second)
	service.clock = func() time.Time { return time.Unix(100, 0) }
	if err := service.SessionConnected("s1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := service.SessionAuthenticated("s1", 9); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := service.MarkPong("s1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := service.MarkLatencyTest("s1", 77); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := service.MarkDesktopView("s1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := service.ActiveAuthenticatedSessions(); len(got) != 1 || got[0] != "s1" {
		t.Fatalf("unexpected sessions: %#v", got)
	}
	service.SessionDisconnected("s1")
	if got := service.ActiveAuthenticatedSessions(); len(got) != 0 {
		t.Fatalf("expected no sessions, got %#v", got)
	}
}

// TestSessionAuthenticatedConcurrentLogin validates concurrent-login eviction.
func TestSessionAuthenticatedConcurrentLogin(t *testing.T) {
	service := NewService(nil, time.Second)
	_ = service.SessionConnected("s1")
	_ = service.SessionConnected("s2")
	if _, err := service.SessionAuthenticated("s1", 99); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	previous, err := service.SessionAuthenticated("s2", 99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if previous != "s1" {
		t.Fatalf("expected previous session s1, got %q", previous)
	}
	if got := service.ActiveAuthenticatedSessions(); len(got) != 1 || got[0] != "s2" {
		t.Fatalf("unexpected sessions: %#v", got)
	}
}

// TestExpirePongTimeoutSessions validates stale session expiration behavior.
func TestExpirePongTimeoutSessions(t *testing.T) {
	service := NewService(nil, time.Second)
	now := time.Unix(200, 0)
	service.clock = func() time.Time { return now.Add(-10 * time.Second) }
	_ = service.SessionConnected("s1")
	_, _ = service.SessionAuthenticated("s1", 1)
	service.clock = func() time.Time { return now }
	_ = service.SessionConnected("s2")
	_, _ = service.SessionAuthenticated("s2", 2)
	expired := service.ExpirePongTimeoutSessions(5*time.Second, now)
	if len(expired) != 1 || expired[0] != "s1" {
		t.Fatalf("unexpected expired sessions: %#v", expired)
	}
}

// TestAllowTelemetryRateLimit validates telemetry log throttling.
func TestAllowTelemetryRateLimit(t *testing.T) {
	service := NewService(nil, time.Second)
	now := time.Unix(300, 0)
	service.clock = func() time.Time { return now }
	_ = service.SessionConnected("s1")
	allowed, err := service.AllowTelemetry("s1", 3230)
	if err != nil || !allowed {
		t.Fatalf("unexpected first telemetry result: %v %v", allowed, err)
	}
	allowed, err = service.AllowTelemetry("s1", 3230)
	if err != nil || allowed {
		t.Fatalf("unexpected second telemetry result: %v %v", allowed, err)
	}
	service.clock = func() time.Time { return now.Add(2 * time.Second) }
	allowed, err = service.AllowTelemetry("s1", 3230)
	if err != nil || !allowed {
		t.Fatalf("unexpected third telemetry result: %v %v", allowed, err)
	}
}

// TestRecordPacketEmitsPluginEvent validates packet-level plugin event emission.
func TestRecordPacketEmitsPluginEvent(t *testing.T) {
	events := eventbus.New()
	service := NewService(events, time.Second)
	_ = service.SessionConnected("s1")
	received := 0
	events.On(sessionmessaging.EventPacketReceived, func(event *plugin.Event) error {
		received++
		payload, ok := event.Data.(sessionmessaging.PacketReceivedEventData)
		if !ok || payload.Header != 2596 || payload.PacketName != "client.pong" {
			t.Fatalf("unexpected payload: %#v", event.Data)
		}
		return nil
	})
	if err := service.RecordPacket("s1", 2596, "client.pong"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received != 1 {
		t.Fatalf("expected one event, got %d", received)
	}
}
