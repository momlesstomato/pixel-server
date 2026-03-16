package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// FriendRequestAccepted fires before a friend request acceptance is persisted.
type FriendRequestAccepted struct {
	sdk.BaseCancellable
	// UserID stores the accepting user identifier.
	UserID int
	// FriendUserID stores the accepted friend identifier.
	FriendUserID int
}
