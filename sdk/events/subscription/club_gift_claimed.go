package subscription

import sdk "github.com/momlesstomato/pixel-sdk"

// ClubGiftClaimed fires after an HC club gift is delivered.
type ClubGiftClaimed struct {
	sdk.BaseEvent
	// ConnID stores the triggering connection identifier.
	ConnID string
	// UserID stores the claiming user identifier.
	UserID int
	// GiftID stores the selected club gift identifier.
	GiftID int
	// GiftName stores the selected club gift name.
	GiftName string
	// ItemID stores the delivered furniture item identifier.
	ItemID int
}