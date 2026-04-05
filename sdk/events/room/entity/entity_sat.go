package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntitySat fires after a room entity sit posture has been toggled.
type EntitySat struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
}
