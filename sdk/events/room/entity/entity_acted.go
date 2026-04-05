package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityActed fires after a room entity has performed a generic action.
type EntityActed struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// ActionID stores the applied action animation identifier.
	ActionID int
}
