package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// PageCreating fires before a catalog page is persisted.
type PageCreating struct {
	sdk.BaseCancellable
	// Caption stores the page caption.
	Caption string
}
