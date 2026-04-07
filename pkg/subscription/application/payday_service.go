package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	sdksubscription "github.com/momlesstomato/pixel-sdk/events/subscription"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// GetPaydayConfig resolves the active HC payday configuration.
func (service *Service) GetPaydayConfig(ctx context.Context) (domain.PaydayConfig, error) {
	cfg, err := service.repository.FindPaydayConfig(ctx)
	if errors.Is(err, domain.ErrPaydayConfigNotFound) {
		return service.repository.SavePaydayConfig(ctx, domain.PaydayConfig{IntervalDays: 31})
	}
	return cfg, err
}

// UpdatePaydayConfig validates and persists the active HC payday configuration.
func (service *Service) UpdatePaydayConfig(ctx context.Context, cfg domain.PaydayConfig) (domain.PaydayConfig, error) {
	if cfg.IntervalDays <= 0 || cfg.FlatCredits < 0 || cfg.MinimumCreditsSpent < 0 || cfg.StreakBonusCredits < 0 || cfg.KickbackPercentage < 0 {
		return domain.PaydayConfig{}, fmt.Errorf("payday config values must be non-negative and interval_days must be positive")
	}
	return service.repository.SavePaydayConfig(ctx, cfg)
}

// GetPaydayStatus resolves the current HC payday snapshot for one user.
func (service *Service) GetPaydayStatus(ctx context.Context, userID int) (domain.PaydayStatus, error) {
	if userID <= 0 {
		return domain.PaydayStatus{}, fmt.Errorf("user id must be positive")
	}
	sub, err := service.FindActiveSubscription(ctx, userID)
	if err != nil {
		return domain.PaydayStatus{}, err
	}
	cfg, err := service.GetPaydayConfig(ctx)
	if err != nil {
		return domain.PaydayStatus{}, err
	}
	state, err := service.ensureBenefitsState(ctx, userID, sub, cfg.IntervalDays)
	if err != nil {
		return domain.PaydayStatus{}, err
	}
	spendReward, streakReward := calculatePaydayRewards(cfg, state)
	return domain.PaydayStatus{Config: cfg, State: state, CurrentHCStreakDays: currentHCStreakDays(sub), SpendRewardCredits: spendReward, StreakRewardCredits: streakReward, TotalRewardCredits: spendReward + streakReward}, nil
}

// TrackCatalogSpend records one successful credit spend toward the current HC payday cycle.
func (service *Service) TrackCatalogSpend(ctx context.Context, userID int, credits int) error {
	if userID <= 0 || credits <= 0 {
		return nil
	}
	status, err := service.GetPaydayStatus(ctx, userID)
	if errors.Is(err, domain.ErrSubscriptionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	status.State.CycleCreditsSpent += credits
	_, err = service.repository.SaveBenefitsState(ctx, status.State)
	return err
}

// TriggerPayday processes one HC payday payout for one user.
func (service *Service) TriggerPayday(ctx context.Context, connID string, userID int, force bool) (domain.PaydayResult, error) {
	if service.creditSpender == nil {
		return domain.PaydayResult{}, fmt.Errorf("credit spender is required")
	}
	status, err := service.GetPaydayStatus(ctx, userID)
	if err != nil {
		return domain.PaydayResult{}, err
	}
	now := time.Now().UTC()
	if !force && now.Before(status.State.NextPaydayAt) {
		return domain.PaydayResult{}, domain.ErrPaydayNotReady
	}
	reward := status.TotalRewardCredits
	if service.fire != nil {
		event := &sdksubscription.PaydayTriggering{ConnID: connID, UserID: userID, RewardCredits: reward, CreditsSpent: status.State.CycleCreditsSpent}
		service.fire(event)
		if event.Cancelled() {
			return domain.PaydayResult{}, fmt.Errorf("payday trigger cancelled by plugin")
		}
	}
	newCredits, err := service.creditSpender.AddCredits(ctx, userID, reward)
	if err != nil {
		return domain.PaydayResult{}, err
	}
	status.State.RewardStreak++
	status.State.TotalCreditsRewarded += reward
	if reward == 0 {
		status.State.TotalCreditsMissed += status.State.CycleCreditsSpent
	}
	status.State.CycleCreditsSpent = 0
	for !status.State.NextPaydayAt.After(now) {
		status.State.NextPaydayAt = status.State.NextPaydayAt.AddDate(0, 0, status.Config.IntervalDays)
	}
	status.State, err = service.repository.SaveBenefitsState(ctx, status.State)
	if err != nil {
		return domain.PaydayResult{}, err
	}
	if service.fire != nil {
		service.fire(&sdksubscription.PaydayTriggered{ConnID: connID, UserID: userID, RewardCredits: reward, CreditsSpent: status.State.CycleCreditsSpent, NewCredits: newCredits})
	}
	return domain.PaydayResult{Status: status, RewardCredits: reward, NewCredits: newCredits}, nil
}

func ensureBenefitsStateDefaults(state *domain.BenefitsState, startedAt time.Time, intervalDays int) bool {
	changed := false
	if state.FirstSubscriptionAt.IsZero() || startedAt.Before(state.FirstSubscriptionAt) {
		state.FirstSubscriptionAt = startedAt
		changed = true
	}
	if state.NextPaydayAt.IsZero() {
		state.NextPaydayAt = startedAt.AddDate(0, 0, intervalDays)
		changed = true
	}
	return changed
}

func calculatePaydayRewards(cfg domain.PaydayConfig, state domain.BenefitsState) (int, int) {
	spendReward := cfg.FlatCredits
	if state.CycleCreditsSpent >= cfg.MinimumCreditsSpent {
		spendReward += int(math.Floor(float64(state.CycleCreditsSpent) * cfg.KickbackPercentage / 100))
	}
	streakReward := 0
	if state.RewardStreak > 0 {
		streakReward = cfg.StreakBonusCredits
	}
	return spendReward, streakReward
}