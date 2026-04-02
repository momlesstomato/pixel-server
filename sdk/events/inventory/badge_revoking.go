package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// BadgeRevoking fires before a badge is revoked from a user.
type BadgeRevoking struct {
	sdk.BaseCancellable
	// UserID stores the user identifier.
	UserID int
	// BadgeCode stores the badge code.
	BadgeCode string
}
