package tests

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
)

// repositoryStub defines deterministic economy repository behavior.
type repositoryStub struct {
	// offer stores deterministic offer return.
	offer domain.MarketplaceOffer
	// tradeLog stores deterministic trade log return.
	tradeLog domain.TradeLog
	// findErr stores deterministic find error.
	findErr error
	// cancelErr stores deterministic cancel error.
	cancelErr error
}

// ListOpenOffers returns deterministic open offer list.
func (s repositoryStub) ListOpenOffers(_ context.Context, f domain.OfferFilter) ([]domain.MarketplaceOffer, int, error) {
	return []domain.MarketplaceOffer{s.offer}, 1, s.findErr
}

// ListOffersBySellerID returns deterministic seller offer list.
func (s repositoryStub) ListOffersBySellerID(_ context.Context, _ int) ([]domain.MarketplaceOffer, error) {
	return []domain.MarketplaceOffer{s.offer}, s.findErr
}

// FindOfferByID returns deterministic offer.
func (s repositoryStub) FindOfferByID(_ context.Context, _ int) (domain.MarketplaceOffer, error) {
	return s.offer, s.findErr
}

// CreateOffer returns deterministic offer.
func (s repositoryStub) CreateOffer(_ context.Context, o domain.MarketplaceOffer) (domain.MarketplaceOffer, error) {
	o.ID = 1
	return o, nil
}

// PurchaseOffer returns deterministic offer.
func (s repositoryStub) PurchaseOffer(_ context.Context, _ int, buyerID int) (domain.MarketplaceOffer, error) {
	o := s.offer
	o.BuyerID = &buyerID
	o.State = domain.OfferStateSold
	return o, s.findErr
}

// CancelOffer returns deterministic error.
func (s repositoryStub) CancelOffer(_ context.Context, _ int) error {
	return s.cancelErr
}

// ExpireOffers returns deterministic expired offers.
func (s repositoryStub) ExpireOffers(_ context.Context, _ int) ([]domain.MarketplaceOffer, error) {
	return []domain.MarketplaceOffer{}, nil
}

// CountActiveOffers returns deterministic count.
func (s repositoryStub) CountActiveOffers(_ context.Context, _ int) (int, error) {
	return 3, s.findErr
}

// GetPriceHistory returns deterministic price history.
func (s repositoryStub) GetPriceHistory(_ context.Context, _ int) ([]domain.PriceHistory, error) {
	return []domain.PriceHistory{}, s.findErr
}

// RecordPriceHistory returns deterministic error.
func (s repositoryStub) RecordPriceHistory(_ context.Context, _ domain.PriceHistory) error {
	return nil
}

// CreateTradeLog returns deterministic trade log.
func (s repositoryStub) CreateTradeLog(_ context.Context, log domain.TradeLog) (domain.TradeLog, error) {
	log.ID = 1
	return log, nil
}
