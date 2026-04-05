package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityLooked fires after a room entity has rotated toward a target.
type EntityLooked struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// TargetX stores the look target horizontal coordinate.
	TargetX int
	// TargetY stores the look target vertical coordinate.
	TargetY int
}
