package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityDanced fires after a room entity dance state has been changed.
type EntityDanced struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// DanceID stores the applied dance animation identifier.
	DanceID int
}
