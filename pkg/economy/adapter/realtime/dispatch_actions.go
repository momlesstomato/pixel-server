package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
	packet "github.com/momlesstomato/pixel-server/pkg/economy/packet"
	"go.uber.org/zap"
)

// handleCancelSale cancels an active marketplace offer.
func (runtime *Runtime) handleCancelSale(ctx context.Context, connID string, userID int, body []byte) error {
	reader := codec.NewReader(body)
	offerID, err := reader.ReadInt32()
	if err != nil {
		return nil
	}
	if cancelErr := runtime.service.CancelOffer(ctx, int(offerID)); cancelErr != nil {
		runtime.logger.Warn("cancel sale failed", zap.Int("user_id", userID), zap.Int32("offer_id", offerID), zap.Error(cancelErr))
		return runtime.sendPacket(connID, packet.MarketplaceCancelResultPacket{OfferID: offerID, Success: false})
	}
	return runtime.sendPacket(connID, packet.MarketplaceCancelResultPacket{OfferID: offerID, Success: true})
}

// handleBuyOffer processes a marketplace purchase request.
func (runtime *Runtime) handleBuyOffer(ctx context.Context, connID string, userID int, body []byte) error {
	reader := codec.NewReader(body)
	offerID, err := reader.ReadInt32()
	if err != nil {
		return nil
	}
	offer, purchaseErr := runtime.service.PurchaseOffer(ctx, int(offerID), userID)
	if purchaseErr != nil {
		runtime.logger.Warn("buy offer failed", zap.Int("user_id", userID), zap.Int32("offer_id", offerID), zap.Error(purchaseErr))
		resultCode := int32(4)
		if purchaseErr == domain.ErrOfferNotOpen {
			resultCode = 2
		}
		return runtime.sendPacket(connID, packet.MarketplaceBuyResultPacket{Result: resultCode, OfferID: offerID})
	}
	return runtime.sendPacket(connID, packet.MarketplaceBuyResultPacket{Result: 1, OfferID: offerID, NewPrice: int32(offer.AskingPrice)})
}

// handleSellItem lists an item for sale on the marketplace.
func (runtime *Runtime) handleSellItem(ctx context.Context, connID string, userID int, body []byte) error {
	reader := codec.NewReader(body)
	askingPrice, err := reader.ReadInt32()
	if err != nil {
		return nil
	}
	_, _ = reader.ReadInt32()
	itemID, err := reader.ReadInt32()
	if err != nil {
		return nil
	}
	_, createErr := runtime.service.CreateOffer(ctx, domain.MarketplaceOffer{
		SellerID: userID, ItemID: int(itemID), AskingPrice: int(askingPrice),
	})
	if createErr != nil {
		runtime.logger.Warn("sell item failed", zap.Int("user_id", userID), zap.Error(createErr))
		return runtime.sendPacket(connID, packet.MarketplaceItemPostedPacket{Result: 2})
	}
	return runtime.sendPacket(connID, packet.MarketplaceItemPostedPacket{Result: 1})
}

// parseSearchFilter reads the search parameters from a search_offers body.
func parseSearchFilter(body []byte) domain.OfferFilter {
	reader := codec.NewReader(body)
	minPrice, _ := reader.ReadInt32()
	maxPrice, _ := reader.ReadInt32()
	query, _ := reader.ReadString()
	sortMode, _ := reader.ReadInt32()
	return domain.OfferFilter{
		MinPrice:    int(minPrice),
		MaxPrice:    int(maxPrice),
		SearchQuery: query,
		SortMode:    int(sortMode),
		Limit:       100,
	}
}
