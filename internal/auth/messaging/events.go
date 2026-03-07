package messaging

// EventTicketValidated is emitted after one SSO ticket is validated.
const EventTicketValidated = "auth.ticket.validated"
const (
	// EventReleaseVersionReceived is emitted when release metadata is received.
	EventReleaseVersionReceived = "auth.handshake.release_version.received"
	// EventDiffieInitialized is emitted when init_diffie succeeds.
	EventDiffieInitialized = "auth.handshake.diffie.initialized"
	// EventDiffieCompleted is emitted when complete_diffie succeeds.
	EventDiffieCompleted = "auth.handshake.diffie.completed"
	// EventMachineIDReceived is emitted when machine-id packet is processed.
	EventMachineIDReceived = "auth.handshake.machine_id.received"
)

// TicketValidatedEventData carries payload for EventTicketValidated.
type TicketValidatedEventData struct {
	// UserID is the authenticated user identifier.
	UserID int32
}

// ReleaseVersionEventData carries payload for EventReleaseVersionReceived.
type ReleaseVersionEventData struct {
	// ReleaseVersion is the reported client release value.
	ReleaseVersion string
	// ClientType is the reported client type value.
	ClientType string
	// Platform is the reported platform identifier.
	Platform int32
	// DeviceCategory is the reported device category identifier.
	DeviceCategory int32
}

// DiffieInitializedEventData carries payload for EventDiffieInitialized.
type DiffieInitializedEventData struct {
	// SessionID identifies the handshake session.
	SessionID string
}

// DiffieCompletedEventData carries payload for EventDiffieCompleted.
type DiffieCompletedEventData struct {
	// SessionID identifies the handshake session.
	SessionID string
}

// MachineIDEventData carries payload for EventMachineIDReceived.
type MachineIDEventData struct {
	// MachineID is the normalized machine id value.
	MachineID string
	// Changed reports whether the incoming machine id required normalization.
	Changed bool
}
