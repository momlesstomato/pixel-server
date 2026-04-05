package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityDancing fires before a room entity changes dance state.
type EntityDancing struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// DanceID stores the requested dance animation identifier.
	DanceID int
}
