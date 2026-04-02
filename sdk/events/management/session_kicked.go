package management

import sdk "github.com/momlesstomato/pixel-sdk"

// SessionKicked fires after a session is administratively disconnected.
type SessionKicked struct {
	sdk.BaseEvent
	// ConnID stores the disconnected connection identifier.
	ConnID string
	// Reason stores the disconnect reason message.
	Reason string
}
