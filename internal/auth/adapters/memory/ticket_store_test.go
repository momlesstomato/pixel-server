package memory

import (
	"errors"
	"testing"
	"time"

	"pixelsv/internal/auth/domain"
)

// TestTicketStoreLifecycle validates create, consume, and revoke flow.
func TestTicketStoreLifecycle(t *testing.T) {
	store := NewTicketStore()
	ticket, err := store.Create(42, time.Minute)
	if err != nil || ticket == "" {
		t.Fatalf("expected ticket, got %q %v", ticket, err)
	}
	userID, err := store.Consume(ticket)
	if err != nil || userID != 42 {
		t.Fatalf("unexpected consume result: %d %v", userID, err)
	}
	if _, err := store.Consume(ticket); !errors.Is(err, domain.ErrTicketNotFound) {
		t.Fatalf("expected consumed ticket missing error, got %v", err)
	}
}

// TestTicketStoreValidation validates invalid input behavior.
func TestTicketStoreValidation(t *testing.T) {
	store := NewTicketStore()
	if _, err := store.Create(0, time.Minute); !errors.Is(err, domain.ErrInvalidUserID) {
		t.Fatalf("expected invalid user id error, got %v", err)
	}
	if _, err := store.Consume(""); !errors.Is(err, domain.ErrInvalidTicket) {
		t.Fatalf("expected invalid ticket error, got %v", err)
	}
	if err := store.Revoke(""); !errors.Is(err, domain.ErrInvalidTicket) {
		t.Fatalf("expected invalid ticket error, got %v", err)
	}
}
