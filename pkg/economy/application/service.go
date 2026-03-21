package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
)

// Service defines economy application use-cases.
type Service struct {
	// repository stores economy persistence contract implementation.
	repository domain.Repository
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
}

// NewService creates one economy service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("economy repository is required")
	}
	return &Service{repository: repository}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// ListOpenOffers resolves paginated open marketplace offers.
func (service *Service) ListOpenOffers(ctx context.Context, filter domain.OfferFilter) ([]domain.MarketplaceOffer, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	return service.repository.ListOpenOffers(ctx, filter)
}

// ListOffersBySellerID resolves all offers for one seller.
func (service *Service) ListOffersBySellerID(ctx context.Context, sellerID int) ([]domain.MarketplaceOffer, error) {
	if sellerID <= 0 {
		return nil, fmt.Errorf("seller id must be positive")
	}
	return service.repository.ListOffersBySellerID(ctx, sellerID)
}

// FindOfferByID resolves one marketplace offer by identifier.
func (service *Service) FindOfferByID(ctx context.Context, id int) (domain.MarketplaceOffer, error) {
	if id <= 0 {
		return domain.MarketplaceOffer{}, fmt.Errorf("offer id must be positive")
	}
	return service.repository.FindOfferByID(ctx, id)
}

// CreateOffer persists one validated marketplace offer.
func (service *Service) CreateOffer(ctx context.Context, offer domain.MarketplaceOffer) (domain.MarketplaceOffer, error) {
	if offer.SellerID <= 0 {
		return domain.MarketplaceOffer{}, fmt.Errorf("seller id must be positive")
	}
	if offer.AskingPrice <= 0 {
		return domain.MarketplaceOffer{}, fmt.Errorf("asking price must be positive")
	}
	return service.repository.CreateOffer(ctx, offer)
}

// PurchaseOffer atomically marks offer as sold.
func (service *Service) PurchaseOffer(ctx context.Context, offerID int, buyerID int) (domain.MarketplaceOffer, error) {
	if offerID <= 0 {
		return domain.MarketplaceOffer{}, fmt.Errorf("offer id must be positive")
	}
	if buyerID <= 0 {
		return domain.MarketplaceOffer{}, fmt.Errorf("buyer id must be positive")
	}
	return service.repository.PurchaseOffer(ctx, offerID, buyerID)
}

// CancelOffer atomically marks offer as cancelled.
func (service *Service) CancelOffer(ctx context.Context, offerID int) error {
	if offerID <= 0 {
		return fmt.Errorf("offer id must be positive")
	}
	return service.repository.CancelOffer(ctx, offerID)
}

// ExpireOffers marks all expired offers and returns affected rows.
func (service *Service) ExpireOffers(ctx context.Context) ([]domain.MarketplaceOffer, error) {
	return service.repository.ExpireOffers(ctx, 0)
}

// GetPriceHistory resolves price history data for one sprite.
func (service *Service) GetPriceHistory(ctx context.Context, spriteID int) ([]domain.PriceHistory, error) {
	if spriteID <= 0 {
		return nil, fmt.Errorf("sprite id must be positive")
	}
	return service.repository.GetPriceHistory(ctx, spriteID)
}
