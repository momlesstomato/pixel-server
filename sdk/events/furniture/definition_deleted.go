package furniture

import sdk "github.com/momlesstomato/pixel-sdk"

// DefinitionDeleted fires after a furniture definition is removed.
type DefinitionDeleted struct {
	sdk.BaseEvent
	// DefinitionID stores the removed definition identifier.
	DefinitionID int
}
