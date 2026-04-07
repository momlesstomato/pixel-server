package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// Service defines subscription application use-cases.
type Service struct {
	// repository stores subscription persistence contract implementation.
	repository domain.Repository
	// creditSpender stores the optional credit reward port used by payday.
	creditSpender domain.CreditSpender
	// itemDeliverer stores the optional furniture delivery port used by club gifts.
	itemDeliverer domain.ItemDeliverer
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
}

// NewService creates one subscription service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("subscription repository is required")
	}
	return &Service{repository: repository}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// SetCreditSpender configures optional credit reward behavior for paydays.
func (service *Service) SetCreditSpender(spender domain.CreditSpender) {
	service.creditSpender = spender
}

// SetItemDeliverer configures optional furniture delivery behavior for club gifts.
func (service *Service) SetItemDeliverer(deliverer domain.ItemDeliverer) {
	service.itemDeliverer = deliverer
}

// FindActiveSubscription resolves active subscription for one user.
func (service *Service) FindActiveSubscription(ctx context.Context, userID int) (domain.Subscription, error) {
	if userID <= 0 {
		return domain.Subscription{}, fmt.Errorf("user id must be positive")
	}
	return service.repository.FindActiveSubscription(ctx, userID)
}

// CreateSubscription persists one validated subscription.
func (service *Service) CreateSubscription(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	if sub.UserID <= 0 {
		return domain.Subscription{}, fmt.Errorf("user id must be positive")
	}
	if sub.DurationDays <= 0 {
		return domain.Subscription{}, fmt.Errorf("duration must be positive")
	}
	return service.repository.CreateSubscription(ctx, sub)
}

// ExtendSubscription adds days to an existing active subscription.
func (service *Service) ExtendSubscription(ctx context.Context, userID int, days int) (domain.Subscription, error) {
	if userID <= 0 {
		return domain.Subscription{}, fmt.Errorf("user id must be positive")
	}
	if days <= 0 {
		return domain.Subscription{}, fmt.Errorf("extension days must be positive")
	}
	return service.repository.ExtendSubscription(ctx, userID, days)
}

// DeactivateSubscription marks one subscription as inactive.
func (service *Service) DeactivateSubscription(ctx context.Context, subscriptionID int) error {
	if subscriptionID <= 0 {
		return fmt.Errorf("subscription id must be positive")
	}
	return service.repository.DeactivateSubscription(ctx, subscriptionID)
}

// ExpireSubscriptions finds and deactivates all elapsed subscriptions.
func (service *Service) ExpireSubscriptions(ctx context.Context) ([]domain.Subscription, error) {
	expired, err := service.repository.FindExpiredActive(ctx)
	if err != nil {
		return nil, err
	}
	for _, sub := range expired {
		if dErr := service.repository.DeactivateSubscription(ctx, sub.ID); dErr != nil {
			return nil, dErr
		}
	}
	return expired, nil
}

// ListClubOffers resolves all enabled club membership offers.
func (service *Service) ListClubOffers(ctx context.Context) ([]domain.ClubOffer, error) {
	return service.repository.ListClubOffers(ctx)
}

// FindClubOfferByID resolves one club offer by identifier.
func (service *Service) FindClubOfferByID(ctx context.Context, id int) (domain.ClubOffer, error) {
	if id <= 0 {
		return domain.ClubOffer{}, fmt.Errorf("club offer id must be positive")
	}
	return service.repository.FindClubOfferByID(ctx, id)
}

// CreateClubOffer persists one validated club offer.
func (service *Service) CreateClubOffer(ctx context.Context, offer domain.ClubOffer) (domain.ClubOffer, error) {
	if offer.Name == "" {
		return domain.ClubOffer{}, fmt.Errorf("club offer name is required")
	}
	if offer.Days <= 0 {
		return domain.ClubOffer{}, fmt.Errorf("club offer days must be positive")
	}
	return service.repository.CreateClubOffer(ctx, offer)
}

// DeleteClubOffer removes one club offer by identifier.
func (service *Service) DeleteClubOffer(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("club offer id must be positive")
	}
	return service.repository.DeleteClubOffer(ctx, id)
}
