package app

import (
	"testing"

	"pixelsv/internal/auth/adapters/memory"
	"pixelsv/pkg/codec"
)

// TestServiceTicketFlow validates create, validate, and revoke behavior.
func TestServiceTicketFlow(t *testing.T) {
	service := NewService(memory.NewTicketStore())
	ticket, ttlSeconds, err := service.CreateTicket(55, 0)
	if err != nil || ticket == "" || ttlSeconds != 300 {
		t.Fatalf("unexpected create result: %q %d %v", ticket, ttlSeconds, err)
	}
	userID, err := service.ValidateTicket(ticket)
	if err != nil || userID != 55 {
		t.Fatalf("unexpected validate result: %d %v", userID, err)
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
