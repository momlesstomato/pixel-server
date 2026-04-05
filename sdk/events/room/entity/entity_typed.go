package entity

import sdk "github.com/momlesstomato/pixel-sdk"

// EntityTyped fires after a room entity typing indicator has been updated.
type EntityTyped struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the acting user identifier.
	UserID int
	// VirtualID stores the entity virtual identifier.
	VirtualID int
	// IsTyping stores whether the entity started (true) or stopped (false) typing.
	IsTyping bool
}
