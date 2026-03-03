package plugin

// EventName identifies an in-process plugin event.
type EventName string

// Standard event names.
const (
	EventPlayerJoined EventName = "player.joined"
	EventPlayerLeft   EventName = "player.left"
	EventPlayerChat   EventName = "player.chat"
	EventPlayerWalk   EventName = "player.walk"

	EventEntitySpawned EventName = "entity.spawned"
	EventEntityRemoved EventName = "entity.removed"

	EventRoomLoaded   EventName = "room.loaded"
	EventRoomUnloaded EventName = "room.unloaded"
	EventRoomTick     EventName = "room.tick"

	EventItemPlaced   EventName = "item.placed"
	EventItemPickedUp EventName = "item.picked_up"

	EventPacketIn  EventName = "packet.in"
	EventPacketOut EventName = "packet.out"
)

// Event carries contextual metadata for plugin handlers.
type Event struct {
	// Name identifies the event kind.
	Name EventName

	// RoomID is the origin room identifier when applicable.
	RoomID int64

	// Tick is the room simulation tick at event emission time.
	Tick uint64

	// Entity identifies the event's primary entity when applicable.
	Entity EntityRef

	// Payload contains event-specific immutable data.
	Payload any

	// cancelled tracks whether processing was cancelled by a handler.
	cancelled bool
}

// Cancel marks the event as cancelled.
func (e *Event) Cancel() {
	e.cancelled = true
}

// IsCancelled returns true if any handler has cancelled this event.
func (e *Event) IsCancelled() bool {
	return e.cancelled
}

// EventHandler handles an in-process event.
type EventHandler func(e *Event)

// CancelFunc unregisters an event or packet hook.
type CancelFunc func()

// EventBus is the synchronous in-process dispatcher.
type EventBus interface {
	// Subscribe registers a handler for a named event.
	Subscribe(event EventName, handler EventHandler) CancelFunc

	// Publish executes handlers in subscription order on the caller goroutine.
	Publish(event *Event)
}
