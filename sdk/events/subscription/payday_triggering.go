package subscription

import sdk "github.com/momlesstomato/pixel-sdk"

// PaydayTriggering fires before an HC payday payout is processed.
type PaydayTriggering struct {
	sdk.BaseCancellable
	// ConnID stores the triggering connection identifier.
	ConnID string
	// UserID stores the target user identifier.
	UserID int
	// RewardCredits stores the credits about to be granted.
	RewardCredits int
	// CreditsSpent stores the cycle credit spend used to compute the reward.
	CreditsSpent int
}
