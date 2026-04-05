package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// WordFiltered fires after a chat message has been filtered.
type WordFiltered struct {
	sdk.BaseEvent
	// UserID stores the user whose message was filtered.
	UserID int
	// RoomID stores the room where filtering occurred.
	RoomID int
	// Original stores the original message text.
	Original string
	// Filtered stores the message text after filtering.
	Filtered string
}
