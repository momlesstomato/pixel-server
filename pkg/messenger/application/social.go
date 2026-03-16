package application

import (
	"context"
	"fmt"

	sdkmessenger "github.com/momlesstomato/pixel-sdk/events/messenger"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// SearchUsers returns users matching a query string.
func (service *Service) SearchUsers(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	return service.repository.SearchUsers(ctx, query, limit)
}

// GetUserProfiles returns profile records for a set of user identifiers.
func (service *Service) GetUserProfiles(ctx context.Context, userIDs []int) ([]domain.SearchResult, error) {
	return service.repository.FindUsersByIDs(ctx, userIDs)
}

// SetRelationship updates the relationship type for one friendship direction.
func (service *Service) SetRelationship(ctx context.Context, userID, friendID int, rel domain.RelationshipType) error {
	if userID <= 0 || friendID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	if !domain.IsValidRelationship(rel) {
		return domain.ErrInvalidRelationship
	}
	friends, err := service.repository.AreFriends(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !friends {
		return domain.ErrNotFriends
	}
	old, err := service.repository.GetRelationship(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if service.fire != nil {
		event := &sdkmessenger.RelationshipChanged{UserID: userID, FriendUserID: friendID, OldType: int(old), NewType: int(rel)}
		service.fire(event)
		if event.Cancelled() {
			return domain.ErrInvalidRelationship
		}
	}
	return service.repository.SetRelationship(ctx, userID, friendID, rel)
}

// GetRelationshipCounts returns grouped relationship counts for one user profile.
func (service *Service) GetRelationshipCounts(ctx context.Context, userID int) ([]domain.RelationshipCount, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.GetRelationshipCounts(ctx, userID)
}

// FollowFriend dispatches a follow-to-room event for one friend.
func (service *Service) FollowFriend(ctx context.Context, userID, friendID int) error {
	if userID <= 0 || friendID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	friends, err := service.repository.AreFriends(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !friends {
		return domain.ErrNotFriends
	}
	_, online := service.sessions.FindByUserID(friendID)
	if !online {
		return domain.ErrFriendNotInRoom
	}
	if service.fire != nil {
		event := &sdkmessenger.FriendFollowed{UserID: userID, FriendUserID: friendID}
		service.fire(event)
	}
	return nil
}
