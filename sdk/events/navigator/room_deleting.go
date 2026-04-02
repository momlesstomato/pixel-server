package navigator

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomDeleting fires before a room is deleted.
type RoomDeleting struct {
	sdk.BaseCancellable
	// RoomID stores the room identifier.
	RoomID int
}
