package application

import (
	"context"

	sdkentity "github.com/momlesstomato/pixel-sdk/events/room/entity"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
)

// StartTyping sets the entity typing indicator in the room engine.
func (s *EntityService) StartTyping(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity) error {
	if s.fire != nil {
		ev := &sdkentity.EntityTyping{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, IsTyping: true}
		s.fire(ev)
		if ev.Cancelled() {
			return domain.ErrAccessDenied
		}
	}
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgTyping, Entity: entity, IntValue: 1, Reply: reply})
	if err := <-reply; err != nil {
		return err
	}
	if s.fire != nil {
		s.fire(&sdkentity.EntityTyped{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, IsTyping: true})
	}
	return nil
}

// StopTyping clears the entity typing indicator in the room engine.
func (s *EntityService) StopTyping(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity) error {
	if s.fire != nil {
		ev := &sdkentity.EntityTyping{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, IsTyping: false}
		s.fire(ev)
		if ev.Cancelled() {
			return domain.ErrAccessDenied
		}
	}
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgTyping, Entity: entity, IntValue: 0, Reply: reply})
	if err := <-reply; err != nil {
		return err
	}
	if s.fire != nil {
		s.fire(&sdkentity.EntityTyped{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, IsTyping: false})
	}
	return nil
}

// LookTo rotates the entity head toward a target coordinate.
func (s *EntityService) LookTo(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, x, y int) error {
	if s.fire != nil {
		ev := &sdkentity.EntityLooking{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, TargetX: x, TargetY: y}
		s.fire(ev)
		if ev.Cancelled() {
			return domain.ErrAccessDenied
		}
	}
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgLookTo, Entity: entity, TargetX: x, TargetY: y, Reply: reply})
	if err := <-reply; err != nil {
		return err
	}
	if s.fire != nil {
		s.fire(&sdkentity.EntityLooked{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, TargetX: x, TargetY: y})
	}
	return nil
}

// Sit toggles entity sit posture on or off.
func (s *EntityService) Sit(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity) error {
	if s.fire != nil {
		ev := &sdkentity.EntitySitting{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID}
		s.fire(ev)
		if ev.Cancelled() {
			return domain.ErrAccessDenied
		}
	}
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgSit, Entity: entity, Reply: reply})
	if err := <-reply; err != nil {
		return err
	}
	if s.fire != nil {
		s.fire(&sdkentity.EntitySat{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID})
	}
	return nil
}
