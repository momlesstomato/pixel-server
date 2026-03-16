package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// PrivateMessageSent fires before a private message is routed or stored.
type PrivateMessageSent struct {
	sdk.BaseCancellable
	// ConnID stores the sender connection identifier.
	ConnID string
	// FromUserID stores the sender user identifier.
	FromUserID int
	// ToUserID stores the recipient user identifier.
	ToUserID int
	// Message stores the message content.
	Message string
}
