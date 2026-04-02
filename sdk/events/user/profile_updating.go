package user

import sdk "github.com/momlesstomato/pixel-sdk"

// ProfileUpdating fires before a user profile patch is persisted.
type ProfileUpdating struct {
	sdk.BaseCancellable
	// UserID stores the user identifier.
	UserID int
}
