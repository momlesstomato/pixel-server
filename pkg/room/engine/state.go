package engine

import (
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

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

// hasSeatPosture reports whether the entity is currently in a sit or lay posture.
func hasSeatPosture(entity *domain.RoomEntity) bool {
	_, hasSit := entity.Statuses["sit"]
	_, hasLay := entity.Statuses["lay"]
	return entity.IsSitting || hasSit || hasLay
}

// clearSeatPosture removes sit and lay posture state from an entity.
func clearSeatPosture(entity *domain.RoomEntity) {
	entity.IsSitting = false
	entity.IsSittingAuto = false
	delete(entity.Statuses, "sit")
	delete(entity.Statuses, "lay")
}

// EjectSittingEntitiesAt clears the sit/lay state of every seated entity at
// tile (x, y) and leaves them standing in place so they can freely navigate afterward.
// It returns updated entity snapshots so callers can broadcast the change.
func (inst *Instance) EjectSittingEntitiesAt(x, y int) []domain.RoomEntity {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	var updated []domain.RoomEntity
	for _, entity := range inst.entities {
		if entity.Position.X != x || entity.Position.Y != y || entity.IsWalking || !hasSeatPosture(entity) {
			continue
		}
		clearSeatPosture(entity)
		if y >= 0 && y < len(inst.Layout.Grid) && x >= 0 && x < len(inst.Layout.Grid[y]) {
			entity.Position.Z = inst.Layout.Grid[y][x].Z
		}
		entity.IsIdle = false
		entity.IdleTimer = 0
		entity.UpdateNeeded = true
		updated = append(updated, *entity)
	}
	return updated
}

// RotateSittingEntitiesAt updates the body and head rotation of every seated entity
// at tile (x, y) to match the new furniture direction.
// It returns the updated entity snapshots so callers can broadcast the change.
func (inst *Instance) RotateSittingEntitiesAt(x, y, dir int) []domain.RoomEntity {
	if dir%2 != 0 {
		dir--
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	var updated []domain.RoomEntity
	for _, e := range inst.entities {
		if e.Position.X == x && e.Position.Y == y && !e.IsWalking && hasSeatPosture(e) {
			e.BodyRotation = dir
			e.HeadRotation = dir
			e.UpdateNeeded = true
			updated = append(updated, *e)
		}
	}
	return updated
}
