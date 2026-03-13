package user

import sdk "github.com/momlesstomato/pixel-sdk"

// MottoChanged fires before a user motto update is persisted.
type MottoChanged struct {
	sdk.BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the user identifier.
	UserID int
	// OldMotto stores previous motto value.
	OldMotto string
	// NewMotto stores requested motto value.
	NewMotto string
}
