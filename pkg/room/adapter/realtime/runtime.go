package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	roomapplication "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by room realtime runtime.
type Transport interface {
	// Send writes one encoded packet to one connection identifier.
	Send(string, uint16, []byte) error
}

// UsernameResolver resolves a display name for one authenticated user identifier.
type UsernameResolver func(ctx context.Context, userID int) (string, error)

// ProfileResolver resolves the full display profile for one authenticated user identifier.
type ProfileResolver func(ctx context.Context, userID int) (username, look, motto, gender string, err error)

// VisitRecorder records one room visit.
type VisitRecorder interface {
	// RecordVisit persists a room visit entry.
	RecordVisit(ctx context.Context, userID int, roomID int) error
}

// Runtime defines room realm websocket packet behavior.
type Runtime struct {
	// service stores room application behavior.
	service *roomapplication.Service
	// entitySvc stores room entity mutation behavior.
	entitySvc *roomapplication.EntityService
	// chatSvc stores room chat behavior.
	chatSvc *roomapplication.ChatService
	// sessions stores authenticated connection lookup.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior for the local connection only.
	transport Transport
	// broadcaster publishes packet frames to distributed per-user channels.
	broadcaster broadcast.Broadcaster
	// logger stores runtime logging behavior.
	logger *zap.Logger
	// connRooms tracks which room each connection is in.
	connRooms map[string]int
	// pendingDoorbell tracks visitor connections waiting for doorbell approval.
	pendingDoorbell map[string]doorbellEntry
	// passwordAttempts tracks consecutive wrong password attempts per connection.
	passwordAttempts map[string]int
	// passwordCooldown tracks when the password cooldown expires per connection.
	passwordCooldown map[string]time.Time
	// usernameResolver resolves display names for user identifiers.
	usernameResolver UsernameResolver
	// profileResolver resolves full user profile for entity creation.
	profileResolver ProfileResolver
	// floorItemSender sends the room floor item list to one arriving connection.
	floorItemSender func(ctx context.Context, connID string, roomID int) error
	// voteRepo stores optional vote persistence for room scoring.
	voteRepo domain.VoteRepository
	// visitRecorder stores optional room visit tracking behavior.
	visitRecorder VisitRecorder
	// permissions stores optional permission check behavior.
	permissions PermissionChecker
}

// PermissionChecker defines permission resolution behavior for room actions.
type PermissionChecker interface {
	// HasPermission checks if a user holds a specific permission scope.
	HasPermission(ctx context.Context, userID int, scope string) (bool, error)
}

// doorbellEntry tracks a visitor waiting for doorbell approval.
type doorbellEntry struct {
	// connID stores the visitor connection identifier.
	connID string
	// roomID stores the target room identifier.
	roomID int
}

// NewRuntime creates one room realtime runtime instance.
func NewRuntime(service *roomapplication.Service, entitySvc *roomapplication.EntityService, chatSvc *roomapplication.ChatService, sessions coreconnection.SessionRegistry, transport Transport, broadcaster broadcast.Broadcaster, logger *zap.Logger) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("room service is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	if broadcaster == nil {
		return nil, fmt.Errorf("broadcaster is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Runtime{
		service: service, entitySvc: entitySvc, chatSvc: chatSvc,
		sessions: sessions, transport: transport, broadcaster: broadcaster,
		logger: logger, connRooms: make(map[string]int),
		pendingDoorbell:  make(map[string]doorbellEntry),
		passwordAttempts: make(map[string]int),
		passwordCooldown: make(map[string]time.Time),
	}, nil
}

// userID resolves authenticated user identifier for one connection.
func (rt *Runtime) userID(connID string) (int, bool) {
	session, found := rt.sessions.FindByConnID(connID)
	if !found || session.UserID <= 0 {
		return 0, false
	}
	return session.UserID, true
}

// sendPacket encodes and transmits one outgoing packet.
func (rt *Runtime) sendPacket(connID string, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return rt.transport.Send(connID, pkt.PacketID(), body)
}

// findEntityByConnID returns the room instance and entity for a connection.
func (rt *Runtime) findEntityByConnID(connID string, userID int) (*engine.Instance, *domain.RoomEntity) {
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil, nil
	}
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return nil, nil
	}
	entities := inst.Entities()
	for i := range entities {
		if entities[i].UserID == userID {
			return inst, &entities[i]
		}
	}
	return nil, nil
}

// leaveCurrentRoom removes all entities for connID from its current room and broadcasts each removal.
// It is safe to call when the connection is not in any room.
func (rt *Runtime) leaveCurrentRoom(connID string) {
	userID, ok := rt.userID(connID)
	if ok {
		for {
			inst, entity := rt.findEntityByConnID(connID, userID)
			if inst == nil || entity == nil {
				break
			}
			reply := make(chan error, 1)
			if !inst.Send(engine.Message{Type: engine.MsgLeave, Entity: entity, Reply: reply}) {
				break
			}
			<-reply
			body, encErr := packet.UserRemoveComposer{VirtualID: int32(entity.VirtualID)}.Encode()
			if encErr == nil {
				frame := codec.EncodeFrame(packet.UserRemoveComposerID, body)
				ctx := context.Background()
				for _, e := range inst.Entities() {
					if e.Type == domain.EntityPlayer && e.UserID != 0 {
						_ = rt.broadcaster.Publish(ctx, sessionnotification.UserChannel(e.UserID), frame)
					}
				}
			}
		}
	}
	delete(rt.connRooms, connID)
}

// Dispose releases per-connection resources and removes the entity from its room.
func (rt *Runtime) Dispose(connID string) {
	for username, entry := range rt.pendingDoorbell {
		if entry.connID == connID {
			delete(rt.pendingDoorbell, username)
		}
	}
	delete(rt.passwordAttempts, connID)
	delete(rt.passwordCooldown, connID)
	rt.leaveCurrentRoom(connID)
}

// SetFloorItemSender configures the function used to send floor items on room entry.
func (rt *Runtime) SetFloorItemSender(fn func(ctx context.Context, connID string, roomID int) error) {
	rt.floorItemSender = fn
}

// SetUsernameResolver configures the optional username lookup function.
func (rt *Runtime) SetUsernameResolver(fn UsernameResolver) {
	rt.usernameResolver = fn
}

// SetProfileResolver configures the full user profile lookup function.
func (rt *Runtime) SetProfileResolver(fn ProfileResolver) {
	rt.profileResolver = fn
}

// SetVoteRepository configures the optional vote persistence layer.
func (rt *Runtime) SetVoteRepository(repo domain.VoteRepository) {
	rt.voteRepo = repo
}

// SetVisitRecorder configures optional room visit tracking.
func (rt *Runtime) SetVisitRecorder(recorder VisitRecorder) {
	rt.visitRecorder = recorder
}

// SetPermissionChecker configures optional moderator permission checks for room actions.
func (rt *Runtime) SetPermissionChecker(checker PermissionChecker) {
	rt.permissions = checker
}

// resolveUsername looks up the display name for a user identifier.
func (rt *Runtime) resolveUsername(ctx context.Context, userID int) string {
	if rt.usernameResolver == nil {
		return fmt.Sprintf("%d", userID)
	}
	name, err := rt.usernameResolver(ctx, userID)
	if err != nil {
		return fmt.Sprintf("%d", userID)
	}
	return name
}
