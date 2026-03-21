package economy

import sdk "github.com/momlesstomato/pixel-sdk"

// TradeCompleted fires after a trade between two users is completed.
type TradeCompleted struct {
	sdk.BaseEvent
	// UserOneID stores the first user identifier.
	UserOneID int
	// UserTwoID stores the second user identifier.
	UserTwoID int
	// TradeLogID stores the trade log identifier.
	TradeLogID int
}
