package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// PageCreated fires after a catalog page is created.
type PageCreated struct {
	sdk.BaseEvent
	// PageID stores the page identifier.
	PageID int
}
