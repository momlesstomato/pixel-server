package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// CreditsUpdated fires after a user credits balance is updated.
type CreditsUpdated struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// OldAmount stores the previous credits balance.
	OldAmount int
	// NewAmount stores the new credits balance.
	NewAmount int
}
