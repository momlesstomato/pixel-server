package navigator

import sdk "github.com/momlesstomato/pixel-sdk"

// FavouriteAdded fires after a room is added to favourites.
type FavouriteAdded struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// RoomID stores the room identifier.
	RoomID int
}
