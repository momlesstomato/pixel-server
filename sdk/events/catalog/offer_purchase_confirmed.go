package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// OfferPurchaseConfirmed fires after a catalog offer purchase is committed.
type OfferPurchaseConfirmed struct {
	sdk.BaseEvent
	// ConnID stores the connection identifier of the buyer.
	ConnID string
	// UserID stores the buyer user identifier.
	UserID int
	// OfferID stores the purchased offer identifier.
	OfferID int
	// Quantity stores the quantity purchased.
	Quantity int
}
