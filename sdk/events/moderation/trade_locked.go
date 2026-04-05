package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// TradeLocked fires after a trade lock has been applied.
type TradeLocked struct {
	sdk.BaseEvent
	// UserID stores the user who was trade-locked.
	UserID int
	// IssuerID stores the staff member who applied the lock.
	IssuerID int
	// Reason stores the lock justification.
	Reason string
	// DurationMinutes stores the lock duration.
	DurationMinutes int
}
