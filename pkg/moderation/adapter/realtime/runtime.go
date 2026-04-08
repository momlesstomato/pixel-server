package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	moderationapplication "github.com/momlesstomato/pixel-server/pkg/moderation/application"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	roomdomain "github.com/momlesstomato/pixel-server/pkg/room/domain"
	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	"go.uber.org/zap"
)

// Transport defines packet write behavior.
type Transport interface {
	// Send writes one encoded packet to one connection identifier.
	Send(string, uint16, []byte) error
}

// SessionCloser defines cross-instance session close signal behavior.
type SessionCloser interface {
	// Close publishes a close signal for one connection identifier.
	Close(ctx context.Context, connID string, code int, reason string) error
}

// Runtime defines moderation realm websocket packet behavior.
type Runtime struct {
	// service stores moderation application behavior.
	service *moderationapplication.Service
	// sessions stores authenticated connection lookup.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// broadcaster publishes packet frames to distributed per-user channels.
	broadcaster broadcast.Broadcaster
	// closer publishes session close signals.
	closer SessionCloser
	// logger stores runtime logging behavior.
	logger *zap.Logger
	// tickets stores optional ticket service for CFH handling.
	tickets *moderationapplication.TicketService
	// presets stores optional preset service for mod tool init.
	presets *moderationapplication.PresetService
	// visits stores optional visit service for room tracking.
	visits *moderationapplication.VisitService
	// permissions stores optional permission check behavior.
	permissions PermissionChecker
	// roomLeaveNotifier removes a target connection from its room before disconnect.
	roomLeaveNotifier RoomLeaveNotifier
	// roomAlertSender broadcasts room alerts from the moderation tool.
	roomAlertSender RoomAlertSender
	// users stores optional user lookup behavior for moderator info queries.
	users UserLookup
	// rooms stores optional room lookup behavior for moderator room queries.
	rooms RoomLookup
	// chatLogs stores optional room chat history lookup.
	chatLogs RoomChatLogLookup
	// currentRoomID resolves the current room for one connection.
	currentRoomID CurrentRoomIDResolver
	// roomUserCount resolves live player counts for one room.
	roomUserCount RoomUserCounter
	// roomMuteToggler toggles the room-wide mute state.
	roomMuteToggler RoomMuteToggler
	// roomSettingsUpdater persists moderator room setting changes.
	roomSettingsUpdater RoomSettingsUpdater
}

// PermissionChecker defines permission resolution behavior.
type PermissionChecker interface {
	// HasPermission checks if a user holds a specific permission scope.
	HasPermission(ctx context.Context, userID int, scope string) (bool, error)
}

// RoomLeaveNotifier removes a connection from its active room immediately.
type RoomLeaveNotifier func(connID string)

// RoomAlertSender broadcasts a moderation alert to users in the issuer's current room.
type RoomAlertSender func(ctx context.Context, connID string, message string) error

// UserLookup defines user lookup behavior for moderation tool details.
type UserLookup interface {
	// FindByID resolves one user by identifier.
	FindByID(ctx context.Context, id int) (userdomain.User, error)
}

// RoomLookup defines room lookup behavior for moderation tool details.
type RoomLookup interface {
	// FindRoom resolves one room by identifier.
	FindRoom(ctx context.Context, roomID int) (roomdomain.Room, error)
}

// RoomChatLogLookup defines room chat history lookup behavior.
type RoomChatLogLookup interface {
	// ListByRoom resolves room chat entries for one time window.
	ListByRoom(ctx context.Context, roomID int, from time.Time, to time.Time) ([]roomdomain.ChatLogEntry, error)
}

// CurrentRoomIDResolver resolves the current room for one connection identifier.
type CurrentRoomIDResolver func(connID string) (int, bool)

// RoomUserCounter resolves live player counts for one room.
type RoomUserCounter func(roomID int) int

// RoomMuteToggler toggles room-wide mute state and returns the new state.
type RoomMuteToggler func(ctx context.Context, roomID int) (bool, error)

