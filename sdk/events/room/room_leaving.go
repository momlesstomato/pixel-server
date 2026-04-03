package room

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomLeaving fires before a user leaves a room instance.
type RoomLeaving struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the leaving user identifier.
	UserID int
}
