package room

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomEntered fires after a user has entered a room instance.
type RoomEntered struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the user identifier.
	UserID int
	// VirtualID stores the assigned room entity identifier.
	VirtualID int
}
