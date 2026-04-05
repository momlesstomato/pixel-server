package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityActing fires before a room entity performs a generic action.
type EntityActing struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// ActionID stores the requested action animation identifier.
	ActionID int
}
