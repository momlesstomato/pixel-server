package realtime

import (
	"context"
	"errors"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	subdomain "github.com/momlesstomato/pixel-server/pkg/subscription/domain"
	"github.com/momlesstomato/pixel-server/pkg/subscription/packet"
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
		return true, runtime.handleGetSubscription(ctx, connID, userID, body)
	case packet.GetClubOffersPacketID:
		return true, runtime.handleGetClubOffers(ctx, connID, userID)
	case packet.GetClubGiftInfoPacketID:
		return true, runtime.handleGetClubGiftInfo(connID)
	default:
		return false, nil
	}
}

// handleGetSubscription responds with active subscription data or a zero-state packet.
func (runtime *Runtime) handleGetSubscription(ctx context.Context, connID string, userID int, body []byte) error {
	productName := parseProductName(body)
	sub, err := runtime.service.FindActiveSubscription(ctx, userID)
	if err != nil {
		if !errors.Is(err, subdomain.ErrSubscriptionNotFound) {
			runtime.logger.Error("get subscription failed", zap.Int("user_id", userID), zap.Error(err))
			return err
		}
		return runtime.sendPacket(connID, packet.SubscriptionResponsePacket{ProductName: productName, ResponseType: 1})
	}
	return runtime.sendPacket(connID, buildSubscriptionPacket(sub, productName))
}

// handleGetClubOffers responds with available club offers.
func (runtime *Runtime) handleGetClubOffers(ctx context.Context, connID string, userID int) error {
	offers, err := runtime.service.ListClubOffers(ctx)
	if err != nil {
		runtime.logger.Error("get club offers failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	return runtime.sendPacket(connID, packet.ClubOffersPacket{Offers: offers, WindowID: 0})
}

// handleGetClubGiftInfo responds with club gift eligibility.
func (runtime *Runtime) handleGetClubGiftInfo(connID string) error {
	return runtime.sendPacket(connID, packet.ClubGiftInfoPacket{DaysUntilNextGift: 0, GiftsAvailable: 0})
}

// parseProductName reads the product name string from a get_subscription body.
func parseProductName(body []byte) string {
	reader := codec.NewReader(body)
	s, err := reader.ReadString()
	if err != nil || s == "" {
		return "club_habbo"
	}
	return s
}

// buildSubscriptionPacket converts a domain subscription into a response packet.
func buildSubscriptionPacket(sub subdomain.Subscription, productName string) packet.SubscriptionResponsePacket {
	expiry := sub.StartedAt.Add(time.Duration(sub.DurationDays) * 24 * time.Hour)
	remaining := int32(time.Until(expiry).Hours() / 24)
	if remaining < 0 {
		remaining = 0
	}
	minutesLeft := int32(time.Until(expiry).Minutes())
	if minutesLeft < 0 {
		minutesLeft = 0
	}
	return packet.SubscriptionResponsePacket{
		ProductName:            productName,
		DaysToPeriodEnd:        remaining,
		MemberPeriods:          1,
		PeriodsAhead:           0,
		ResponseType:           1,
		HasEverBeenMember:      true,
		IsVIP:                  sub.SubscriptionType == subdomain.SubscriptionHabboClub,
		PastClubDays:           int32(sub.DurationDays),
		MinutesUntilExpiration: minutesLeft,
	}
}
