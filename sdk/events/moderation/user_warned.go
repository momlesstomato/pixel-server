package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserWarned fires after a user has been warned.
type UserWarned struct {
	sdk.BaseEvent
	// TargetID stores the user who was warned.
	TargetID int
	// IssuerID stores the staff member who performed the warning.
	IssuerID int
	// Message stores the warning message.
	Message string
}
