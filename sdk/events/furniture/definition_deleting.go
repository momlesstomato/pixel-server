package furniture

import sdk "github.com/momlesstomato/pixel-sdk"

// DefinitionDeleting fires before a furniture definition is removed.
type DefinitionDeleting struct {
	sdk.BaseCancellable
	// DefinitionID stores the definition identifier.
	DefinitionID int
}
