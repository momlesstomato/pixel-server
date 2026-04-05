package moderation

import sdk "github.com/momlesstomato/pixel-sdk"

// VisitRecorded fires after a room visit has been recorded.
type VisitRecorded struct {
	sdk.BaseEvent
	// UserID stores the user who visited the room.
	UserID int
	// RoomID stores the room that was visited.
	RoomID int
}
