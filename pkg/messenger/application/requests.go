package application

import (
	"context"
	"fmt"

	sdkmessenger "github.com/momlesstomato/pixel-sdk/events/messenger"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// SendRequest creates a friend request between two users.
// If a reverse request exists, the friendship is accepted immediately.
func (service *Service) SendRequest(ctx context.Context, connID string, fromUserID int, toUsername string) (domain.FriendRequest, bool, error) {
	if fromUserID <= 0 {
		return domain.FriendRequest{}, false, fmt.Errorf("user id must be positive")
	}
	toUserID, found, err := service.repository.FindUserIDByUsername(ctx, toUsername)
	if err != nil {
		return domain.FriendRequest{}, false, err
	}
	if !found {
		return domain.FriendRequest{}, false, domain.ErrTargetNotFound
	}
	if toUserID == fromUserID {
		return domain.FriendRequest{}, false, domain.ErrSelfRequest
	}
	alreadyFriends, err := service.repository.AreFriends(ctx, fromUserID, toUserID)
	if err != nil {
		return domain.FriendRequest{}, false, err
	}
	if alreadyFriends {
		return domain.FriendRequest{}, false, domain.ErrAlreadyFriends
	}
	if service.fire != nil {
		event := &sdkmessenger.FriendRequestSent{ConnID: connID, FromUserID: fromUserID, ToUserID: toUserID, ToUsername: toUsername}
		service.fire(event)
		if event.Cancelled() {
			return domain.FriendRequest{}, false, domain.ErrTargetNotAccepting
		}
	}
	reverse, reverseFound, err := service.repository.FindRequestByUsers(ctx, toUserID, fromUserID)
	if err != nil {
		return domain.FriendRequest{}, false, err
	}
	if reverseFound {
		if err = service.checkFriendLimits(ctx, fromUserID, toUserID); err != nil {
			return domain.FriendRequest{}, false, err
		}
		if err = service.repository.DeleteRequest(ctx, reverse.ID); err != nil {
			return domain.FriendRequest{}, false, err
		}
		if err = service.repository.AddFriendship(ctx, fromUserID, toUserID); err != nil {
			return domain.FriendRequest{}, false, err
		}
		return domain.FriendRequest{FromUserID: fromUserID, ToUserID: toUserID}, true, nil
	}
	_, alreadyFound, err := service.repository.FindRequestByUsers(ctx, fromUserID, toUserID)
	if err != nil {
		return domain.FriendRequest{}, false, err
	}
	if alreadyFound {
		return domain.FriendRequest{}, false, domain.ErrDuplicateRequest
	}
	req, err := service.repository.CreateRequest(ctx, fromUserID, toUserID)
	return req, false, err
}

// checkFriendLimits verifies that neither user has reached their permitted friend cap.
func (service *Service) checkFriendLimits(ctx context.Context, userOneID, userTwoID int) error {
	limitOne := service.ResolvedFriendLimit(ctx, userOneID)
	if limitOne > 0 {
		count, err := service.repository.CountFriends(ctx, userOneID)
		if err != nil {
			return err
		}
		if count >= limitOne {
			return domain.ErrFriendListFull
		}
	}
	limitTwo := service.ResolvedFriendLimit(ctx, userTwoID)
	if limitTwo > 0 {
		count, err := service.repository.CountFriends(ctx, userTwoID)
		if err != nil {
			return err
		}
		if count >= limitTwo {
			return domain.ErrTargetFriendListFull
		}
	}
	return nil
}

// AcceptRequest accepts one friend request by from-user identifier.
// If the request row was already removed (e.g. by a concurrent decline) the
// method still proceeds: it skips deletion and falls back to an AreFriends
// check before calling AddFriendship, mirroring the Arcturus accept pattern.
func (service *Service) AcceptRequest(ctx context.Context, userID, fromUserID int) error {
	req, found, err := service.repository.FindRequestByUsers(ctx, fromUserID, userID)
	if err != nil {
		return err
	}
	reverse, reverseFound, err := service.repository.FindRequestByUsers(ctx, userID, fromUserID)
	if err != nil {
		return err
	}
	if err = service.checkFriendLimits(ctx, userID, fromUserID); err != nil {
		return err
	}
	if service.fire != nil {
		event := &sdkmessenger.FriendRequestAccepted{UserID: userID, FriendUserID: fromUserID}
		service.fire(event)
		if event.Cancelled() {
			return domain.ErrTargetNotAccepting
		}
	}
	if found {
		if err = service.repository.DeleteRequest(ctx, req.ID); err != nil {
			return err
		}
	}
	if reverseFound {
		if err = service.repository.DeleteRequest(ctx, reverse.ID); err != nil {
			return err
		}
	}
	if !found && !reverseFound {
		already, err := service.repository.AreFriends(ctx, userID, fromUserID)
		if err != nil {
			return err
		}
		if already {
			return nil
		}
	}
	return service.repository.AddFriendship(ctx, userID, fromUserID)
}

// DeclineRequest declines one friend request by from-user identifier.
// Returns nil if the request no longer exists (idempotent).
func (service *Service) DeclineRequest(ctx context.Context, userID, fromUserID int) error {
	req, found, err := service.repository.FindRequestByUsers(ctx, fromUserID, userID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return service.repository.DeleteRequest(ctx, req.ID)
}

// DeclineAllRequests declines all pending friend requests for one user.
func (service *Service) DeclineAllRequests(ctx context.Context, userID int) error {
	if userID <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	return service.repository.DeleteAllRequests(ctx, userID)
}

// ListPendingRequests returns all pending requests for one user.
func (service *Service) ListPendingRequests(ctx context.Context, userID int) ([]domain.FriendRequest, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListRequests(ctx, userID)
}
