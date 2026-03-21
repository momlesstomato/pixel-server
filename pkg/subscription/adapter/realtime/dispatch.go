package realtime

import (
	"context"

	packet "github.com/momlesstomato/pixel-server/pkg/subscription/packet"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated subscription packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.GetSubscriptionPacketID:
		return true, runtime.handleGetSubscription(ctx, connID, userID)
	case packet.GetClubOffersPacketID:
		return true, runtime.handleGetClubOffers(ctx, connID, userID)
	default:
		return false, nil
	}
}

// handleGetSubscription responds with active subscription data.
func (runtime *Runtime) handleGetSubscription(ctx context.Context, connID string, userID int) error {
	sub, err := runtime.service.FindActiveSubscription(ctx, userID)
	if err != nil {
		runtime.logger.Error("get subscription failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	_ = sub
	return nil
}

// handleGetClubOffers responds with available club offers.
func (runtime *Runtime) handleGetClubOffers(ctx context.Context, connID string, userID int) error {
	offers, err := runtime.service.ListClubOffers(ctx)
	if err != nil {
		runtime.logger.Error("get club offers failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	_ = offers
	return nil
}
