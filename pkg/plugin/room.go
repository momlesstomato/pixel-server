package plugin

// RoomService defines ECS-safe room read/write operations.
type RoomService interface {
	// Snapshot reads immutable room state.
	Snapshot(roomID int64) (*RoomSnapshot, error)
	// SendCommand queues a room command for worker-owned mutation.
	SendCommand(roomID int64, command RoomCommand) error
	// BroadcastPacket sends one encoded packet to all room sessions.
	BroadcastPacket(roomID int64, headerID uint16, payload []byte) error
}

// RoomSnapshot is an immutable serialized view of room state.
type RoomSnapshot struct {
	// RoomID identifies the room.
	RoomID int64
	// Tick identifies the worker tick captured in this snapshot.
	Tick uint64
	// EntityCount reports total entities in room state.
	EntityCount int
	// Avatars holds avatar entity snapshots.
	Avatars []AvatarSnapshot
	// Items holds furniture entity snapshots.
	Items []ItemSnapshot
	// Bots holds bot entity snapshots.
	Bots []BotSnapshot
	// Pets holds pet entity snapshots.
	Pets []PetSnapshot
}

// AvatarSnapshot captures avatar position and posture state.
type AvatarSnapshot struct {
	// UserID is the account identifier.
	UserID int64
	// RoomUnit is the room unit identifier.
	RoomUnit int32
	// X is the horizontal X tile coordinate.
	X float32
	// Y is the horizontal Y tile coordinate.
	Y float32
	// Z is the vertical tile coordinate.
	Z float32
	// Posture is the encoded posture identifier.
	Posture uint8
}

// ItemSnapshot captures immutable furniture entity state.
type ItemSnapshot struct {
	// ID is the room item identifier.
	ID int64
	// Type identifies the furniture item type.
	Type string
	// X is the horizontal X tile coordinate.
	X float32
	// Y is the horizontal Y tile coordinate.
	Y float32
	// Z is the vertical tile coordinate.
	Z float32
}

// BotSnapshot captures immutable bot entity state.
type BotSnapshot struct {
	// ID is the bot identifier.
	ID int64
	// Name is the visible bot name.
	Name string
	// X is the horizontal X tile coordinate.
	X float32
	// Y is the horizontal Y tile coordinate.
	Y float32
	// Z is the vertical tile coordinate.
	Z float32
}

// PetSnapshot captures immutable pet entity state.
type PetSnapshot struct {
	// ID is the pet identifier.
	ID int64
	// Name is the visible pet name.
	Name string
	// X is the horizontal X tile coordinate.
	X float32
	// Y is the horizontal Y tile coordinate.
	Y float32
	// Z is the vertical tile coordinate.
	Z float32
}

// RoomCommand defines a typed message enqueued to room workers.
type RoomCommand struct {
	// Type identifies command intent.
	Type string
	// Payload carries command-specific data.
	Payload any
}
