package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// TradeLocking fires before a trade lock is applied.
type TradeLocking struct {
	sdk.BaseCancellable
	// UserID stores the user being trade-locked.
	UserID int
	// IssuerID stores the staff member applying the lock.
	IssuerID int
	// Reason stores the lock justification.
	Reason string
	// DurationMinutes stores the lock duration.
	DurationMinutes int
}
