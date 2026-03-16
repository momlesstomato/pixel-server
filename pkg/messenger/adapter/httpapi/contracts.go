package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// Service defines messenger API behavior required by HTTP routes.
type Service interface {
	// ListFriends returns all friendships for one user.
	ListFriends(context.Context, int) ([]domain.Friendship, error)
	// AddFriendship force-adds a friendship between two users.
	AddFriendship(context.Context, int, int) error
	// RemoveFriendship removes a friendship between two users.
	RemoveFriendship(context.Context, int, int) error
	// ListPendingRequests returns all pending friend requests for one user.
	ListPendingRequests(context.Context, int) ([]domain.FriendRequest, error)
	// GetRelationshipCounts returns grouped relationship counts for one user profile.
	GetRelationshipCounts(context.Context, int) ([]domain.RelationshipCount, error)
	// SetRelationship updates the relationship type for one friendship direction.
	SetRelationship(context.Context, int, int, domain.RelationshipType) error
}
