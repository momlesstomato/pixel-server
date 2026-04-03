package room

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomEntering fires before a user enters a room instance.
type RoomEntering struct {
	sdk.BaseCancellable
	// RoomID stores the target room identifier.
	RoomID int
	// UserID stores the entering user identifier.
	UserID int
}
