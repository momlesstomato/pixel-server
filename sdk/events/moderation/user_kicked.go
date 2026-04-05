package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// UserKicked fires after a user has been kicked.
type UserKicked struct {
	sdk.BaseEvent
	// TargetID stores the user who was kicked.
	TargetID int
	// IssuerID stores the staff member who performed the kick.
	IssuerID int
	// RoomID stores the room identifier (0 for hotel scope).
	RoomID int
	// Scope stores "room" or "hotel".
	Scope string
}
