package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// CurrencyUpdating fires before a user activity-point balance modification is committed.
type CurrencyUpdating struct {
	sdk.BaseCancellable
	// UserID stores the user identifier.
	UserID int
	// CurrencyType stores the currency type identifier.
	CurrencyType int
	// Amount stores the signed currency delta.
	Amount int
}
