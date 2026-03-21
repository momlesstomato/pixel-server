package economy

import sdk "github.com/momlesstomato/pixel-sdk"

// MarketplaceSold fires after a marketplace offer is purchased.
type MarketplaceSold struct {
	sdk.BaseEvent
	// SellerID stores the seller identifier.
	SellerID int
	// BuyerID stores the buyer identifier.
	BuyerID int
	// OfferID stores the offer identifier.
	OfferID int
	// Price stores the sale price.
	Price int
}
