package messaging

const (
	// EventSessionConnected is emitted when a websocket session is connected.
	EventSessionConnected = "sessionconnection.session.connected"
	// EventSessionDisconnected is emitted when a websocket session is disconnected.
	EventSessionDisconnected = "sessionconnection.session.disconnected"
	// EventSessionAuthenticated is emitted when auth marks a session as authenticated.
	EventSessionAuthenticated = "sessionconnection.session.authenticated"
	// EventClientPongReceived is emitted when one client pong packet is processed.
	EventClientPongReceived = "sessionconnection.client.pong.received"
	// EventLatencyTestReceived is emitted when one latency test packet is processed.
	EventLatencyTestReceived = "sessionconnection.client.latency_test.received"
	// EventDesktopViewReceived is emitted when one desktop-view packet is processed.
	EventDesktopViewReceived = "sessionconnection.session.desktop_view.received"
	// EventPacketReceived is emitted when one session-connection packet is decoded.
	EventPacketReceived = "sessionconnection.packet.received"
)

// SessionConnectedEventData carries connected event payload.
type SessionConnectedEventData struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
}

// SessionDisconnectedEventData carries disconnected event payload.
type SessionDisconnectedEventData struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
}

// SessionAuthenticatedEventData carries authenticated event payload.
type SessionAuthenticatedEventData struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
	// UserID is the authenticated user identifier.
	UserID int32
}

// LatencyTestEventData carries latency test payload.
type LatencyTestEventData struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
	// RequestID is the latency probe request id.
	RequestID int32
}

// ClientPongEventData carries client pong payload.
type ClientPongEventData struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
}

// DesktopViewEventData carries desktop-view payload.
type DesktopViewEventData struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
}

// PacketReceivedEventData carries packet receive event payload.
type PacketReceivedEventData struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
	// Header is the decoded packet header.
	Header uint16
	// PacketName is the decoded packet canonical name.
	PacketName string
}
