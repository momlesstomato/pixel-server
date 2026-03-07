package plugin

// Registration represents a reversible runtime subscription handle.
type Registration interface {
	// Unsubscribe removes the registration and is safe to call multiple times.
	Unsubscribe()
}

// EventHandler processes an emitted event.
type EventHandler func(event *Event) error

// Event carries realm domain change notifications to plugins.
type Event struct {
	// Name uses the <realm>.<entity>.<action> naming convention.
	Name string
	// RoomID is set for room-scoped events; negative means not room-scoped.
	RoomID int64
	// SessionID is set when the event belongs to one connected session.
	SessionID string
	// Tick is set for ECS-originated events.
	Tick uint64
	// Data carries event-specific payload data.
	Data any

	// cancelled tracks event cancellation state.
	cancelled bool
}

// Cancel marks the event as cancelled.
func (e *Event) Cancel() {
	if e == nil {
		return
	}
	e.cancelled = true
}

// Cancelled reports whether the event has been cancelled.
func (e *Event) Cancelled() bool {
	if e == nil {
		return false
	}
	return e.cancelled
}
