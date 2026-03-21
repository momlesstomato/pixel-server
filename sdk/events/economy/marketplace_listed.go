package economy

import sdk "github.com/momlesstomato/pixel-sdk"

// MarketplaceListed fires after an item is listed on the marketplace.
type MarketplaceListed struct {
	sdk.BaseEvent
	// SellerID stores the seller identifier.
	SellerID int
	// OfferID stores the marketplace offer identifier.
	OfferID int
	// Price stores the listing price.
	Price int
}
