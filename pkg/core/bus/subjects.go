// Package bus provides a thin wrapper around NATS JetStream and defines
// infrastructure-level NATS subjects shared across all services.
// Domain-specific subjects are owned by their respective feature packages.
package bus

// Infrastructure-level NATS subjects for session lifecycle.
const (
	SubjHandshakeC2S         = "session.handshake.c2s"
	SubjSessionAuthenticated = "session.authenticated"
	SubjSessionDisconnected  = "session.disconnected"
	SubjSessionOutput        = "session.output"
)
