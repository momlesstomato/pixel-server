package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// TestSearchUsers_ReturnsResults verifies that search returns repository result.
func TestSearchUsers_ReturnsResults(t *testing.T) {
	expectations := []domain.SearchResult{{ID: 5, Username: "mario"}}
	repo := repositoryStub{searchResults: expectations}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	results, err := service.SearchUsers(context.Background(), "mario", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

// TestSearchUsers_EmptyQuery verifies error for empty search query.
func TestSearchUsers_EmptyQuery(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	_, err := service.SearchUsers(context.Background(), "", 20)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

// TestSetRelationship_ValidType verifies that a registered type is accepted for friends.
func TestSetRelationship_ValidType(t *testing.T) {
	repo := repositoryStub{areFriends: true}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.SetRelationship(context.Background(), 1, 2, domain.RelationshipHeart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSetRelationship_InvalidType verifies that an unregistered type is rejected.
func TestSetRelationship_InvalidType(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.SetRelationship(context.Background(), 1, 2, domain.RelationshipType(99))
	if err != domain.ErrInvalidRelationship {
		t.Fatalf("expected ErrInvalidRelationship, got %v", err)
	}
}

// TestGetRelationshipCounts_Positive verifies delegation to repository.
func TestGetRelationshipCounts_Positive(t *testing.T) {
	counts := []domain.RelationshipCount{{Type: domain.RelationshipHeart, Count: 3}}
	repo := repositoryStub{relCounts: counts}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	result, err := service.GetRelationshipCounts(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 count, got %d", len(result))
	}
}

// TestFollowFriend_NotFriends verifies error when users are not friends.
func TestFollowFriend_NotFriends(t *testing.T) {
	repo := repositoryStub{areFriends: false}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	err := service.FollowFriend(context.Background(), 1, 2)
	if err != domain.ErrNotFriends {
		t.Fatalf("expected ErrNotFriends, got %v", err)
	}
}

// TestRegisterRelationship_ExtendedType verifies that custom types can be registered.
func TestRegisterRelationship_ExtendedType(t *testing.T) {
	customType := domain.RelationshipType(100)
	domain.RegisterRelationship(customType, "custom")
	if !domain.IsValidRelationship(customType) {
		t.Fatal("expected custom type to be valid after registration")
	}
	delete(domain.KnownRelationships, customType)
}

// TestGetUserProfiles_ReturnsBulkResults verifies that profile lookup delegates to repository.
func TestGetUserProfiles_ReturnsBulkResults(t *testing.T) {
	expected := []domain.SearchResult{{ID: 1, Username: "alice"}, {ID: 2, Username: "bob"}}
	repo := repositoryStub{searchResults: expected}
	service := newTestService(repo, &sessionRegistryStub{}, &broadcasterStub{})
	results, err := service.GetUserProfiles(context.Background(), []int{1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(results))
	}
}

// TestGetUserProfiles_EmptyInput verifies that empty ids return empty result without error.
func TestGetUserProfiles_EmptyInput(t *testing.T) {
	service := newTestService(repositoryStub{}, &sessionRegistryStub{}, &broadcasterStub{})
	results, err := service.GetUserProfiles(context.Background(), []int{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty result, got %d", len(results))
	}
}
