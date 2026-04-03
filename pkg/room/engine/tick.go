package engine

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/pathfinding"
)

// processTick advances the room state by one tick cycle.
func (inst *Instance) processTick() {
	inst.processIdleCheck()
	inst.processEntityMovement()
	inst.processEntityIdle()
	inst.broadcastDirtyEntities()
}

// processIdleCheck increments or resets idle counter and triggers unload.
func (inst *Instance) processIdleCheck() {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	if len(inst.entities) == 0 {
		inst.idleTicks++
		inst.state = StateIdle
		if inst.idleTicks >= idleUnloadTicks {
			inst.state = StateStopped
		}
		return
	}
	inst.idleTicks = 0
	inst.state = StateActive
}

// processEntityMovement advances each walking entity one step along its path.
func (inst *Instance) processEntityMovement() {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	for _, entity := range inst.entities {
		if !entity.IsWalking || len(entity.Path) == 0 {
			if entity.IsWalking {
				entity.IsWalking = false
				delete(entity.Statuses, "mv")
				entity.UpdateNeeded = true
			}
			continue
		}
		next := entity.Path[0]
		entity.Path = entity.Path[1:]
		entity.Position = next
		entity.BodyRotation = calcRotation(entity.Position.X-next.X, entity.Position.Y-next.Y, next.X, next.Y, entity.Position.X, entity.Position.Y)
		entity.HeadRotation = entity.BodyRotation
		entity.Statuses["mv"] = fmt.Sprintf("%d,%d,%g", next.X, next.Y, next.Z)
		entity.UpdateNeeded = true
		entity.IsIdle = false
		entity.IdleTimer = 0
		if len(entity.Path) == 0 {
			entity.IsWalking = false
			entity.GoalPosition = nil
			delete(entity.Statuses, "mv")
		}
	}
}

// processEntityIdle increments idle timers for stationary entities.
func (inst *Instance) processEntityIdle() {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	for _, entity := range inst.entities {
		if entity.IsWalking {
			continue
		}
		entity.IdleTimer++
		if entity.IdleTimer >= 600 && !entity.IsIdle {
			entity.IsIdle = true
			entity.UpdateNeeded = true
		}
		if entity.CarryTimer > 0 {
			entity.CarryTimer--
			if entity.CarryTimer == 0 {
				delete(entity.Statuses, "sign")
				entity.CarryItem = 0
				entity.UpdateNeeded = true
			}
		}
	}
}

// broadcastDirtyEntities sends status updates for changed entities.
func (inst *Instance) broadcastDirtyEntities() {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	dirty := make([]domain.RoomEntity, 0)
	for _, entity := range inst.entities {
		if entity.UpdateNeeded {
			dirty = append(dirty, *entity)
			entity.UpdateNeeded = false
		}
	}
	if len(dirty) > 0 && inst.broadcaster != nil {
		inst.broadcaster(inst.RoomID, dirty, nil)
	}
}

// calcRotation computes the facing direction between two positions.
func calcRotation(_, _ int, toX, toY, fromX, fromY int) int {
	dx := toX - fromX
	dy := toY - fromY
	if dx == 0 && dy == -1 {
		return 0
	}
	if dx == 1 && dy == -1 {
		return 1
	}
	if dx == 1 && dy == 0 {
		return 2
	}
	if dx == 1 && dy == 1 {
		return 3
	}
	if dx == 0 && dy == 1 {
		return 4
	}
	if dx == -1 && dy == 1 {
		return 5
	}
	if dx == -1 && dy == 0 {
		return 6
	}
	if dx == -1 && dy == -1 {
		return 7
	}
	return 2
}

// startWalk computes a path and initiates entity movement.
func (inst *Instance) startWalk(entity *domain.RoomEntity, targetX, targetY int) error {
	grid := pathfinding.NewGrid(inst.Layout.Grid)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(grid, entity.Position.X, entity.Position.Y, targetX, targetY, opts)
	if path == nil {
		return domain.ErrPathBlocked
	}
	entity.Path = path
	entity.GoalPosition = &domain.Tile{X: targetX, Y: targetY}
	entity.IsWalking = true
	entity.IsIdle = false
	entity.IdleTimer = 0
	return nil
}
