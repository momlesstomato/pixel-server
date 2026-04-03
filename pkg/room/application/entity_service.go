package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkentity "github.com/momlesstomato/pixel-sdk/events/room/entity"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"go.uber.org/zap"
)

// EntityService manages room entity state mutations.
type EntityService struct {
	// manager stores the room instance registry.
	manager *engine.Manager
	// logger stores structured logging behavior.
	logger *zap.Logger
	// fire stores optional plugin event dispatch.
	fire func(sdk.Event)
}

// NewEntityService creates one room entity service.
func NewEntityService(manager *engine.Manager, logger *zap.Logger) (*EntityService, error) {
	if manager == nil {
		return nil, fmt.Errorf("manager is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EntityService{manager: manager, logger: logger}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (s *EntityService) SetEventFirer(fire func(sdk.Event)) {
	s.fire = fire
}

// Walk requests a walk path for an entity toward target coordinates.
func (s *EntityService) Walk(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, toX, toY int) error {
	if s.fire != nil {
		ev := &sdkentity.EntityMoving{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, ToX: toX, ToY: toY}
		s.fire(ev)
		if ev.Cancelled() {
			return domain.ErrAccessDenied
		}
	}
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: toX, TargetY: toY, Reply: reply})
	if err := <-reply; err != nil {
		return err
	}
	if s.fire != nil {
		s.fire(&sdkentity.EntityMoved{RoomID: inst.RoomID, UserID: entity.UserID, VirtualID: entity.VirtualID, ToX: toX, ToY: toY})
	}
	return nil
}

// Dance sends a dance state change into the room engine.
func (s *EntityService) Dance(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, danceID int) error {
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgDance, Entity: entity, IntValue: danceID, Reply: reply})
	return <-reply
}

// Action sends a generic user action into the room engine.
func (s *EntityService) Action(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, actionID int) error {
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgAction, Entity: entity, IntValue: actionID, Reply: reply})
	return <-reply
}

// Sign sends a sign display request into the room engine.
func (s *EntityService) Sign(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, signID int) error {
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgSign, Entity: entity, IntValue: signID, Reply: reply})
	return <-reply
}

// StartTyping sets the entity typing indicator in the room engine.
func (s *EntityService) StartTyping(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity) error {
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgTyping, Entity: entity, IntValue: 1, Reply: reply})
	return <-reply
}

// StopTyping clears the entity typing indicator in the room engine.
func (s *EntityService) StopTyping(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity) error {
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgTyping, Entity: entity, IntValue: 0, Reply: reply})
	return <-reply
}

// LookTo rotates the entity head toward a target coordinate.
func (s *EntityService) LookTo(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, x, y int) error {
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgLookTo, Entity: entity, TargetX: x, TargetY: y, Reply: reply})
	return <-reply
}
