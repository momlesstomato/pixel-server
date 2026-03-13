package sdk

// Event is the base contract for all dispatchable events.
type Event interface {
	event()
}

// Cancellable extends Event with cancellation support.
type Cancellable interface {
	Event
	// Cancelled reports whether the event has been cancelled.
	Cancelled() bool
	// Cancel marks the event as cancelled.
	Cancel()
}

// BaseEvent is embedded by all non-cancellable concrete event types.
type BaseEvent struct{}

func (BaseEvent) event() {}

// BaseCancellable is embedded by cancellable concrete event types.
type BaseCancellable struct {
	cancelled bool
}

func (BaseCancellable) event() {}

// Cancelled reports whether the event has been cancelled.
func (e *BaseCancellable) Cancelled() bool { return e.cancelled }

// Cancel marks the event as cancelled.
func (e *BaseCancellable) Cancel() { e.cancelled = true }

// ConnectionOpened fires when a WebSocket connection is established.
type ConnectionOpened struct {
	BaseEvent
	// ConnID stores the connection identifier.
	ConnID string
}

// ConnectionClosed fires after a WebSocket connection is fully closed.
type ConnectionClosed struct {
	BaseEvent
	// ConnID stores the connection identifier.
	ConnID string
	// Reason stores the disconnect reason code.
	Reason int32
}

// AuthValidating fires after SSO ticket is validated but before authentication.ok is sent.
type AuthValidating struct {
	BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the authenticated user identifier.
	UserID int
	// Ticket stores the raw SSO ticket value.
	Ticket string
}

// AuthCompleted fires after authentication.ok is sent.
type AuthCompleted struct {
	BaseEvent
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the authenticated user identifier.
	UserID int
}

// DuplicateKick fires before an existing session is kicked due to duplicate login.
type DuplicateKick struct {
	BaseCancellable
	// OldConnID stores the existing connection identifier.
	OldConnID string
	// NewConnID stores the new connection identifier.
	NewConnID string
	// UserID stores the user identifier.
	UserID int
}

// SessionDisconnecting fires before a graceful disconnect is processed.
type SessionDisconnecting struct {
	BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the user identifier.
	UserID int
	// Reason stores the disconnect reason code.
	Reason int32
}

// PongTimeout fires after a heartbeat timeout is detected.
type PongTimeout struct {
	BaseEvent
	// ConnID stores the connection identifier.
	ConnID string
	// UserID stores the user identifier.
	UserID int
}

// HotelStatusChanged fires when the hotel state machine transitions.
type HotelStatusChanged struct {
	BaseEvent
	// OldState stores the previous hotel state string.
	OldState string
	// NewState stores the new hotel state string.
	NewState string
}

// PacketReceived fires before an inbound packet is dispatched to its handler.
type PacketReceived struct {
	BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// PacketID stores the protocol packet identifier.
	PacketID uint16
	// Body stores the raw packet payload bytes.
	Body []byte
}

// PacketSending fires before an outbound packet is written to the socket.
type PacketSending struct {
	BaseCancellable
	// ConnID stores the connection identifier.
	ConnID string
	// PacketID stores the protocol packet identifier.
	PacketID uint16
	// Body stores the raw packet payload bytes.
	Body []byte
}
