package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityMoved fires after a room entity walk has been successfully initiated.
type EntityMoved struct {
	sdk.BaseEvent
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
