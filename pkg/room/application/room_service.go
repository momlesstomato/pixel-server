package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkroom "github.com/momlesstomato/pixel-sdk/events/room"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/momlesstomato/pixel-server/pkg/room/heightmap"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Service defines room application use-cases.
type Service struct {
	// rooms stores optional room data persistence.
	rooms domain.RoomRepository
	// models stores room model persistence.
	models domain.ModelRepository
	// bans stores room ban persistence.
	bans domain.BanRepository
	// rights stores room rights persistence.
	rights domain.RightsRepository
	// manager stores room instance registry.
	manager *engine.Manager
	// logger stores structured logger.
	logger *zap.Logger
	// fire stores optional plugin event dispatch.
	fire func(sdk.Event)
}

// NewService creates one room application service.
func NewService(models domain.ModelRepository, bans domain.BanRepository, rights domain.RightsRepository, manager *engine.Manager, logger *zap.Logger) (*Service, error) {
	if models == nil {
		return nil, fmt.Errorf("model repository is required")
	}
	if bans == nil {
		return nil, fmt.Errorf("ban repository is required")
	}
	if rights == nil {
		return nil, fmt.Errorf("rights repository is required")
	}
	if manager == nil {
		return nil, fmt.Errorf("room manager is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{models: models, bans: bans, rights: rights, manager: manager, logger: logger}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (s *Service) SetEventFirer(fire func(sdk.Event)) {
	s.fire = fire
}

// LoadRoom loads or retrieves a room instance by room metadata.
func (s *Service) LoadRoom(ctx context.Context, room domain.Room) (*engine.Instance, error) {
	if inst, ok := s.manager.Get(room.ID); ok {
		return inst, nil
	}
	if s.fire != nil {
		ev := &sdkroom.RoomLoading{RoomID: room.ID}
		s.fire(ev)
		if ev.Cancelled() {
			return nil, domain.ErrAccessDenied
		}
	}
	model, err := s.models.FindModelBySlug(ctx, room.ModelSlug)
	if err != nil {
		return nil, domain.ErrRoomModelNotFound
	}
	grid, err := heightmap.Parse(model.Heightmap)
	if err != nil {
		return nil, domain.ErrInvalidHeightmap
	}
	layout := domain.Layout{
		Slug: model.Slug, DoorX: model.DoorX, DoorY: model.DoorY,
		DoorZ: model.DoorZ, DoorDir: model.DoorDir,
		WallHeight: model.WallHeight, Grid: grid,
	}
	inst := s.manager.Load(room.ID, layout)
	if s.fire != nil {
		s.fire(&sdkroom.RoomLoaded{RoomID: room.ID})
	}
	return inst, nil
}

// EnterRoom places a player entity into a room instance.
func (s *Service) EnterRoom(ctx context.Context, inst *engine.Instance, entity *domain.RoomEntity, roomID int, userID int) error {
	if s.fire != nil {
		ev := &sdkroom.RoomEntering{RoomID: roomID, UserID: userID}
		s.fire(ev)
		if ev.Cancelled() {
			return domain.ErrAccessDenied
		}
	}
	reply := make(chan error, 1)
	if !inst.Send(engine.Message{Type: engine.MsgEnter, Entity: entity, Reply: reply}) {
		return domain.ErrRoomFull
	}
	if err := <-reply; err != nil {
		return err
	}
	if s.fire != nil {
		s.fire(&sdkroom.RoomEntered{RoomID: roomID, UserID: userID, VirtualID: entity.VirtualID})
	}
	return nil
}

// LeaveRoom removes a player entity from a room instance.
func (s *Service) LeaveRoom(_ context.Context, inst *engine.Instance, entity *domain.RoomEntity, roomID int, userID int) error {
	if s.fire != nil {
		ev := &sdkroom.RoomLeaving{RoomID: roomID, UserID: userID}
		s.fire(ev)
		if ev.Cancelled() {
			return domain.ErrAccessDenied
		}
	}
	reply := make(chan error, 1)
	if !inst.Send(engine.Message{Type: engine.MsgLeave, Entity: entity, Reply: reply}) {
		return domain.ErrEntityNotFound
	}
	if err := <-reply; err != nil {
		return err
	}
	if s.fire != nil {
		s.fire(&sdkroom.RoomLeft{RoomID: roomID, UserID: userID})
	}
	return nil
}

// CheckBan reports whether a user is banned from a room.
func (s *Service) CheckBan(ctx context.Context, roomID, userID int) bool {
	ban, err := s.bans.FindActiveBan(ctx, roomID, userID)
	if err != nil {
		return false
	}
	return ban != nil
}

// HasRights reports whether a user has room rights.
func (s *Service) HasRights(ctx context.Context, roomID, userID int) bool {
	has, _ := s.rights.HasRights(ctx, roomID, userID)
	return has
}

// Manager returns the underlying room instance manager.
func (s *Service) Manager() *engine.Manager { return s.manager }

// SetRoomRepository configures the optional room data persistence layer.
func (s *Service) SetRoomRepository(repo domain.RoomRepository) {
	s.rooms = repo
}

// FindRoom resolves full room data by identifier.
func (s *Service) FindRoom(ctx context.Context, roomID int) (domain.Room, error) {
	if s.rooms == nil {
		return domain.Room{}, domain.ErrRoomNotFound
	}
	return s.rooms.FindByID(ctx, roomID)
}

// CheckAccess validates whether a requester may enter a room.
func (s *Service) CheckAccess(ctx context.Context, room domain.Room, password string, requesterID int) error {
	if room.OwnerID == requesterID {
		return nil
	}
	if room.State == domain.AccessLocked && s.HasRights(ctx, room.ID, requesterID) {
		return nil
	}
	switch room.State {
	case domain.AccessPassword:
		if err := bcrypt.CompareHashAndPassword([]byte(room.Password), []byte(password)); err != nil {
			return domain.ErrInvalidPassword
		}
	case domain.AccessLocked:
		return domain.ErrAccessDenied
	}
	return nil
}

// SaveSettings persists updated room settings after ownership validation.
func (s *Service) SaveSettings(ctx context.Context, roomID, ownerID int, updated domain.Room) error {
	if s.rooms == nil {
		return domain.ErrRoomNotFound
	}
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room.OwnerID != ownerID {
		return domain.ErrAccessDenied
	}
	updated.ID = roomID
	return s.rooms.SaveSettings(ctx, updated)
}
