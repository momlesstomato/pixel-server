package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserMuted fires after a user has been muted.
type UserMuted struct {
	sdk.BaseEvent
	// TargetID stores the user who was muted.
	TargetID int
	// IssuerID stores the staff member who performed the mute.
	IssuerID int
	// Scope stores "room" or "hotel".
	Scope string
	// DurationMinutes stores the mute duration (0 for permanent).
	DurationMinutes int
	// Reason stores the mute justification.
	Reason string
}
