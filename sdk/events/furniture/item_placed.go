package furniture

import sdk "github.com/momlesstomato/pixel-sdk"

// ItemPlaced fires before a furniture item is placed in a room.
type ItemPlaced struct {
	sdk.BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the user identifier.
	UserID int
	// ItemID stores the item identifier.
	ItemID int
	// RoomID stores the room identifier.
	RoomID int
}
