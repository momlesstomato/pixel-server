package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserKicking fires before a user is kicked.
type UserKicking struct {
	sdk.BaseCancellable
	// TargetID stores the user being kicked.
	TargetID int
	// IssuerID stores the staff member performing the kick.
	IssuerID int
	// RoomID stores the room identifier (0 for hotel scope).
	RoomID int
	// Scope stores "room" or "hotel".
	Scope string
}
