package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserBanned fires after a user has been banned.
type UserBanned struct {
	sdk.BaseEvent
	// TargetID stores the user who was banned.
	TargetID int
	// IssuerID stores the staff member who performed the ban.
	IssuerID int
	// Scope stores "room" or "hotel".
	Scope string
	// Reason stores the ban justification.
	Reason string
	// DurationMinutes stores the ban duration (0 for permanent).
	DurationMinutes int
}
