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
	// paydayConfig stores deterministic payday configuration.
	paydayConfig domain.PaydayConfig
	// benefitsState stores deterministic benefits progress.
	benefitsState domain.BenefitsState
	// clubGifts stores deterministic club gifts.
	clubGifts []domain.ClubGift
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

// FindPaydayConfig returns deterministic payday config.
func (s repositoryStub) FindPaydayConfig(_ context.Context) (domain.PaydayConfig, error) {
	if s.paydayConfig.IntervalDays == 0 {
		return domain.PaydayConfig{}, domain.ErrPaydayConfigNotFound
	}
	return s.paydayConfig, nil
}

// SavePaydayConfig returns deterministic payday config.
func (s repositoryStub) SavePaydayConfig(_ context.Context, cfg domain.PaydayConfig) (domain.PaydayConfig, error) {
	return cfg, nil
}

// FindBenefitsState returns deterministic benefits state.
func (s repositoryStub) FindBenefitsState(_ context.Context, _ int) (domain.BenefitsState, error) {
	if s.benefitsState.UserID == 0 {
		return domain.BenefitsState{}, domain.ErrBenefitsStateNotFound
	}
	return s.benefitsState, nil
}

// SaveBenefitsState returns deterministic benefits state.
func (s repositoryStub) SaveBenefitsState(_ context.Context, state domain.BenefitsState) (domain.BenefitsState, error) {
	return state, nil
}

// ListClubGifts returns deterministic club gift list.
func (s repositoryStub) ListClubGifts(_ context.Context) ([]domain.ClubGift, error) {
	return s.clubGifts, nil
}

// FindClubGiftByName returns deterministic club gift by name.
func (s repositoryStub) FindClubGiftByName(_ context.Context, name string) (domain.ClubGift, error) {
	for _, gift := range s.clubGifts {
		if gift.Name == name {
			return gift, nil
		}
	}
	return domain.ClubGift{}, domain.ErrClubGiftNotFound
}
