package engine

import "github.com/momlesstomato/pixel-server/pkg/room/domain"

// State returns the current lifecycle phase.
func (inst *Instance) State() RoomState {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.state
}

// EntityCount returns the number of entities in the room.
func (inst *Instance) EntityCount() int {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return len(inst.entities)
}

// Entities returns a snapshot of all room entities.
func (inst *Instance) Entities() []domain.RoomEntity {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	out := make([]domain.RoomEntity, 0, len(inst.entities))
	for _, e := range inst.entities {
		out = append(out, *e)
	}
	return out
}

// Entity returns a snapshot of a single entity by virtual ID.
func (inst *Instance) Entity(virtualID int) (domain.RoomEntity, bool) {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	e, ok := inst.entities[virtualID]
	if !ok {
		return domain.RoomEntity{}, false
	}
	return *e, true
}

// Done returns a channel that closes when the room goroutine exits.
func (inst *Instance) Done() <-chan struct{} {
	return inst.done
}
