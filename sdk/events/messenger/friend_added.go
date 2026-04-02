package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// FriendAdded fires after a forced friendship is created.
type FriendAdded struct {
	sdk.BaseEvent
	// UserOneID stores the first user identifier.
	UserOneID int
	// UserTwoID stores the second user identifier.
	UserTwoID int
}
