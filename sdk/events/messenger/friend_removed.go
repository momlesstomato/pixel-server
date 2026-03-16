package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// FriendRemoved fires before a friendship removal is persisted.
type FriendRemoved struct {
	sdk.BaseCancellable
	// UserID stores the initiating user identifier.
	UserID int
	// FriendUserID stores the removed friend identifier.
	FriendUserID int
}
