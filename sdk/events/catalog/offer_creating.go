package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// OfferCreating fires before a catalog offer is persisted.
type OfferCreating struct {
	sdk.BaseCancellable
	// PageID stores the parent page identifier.
	PageID int
	// OfferID stores the offer identifier after creation.
	OfferID int
}
