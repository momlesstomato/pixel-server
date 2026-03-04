// Package bus provides a thin wrapper around NATS JetStream and defines
// infrastructure-level NATS subjects shared across all services.
// Domain-specific subjects are owned by their respective feature packages.
package bus

// Infrastructure-level NATS subjects for session lifecycle.
// Subjects with a trailing dot are prefixes — append a session/user ID.
const (
	// SubjHandshakeC2S is the prefix for pre-auth packets.
	// Full subject: handshake.c2s.<sessionID>
	SubjHandshakeC2S = "handshake.c2s"

	// SubjSessionAuthenticated is published by auth after SSO validation.
	SubjSessionAuthenticated = "session.authenticated"

	// SubjSessionDisconnected is published by gateway on connection close.
	SubjSessionDisconnected = "session.disconnected"

	// SubjSessionOutput is the prefix for outbound per-session delivery.
	// Full subject: session.output.<sessionID>
	SubjSessionOutput = "session.output"

	// SubjRoomInput is the prefix for post-auth packets to the game service.
	// Full subject: room.input.<sessionID>
	SubjRoomInput = "room.input"
)
