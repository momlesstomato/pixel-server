package subscription

import sdk "github.com/momlesstomato/pixel-sdk"

// SubscriptionActivated fires after a subscription is activated for a user.
type SubscriptionActivated struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// SubscriptionID stores the subscription identifier.
	SubscriptionID int
	// Type stores the subscription type name.
	Type string
}
