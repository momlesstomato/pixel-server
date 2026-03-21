package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// Service defines subscription API behavior required by HTTP routes.
type Service interface {
	// FindActiveSubscription resolves active subscription for one user.
	FindActiveSubscription(context.Context, int) (domain.Subscription, error)
	// ListClubOffers resolves all enabled club membership offers.
	ListClubOffers(context.Context) ([]domain.ClubOffer, error)
	// FindClubOfferByID resolves one club offer by identifier.
	FindClubOfferByID(context.Context, int) (domain.ClubOffer, error)
	// CreateClubOffer persists one validated club offer.
	CreateClubOffer(context.Context, domain.ClubOffer) (domain.ClubOffer, error)
	// DeleteClubOffer removes one club offer by identifier.
	DeleteClubOffer(context.Context, int) error
}
