package room

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomLeft fires after a user has left a room instance.
type RoomLeft struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the user identifier.
	UserID int
}
