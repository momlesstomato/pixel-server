package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// ActionDeactivating fires before a moderation action is deactivated.
type ActionDeactivating struct {
	sdk.BaseCancellable
	// ActionID stores the action being deactivated.
	ActionID int64
	// DeactivatedBy stores the staff member performing the deactivation.
	DeactivatedBy int
}
