package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// ActionDeactivated fires after a moderation action has been deactivated.
type ActionDeactivated struct {
	sdk.BaseEvent
	// ActionID stores the action that was deactivated.
	ActionID int64
	// DeactivatedBy stores the staff member who performed the deactivation.
	DeactivatedBy int
}
