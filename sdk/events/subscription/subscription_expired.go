package subscription

import sdk "github.com/momlesstomato/pixel-sdk"

// SubscriptionExpired fires after a subscription expires for a user.
type SubscriptionExpired struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// SubscriptionID stores the subscription identifier.
	SubscriptionID int
}
