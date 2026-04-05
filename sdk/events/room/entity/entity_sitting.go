package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntitySitting fires before a room entity toggles sit posture.
type EntitySitting struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
}
