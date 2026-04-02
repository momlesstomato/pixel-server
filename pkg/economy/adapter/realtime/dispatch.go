package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
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
		return true, runtime.handleGetConfig(connID)
	case packet.MarketplaceSearchOffersPacketID:
		return true, runtime.handleSearchOffers(ctx, connID, userID, body)
	case packet.MarketplaceGetOwnItemsPacketID:
		return true, runtime.handleGetOwnItems(ctx, connID, userID)
	case packet.MarketplaceGetItemStatsPacketID:
		return true, runtime.handleGetItemStats(ctx, connID, userID, body)
	case packet.MarketplaceCancelSalePacketID:
		return true, runtime.handleCancelSale(ctx, connID, userID, body)
	case packet.MarketplaceBuyOfferPacketID:
		return true, runtime.handleBuyOffer(ctx, connID, userID, body)
	case packet.MarketplaceGetCanSellPacketID:
		return true, runtime.handleGetCanSell(connID)
	case packet.MarketplaceSellItemPacketID:
		return true, runtime.handleSellItem(ctx, connID, userID, body)
	default:
		return false, nil
	}
}

// handleGetConfig responds with marketplace configuration.
func (runtime *Runtime) handleGetConfig(connID string) error {
	return runtime.sendPacket(connID, packet.MarketplaceConfigPacket{
		Enabled: true, Commission: 1, TokenTax: 0,
		OfferMinPrice: 1, OfferMaxPrice: 999999,
		OfferExpireHours: 48, AverageDays: 7,
	})
}

// handleGetCanSell responds with marketplace sell permission.
func (runtime *Runtime) handleGetCanSell(connID string) error {
	return runtime.sendPacket(connID, packet.MarketplaceCanSellPacket{ErrorCode: 1})
}

// handleSearchOffers responds with matching marketplace offers.
func (runtime *Runtime) handleSearchOffers(ctx context.Context, connID string, userID int, body []byte) error {
	filter := parseSearchFilter(body)
	offers, total, err := runtime.service.ListOpenOffers(ctx, filter)
	if err != nil {
		runtime.logger.Error("search offers failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	return runtime.sendPacket(connID, packet.MarketplaceSearchResultsPacket{Offers: offers, TotalResults: total})
}

// handleGetOwnItems responds with user marketplace offers.
func (runtime *Runtime) handleGetOwnItems(ctx context.Context, connID string, userID int) error {
	offers, err := runtime.service.ListOffersBySellerID(ctx, userID)
	if err != nil {
		runtime.logger.Error("get own marketplace items failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	return runtime.sendPacket(connID, packet.MarketplaceOwnItemsPacket{Offers: offers, CreditsWaiting: 0})
}

// handleGetItemStats responds with marketplace price stats.
func (runtime *Runtime) handleGetItemStats(ctx context.Context, connID string, userID int, body []byte) error {
	reader := codec.NewReader(body)
	_, _ = reader.ReadInt32()
	spriteID, err := reader.ReadInt32()
	if err != nil {
		return nil
	}
	history, hErr := runtime.service.GetPriceHistory(ctx, int(spriteID))
	if hErr != nil {
		runtime.logger.Error("get item stats failed", zap.Int("user_id", userID), zap.Error(hErr))
		return hErr
	}
	avg, count := int32(0), int32(len(history))
	for _, h := range history {
		avg += int32(h.AvgPrice)
	}
	if count > 0 {
		avg /= count
	}
	return runtime.sendPacket(connID, packet.MarketplaceItemStatsPacket{AvgPrice: avg, OfferCount: count, HistoryLength: count})
}
