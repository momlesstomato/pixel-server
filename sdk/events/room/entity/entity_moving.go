package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityMoving fires before a room entity initiates a walk toward a destination.
type EntityMoving struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// ToX stores the walk target horizontal coordinate.
	ToX int
	// ToY stores the walk target vertical coordinate.
	ToY int
}
