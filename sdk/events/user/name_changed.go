package user

import sdk "github.com/momlesstomato/pixel-sdk"

// NameChanged fires before a username change is persisted.
type NameChanged struct {
	sdk.BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the user identifier.
	UserID int
	// OldName stores previous username.
	OldName string
	// NewName stores requested username.
	NewName string
}
