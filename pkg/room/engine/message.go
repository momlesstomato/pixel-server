package engine

import "github.com/momlesstomato/pixel-server/pkg/room/domain"

// MessageType identifies the kind of room message.
type MessageType int

const (
	// MsgEnter requests an entity to join the room.
	MsgEnter MessageType = iota
	// MsgLeave requests an entity to leave the room.
	MsgLeave
	// MsgWalk requests an entity to walk to a destination.
	MsgWalk
	// MsgWarp requests an entity to move directly to a destination tile.
	MsgWarp
	// MsgChat delivers a chat message to the room.
	MsgChat
	// MsgAction delivers an entity action to the room.
	MsgAction
	// MsgDance updates entity dance style.
	MsgDance
	// MsgSign displays a sign above the entity head.
	MsgSign
	// MsgTyping sets or clears the entity typing indicator.
	MsgTyping
	// MsgLookTo rotates entity head toward a coordinate.
	MsgLookTo
	// MsgSit toggles entity sit posture.
	MsgSit
	// MsgStop forces a graceful room shutdown.
	MsgStop
)

// Message carries a command into the room goroutine channel.
type Message struct {
	// Type identifies the message kind.
	Type MessageType
	// Entity stores the entity performing the action.
	Entity *domain.RoomEntity
	// Tile stores an optional target tile payload.
	Tile *domain.Tile
	// TargetX stores walk destination horizontal coordinate.
	TargetX int
	// TargetY stores walk destination vertical coordinate.
	TargetY int
	// Text stores chat or action payload.
	Text string
	// IntValue stores numeric payload for action commands.
	IntValue int
	// Dir stores the target facing direction when applicable.
	Dir int
	// Silent suppresses a follow-up user update broadcast for direct moves.
	Silent bool
	// Animate requests a one-step client movement update instead of an instant relocation.
	Animate bool
	// Reply receives a response after processing.
	Reply chan error
}
