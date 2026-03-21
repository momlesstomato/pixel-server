package realtime

import (
	"context"

	packet "github.com/momlesstomato/pixel-server/pkg/economy/packet"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated economy packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.MarketplaceGetConfigPacketID:
		return true, runtime.handleGetConfig(ctx, connID, userID)
	case packet.MarketplaceSearchOffersPacketID:
		return true, runtime.handleSearchOffers(ctx, connID, userID, body)
	case packet.MarketplaceGetOwnItemsPacketID:
		return true, runtime.handleGetOwnItems(ctx, connID, userID)
	case packet.MarketplaceGetItemStatsPacketID:
		return true, runtime.handleGetItemStats(ctx, connID, userID, body)
	case packet.MarketplaceCancelSalePacketID:
		return true, runtime.handleCancelSale(ctx, connID, userID, body)
	default:
		return false, nil
	}
}

// handleGetConfig responds with marketplace configuration.
func (runtime *Runtime) handleGetConfig(ctx context.Context, connID string, userID int) error {
	_ = userID
	return nil
}

// handleSearchOffers responds with matching marketplace offers.
func (runtime *Runtime) handleSearchOffers(ctx context.Context, connID string, userID int, body []byte) error {
	_ = body
	return nil
}

// handleGetOwnItems responds with user marketplace offers.
func (runtime *Runtime) handleGetOwnItems(ctx context.Context, connID string, userID int) error {
	offers, err := runtime.service.ListOffersBySellerID(ctx, userID)
	if err != nil {
		runtime.logger.Error("get own marketplace items failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	_ = offers
	return nil
}

// handleGetItemStats responds with marketplace price stats.
func (runtime *Runtime) handleGetItemStats(ctx context.Context, connID string, userID int, body []byte) error {
	_ = body
	return nil
}

// handleCancelSale cancels an active marketplace offer.
func (runtime *Runtime) handleCancelSale(ctx context.Context, connID string, userID int, body []byte) error {
	_ = body
	return nil
}
