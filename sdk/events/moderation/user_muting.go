package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserMuting fires before a user is muted.
type UserMuting struct {
	sdk.BaseCancellable
	// TargetID stores the user being muted.
	TargetID int
	// IssuerID stores the staff member performing the mute.
	IssuerID int
	// Scope stores "room" or "hotel".
	Scope string
	// DurationMinutes stores the mute duration (0 for permanent).
	DurationMinutes int
	// Reason stores the mute justification.
	Reason string
}
