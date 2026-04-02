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

// handleGetCurrency responds with credit balance and activity-point currency balances.
// Credits are sent via CreditsResponsePacketID (3475); activity-point currencies via CurrencyResponsePacketID (2018).
func (runtime *Runtime) handleGetCurrency(ctx context.Context, connID string, userID int) error {
	credits, err := runtime.service.GetCredits(ctx, userID)
	if err != nil {
		runtime.logger.Error("get credits failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	credBody, encErr := packet.CreditBalancePacket{Balance: credits}.Encode()
	if encErr != nil {
		return encErr
	}
	if sendErr := runtime.transport.Send(connID, packet.CreditsResponsePacketID, credBody); sendErr != nil {
		return sendErr
	}
	currencies, listErr := runtime.service.ListCurrencies(ctx, userID)
	if listErr != nil {
		runtime.logger.Error("get currencies failed", zap.Int("user_id", userID), zap.Error(listErr))
		return listErr
	}
	currBody, encErr := packet.CurrencyBalancePacket{Currencies: currencies}.Encode()
	if encErr != nil {
		return encErr
	}
	return runtime.transport.Send(connID, packet.CurrencyResponsePacketID, currBody)
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
