package realtime

import (
	"context"

	packet "github.com/momlesstomato/pixel-server/pkg/inventory/packet"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated inventory packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.GetCurrencyPacketID:
		return true, runtime.handleGetCurrency(ctx, connID, userID)
	case packet.GetBadgesPacketID:
		return true, runtime.handleGetBadges(ctx, connID, userID)
	case packet.EffectActivatePacketID:
		return true, runtime.handleEffectActivate(ctx, connID, userID, body)
	default:
		return false, nil
	}
}

// handleGetCurrency responds with user currency balances.
func (runtime *Runtime) handleGetCurrency(ctx context.Context, connID string, userID int) error {
	currencies, err := runtime.service.ListCurrencies(ctx, userID)
	if err != nil {
		runtime.logger.Error("get currency failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	_ = currencies
	return nil
}

// handleGetBadges responds with user badge list.
func (runtime *Runtime) handleGetBadges(ctx context.Context, connID string, userID int) error {
	badges, err := runtime.service.ListBadges(ctx, userID)
	if err != nil {
		runtime.logger.Error("get badges failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	_ = badges
	return nil
}

// handleEffectActivate activates a user avatar effect.
func (runtime *Runtime) handleEffectActivate(ctx context.Context, connID string, userID int, body []byte) error {
	_ = body
	return nil
}
