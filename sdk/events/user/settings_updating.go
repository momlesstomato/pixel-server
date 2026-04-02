package user

import sdk "github.com/momlesstomato/pixel-sdk"

// SettingsUpdating fires before user settings are persisted.
type SettingsUpdating struct {
	sdk.BaseCancellable
	// UserID stores the user identifier.
	UserID int
}
