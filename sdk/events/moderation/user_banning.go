package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserBanning fires before a user is banned.
type UserBanning struct {
	sdk.BaseCancellable
	// TargetID stores the user being banned.
	TargetID int
	// IssuerID stores the staff member performing the ban.
	IssuerID int
	// Scope stores "room" or "hotel".
	Scope string
	// Reason stores the ban justification.
	Reason string
	// DurationMinutes stores the ban duration (0 for permanent).
	DurationMinutes int
}
