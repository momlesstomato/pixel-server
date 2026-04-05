package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntitySigning fires before a room entity displays a sign.
type EntitySigning struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// SignID stores the requested sign display identifier.
	SignID int
}
