package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityLooking fires before a room entity rotates toward a target.
type EntityLooking struct {
	sdk.BaseCancellable
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
