package engine

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// handleMessage dispatches a room message to the correct handler.
func (inst *Instance) handleMessage(msg Message) {
	var err error
	switch msg.Type {
	case MsgEnter:
		err = inst.handleEnter(msg)
	case MsgLeave:
		err = inst.handleLeave(msg)
	case MsgWalk:
		err = inst.handleWalk(msg)
	case MsgAction:
		err = inst.handleAction(msg)
	case MsgDance:
		err = inst.handleDance(msg)
	case MsgSign:
		err = inst.handleSign(msg)
	case MsgTyping:
		err = inst.handleTyping(msg)
	case MsgLookTo:
		err = inst.handleLookTo(msg)
	case MsgSit:
		err = inst.handleSit(msg)
	case MsgStop:
		inst.handleStop()
		return
	default:
		return
	}
	if msg.Reply != nil {
		msg.Reply <- err
	}
}

// handleEnter adds a new entity to the room.
func (inst *Instance) handleEnter(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	entity := *msg.Entity
	entity.VirtualID = inst.nextVID
	inst.nextVID++
	entity.Position = domain.Tile{
		X: inst.Layout.DoorX, Y: inst.Layout.DoorY,
		Z: inst.Layout.DoorZ, State: domain.TileOpen,
	}
	entity.BodyRotation = inst.Layout.DoorDir
	entity.HeadRotation = inst.Layout.DoorDir
	entity.Statuses = make(map[string]string)
	entity.CanWalk = true
	entity.UpdateNeeded = true
	inst.entities[entity.VirtualID] = &entity
	inst.idleTicks = 0
	inst.state = StateActive
	*msg.Entity = entity
	return nil
}

// handleLeave removes an entity from the room by virtual ID.
func (inst *Instance) handleLeave(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	if _, ok := inst.entities[msg.Entity.VirtualID]; !ok {
		return domain.ErrEntityNotFound
	}
	delete(inst.entities, msg.Entity.VirtualID)
	return nil
}

// handleWalk initiates pathfinding for the requesting entity.
func (inst *Instance) handleWalk(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	entity, ok := inst.entities[msg.Entity.VirtualID]
	inst.mu.Unlock()
	if !ok {
		return domain.ErrEntityNotFound
	}
	if !entity.CanWalk {
		return domain.ErrAccessDenied
	}
	return inst.startWalk(entity, msg.TargetX, msg.TargetY)
}

// handleAction marks entity as needing an update for expression broadcast.
func (inst *Instance) handleAction(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	_, ok := inst.entities[msg.Entity.VirtualID]
	if !ok {
		return domain.ErrEntityNotFound
	}
	return nil
}

// handleDance updates entity dance animation style.
func (inst *Instance) handleDance(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	entity, ok := inst.entities[msg.Entity.VirtualID]
	if !ok {
		return domain.ErrEntityNotFound
	}
	entity.DanceID = msg.IntValue
	if entity.DanceID > 0 {
		entity.Statuses["dance"] = fmt.Sprintf("%d", entity.DanceID)
	} else {
		delete(entity.Statuses, "dance")
	}
	resetEntityIdle(entity)
	entity.UpdateNeeded = true
	return nil
}

// handleSign updates entity sign display with a timed carry.
func (inst *Instance) handleSign(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	entity, ok := inst.entities[msg.Entity.VirtualID]
	if !ok {
		return domain.ErrEntityNotFound
	}
	entity.CarryItem = msg.IntValue
	entity.CarryTimer = 5
	entity.Statuses["sign"] = fmt.Sprintf("%d", msg.IntValue)
	resetEntityIdle(entity)
	entity.UpdateNeeded = true
	return nil
}

// handleTyping sets or clears entity typing status indicator.
func (inst *Instance) handleTyping(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	entity, ok := inst.entities[msg.Entity.VirtualID]
	if !ok {
		return domain.ErrEntityNotFound
	}
	if msg.IntValue == 1 {
		entity.Statuses["trd"] = ""
	} else {
		delete(entity.Statuses, "trd")
	}
	resetEntityIdle(entity)
	entity.UpdateNeeded = true
	return nil
}

// handleLookTo rotates entity head toward a target coordinate.
func (inst *Instance) handleLookTo(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	entity, ok := inst.entities[msg.Entity.VirtualID]
	if !ok {
		return domain.ErrEntityNotFound
	}
	entity.HeadRotation = calcRotation(0, 0, msg.TargetX, msg.TargetY, entity.Position.X, entity.Position.Y)
	resetEntityIdle(entity)
	entity.UpdateNeeded = true
	return nil
}

// handleStop triggers graceful room shutdown.
func (inst *Instance) handleStop() {
	inst.mu.Lock()
	inst.state = StateStopped
	inst.mu.Unlock()
	if inst.cancel != nil {
		inst.cancel()
	}
}

// handleSit toggles the entity sit posture on or off.
func (inst *Instance) handleSit(msg Message) error {
	if msg.Entity == nil {
		return domain.ErrEntityNotFound
	}
	inst.mu.Lock()
	defer inst.mu.Unlock()
	entity, ok := inst.entities[msg.Entity.VirtualID]
	if !ok {
		return domain.ErrEntityNotFound
	}
	if entity.IsWalking {
		return nil
	}
	if !entity.IsSitting {
		if entity.BodyRotation%2 != 0 {
			entity.BodyRotation--
		}
		entity.Statuses["sit"] = "1.0"
		entity.Position.Z -= 0.35
		entity.IsSitting = true
		entity.IsSittingAuto = false
	} else {
		if !entity.IsSittingAuto {
			entity.Position.Z += 0.35
		}
		delete(entity.Statuses, "sit")
		delete(entity.Statuses, "lay")
		entity.IsSitting = false
		entity.IsSittingAuto = false
	}
	resetEntityIdle(entity)
	entity.UpdateNeeded = true
	return nil
}

// resetEntityIdle clears idle and sleep state for an entity that performed an action.
func resetEntityIdle(entity *domain.RoomEntity) {
	entity.IdleTimer = 0
	entity.IsIdle = false
}
