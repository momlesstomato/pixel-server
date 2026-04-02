package navigator

import sdk "github.com/momlesstomato/pixel-sdk"

// FavouriteAdding fires before a room is added to favourites.
type FavouriteAdding struct {
	sdk.BaseCancellable
	// UserID stores the user identifier.
	UserID int
	// RoomID stores the room identifier.
	RoomID int
}
