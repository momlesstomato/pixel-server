package domain

import (
	"context"
	"time"
)

// CreditSpender adjusts one user's credit balance.
type CreditSpender interface {
	// AddCredits atomically adjusts credit balance by a signed amount.
	AddCredits(ctx context.Context, userID int, amount int) (int, error)
}

// PaydayConfig defines configurable HC payday behavior.
type PaydayConfig struct {
	// IntervalDays stores the number of days between paydays.
	IntervalDays int
	// KickbackPercentage stores the percentage of spent credits returned on payday.
	KickbackPercentage float64
	// FlatCredits stores the fixed payday credit reward.
	FlatCredits int
	// MinimumCreditsSpent stores the minimum cycle spend required for kickback rewards.
	MinimumCreditsSpent int
	// StreakBonusCredits stores the extra reward granted after consecutive paydays.
	StreakBonusCredits int
	// UpdatedAt stores the last configuration update timestamp.
	UpdatedAt time.Time
}

// BenefitsState stores per-user HC payday and club-gift progress.
type BenefitsState struct {
	// UserID stores the owning user identifier.
	UserID int
	// FirstSubscriptionAt stores the first recorded HC activation time.
	FirstSubscriptionAt time.Time
	// NextPaydayAt stores the next time the user becomes payday-eligible.
	NextPaydayAt time.Time
	// CycleCreditsSpent stores spent credits since the last processed payday.
	CycleCreditsSpent int
	// RewardStreak stores the count of consecutive processed paydays.
	RewardStreak int
	// TotalCreditsRewarded stores lifetime payday credits granted.
	TotalCreditsRewarded int
	// TotalCreditsMissed stores lifetime cycle spend that produced no reward.
	TotalCreditsMissed int
	// ClubGiftsClaimed stores the total claimed HC monthly gifts.
	ClubGiftsClaimed int
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// PaydayStatus defines the current HC payday snapshot shown to clients and APIs.
type PaydayStatus struct {
	// Config stores the active payday configuration.
	Config PaydayConfig
	// State stores the per-user benefits progress state.
	State BenefitsState
	// CurrentHCStreakDays stores the current continuous HC streak length in days.
	CurrentHCStreakDays int
	// SpendRewardCredits stores the spend-based reward currently accrued.
	SpendRewardCredits int
	// StreakRewardCredits stores the streak-bonus reward currently accrued.
	StreakRewardCredits int
	// TotalRewardCredits stores the full payday reward currently accrued.
	TotalRewardCredits int
}

// PaydayResult defines one processed payday payout.
type PaydayResult struct {
	// Status stores the post-trigger payday snapshot.
	Status PaydayStatus
	// RewardCredits stores the reward granted by the trigger.
	RewardCredits int
	// NewCredits stores the user's credit balance after the payout.
	NewCredits int
}
