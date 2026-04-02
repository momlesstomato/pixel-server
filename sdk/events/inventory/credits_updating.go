package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// CreditsUpdating fires before a user credits balance modification is committed.
type CreditsUpdating struct {
	sdk.BaseCancellable
	// UserID stores the user identifier.
	UserID int
	// Amount stores the signed credit delta.
	Amount int
}
