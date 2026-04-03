package engine

import (
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/pathfinding"
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

// EjectSittingEntitiesAt clears the sit/lay state of every auto-sitting entity at
// tile (x, y) and initiates a walk toward the room door.
// It returns updated entity snapshots so callers can broadcast the change.
func (inst *Instance) EjectSittingEntitiesAt(x, y int) []domain.RoomEntity {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	var updated []domain.RoomEntity
	for _, entity := range inst.entities {
		if !entity.IsSittingAuto || entity.Position.X != x || entity.Position.Y != y {
			continue
		}
		entity.IsSitting = false
		entity.IsSittingAuto = false
		delete(entity.Statuses, "sit")
		delete(entity.Statuses, "lay")
		grid := pathfinding.NewGrid(inst.Layout.Grid)
		path := pathfinding.FindPath(grid, entity.Position.X, entity.Position.Y, inst.Layout.DoorX, inst.Layout.DoorY, pathfinding.DefaultOptions())
		if path != nil {
			entity.Path = path
			entity.GoalPosition = &domain.Tile{X: inst.Layout.DoorX, Y: inst.Layout.DoorY}
			entity.IsWalking = true
		}
		entity.IsIdle = false
		entity.IdleTimer = 0
		entity.UpdateNeeded = true
		updated = append(updated, *entity)
	}
	return updated
}

// RotateSittingEntitiesAt updates the body and head rotation of every entity
// that is auto-sitting at tile (x, y) to match the new furniture direction.
// It returns the updated entity snapshots so callers can broadcast the change.
func (inst *Instance) RotateSittingEntitiesAt(x, y, dir int) []domain.RoomEntity {
	if dir%2 != 0 {
		dir--
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	var updated []domain.RoomEntity
	for _, e := range inst.entities {
		if e.IsSittingAuto && e.Position.X == x && e.Position.Y == y {
			e.BodyRotation = dir
			e.HeadRotation = dir
			e.UpdateNeeded = true
			updated = append(updated, *e)
		}
	}
	return updated
}
