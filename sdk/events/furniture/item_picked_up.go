package furniture

import sdk "github.com/momlesstomato/pixel-sdk"

// ItemPickedUp fires before a furniture item is picked up.
type ItemPickedUp struct {
	sdk.BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the user identifier.
	UserID int
	// ItemID stores the item identifier.
	ItemID int
}
