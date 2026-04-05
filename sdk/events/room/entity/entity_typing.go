package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityTyping fires before a room entity starts showing the typing indicator.
type EntityTyping struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// IsTyping stores whether the entity is starting (true) or stopping (false) typing.
	IsTyping bool
}
