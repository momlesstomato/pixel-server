package navigator

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomDeleted fires after a room is deleted.
type RoomDeleted struct {
	sdk.BaseEvent
	// RoomID stores the deleted room identifier.
	RoomID int
}
