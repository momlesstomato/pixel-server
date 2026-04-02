package navigator

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomCreating fires before a room is persisted.
type RoomCreating struct {
	sdk.BaseCancellable
	// OwnerID stores the room owner identifier.
	OwnerID int
	// Name stores the requested room name.
	Name string
}
