package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
)

// Service defines economy API behavior required by HTTP routes.
type Service interface {
	// ListOpenOffers resolves paginated open marketplace offers.
	ListOpenOffers(context.Context, domain.OfferFilter) ([]domain.MarketplaceOffer, int, error)
	// FindOfferByID resolves one marketplace offer by identifier.
	FindOfferByID(context.Context, int) (domain.MarketplaceOffer, error)
	// CreateOffer persists one validated marketplace offer.
	CreateOffer(context.Context, domain.MarketplaceOffer) (domain.MarketplaceOffer, error)
	// CancelOffer atomically marks offer as cancelled.
	CancelOffer(context.Context, int) error
	// GetPriceHistory resolves price history data for one sprite.
	GetPriceHistory(context.Context, int) ([]domain.PriceHistory, error)
	// ListOffersBySellerID resolves all offers for one seller.
	ListOffersBySellerID(context.Context, int) ([]domain.MarketplaceOffer, error)
}
