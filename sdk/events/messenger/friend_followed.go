package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// FriendFollowed fires after a user follows a friend to their room.
type FriendFollowed struct {
	sdk.BaseEvent
	// UserID stores the following user identifier.
	UserID int
	// FriendUserID stores the followed friend identifier.
	FriendUserID int
}
