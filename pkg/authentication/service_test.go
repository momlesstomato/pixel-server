package authentication

import (
	"context"
	"testing"
	"time"
)

// TestServiceIssueStoresTicketWithDefaultTTL verifies issue fallback TTL behavior.
func TestServiceIssueStoresTicketWithDefaultTTL(t *testing.T) {
	store := &stubStore{}
	service := NewService(store, Config{DefaultTTLSeconds: 300, MaxTTLSeconds: 1800})
	result, err := service.Issue(context.Background(), IssueRequest{UserID: 10})
	if err != nil {
		t.Fatalf("expected issue success, got %v", err)
	}
	if len(result.Ticket) != 32 || store.lastTTL != 5*time.Minute || store.lastUserID != 10 {
		t.Fatalf("unexpected issue result: %+v store=%+v", result, store)
	}
}

// TestServiceIssueRejectsInvalidInputs verifies issue validation behavior.
func TestServiceIssueRejectsInvalidInputs(t *testing.T) {
	service := NewService(&stubStore{}, Config{DefaultTTLSeconds: 300, MaxTTLSeconds: 1800})
	if _, err := service.Issue(context.Background(), IssueRequest{UserID: 0}); err == nil {
		t.Fatalf("expected issue failure for invalid user id")
	}
	if _, err := service.Issue(context.Background(), IssueRequest{UserID: 1, TTL: 31 * time.Minute}); err == nil {
		t.Fatalf("expected issue failure for ttl above maximum")
	}
}

// TestServiceValidateRejectsEmptyTicket verifies validation precondition checks.
func TestServiceValidateRejectsEmptyTicket(t *testing.T) {
	service := NewService(&stubStore{}, Config{DefaultTTLSeconds: 300, MaxTTLSeconds: 1800})
	if _, err := service.Validate(context.Background(), "   "); err == nil {
		t.Fatalf("expected validate failure for empty ticket")
	}
}

// stubStore defines in-memory behavior for service unit tests.
type stubStore struct {
	// lastUserID stores the last user id received by Store.
	lastUserID int
	// lastTTL stores the last ttl received by Store.
	lastTTL time.Duration
}

// Store captures issued ticket inputs.
func (store *stubStore) Store(_ context.Context, _ string, userID int, ttl time.Duration) error {
	store.lastUserID = userID
	store.lastTTL = ttl
	return nil
}

// Validate returns one static user identifier.
func (store *stubStore) Validate(_ context.Context, _ string) (int, error) {
	return 1, nil
}
