package management

import sdk "github.com/momlesstomato/pixel-sdk"

// HotelReopening fires before a hotel reopen is processed.
type HotelReopening struct {
	sdk.BaseCancellable
}
