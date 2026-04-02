package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// OfferCreated fires after a catalog offer is persisted.
type OfferCreated struct {
	sdk.BaseEvent
	// PageID stores the parent page identifier.
	PageID int
	// OfferID stores the created offer identifier.
	OfferID int
}
