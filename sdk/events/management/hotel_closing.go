package management

import sdk "github.com/momlesstomato/pixel-sdk"

// HotelClosing fires before a hotel close is scheduled.
type HotelClosing struct {
	sdk.BaseCancellable
	// MinutesUntilClose stores countdown minutes.
	MinutesUntilClose int32
	// DurationMinutes stores maintenance duration.
	DurationMinutes int32
	// ThrowUsers stores whether users are disconnected.
	ThrowUsers bool
}
