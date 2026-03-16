package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// FriendRequestSent fires before a friend request is persisted.
type FriendRequestSent struct {
	sdk.BaseCancellable
	// ConnID stores the requesting user connection identifier.
	ConnID string
	// FromUserID stores the requesting user identifier.
	FromUserID int
	// ToUserID stores the target user identifier.
	ToUserID int
	// ToUsername stores the target username.
	ToUsername string
}
