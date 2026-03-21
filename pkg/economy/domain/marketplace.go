package domain

import "time"

// OfferState represents the lifecycle state of a marketplace offer.
type OfferState string

const (
	// OfferStateOpen identifies an active listing.
	OfferStateOpen OfferState = "open"
	// OfferStateSold identifies a completed sale.
	OfferStateSold OfferState = "sold"
	// OfferStateCancelled identifies a seller-cancelled listing.
	OfferStateCancelled OfferState = "cancelled"
	// OfferStateExpired identifies a time-expired listing.
	OfferStateExpired OfferState = "expired"
)

// MarketplaceOffer defines one marketplace listing entry.
type MarketplaceOffer struct {
	// ID stores stable marketplace offer identifier.
	ID int
	// SellerID stores the listing owner identifier.
	SellerID int
	// ItemID stores the offered item identifier.
	ItemID int
	// DefinitionID stores the item definition reference.
	DefinitionID int
	// AskingPrice stores the seller requested price in credits.
	AskingPrice int
	// State stores the current offer lifecycle state.
	State OfferState
	// BuyerID stores the purchaser identifier when sold.
	BuyerID *int
	// SoldAt stores the sale completion timestamp.
	SoldAt *time.Time
	// ExpireAt stores the listing expiration timestamp.
	ExpireAt time.Time
	// CreatedAt stores listing creation timestamp.
	CreatedAt time.Time
}

// PriceHistory stores one aggregated price data point for charts.
type PriceHistory struct {
	// ID stores stable price history entry identifier.
	ID int
	// SpriteID stores the item sprite being tracked.
	SpriteID int
	// DayOffset stores how many days ago this entry represents.
	DayOffset int
	// AvgPrice stores average sale price for that day.
	AvgPrice int
	// SoldCount stores total units sold for that day.
	SoldCount int
	// RecordedAt stores when this aggregate was computed.
	RecordedAt time.Time
}
