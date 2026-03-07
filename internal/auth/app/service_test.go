package app

import (
	"testing"

	"pixelsv/internal/auth/adapters/memory"
	authmessaging "pixelsv/internal/auth/messaging"
	"pixelsv/pkg/codec"
	"pixelsv/pkg/plugin"
	"pixelsv/pkg/plugin/eventbus"
	"pixelsv/pkg/protocol"
)

// TestServiceTicketFlow validates create, validate, and revoke behavior.
func TestServiceTicketFlow(t *testing.T) {
	events := eventbus.New()
	service := NewService(memory.NewTicketStore(), events)
	received := 0
	events.On(authmessaging.EventTicketValidated, func(event *plugin.Event) error {
		received++
		if event.SessionID != "s-1" {
			t.Fatalf("unexpected session id: %s", event.SessionID)
		}
		payload, ok := event.Data.(authmessaging.TicketValidatedEventData)
		if !ok || payload.UserID != 55 {
			t.Fatalf("unexpected payload: %#v", event.Data)
		}
		return nil
	})
	ticket, ttlSeconds, err := service.CreateTicket(55, 0)
	if err != nil || ticket == "" || ttlSeconds != 300 {
		t.Fatalf("unexpected create result: %q %d %v", ticket, ttlSeconds, err)
	}
	if err := service.RecordReleaseVersion("s-1", &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	userID, err := service.ValidateTicket("s-1", ticket)
	if err != nil || userID != 55 {
		t.Fatalf("unexpected validate result: %d %v", userID, err)
	}
	if received != 1 {
		t.Fatalf("expected one event, got %d", received)
	}
	if err := service.RevokeTicket("revoked"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestEncodeAuthenticatedEvent validates event encoding format.
func TestEncodeAuthenticatedEvent(t *testing.T) {
	raw := EncodeAuthenticatedEvent("s1", 9)
	reader := codec.NewReader(raw)
	sessionID, err := reader.ReadString()
	if err != nil || sessionID != "s1" {
		t.Fatalf("unexpected session id: %q %v", sessionID, err)
	}
	userID, err := reader.ReadInt32()
	if err != nil || userID != 9 {
		t.Fatalf("unexpected user id: %d %v", userID, err)
	}
}
