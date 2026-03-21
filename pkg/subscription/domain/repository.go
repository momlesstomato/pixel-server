package domain

import "context"

// Repository defines subscription persistence behavior.
type Repository interface {
	// FindActiveSubscription resolves active subscription for one user.
	FindActiveSubscription(ctx context.Context, userID int) (Subscription, error)
	// CreateSubscription persists one subscription row.
	CreateSubscription(context.Context, Subscription) (Subscription, error)
	// ExtendSubscription adds days to an existing active subscription.
	ExtendSubscription(ctx context.Context, userID int, days int) (Subscription, error)
	// DeactivateSubscription marks one subscription as inactive.
	DeactivateSubscription(ctx context.Context, subscriptionID int) error
	// FindExpiredActive resolves all active but elapsed subscriptions.
	FindExpiredActive(context.Context) ([]Subscription, error)
	// ListClubOffers resolves all enabled club membership offers.
	ListClubOffers(context.Context) ([]ClubOffer, error)
	// FindClubOfferByID resolves one club offer by identifier.
	FindClubOfferByID(context.Context, int) (ClubOffer, error)
	// CreateClubOffer persists one club offer row.
	CreateClubOffer(context.Context, ClubOffer) (ClubOffer, error)
	// DeleteClubOffer removes one club offer by identifier.
	DeleteClubOffer(context.Context, int) error
}
