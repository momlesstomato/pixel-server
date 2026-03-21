package domain

import "context"

// Repository defines economy persistence behavior.
type Repository interface {
	// ListOpenOffers resolves paginated open marketplace offers.
	ListOpenOffers(ctx context.Context, filter OfferFilter) ([]MarketplaceOffer, int, error)
	// ListOffersBySellerID resolves all offers for one seller.
	ListOffersBySellerID(ctx context.Context, sellerID int) ([]MarketplaceOffer, error)
	// FindOfferByID resolves one marketplace offer by identifier.
	FindOfferByID(context.Context, int) (MarketplaceOffer, error)
	// CreateOffer persists one marketplace listing atomically.
	CreateOffer(context.Context, MarketplaceOffer) (MarketplaceOffer, error)
	// PurchaseOffer atomically marks offer as sold.
	PurchaseOffer(ctx context.Context, offerID int, buyerID int) (MarketplaceOffer, error)
	// CancelOffer atomically marks offer as cancelled.
	CancelOffer(ctx context.Context, offerID int) error
	// ExpireOffers marks all expired offers and returns affected rows.
	ExpireOffers(ctx context.Context, maxAgeHours int) ([]MarketplaceOffer, error)
	// CountActiveOffers returns active offer count for one seller.
	CountActiveOffers(ctx context.Context, sellerID int) (int, error)
	// GetPriceHistory resolves price history data for one sprite.
	GetPriceHistory(ctx context.Context, spriteID int) ([]PriceHistory, error)
	// RecordPriceHistory persists one aggregated price data point.
	RecordPriceHistory(context.Context, PriceHistory) error
	// CreateTradeLog persists one trade audit row.
	CreateTradeLog(context.Context, TradeLog) (TradeLog, error)
}

// OfferFilter defines marketplace search filter parameters.
type OfferFilter struct {
	// MinPrice stores minimum asking price filter.
	MinPrice int
	// MaxPrice stores maximum asking price filter.
	MaxPrice int
	// SearchQuery stores text search filter.
	SearchQuery string
	// SortMode stores result sort order identifier.
	SortMode int
	// Offset stores pagination offset.
	Offset int
	// Limit stores pagination page size.
	Limit int
}
