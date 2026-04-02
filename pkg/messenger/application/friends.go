package application

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
	sdkmessenger "github.com/momlesstomato/pixel-sdk/events/messenger"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// ListFriends returns all friendships for one user.
func (service *Service) ListFriends(ctx context.Context, userID int) ([]domain.Friendship, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListFriendships(ctx, userID)
}

// AreFriends reports whether two users share a friendship.
func (service *Service) AreFriends(ctx context.Context, userOneID, userTwoID int) (bool, error) {
	if userOneID <= 0 || userTwoID <= 0 {
		return false, fmt.Errorf("user ids must be positive")
	}
	return service.repository.AreFriends(ctx, userOneID, userTwoID)
}

// FriendCount returns the number of friends for one user.
func (service *Service) FriendCount(ctx context.Context, userID int) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user id must be positive")
	}
	return service.repository.CountFriends(ctx, userID)
}

// AddFriendship force-adds a friendship between two users, bypassing limits.
func (service *Service) AddFriendship(ctx context.Context, userOneID, userTwoID int) error {
	if userOneID <= 0 || userTwoID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	if userOneID == userTwoID {
		return fmt.Errorf("user ids must be different")
	}
	err := service.repository.AddFriendship(ctx, userOneID, userTwoID)
	if err == nil && service.fire != nil {
		service.fire(&sdkmessenger.FriendAdded{UserOneID: userOneID, UserTwoID: userTwoID})
	}
	return err
}

// RemoveFriendship removes a friendship between two users.
func (service *Service) RemoveFriendship(ctx context.Context, userID, friendID int) error {
	if userID <= 0 || friendID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	if service.fire != nil {
		event := &sdkmessenger.FriendRemoved{UserID: userID, FriendUserID: friendID}
		service.fire(event)
		if event.Cancelled() {
			return domain.ErrNotFriends
		}
	}
	return service.repository.RemoveFriendship(ctx, userID, friendID)
}

// NotifyFriendUpdate publishes an encoded friend update packet to one friend's channel.
func (service *Service) NotifyFriendUpdate(ctx context.Context, friendID int, frame []byte) error {
	return service.broadcaster.Publish(ctx, userChannel(friendID), frame)
}

// BuildFriendUpdateFrame encodes a raw friend update payload frame for broadcast.
func BuildFriendUpdateFrame(packetID uint16, body []byte) []byte {
	return codec.EncodeFrame(packetID, body)
}
