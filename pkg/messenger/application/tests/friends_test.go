package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// TestListFriends_Positive verifies that ListFriends delegates to the repository.
func TestListFriends_Positive(t *testing.T) {
	repo := repositoryStub{friendships: []domain.Friendship{{UserOneID: 1, UserTwoID: 2}}}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	result, err := service.ListFriends(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].UserTwoID != 2 {
		t.Fatalf("unexpected result: %v", result)
	}
}

// TestListFriends_InvalidID verifies that a zero user ID returns an error.
func TestListFriends_InvalidID(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	_, err := service.ListFriends(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero userID")
	}
}

// TestAddFriendship_SameUserIDs verifies that same-user add returns an error.
func TestAddFriendship_SameUserIDs(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.AddFriendship(context.Background(), 5, 5)
	if err == nil {
		t.Fatal("expected error for same userID")
	}
}

// TestAddFriendship_Positive verifies that valid user IDs call the repository.
func TestAddFriendship_Positive(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.AddFriendship(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRemoveFriendship_Positive verifies that removal calls the repository.
func TestRemoveFriendship_Positive(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.RemoveFriendship(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAreFriends_Positive verifies that AreFriends returns the repository value.
func TestAreFriends_Positive(t *testing.T) {
	repo := repositoryStub{areFriends: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	result, err := service.AreFriends(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Fatal("expected areFriends=true")
	}
}

// TestFriendCount_InvalidID verifies that a negative user ID returns an error.
func TestFriendCount_InvalidID(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	_, err := service.FriendCount(context.Background(), -1)
	if err == nil {
		t.Fatal("expected error for negative userID")
	}
}
