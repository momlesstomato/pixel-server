package tests

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// repositoryStub defines deterministic subscription repository behavior.
type repositoryStub struct {
	// subscription stores deterministic subscription return.
	subscription domain.Subscription
	// clubOffer stores deterministic club offer return.
	clubOffer domain.ClubOffer
	// findErr stores deterministic find error.
	findErr error
	// deleteErr stores deterministic delete error.
	deleteErr error
	// deactivateErr stores deterministic deactivate error.
	deactivateErr error
	// expired stores deterministic expired subscriptions.
	expired []domain.Subscription
}

// FindActiveSubscription returns deterministic subscription.
func (s repositoryStub) FindActiveSubscription(_ context.Context, _ int) (domain.Subscription, error) {
	return s.subscription, s.findErr
}

// CreateSubscription returns deterministic subscription.
func (s repositoryStub) CreateSubscription(_ context.Context, sub domain.Subscription) (domain.Subscription, error) {
	sub.ID = 1
	return sub, nil
}

// ExtendSubscription returns deterministic subscription.
func (s repositoryStub) ExtendSubscription(_ context.Context, _ int, days int) (domain.Subscription, error) {
	sub := s.subscription
	sub.DurationDays += days
	return sub, s.findErr
}

// DeactivateSubscription returns deterministic error.
func (s repositoryStub) DeactivateSubscription(_ context.Context, _ int) error {
	return s.deactivateErr
}

// FindExpiredActive returns deterministic expired subscriptions.
func (s repositoryStub) FindExpiredActive(_ context.Context) ([]domain.Subscription, error) {
	return s.expired, nil
}

// ListClubOffers returns deterministic club offer list.
func (s repositoryStub) ListClubOffers(_ context.Context) ([]domain.ClubOffer, error) {
	return []domain.ClubOffer{s.clubOffer}, nil
}

// FindClubOfferByID returns deterministic club offer.
func (s repositoryStub) FindClubOfferByID(_ context.Context, _ int) (domain.ClubOffer, error) {
	return s.clubOffer, s.findErr
}

// CreateClubOffer returns deterministic club offer.
func (s repositoryStub) CreateClubOffer(_ context.Context, o domain.ClubOffer) (domain.ClubOffer, error) {
	o.ID = 1
	return o, nil
}

// DeleteClubOffer returns deterministic error.
func (s repositoryStub) DeleteClubOffer(_ context.Context, _ int) error {
	return s.deleteErr
}
