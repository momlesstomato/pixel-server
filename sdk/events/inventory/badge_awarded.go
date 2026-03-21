package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// BadgeAwarded fires after a badge is awarded to a user.
type BadgeAwarded struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// BadgeCode stores the badge code.
	BadgeCode string
}
