package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdksubscription "github.com/momlesstomato/pixel-sdk/events/subscription"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// GetClubGiftInfo resolves the current HC gift selector state for one user.
func (service *Service) GetClubGiftInfo(ctx context.Context, userID int) (domain.ClubGiftInfo, error) {
	if userID <= 0 {
		return domain.ClubGiftInfo{}, fmt.Errorf("user id must be positive")
	}
	status, err := service.GetPaydayStatus(ctx, userID)
	if err != nil {
		return domain.ClubGiftInfo{}, err
	}
	gifts, err := service.repository.ListClubGifts(ctx)
	if err != nil {
		return domain.ClubGiftInfo{}, err
	}
	available := eligibleClubGiftCount(status.CurrentHCStreakDays) - status.State.ClubGiftsClaimed
	if available < 0 {
		available = 0
	}
	return domain.ClubGiftInfo{DaysUntilNextGift: daysUntilNextGift(status.CurrentHCStreakDays, available), GiftsAvailable: available, ActiveDays: status.CurrentHCStreakDays, Gifts: gifts}, nil
}

// ClaimClubGift validates and delivers one HC club gift.
func (service *Service) ClaimClubGift(ctx context.Context, connID string, userID int, name string) (domain.ClubGiftClaimResult, error) {
	if service.itemDeliverer == nil {
		return domain.ClubGiftClaimResult{}, fmt.Errorf("club gift deliverer is required")
	}
	giftName := strings.TrimSpace(name)
	if giftName == "" || userID <= 0 {
		return domain.ClubGiftClaimResult{}, fmt.Errorf("gift name and user id are required")
	}
	info, err := service.GetClubGiftInfo(ctx, userID)
	if err != nil {
		return domain.ClubGiftClaimResult{}, err
	}
	if info.GiftsAvailable < 1 {
		return domain.ClubGiftClaimResult{}, domain.ErrClubGiftUnavailable
	}
	gift, err := service.repository.FindClubGiftByName(ctx, giftName)
	if err != nil {
		return domain.ClubGiftClaimResult{}, err
	}
	if info.ActiveDays < gift.DaysRequired {
		return domain.ClubGiftClaimResult{}, domain.ErrClubGiftUnavailable
	}
	if service.fire != nil {
		event := &sdksubscription.ClubGiftClaiming{ConnID: connID, UserID: userID, GiftID: gift.ID, GiftName: gift.Name}
		service.fire(event)
		if event.Cancelled() {
			return domain.ClubGiftClaimResult{}, fmt.Errorf("club gift claim cancelled by plugin")
		}
	}
	itemID, err := service.itemDeliverer.DeliverItem(ctx, userID, gift.ItemDefinitionID, gift.ExtraData, 0, 0)
	if err != nil {
		return domain.ClubGiftClaimResult{}, err
	}
	status, err := service.GetPaydayStatus(ctx, userID)
	if err != nil {
		return domain.ClubGiftClaimResult{}, err
	}
	status.State.ClubGiftsClaimed++
	if _, err = service.repository.SaveBenefitsState(ctx, status.State); err != nil {
		return domain.ClubGiftClaimResult{}, err
	}
	if service.fire != nil {
		service.fire(&sdksubscription.ClubGiftClaimed{ConnID: connID, UserID: userID, GiftID: gift.ID, GiftName: gift.Name, ItemID: itemID})
	}
	return domain.ClubGiftClaimResult{Gift: gift, ItemID: itemID}, nil
}

func (service *Service) ensureBenefitsState(ctx context.Context, userID int, sub domain.Subscription, intervalDays int) (domain.BenefitsState, error) {
	state, err := service.repository.FindBenefitsState(ctx, userID)
	if err == nil {
		if ensureBenefitsStateDefaults(&state, sub.StartedAt.UTC(), intervalDays) {
			return service.repository.SaveBenefitsState(ctx, state)
		}
		return state, nil
	}
	if err != domain.ErrBenefitsStateNotFound {
		return domain.BenefitsState{}, err
	}
	state = domain.BenefitsState{UserID: userID, FirstSubscriptionAt: sub.StartedAt.UTC(), NextPaydayAt: sub.StartedAt.UTC().AddDate(0, 0, intervalDays)}
	return service.repository.SaveBenefitsState(ctx, state)
}

func currentHCStreakDays(sub domain.Subscription) int {
	days := int(time.Since(sub.StartedAt).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

func eligibleClubGiftCount(activeDays int) int {
	if activeDays < 31 {
		return 0
	}
	return activeDays / 31
}

func daysUntilNextGift(activeDays int, giftsAvailable int) int {
	if giftsAvailable > 0 {
		return 0
	}
	remaining := 31 - (activeDays % 31)
	if remaining == 0 {
		return 31
	}
	return remaining
}