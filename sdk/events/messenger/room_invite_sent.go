package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// RoomInviteSent fires before room invites are routed.
type RoomInviteSent struct {
	sdk.BaseCancellable
	// ConnID stores the sender connection identifier.
	ConnID string
	// FromUserID stores the sender user identifier.
	FromUserID int
	// ToUserIDs stores all recipient user identifiers.
	ToUserIDs []int
	// Message stores the invite message content.
	Message string
}
