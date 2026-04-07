package subscription

import sdk "github.com/momlesstomato/pixel-sdk"

// PaydayTriggered fires after an HC payday payout is processed.
type PaydayTriggered struct {
	sdk.BaseEvent
	// ConnID stores the triggering connection identifier.
	ConnID string
	// UserID stores the target user identifier.
	UserID int
	// RewardCredits stores the credits granted by the payout.
	RewardCredits int
	// CreditsSpent stores the cycle credit spend used to compute the reward.
	CreditsSpent int
	// NewCredits stores the resulting user credit balance.
	NewCredits int
}
