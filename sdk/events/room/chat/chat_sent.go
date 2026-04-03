package chat

import sdk "github.com/momlesstomato/pixel-sdk"

// ChatSent fires after a chat message has been distributed to room entities.
type ChatSent struct {
	sdk.BaseEvent
	// RoomID stores the room identifier.
	RoomID int
	// UserID stores the sender user identifier.
	UserID int
	// VirtualID stores the sender entity virtual identifier.
	VirtualID int
	// Message stores the chat text payload.
	Message string
	// ChatType stores the message kind: talk, shout, or whisper.
	ChatType string
}
