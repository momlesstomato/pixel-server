package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntitySigned fires after a room entity has displayed a sign.
type EntitySigned struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// SignID stores the applied sign display identifier.
	SignID int
}
