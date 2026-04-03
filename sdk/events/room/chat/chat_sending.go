package chat

import sdk "github.com/momlesstomato/pixel-sdk"

// ChatSending fires before a chat message is distributed to room entities.
type ChatSending struct {
	sdk.BaseCancellable
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
