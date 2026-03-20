package tests

import (
	"context"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// TestSendMessage_OfflineRouting verifies that messages to offline users are saved offline.
func TestSendMessage_OfflineRouting(t *testing.T) {
	repo := repositoryStub{areFriends: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.SendMessage(context.Background(), "conn1", 1, 2, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSendMessage_InvalidUserID verifies error for zero sender ID.
func TestSendMessage_InvalidUserID(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.SendMessage(context.Background(), "conn1", 0, 2, "hello")
	if err == nil {
		t.Fatal("expected error for zero fromUserID")
	}
}

// TestSendMessage_NotFriends verifies that sending to a non-friend returns ErrNotFriends.
func TestSendMessage_NotFriends(t *testing.T) {
	repo := repositoryStub{areFriends: false}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.SendMessage(context.Background(), "conn1", 1, 2, "hello")
	if err != domain.ErrNotFriends {
		t.Fatalf("expected ErrNotFriends, got %v", err)
	}
}

// TestDeliverOfflineMessages_Positive verifies offline message retrieval.
func TestDeliverOfflineMessages_Positive(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	_, err := service.DeliverOfflineMessages(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDeliverOfflineMessages_InvalidUser verifies error for zero user ID.
func TestDeliverOfflineMessages_InvalidUser(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	_, err := service.DeliverOfflineMessages(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero userID")
	}
}

// TestSendRoomInvite_Positive verifies room invite with valid recipients.
func TestSendRoomInvite_Positive(t *testing.T) {
	sessions := &sessionRegistryStub{byConnID: map[string]coreconnection.Session{
		"conn1": {UserID: 1, ConnID: "conn1"},
	}}
	service := newTestService(repositoryStub{}, sessions, &broadcasterStub{})
	err := service.SendRoomInvite(context.Background(), "conn1", 1, []int{2, 3}, "join us!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPurgeOldOfflineMessages_Positive verifies TTL purge runs without error.
func TestPurgeOldOfflineMessages_Positive(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.PurgeOldOfflineMessages(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSendMessage_FloodBlocked verifies that rapid messages trigger ErrSenderMuted.
func TestSendMessage_FloodBlocked(t *testing.T) {
	repo := repositoryStub{areFriends: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	for i := 0; i < 5; i++ {
		_ = service.SendMessage(context.Background(), "conn1", 1, 2, "hi")
	}
	err := service.SendMessage(context.Background(), "conn1", 1, 2, "hi")
	if err != domain.ErrSenderMuted {
		t.Fatalf("expected ErrSenderMuted after flood, got %v", err)
	}
}

// TestSendMessage_FloodBypass_SkipsRateLimit verifies messenger.flood.bypass skips flood enforcement.
func TestSendMessage_FloodBypass_SkipsRateLimit(t *testing.T) {
	repo := repositoryStub{areFriends: true}
	checker := &permissionCheckerStub{grants: map[string]bool{domain.PermFloodBypass: true}}
	service := newTestServiceWithChecker(repo, &sessionRegistryStub{}, &broadcasterStub{}, checker)
	for i := 0; i < 10; i++ {
		if err := service.SendMessage(context.Background(), "conn1", 1, 2, "hi"); err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
	}
}
