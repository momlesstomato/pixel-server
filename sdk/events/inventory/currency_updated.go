package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// CurrencyUpdated fires after a user currency balance is updated.
type CurrencyUpdated struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// CurrencyType stores the currency type identifier.
	CurrencyType int
	// OldAmount stores the previous amount.
	OldAmount int
	// NewAmount stores the new amount.
	NewAmount int
}
