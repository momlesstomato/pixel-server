package sdk

// SessionInfo provides read-only session data.
type SessionInfo struct {
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the authenticated user identifier.
	UserID int
	// MachineID stores the machine fingerprint.
	MachineID string
	// Encrypted indicates whether the connection uses RC4 encryption.
	Encrypted bool
	// InstanceID stores the server instance identifier.
	InstanceID string
}

// SessionAPI provides session query and control.
type SessionAPI interface {
	// FindByUserID returns session info for an online user.
	FindByUserID(userID int) (SessionInfo, bool)
	// FindByConnID returns session info for a connection.
	FindByConnID(connID string) (SessionInfo, bool)
	// Kick disconnects a session with a reason code.
	Kick(connID string, reason int32) error
	// Count returns the number of authenticated sessions.
	Count() int
}

// PacketAPI provides packet injection and custom handler registration.
type PacketAPI interface {
	// Send writes an encoded packet to a specific connection.
	Send(connID string, packetID uint16, body []byte) error
	// Broadcast sends a packet to all authenticated sessions.
	Broadcast(packetID uint16, body []byte) error
	// Handle registers a handler for a custom inbound packet ID.
	Handle(packetID uint16, handler PacketHandler) error
}

// PacketHandler processes an inbound packet from a connection.
type PacketHandler func(connID string, body []byte) error

// Logger provides structured logging for plugins.
type Logger interface {
	// Printf writes an informational log entry.
	Printf(format string, args ...any)
	// Errorf writes an error log entry.
	Errorf(format string, args ...any)
}
