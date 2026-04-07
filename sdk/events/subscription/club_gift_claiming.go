package subscription

import sdk "github.com/momlesstomato/pixel-sdk"

// ClubGiftClaiming fires before an HC club gift is delivered.
type ClubGiftClaiming struct {
	sdk.BaseCancellable
	// ConnID stores the triggering connection identifier.
	ConnID string
	// UserID stores the claiming user identifier.
	UserID int
	// GiftID stores the selected club gift identifier.
	GiftID int
	// GiftName stores the selected club gift name.
	GiftName string
}
