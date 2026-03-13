package user

import sdk "github.com/momlesstomato/pixel-sdk"

// Ignored fires before a user ignore relation is persisted.
type Ignored struct {
	sdk.BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the actor user identifier.
	UserID int
	// IgnoredUserID stores the ignored user identifier.
	IgnoredUserID int
}
