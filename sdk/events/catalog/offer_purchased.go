package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// OfferPurchased fires before a catalog offer purchase is committed.
type OfferPurchased struct {
	sdk.BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the buyer identifier.
	UserID int
	// OfferID stores the offer identifier.
	OfferID int
	// Quantity stores the quantity purchased.
	Quantity int
}