// RoomSettingsUpdater persists moderator-applied room metadata changes.
type RoomSettingsUpdater func(ctx context.Context, room roomdomain.Room) error

// NewRuntime creates one moderation realtime runtime instance.
func NewRuntime(svc *moderationapplication.Service, sessions coreconnection.SessionRegistry, transport Transport, broadcaster broadcast.Broadcaster, closer SessionCloser, logger *zap.Logger) (*Runtime, error) {
	if svc == nil {
		return nil, fmt.Errorf("moderation service is required")
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
		service: svc, sessions: sessions, transport: transport,
		broadcaster: broadcaster, closer: closer, logger: logger,
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

// Dispose releases per-connection resources.
func (rt *Runtime) Dispose(_ string) {}

// SetTicketService configures optional ticket handling.
func (rt *Runtime) SetTicketService(svc *moderationapplication.TicketService) {
	rt.tickets = svc
}

// SetPresetService configures optional preset/mod tool support.
func (rt *Runtime) SetPresetService(svc *moderationapplication.PresetService) {
	rt.presets = svc
}

// SetVisitService configures optional room visit tracking.
func (rt *Runtime) SetVisitService(svc *moderationapplication.VisitService) {
	rt.visits = svc
}

// SetPermissionChecker configures optional permission checking.
func (rt *Runtime) SetPermissionChecker(checker PermissionChecker) {
	rt.permissions = checker
}

// SetRoomLeaveNotifier configures the callback used to remove users from rooms before closing sessions.
func (rt *Runtime) SetRoomLeaveNotifier(notifier RoomLeaveNotifier) {
	rt.roomLeaveNotifier = notifier
}

// SetRoomAlertSender configures the callback used to broadcast current-room alerts.
func (rt *Runtime) SetRoomAlertSender(sender RoomAlertSender) {
	rt.roomAlertSender = sender
}

// SetUserLookup configures user lookup behavior for moderation scene requests.
func (rt *Runtime) SetUserLookup(lookup UserLookup) {
	rt.users = lookup
}

// SetRoomLookup configures room lookup behavior for moderation scene requests.
func (rt *Runtime) SetRoomLookup(lookup RoomLookup) {
	rt.rooms = lookup
}

// SetRoomChatLogLookup configures room chat history lookup behavior.
func (rt *Runtime) SetRoomChatLogLookup(lookup RoomChatLogLookup) {
	rt.chatLogs = lookup
}

// SetCurrentRoomIDResolver configures current-room resolution for one connection.
func (rt *Runtime) SetCurrentRoomIDResolver(resolver CurrentRoomIDResolver) {
	rt.currentRoomID = resolver
}

// SetRoomUserCounter configures live player count resolution for room info queries.
func (rt *Runtime) SetRoomUserCounter(counter RoomUserCounter) {
	rt.roomUserCount = counter
}

// SetRoomMuteToggler configures room-wide mute toggling behavior.
func (rt *Runtime) SetRoomMuteToggler(toggler RoomMuteToggler) {
	rt.roomMuteToggler = toggler
}

// SetRoomSettingsUpdater configures moderator room setting persistence behavior.
func (rt *Runtime) SetRoomSettingsUpdater(updater RoomSettingsUpdater) {
	rt.roomSettingsUpdater = updater
}

// RecordVisit logs a room visit for moderation tracking.
func (rt *Runtime) RecordVisit(ctx context.Context, userID int, roomID int) {
	if rt.visits == nil {
		return
	}
	_ = rt.visits.RecordVisit(ctx, userID, roomID)
}

// alertAmbassadors broadcasts a moderation alert to ambassador sessions.
func (rt *Runtime) alertAmbassadors(ctx context.Context, message string) {
	if rt.permissions == nil {
		return
	}
	sessions, err := rt.sessions.ListAll()
	if err != nil {
		return
	}
	for _, s := range sessions {
		ok, _ := rt.permissions.HasPermission(ctx, s.UserID, domain.PermAmbassador)
		if ok {
			rt.sendCautionToUser(ctx, s.UserID, message)
		}
	}
}

// _ suppresses unused import warning.
var _ = domain.ScopeHotel
