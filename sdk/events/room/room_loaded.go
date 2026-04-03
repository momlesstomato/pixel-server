package room

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomLoaded fires after a room instance has been loaded into memory.
type RoomLoaded struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
}
