package application

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
)

// CreateTradeLog persists one validated trade audit log.
func (service *Service) CreateTradeLog(ctx context.Context, log domain.TradeLog) (domain.TradeLog, error) {
	if log.UserOneID <= 0 || log.UserTwoID <= 0 {
		return domain.TradeLog{}, fmt.Errorf("both trade participant ids must be positive")
	}
	if log.UserOneID == log.UserTwoID {
		return domain.TradeLog{}, fmt.Errorf("cannot trade with yourself")
	}
	return service.repository.CreateTradeLog(ctx, log)
}

// CountActiveOffers returns active offer count for one seller.
func (service *Service) CountActiveOffers(ctx context.Context, sellerID int) (int, error) {
	if sellerID <= 0 {
		return 0, fmt.Errorf("seller id must be positive")
	}
	return service.repository.CountActiveOffers(ctx, sellerID)
}

// RecordPriceHistory persists one aggregated price data point.
func (service *Service) RecordPriceHistory(ctx context.Context, ph domain.PriceHistory) error {
	if ph.SpriteID <= 0 {
		return fmt.Errorf("sprite id must be positive")
	}
	return service.repository.RecordPriceHistory(ctx, ph)
}
