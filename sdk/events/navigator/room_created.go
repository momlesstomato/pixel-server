package navigator

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomCreated fires after a room is persisted.
type RoomCreated struct {
	sdk.BaseEvent
	// RoomID stores the created room identifier.
	RoomID int
	// OwnerID stores the room owner identifier.
	OwnerID int
	// Name stores the room name.
	Name string
}
