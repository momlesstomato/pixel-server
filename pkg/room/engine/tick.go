package engine

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/pathfinding"
)

// processTick advances the room state by one tick cycle.
func (inst *Instance) processTick() {
	inst.processIdleCheck()
	doorExits := inst.processEntityMovement()
	newlySlept, kicked := inst.processEntityIdle()
	inst.broadcastDirtyEntities()
	inst.removeDoorExits(doorExits)
	for _, e := range newlySlept {
		if inst.sleepNotifier != nil {
			inst.sleepNotifier(inst.RoomID, e.VirtualID, true)
		}
	}
	for _, e := range kicked {
		if inst.kickNotifier != nil {
			inst.kickNotifier(inst.RoomID, e)
		}
	}
	for _, e := range doorExits {
		if inst.doorExitNotifier != nil {
			inst.doorExitNotifier(inst.RoomID, e)
		}
	}
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

// processEntityMovement advances each walking entity one step along its path
// and returns entities that reached the door tile this tick.
func (inst *Instance) processEntityMovement() (doorExits []domain.RoomEntity) {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	for _, entity := range inst.entities {
		if !entity.IsWalking && entity.StepFrom != nil {
			entity.StepFrom = nil
			delete(entity.Statuses, "mv")
			if inst.seatChecker != nil && !entity.IsSitting {
				if h, seatDir, isSit, _ := inst.seatChecker(inst.RoomID, entity.Position.X, entity.Position.Y); isSit {
					if seatDir%2 != 0 {
						seatDir--
					}
					entity.BodyRotation = seatDir
					entity.HeadRotation = seatDir
					entity.Statuses["sit"] = fmt.Sprintf("%.2f", h)
					entity.IsSitting = true
					entity.IsSittingAuto = true
				}
			}
			entity.UpdateNeeded = true
			continue
		}
		if !entity.IsWalking || len(entity.Path) == 0 {
			continue
		}
		next := entity.Path[0]
		entity.Path = entity.Path[1:]
		prevPos := entity.Position
		dir := calcRotation(0, 0, next.X, next.Y, prevPos.X, prevPos.Y)
		entity.Position = next
		entity.StepFrom = &prevPos
		entity.BodyRotation = dir
		entity.HeadRotation = dir
		entity.Statuses["mv"] = fmt.Sprintf("%d,%d,%g", next.X, next.Y, next.Z)
		entity.UpdateNeeded = true
		entity.IsIdle = false
		entity.IdleTimer = 0
		if len(entity.Path) == 0 {
			entity.IsWalking = false
			entity.GoalPosition = nil
			if entity.Position.X == inst.Layout.DoorX && entity.Position.Y == inst.Layout.DoorY {
				doorExits = append(doorExits, *entity)
			}
		}
	}
	return doorExits
}

// processEntityIdle increments idle timers for stationary entities and returns
// entities that just fell asleep and entities kicked due to idle timeout.
func (inst *Instance) processEntityIdle() (newlySlept []domain.RoomEntity, kicked []domain.RoomEntity) {
	inst.mu.Lock()
	defer inst.mu.Unlock()
	for _, entity := range inst.entities {
		if entity.IsWalking {
			continue
		}
		entity.IdleTimer++
		if entity.IdleTimer == idleSleepTicks && !entity.IsIdle {
			entity.IsIdle = true
			entity.UpdateNeeded = true
			newlySlept = append(newlySlept, *entity)
		}
		if entity.IdleTimer >= idleKickTicks {
			kicked = append(kicked, *entity)
			delete(inst.entities, entity.VirtualID)
			continue
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
	return newlySlept, kicked
}

// removeDoorExits removes door-exited entities from the instance after their final update is broadcast.
func (inst *Instance) removeDoorExits(exits []domain.RoomEntity) {
	if len(exits) == 0 {
		return
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	for _, e := range exits {
		delete(inst.entities, e.VirtualID)
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

// calcRotation computes the facing direction from (fromX, fromY) toward (toX, toY).
// Direction mapping follows the Habbo isometric convention (0=N, 2=E, 4=S, 6=W).
// Sign-based comparison is used so the function works for any tile distance.
func calcRotation(_, _ int, toX, toY, fromX, fromY int) int {
	dx := toX - fromX
	dy := toY - fromY
	switch {
	case dx < 0 && dy < 0:
		return 7
	case dx > 0 && dy < 0:
		return 1
	case dx > 0 && dy > 0:
		return 3
	case dx < 0 && dy > 0:
		return 5
	case dx < 0:
		return 6
	case dx > 0:
		return 2
	case dy > 0:
		return 4
	default:
		return 0
	}
}

// startWalk computes a path and initiates entity movement.
func (inst *Instance) startWalk(entity *domain.RoomEntity, targetX, targetY int) error {
	grid := pathfinding.NewGrid(inst.Layout.Grid)
	opts := pathfinding.DefaultOptions()
	path := pathfinding.FindPath(grid, entity.Position.X, entity.Position.Y, targetX, targetY, opts)
	if path == nil {
		return domain.ErrPathBlocked
	}
	if entity.IsSitting {
		if !entity.IsSittingAuto {
			entity.Position.Z += 0.35
		}
		entity.IsSitting = false
		entity.IsSittingAuto = false
		delete(entity.Statuses, "sit")
	}
	delete(entity.Statuses, "lay")
	entity.Path = path
	entity.GoalPosition = &domain.Tile{X: targetX, Y: targetY}
	entity.IsWalking = true
	entity.IsIdle = false
	entity.IdleTimer = 0
	return nil
}
