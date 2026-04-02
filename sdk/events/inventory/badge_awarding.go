package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// BadgeAwarding fires before a badge is awarded to a user.
type BadgeAwarding struct {
	sdk.BaseCancellable
	// UserID stores the user identifier.
	UserID int
	// BadgeCode stores the badge code.
	BadgeCode string
}
