package furniture

import sdk "github.com/momlesstomato/pixel-sdk"

// DefinitionCreating fires before a furniture definition is persisted.
type DefinitionCreating struct {
	sdk.BaseCancellable
	// ItemName stores the definition item name.
	ItemName string
}
