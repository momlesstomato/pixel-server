package room

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomLoading fires before a room instance is loaded into memory.
type RoomLoading struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
}
