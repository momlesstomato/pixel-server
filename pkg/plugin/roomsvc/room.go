// Package roomsvc provides the ECS-safe room facade exposed to plugins.
// Implementations route operations through the owning room goroutine.
package roomsvc

import (
	"fmt"

	"pixel-server/pkg/plugin/event"
)

// EntityKind describes a high-level ECS entity category.
type EntityKind string

const (
	EntityKindAvatar EntityKind = "avatar"
	EntityKindBot    EntityKind = "bot"
	EntityKindPet    EntityKind = "pet"
	EntityKindItem   EntityKind = "item"
)

// EntityRef identifies an entity without exposing mutable ECS internals.
type EntityRef struct {
	// ID is a room-scoped stable entity identifier.
	ID int64

	// Kind is the entity category.
	Kind EntityKind
}

// Snapshot is a read-only view of room simulation state.
type Snapshot struct {
	// RoomID is the unique room identifier.
	RoomID int64

	// Tick is the current simulation tick count.
	Tick uint64

	// Population is the number of active avatars in the room.
	Population int

	// Entities is a lightweight list of known entities.
	Entities []EntityRef
}

// Service is the ECS-safe facade exposed to plugins.
type Service interface {
	// Snapshot returns an immutable room snapshot.
	Snapshot(roomID int64) (Snapshot, error)

	// BroadcastPacket sends a packet to all sessions in the room.
	BroadcastPacket(roomID int64, headerID uint16, payload []byte) error

	// EmitEvent publishes a plugin or domain event into the in-process bus.
	EmitEvent(e *event.Event)
}

// NopService is a default implementation for non-game services.
type NopService struct{}

// Snapshot returns an error because room state is unavailable.
func (NopService) Snapshot(roomID int64) (Snapshot, error) {
	return Snapshot{}, fmt.Errorf("room %d: room snapshots are unavailable in this service", roomID)
}

// BroadcastPacket returns an error because room packet broadcasting is unavailable.
func (NopService) BroadcastPacket(roomID int64, _ uint16, _ []byte) error {
	return fmt.Errorf("room %d: packet broadcast is unavailable in this service", roomID)
}

// EmitEvent is a no-op for non-game services.
func (NopService) EmitEvent(_ *event.Event) {}
