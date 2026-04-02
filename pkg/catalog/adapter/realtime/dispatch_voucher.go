package realtime

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	"github.com/momlesstomato/pixel-server/pkg/catalog/packet"
	"go.uber.org/zap"
)

// handleRedeemVoucher processes a voucher redemption request.
func (runtime *Runtime) handleRedeemVoucher(ctx context.Context, connID string, userID int, body []byte) error {
	code := parseVoucherCode(body)
	if code == "" {
		return runtime.sendPacket(connID, packet.VoucherRedeemErrorPacket{ErrorCode: "1"})
	}
	voucher, err := runtime.service.RedeemVoucher(ctx, code, userID)
	if err != nil {
		runtime.logger.Warn("voucher redeem failed", zap.Int("user_id", userID), zap.String("code", code), zap.Error(err))
		return runtime.sendVoucherError(connID, err)
	}
	runtime.logger.Info("voucher redeemed", zap.Int("user_id", userID), zap.String("code", voucher.Code))
	return runtime.sendPacket(connID, packet.VoucherRedeemOKPacket{ProductName: voucher.RewardType, IsHC: false})
}

// handleCheckGiftable checks whether a specific offer can be gifted.
func (runtime *Runtime) handleCheckGiftable(ctx context.Context, connID string, userID int, body []byte) error {
	offerID, err := parseOfferID(body)
	if err != nil {
		return err
	}
	offer, findErr := runtime.service.FindOfferByID(ctx, int(offerID))
	if findErr != nil {
		runtime.logger.Warn("check giftable failed", zap.Int("user_id", userID), zap.Int32("offer_id", offerID), zap.Error(findErr))
		return runtime.sendPacket(connID, packet.IsOfferGiftablePacket{OfferID: offerID, Giftable: false})
	}
	return runtime.sendPacket(connID, packet.IsOfferGiftablePacket{OfferID: offerID, Giftable: !offer.IsLimited()})
}

// sendVoucherError maps a voucher error to a client error packet.
func (runtime *Runtime) sendVoucherError(connID string, err error) error {
	code := "0"
	switch {
	case errors.Is(err, domain.ErrVoucherNotFound):
		code = "1"
	case errors.Is(err, domain.ErrVoucherAlreadyRedeemed):
		code = "1"
	case errors.Is(err, domain.ErrVoucherExhausted):
		code = "1"
	case errors.Is(err, domain.ErrVoucherDisabled):
		code = "1"
	}
	return runtime.sendPacket(connID, packet.VoucherRedeemErrorPacket{ErrorCode: code})
}

// parseVoucherCode reads the voucher code string from a redeem_voucher body.
func parseVoucherCode(body []byte) string {
	reader := codec.NewReader(body)
	s, err := reader.ReadString()
	if err != nil {
		return ""
	}
	return s
}

// parseOfferID reads the offer identifier from a check_giftable body.
func parseOfferID(body []byte) (int32, error) {
	reader := codec.NewReader(body)
	return reader.ReadInt32()
}
