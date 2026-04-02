package realtime

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	"github.com/momlesstomato/pixel-server/pkg/catalog/packet"
	"go.uber.org/zap"
)

// handleGetGiftWrappingConfig responds with gift wrapping configuration.
func (runtime *Runtime) handleGetGiftWrappingConfig(connID string) error {
	return runtime.sendPacket(connID, packet.DefaultGiftWrappingConfig())
}

// handlePurchase processes a catalog purchase request from the client.
func (runtime *Runtime) handlePurchase(ctx context.Context, connID string, userID int, body []byte) error {
	offerID, extraData, amount, err := parsePurchaseRequest(body)
	if err != nil {
		return err
	}
	offer, purchaseErr := runtime.service.PurchaseOffer(ctx, connID, userID, int(offerID), extraData, int(amount))
	if purchaseErr != nil {
		return runtime.sendPurchaseError(connID, purchaseErr)
	}
	runtime.logger.Info("catalog purchase successful", zap.Int("user_id", userID), zap.Int32("offer_id", offerID))
	return runtime.sendPacket(connID, packet.PurchaseOKPacket{Offer: buildOfferEntry(offer)})
}

// handlePurchaseGift processes a catalog gift purchase request from the client.
func (runtime *Runtime) handlePurchaseGift(ctx context.Context, connID string, userID int, body []byte) error {
	offerID, extraData, recipientName, err := parseGiftRequest(body)
	if err != nil {
		return err
	}
	offer, purchaseErr := runtime.service.PurchaseGift(ctx, connID, userID, int(offerID), extraData, recipientName)
	if purchaseErr != nil {
		return runtime.sendPurchaseError(connID, purchaseErr)
	}
	runtime.logger.Info("catalog gift purchase successful", zap.Int("user_id", userID), zap.Int32("offer_id", offerID), zap.String("recipient", recipientName))
	return runtime.sendPacket(connID, packet.PurchaseOKPacket{Offer: buildOfferEntry(offer)})
}

// sendPurchaseError maps a purchase error to the appropriate client error packet.
func (runtime *Runtime) sendPurchaseError(connID string, err error) error {
	switch {
	case errors.Is(err, domain.ErrInsufficientCredits):
		return runtime.sendPacket(connID, packet.PurchaseErrorPacket{Code: packet.PurchaseErrorInsufficientCredits})
	case errors.Is(err, domain.ErrInsufficientActivityPoints):
		return runtime.sendPacket(connID, packet.PurchaseErrorPacket{Code: packet.PurchaseErrorInsufficientPoints})
	case errors.Is(err, domain.ErrOfferInactive), errors.Is(err, domain.ErrOfferSoldOut),
		errors.Is(err, domain.ErrPageDisabled), errors.Is(err, domain.ErrOfferNotFound):
		return runtime.sendPacket(connID, packet.PurchaseNotAllowedPacket{Code: 0})
	default:
		return runtime.sendPacket(connID, packet.PurchaseErrorPacket{Code: packet.PurchaseErrorGeneric})
	}
}

// parsePurchaseRequest reads pageId, offerId, extraData and amount from a purchase body.
func parsePurchaseRequest(body []byte) (int32, string, int32, error) {
	reader := codec.NewReader(body)
	_, err := reader.ReadInt32()
	if err != nil {
		return 0, "", 0, err
	}
	offerID, err := reader.ReadInt32()
	if err != nil {
		return 0, "", 0, err
	}
	extraData, err := reader.ReadString()
	if err != nil {
		extraData = ""
	}
	amount, err := reader.ReadInt32()
	if err != nil || amount < 1 {
		amount = 1
	}
	return offerID, extraData, amount, nil
}

// parseGiftRequest reads pageId, itemId, extraData, recipientName and gift fields from a purchase_gift body.
func parseGiftRequest(body []byte) (int32, string, string, error) {
	reader := codec.NewReader(body)
	_, err := reader.ReadInt32()
	if err != nil {
		return 0, "", "", err
	}
	offerID, err := reader.ReadInt32()
	if err != nil {
		return 0, "", "", err
	}
	extraData, _ := reader.ReadString()
	recipientName, err := reader.ReadString()
	if err != nil {
		return 0, "", "", err
	}
	return offerID, extraData, recipientName, nil
}
