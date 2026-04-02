package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkmessenger "github.com/momlesstomato/pixel-sdk/events/messenger"
)

// TestFriendAddedEventFires verifies FriendAdded event fires after successful friendship add.
func TestFriendAddedEventFires(t *testing.T) {
	var afterFired bool
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkmessenger.FriendAdded); ok {
			afterFired = true
		}
	})
	if err := service.AddFriendship(context.Background(), 1, 2); err != nil {
		t.Fatalf("unexpected add friendship error: %v", err)
	}
	if !afterFired {
		t.Fatalf("expected FriendAdded event to fire")
	}
}

// TestFriendRemovedEventCancelsRemoval verifies FriendRemoved cancellation aborts removal.
func TestFriendRemovedEventCancelsRemoval(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if value, ok := event.(*sdkmessenger.FriendRemoved); ok {
			value.Cancel()
		}
	})
	if err := service.RemoveFriendship(context.Background(), 1, 2); err == nil {
		t.Fatalf("expected friend removal to be cancelled")
	}
}

// TestFriendRemovedEventAllowsRemoval verifies FriendRemoved passes through without cancellation.
func TestFriendRemovedEventAllowsRemoval(t *testing.T) {
	var fired bool
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	service.SetEventFirer(func(event sdk.Event) {
		if _, ok := event.(*sdkmessenger.FriendRemoved); ok {
			fired = true
		}
	})
	if err := service.RemoveFriendship(context.Background(), 1, 2); err != nil {
		t.Fatalf("unexpected removal error: %v", err)
	}
	if !fired {
		t.Fatalf("expected FriendRemoved event to fire")
	}
}
