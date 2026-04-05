package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// AmbassadorAlerted fires after an ambassador alert has been sent.
type AmbassadorAlerted struct {
	sdk.BaseEvent
	// Message stores the alert message content.
	Message string
	// RecipientCount stores the number of ambassadors notified.
	RecipientCount int
}
