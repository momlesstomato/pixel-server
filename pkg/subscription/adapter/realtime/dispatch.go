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
	case packet.GetHCExtendOfferPacketID:
		return true, runtime.handleGetHCExtendOffer(ctx, connID, userID)
	case packet.GetClubGiftInfoPacketID:
		return true, runtime.handleGetClubGiftInfo(ctx, connID, userID)
	case packet.GetKickbackInfoPacketID:
		return true, runtime.handleGetKickbackInfo(ctx, connID, userID)
	case packet.SelectClubGiftPacketID:
		return true, runtime.handleSelectClubGift(ctx, connID, userID, body)
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
	if err = runtime.sendPacket(connID, packet.ClubOffersPacket{Offers: offers}); err != nil {
		return err
	}
	directDays := int32(0)
	if len(offers) > 0 {
		directDays = int32(offers[0].Days)
	}
	return runtime.sendPacket(connID, packet.DirectClubBuyAvailablePacket{LengthInDays: directDays})
}

// handleGetHCExtendOffer responds with one extend offer for a user with an active subscription.
func (runtime *Runtime) handleGetHCExtendOffer(ctx context.Context, connID string, userID int) error {
	sub, err := runtime.service.FindActiveSubscription(ctx, userID)
	if err != nil {
		return runtime.handleGetClubOffers(ctx, connID, userID)
	}
	offers, err := runtime.service.ListClubOffers(ctx)
	if err != nil || len(offers) == 0 {
		return runtime.handleGetClubOffers(ctx, connID, userID)
	}
	offer := offers[0]
	expiry := sub.StartedAt.Add(time.Duration(sub.DurationDays) * 24 * time.Hour)
	daysLeft := int32(time.Until(expiry).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}
	return runtime.sendPacket(connID, packet.HCExtendOfferPacket{Offer: offer, SubscriptionDaysLeft: daysLeft})
}

// handleGetClubGiftInfo responds with club gift eligibility and selectable gifts.
func (runtime *Runtime) handleGetClubGiftInfo(ctx context.Context, connID string, userID int) error {
	info, err := runtime.service.GetClubGiftInfo(ctx, userID)
	if err != nil {
		runtime.logger.Error("get club gift info failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	return runtime.sendPacket(connID, packet.ClubGiftInfoPacket{DaysUntilNextGift: int32(info.DaysUntilNextGift), GiftsAvailable: int32(info.GiftsAvailable), ActiveDays: int32(info.ActiveDays), Gifts: info.Gifts})
}

// handleGetKickbackInfo responds with HC payday and streak metadata.
func (runtime *Runtime) handleGetKickbackInfo(ctx context.Context, connID string, userID int) error {
	status, err := runtime.service.GetPaydayStatus(ctx, userID)
	if err != nil {
		if !errors.Is(err, subdomain.ErrSubscriptionNotFound) {
			runtime.logger.Error("get kickback info failed", zap.Int("user_id", userID), zap.Error(err))
			return err
		}
		return runtime.sendPacket(connID, packet.KickbackInfoPacket{})
	}
	secondsUntilPayday := int32(time.Until(status.State.NextPaydayAt).Seconds())
	if secondsUntilPayday < 0 {
		secondsUntilPayday = 0
	}
	return runtime.sendPacket(connID, packet.KickbackInfoPacket{
		CurrentHCStreak:             int32(status.CurrentHCStreakDays),
		FirstSubscriptionDate:       status.State.FirstSubscriptionAt.UTC().Format("2006-01-02"),
		KickbackPercentage:          status.Config.KickbackPercentage,
		TotalCreditsMissed:          int32(status.State.TotalCreditsMissed),
		TotalCreditsRewarded:        int32(status.State.TotalCreditsRewarded),
		TotalCreditsSpent:           int32(status.State.CycleCreditsSpent),
		CreditRewardForStreakBonus:  int32(status.StreakRewardCredits),
		CreditRewardForMonthlySpent: int32(status.SpendRewardCredits),
		TimeUntilPayday:             secondsUntilPayday,
	})
}

// handleSelectClubGift validates and delivers one HC club gift.
func (runtime *Runtime) handleSelectClubGift(ctx context.Context, connID string, userID int, body []byte) error {
	reader := codec.NewReader(body)
	name, err := reader.ReadString()
	if err != nil {
		return err
	}
	result, err := runtime.service.ClaimClubGift(ctx, connID, userID, name)
	if err != nil {
		runtime.logger.Error("claim club gift failed", zap.Int("user_id", userID), zap.String("gift_name", name), zap.Error(err))
		return err
	}
	if err = runtime.sendPacket(connID, packet.ClubGiftSelectedPacket{ProductCode: result.Gift.Name, SpriteID: int32(result.Gift.SpriteID), ExtraData: result.Gift.ExtraData}); err != nil {
		return err
	}
	if result.ItemID > 0 && runtime.inventoryItemSender != nil {
		return runtime.inventoryItemSender(ctx, connID, userID, result.ItemID)
	}
	return nil
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
	elapsedDays := int32(time.Since(sub.StartedAt).Hours() / 24)
	if elapsedDays < 0 {
		elapsedDays = 0
	}
	minutesSince := int32(time.Since(sub.StartedAt).Minutes())
	if minutesSince < 0 {
		minutesSince = 0
	}
	return packet.SubscriptionResponsePacket{
		ProductName:              productName,
		DaysToPeriodEnd:          remaining,
		MemberPeriods:            elapsedDays / 31,
		PeriodsAhead:             0,
		ResponseType:             1,
		HasEverBeenMember:        true,
		IsVIP:                    true,
		PastClubDays:             0,
		PastVIPDays:              elapsedDays,
		MinutesUntilExpiration:   minutesLeft,
		MinutesSinceLastModified: minutesSince,
	}
}
