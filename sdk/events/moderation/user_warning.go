package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserWarning fires before a user is warned.
type UserWarning struct {
	sdk.BaseCancellable
	// TargetID stores the user being warned.
	TargetID int
	// IssuerID stores the staff member performing the warning.
	IssuerID int
	// Message stores the warning message.
	Message string
}
