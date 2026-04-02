package authentication

import sdk "github.com/momlesstomato/pixel-sdk"

// SSOGenerated fires after an SSO ticket is issued.
type SSOGenerated struct {
	sdk.BaseEvent
	// UserID stores the authenticated user identifier.
	UserID int
	// Ticket stores the issued SSO token value.
	Ticket string
}
