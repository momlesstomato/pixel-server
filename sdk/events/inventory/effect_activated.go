package inventory

import sdk "github.com/momlesstomato/pixel-sdk"

// EffectActivated fires after an avatar effect is activated.
type EffectActivated struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// EffectID stores the effect identifier.
	EffectID int
}
