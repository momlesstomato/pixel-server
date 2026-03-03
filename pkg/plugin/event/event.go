// Package event provides the in-process synchronous event bus used by
// the plugin system. Events are dispatched on the caller goroutine.
package event

// Name identifies an in-process plugin event.
type Name string

// Standard event names.
const (
	PlayerJoined  Name = "player.joined"
	PlayerLeft    Name = "player.left"
	PlayerChat    Name = "player.chat"
	PlayerWalk    Name = "player.walk"
	EntitySpawned Name = "entity.spawned"
	EntityRemoved Name = "entity.removed"
	RoomLoaded    Name = "room.loaded"
	RoomUnloaded  Name = "room.unloaded"
	RoomTick      Name = "room.tick"
	ItemPlaced    Name = "item.placed"
	ItemPickedUp  Name = "item.picked_up"
	PacketIn      Name = "packet.in"
	PacketOut     Name = "packet.out"
)

// Event carries contextual metadata for plugin handlers.
type Event struct {
	// Name identifies the event kind.
	Name Name

	// RoomID is the origin room identifier when applicable.
	RoomID int64

	// Tick is the room simulation tick at event emission time.
	Tick uint64

	// EntityID is the room-scoped stable entity identifier when applicable.
	EntityID int64

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

// Handler handles an in-process event.
type Handler func(e *Event)

// CancelFunc unregisters an event or packet hook.
type CancelFunc func()

// Bus is the synchronous in-process event dispatcher.
type Bus interface {
	// Subscribe registers a handler for a named event.
	Subscribe(event Name, handler Handler) CancelFunc

	// Publish executes handlers in subscription order on the caller goroutine.
	Publish(event *Event)
}
