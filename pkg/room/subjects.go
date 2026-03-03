package room

// NATS subjects for the room realm.
const (
	// SubjInput routes post-auth packets from gateway to game. Format: roomID.
	SubjInput = "room.input.%d"
)
