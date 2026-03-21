package furniture

import sdk "github.com/momlesstomato/pixel-sdk"

// DefinitionCreated fires after a furniture definition is created.
type DefinitionCreated struct {
	sdk.BaseEvent
	// DefinitionID stores the definition identifier.
	DefinitionID int
}
