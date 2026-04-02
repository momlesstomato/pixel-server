package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// BadgeRevoked fires after a badge is revoked from a user.
type BadgeRevoked struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// BadgeCode stores the badge code.
	BadgeCode string
}
