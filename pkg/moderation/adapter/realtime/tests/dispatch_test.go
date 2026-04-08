package tests

import (
	"context"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/core/connection"
	moderationrealtime "github.com/momlesstomato/pixel-server/pkg/moderation/adapter/realtime"
	moderationapplication "github.com/momlesstomato/pixel-server/pkg/moderation/application"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
	roomdomain "github.com/momlesstomato/pixel-server/pkg/room/domain"
	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type actionRepoStub struct {
	actions []*domain.Action
}

func (repo *actionRepoStub) Create(_ context.Context, action *domain.Action) error {
	copy := *action
	repo.actions = append(repo.actions, &copy)
	return nil
}

func (repo *actionRepoStub) FindByID(_ context.Context, _ int64) (*domain.Action, error) {
	return nil, domain.ErrActionNotFound
}

func (repo *actionRepoStub) List(_ context.Context, filter domain.ListFilter) ([]domain.Action, error) {
	out := make([]domain.Action, 0, len(repo.actions))
	for _, action := range repo.actions {
		if filter.Active != nil && action.Active != *filter.Active {
			continue
		}
		if filter.ActionType != "" && action.ActionType != filter.ActionType {
			continue
		}
		if filter.IssuerID > 0 && action.IssuerID != filter.IssuerID {
			continue
		}
		if filter.TargetUserID > 0 && action.TargetUserID != filter.TargetUserID {
			continue
		}
		if filter.RoomID > 0 && action.RoomID != filter.RoomID {
			continue
		}
		out = append(out, *action)
	}
	return out, nil
}

func (repo *actionRepoStub) Deactivate(_ context.Context, _ int64, _ int) error {
	return nil
}

func (repo *actionRepoStub) Delete(_ context.Context, _ int64) error {
	return nil
}

func (repo *actionRepoStub) HasActiveBan(_ context.Context, _ int, _ domain.ActionScope) (bool, error) {
	return false, nil
}

func (repo *actionRepoStub) HasActiveMute(_ context.Context, _ int, _ domain.ActionScope) (bool, error) {
	for _, action := range repo.actions {
		if action.ActionType == domain.TypeMute && action.Active {
			return true, nil
		}
	}
	return false, nil
}

func (repo *actionRepoStub) HasActiveIPBan(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (repo *actionRepoStub) HasActiveTradeLock(_ context.Context, _ int) (bool, error) {
	for _, action := range repo.actions {
		if action.ActionType == domain.TypeTradeLock && action.Active {
			return true, nil
		}
	}
	return false, nil
}

type sessionRegistryStub struct{}

func (sessionRegistryStub) Register(connection.Session) error { return nil }

func (sessionRegistryStub) FindByConnID(connID string) (connection.Session, bool) {
	switch connID {
	case "conn1":
		return connection.Session{ConnID: "conn1", UserID: 1}, true
	case "conn2":
		return connection.Session{ConnID: "conn2", UserID: 2}, true
	case "conn3":
		return connection.Session{ConnID: "conn3", UserID: 3}, true
	default:
		return connection.Session{}, false
	}
}

func (sessionRegistryStub) FindByUserID(userID int) (connection.Session, bool) {
	switch userID {
	case 1:
		return connection.Session{ConnID: "conn1", UserID: 1}, true
	case 2:
		return connection.Session{ConnID: "conn2", UserID: 2}, true
	case 3:
		return connection.Session{ConnID: "conn3", UserID: 3}, true
	default:
		return connection.Session{}, false
	}
}

func (sessionRegistryStub) Touch(string) error { return nil }

func (sessionRegistryStub) Remove(string) {}

func (sessionRegistryStub) ListAll() ([]connection.Session, error) {
	return []connection.Session{{ConnID: "conn1", UserID: 1}, {ConnID: "conn2", UserID: 2}, {ConnID: "conn3", UserID: 3}}, nil
}

type transportStub struct {
	packetIDs []uint16
	bodies    map[uint16][][]byte
}

func (transport *transportStub) Send(_ string, packetID uint16, body []byte) error {
	transport.packetIDs = append(transport.packetIDs, packetID)
	if transport.bodies == nil {
		transport.bodies = map[uint16][][]byte{}
	}
	copy := append([]byte(nil), body...)
	transport.bodies[packetID] = append(transport.bodies[packetID], copy)
	return nil
}

type broadcasterStub struct{}

func (broadcasterStub) Publish(context.Context, string, []byte) error { return nil }

func (broadcasterStub) Subscribe(context.Context, string) (<-chan []byte, connection.Disposable, error) {
	return nil, nil, nil
}

var _ broadcast.Broadcaster = broadcasterStub{}

type closerStub struct {
	closed []string
}

func (closer *closerStub) Close(_ context.Context, connID string, _ int, _ string) error {
	closer.closed = append(closer.closed, connID)
	return nil
}

type permissionStub struct{}

func (permissionStub) HasPermission(_ context.Context, userID int, scope string) (bool, error) {
	if userID != 1 {
		return false, nil
	}
	switch scope {
	case domain.PermKick, domain.PermMute, domain.PermWarn, domain.PermAmbassador, domain.PermTool, domain.PermHistory:
		return true, nil
	default:
		return false, nil
	}
}

type roomLookupStub struct{}

func (roomLookupStub) FindRoom(_ context.Context, roomID int) (roomdomain.Room, error) {
	return roomdomain.Room{ID: roomID, OwnerID: 2, OwnerName: "beta", Name: "Blue Room", Description: "Testing", State: roomdomain.AccessOpen, Tags: []string{"test"}, MaxUsers: 25}, nil
}

type userLookupStub struct{}

func (userLookupStub) FindByID(_ context.Context, id int) (userdomain.User, error) {
	return userdomain.User{ID: id, Username: "beta", Figure: "hr-1", GroupID: 4}, nil
}

type chatLogLookupStub struct{}

func (chatLogLookupStub) ListByRoom(_ context.Context, roomID int, _ time.Time, _ time.Time) ([]roomdomain.ChatLogEntry, error) {
	return []roomdomain.ChatLogEntry{{RoomID: roomID, UserID: 2, Username: "beta", Message: "hello", CreatedAt: time.Unix(1710000000, 0)}}, nil
}

type ticketRepoStub struct{}

func (ticketRepoStub) Create(_ context.Context, _ *domain.Ticket) error { return nil }

func (ticketRepoStub) FindByID(_ context.Context, id int64) (*domain.Ticket, error) {
	return &domain.Ticket{ID: id, ReporterID: 2, ReportedID: 3, RoomID: 77, Message: "Need help", CreatedAt: time.Unix(1710000000, 0)}, nil
}

func (ticketRepoStub) List(_ context.Context, _ domain.TicketStatus, _ int) ([]domain.Ticket, error) {
	return []domain.Ticket{{ID: 5, ReporterID: 2, ReportedID: 3, RoomID: 77, Message: "Need help", CreatedAt: time.Unix(1710000000, 0)}}, nil
}

func (ticketRepoStub) UpdateStatus(_ context.Context, _ int64, _ domain.TicketStatus, _ int) error {
	return nil
}

func (ticketRepoStub) Delete(_ context.Context, _ int64) error { return nil }

type reporterTicketRepoStub struct{}

func (reporterTicketRepoStub) Create(_ context.Context, _ *domain.Ticket) error { return nil }

func (reporterTicketRepoStub) FindByID(_ context.Context, id int64) (*domain.Ticket, error) {
	return &domain.Ticket{ID: id, ReporterID: 1, ReportedID: 2, RoomID: 77, Message: "Need help", Status: domain.TicketOpen, CreatedAt: time.Now().Add(-2 * time.Minute)}, nil
}

func (reporterTicketRepoStub) List(_ context.Context, status domain.TicketStatus, _ int) ([]domain.Ticket, error) {
	if status != domain.TicketOpen && status != domain.TicketInProgress {
		return nil, nil
	}
	return []domain.Ticket{{ID: 9, ReporterID: 1, ReportedID: 2, RoomID: 77, Message: "Need help", Status: status, CreatedAt: time.Now().Add(-2 * time.Minute)}}, nil
}

func (reporterTicketRepoStub) UpdateStatus(_ context.Context, _ int64, _ domain.TicketStatus, _ int) error {
	return nil
}

func (reporterTicketRepoStub) Delete(_ context.Context, _ int64) error { return nil }

// TestHandleModKickUsesFallbackReasonAndLeavesRoom verifies empty-message kicks still persist, remove the room entity, and close the target session.
func TestHandleModKickUsesFallbackReasonAndLeavesRoom(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	closer := &closerStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, &transportStub{}, broadcasterStub{}, closer, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	leftRooms := make([]string, 0, 1)
	rt.SetRoomLeaveNotifier(func(connID string) {
		leftRooms = append(leftRooms, connID)
	})
	w := codec.NewWriter()
	w.WriteInt32(2)
	require.NoError(t, w.WriteString(""))
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModKickUserPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	require.Len(t, repo.actions, 1)
	assert.Equal(t, "ambassador kick", repo.actions[0].Reason)
	assert.Equal(t, []string{"conn2"}, leftRooms)
	assert.Equal(t, []string{"conn2"}, closer.closed)
}

// TestHandleModRoomAlertForwardsToRoomSender verifies room alerts reach the room alert callback.
func TestHandleModRoomAlertForwardsToRoomSender(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, &transportStub{}, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetCurrentRoomIDResolver(func(connID string) (int, bool) {
		if connID == "conn1" {
			return 77, true
		}
		return 0, false
	})
	called := false
	message := ""
	rt.SetRoomAlertSender(func(_ context.Context, connID string, value string) error {
		called = connID == "conn1"
		message = value
		return nil
	})
	w := codec.NewWriter()
	w.WriteInt32(3)
	require.NoError(t, w.WriteString(""))
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModRoomAlertPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	assert.True(t, called)
	assert.Equal(t, "ambassador room alert", message)
	require.Len(t, repo.actions, 1)
	assert.Equal(t, domain.ScopeRoom, repo.actions[0].Scope)
	assert.Equal(t, 77, repo.actions[0].RoomID)
	assert.Equal(t, 0, repo.actions[0].TargetUserID)
}

// TestHandleModMuteSendsDirectCaution verifies mute actions immediately notify a local target connection.
func TestHandleModMuteSendsDirectCaution(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	w := codec.NewWriter()
	w.WriteInt32(2)
	require.NoError(t, w.WriteString(""))
	w.WriteInt32(60)
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModMuteUserPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	require.Len(t, repo.actions, 1)
	assert.Equal(t, domain.TypeMute, repo.actions[0].ActionType)
	assert.Contains(t, transport.packetIDs, uint16(1890))
}

// TestHandleRoomAmbassadorAlertSendsTargetedCaution verifies the ambassador alert targets a single user.
func TestHandleRoomAmbassadorAlertSendsTargetedCaution(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	w := codec.NewWriter()
	w.WriteInt32(2)
	handled, err := rt.Handle(context.Background(), "conn1", packet.RoomAmbassadorAlertPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	require.Len(t, repo.actions, 1)
	assert.Equal(t, 2, repo.actions[0].TargetUserID)
	assert.Contains(t, transport.packetIDs, uint16(1890))
}

// TestHandleModToolRequestRoomInfoSendsPacket verifies room info scene requests return Nitro room info payloads.
func TestHandleModToolRequestRoomInfoSendsPacket(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetRoomLookup(roomLookupStub{})
	rt.SetRoomUserCounter(func(roomID int) int {
		if roomID == 77 {
			return 4
		}
		return 0
	})
	rt.SetCurrentRoomIDResolver(func(connID string) (int, bool) {
		if connID == "conn2" {
			return 77, true
		}
		return 0, false
	})
	w := codec.NewWriter()
	w.WriteInt32(77)
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModToolRequestRoomInfoPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Contains(t, transport.packetIDs, packet.ModToolRoomInfoComposerID)
}

// TestHandleModToolRequestRoomChatlogSendsPacket verifies room chatlog scene requests return Nitro chatlog payloads.
func TestHandleModToolRequestRoomChatlogSendsPacket(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetRoomLookup(roomLookupStub{})
	rt.SetRoomChatLogLookup(chatLogLookupStub{})
	w := codec.NewWriter()
	w.WriteInt32(0)
	w.WriteInt32(77)
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModToolRequestRoomChatlogPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Contains(t, transport.packetIDs, packet.ModToolRoomChatlogComposerID)
}

// TestHandleModToolUserInfoSendsPacket verifies user info scene requests return Nitro user info payloads.
func TestHandleModToolUserInfoSendsPacket(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetUserLookup(userLookupStub{})
	w := codec.NewWriter()
	w.WriteInt32(2)
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModToolUserInfoPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Contains(t, transport.packetIDs, packet.ModeratorUserInfoComposerID)
}

// TestHandleGetPendingCallsForHelpSendsPacket verifies pending CFH scene requests return ticket list payloads.
func TestHandleGetPendingCallsForHelpSendsPacket(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	ticketSvc, err := moderationapplication.NewTicketService(ticketRepoStub{})
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetTicketService(ticketSvc)
	handled, err := rt.Handle(context.Background(), "conn1", packet.GetPendingCallsForHelpPacketID, nil)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Contains(t, transport.packetIDs, packet.CFHPendingPacketID)
}

// TestHandleGetCFHChatlogSendsPacket verifies CFH chatlog scene requests return Nitro chatlog payloads.
func TestHandleGetCFHChatlogSendsPacket(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	ticketSvc, err := moderationapplication.NewTicketService(ticketRepoStub{})
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetTicketService(ticketSvc)
	rt.SetRoomLookup(roomLookupStub{})
	rt.SetRoomChatLogLookup(chatLogLookupStub{})
	w := codec.NewWriter()
	w.WriteInt32(5)
	handled, err := rt.Handle(context.Background(), "conn1", packet.GetCFHChatlogPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Contains(t, transport.packetIDs, packet.ModeratorCFHChatlogPacketID)
}

// TestHandleModToolPreferencesSendsPacket verifies preferences scene requests return a preferences payload.
func TestHandleModToolPreferencesSendsPacket(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModToolPreferencesPacketID, nil)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Contains(t, transport.packetIDs, packet.ModeratorToolPreferencesComposerID)
}

// TestHandleRoomMuteCallsToggler verifies room mute scene requests toggle the room mute state.
func TestHandleRoomMuteCallsToggler(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, &transportStub{}, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetCurrentRoomIDResolver(func(connID string) (int, bool) {
		if connID == "conn1" {
			return 77, true
		}
		return 0, false
	})
	called := 0
	rt.SetRoomMuteToggler(func(_ context.Context, roomID int) (bool, error) {
		called++
		assert.Equal(t, 77, roomID)
		return true, nil
	})
	handled, err := rt.Handle(context.Background(), "conn1", packet.RoomMutePacketID, nil)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, 1, called)
}

// TestHandleModToolChangeRoomSettingsUpdatesTitleAndDoorMode verifies moderator room settings apply title and locked-door changes.
func TestHandleModToolChangeRoomSettingsUpdatesTitleAndDoorMode(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, &transportStub{}, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetPermissionChecker(permissionStub{})
	rt.SetRoomLookup(roomLookupStub{})
	var saved roomdomain.Room
	called := 0
	rt.SetRoomSettingsUpdater(func(_ context.Context, room roomdomain.Room) error {
		called++
		saved = room
		return nil
	})
	w := codec.NewWriter()
	w.WriteInt32(77)
	w.WriteInt32(1)
	w.WriteInt32(1)
	w.WriteInt32(0)
	handled, err := rt.Handle(context.Background(), "conn1", packet.ModToolChangeRoomSettingsPacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, 1, called)
	assert.Equal(t, 77, saved.ID)
	assert.Equal(t, roomdomain.AccessLocked, saved.State)
	assert.Equal(t, "Inappropriate to hotel staff", saved.Name)
}

// TestHandleGetCFHStatusSendsSanctionStatusPacket verifies sanction-status requests return Nitro sanction data.
func TestHandleGetCFHStatusSendsSanctionStatusPacket(t *testing.T) {
	repo := &actionRepoStub{actions: []*domain.Action{{
		TargetUserID:    1,
		Scope:           domain.ScopeHotel,
		ActionType:      domain.TypeMute,
		Reason:          "spam",
		DurationMinutes: 120,
		Active:          true,
		CreatedAt:       time.Now().Add(-30 * time.Minute),
		ExpiresAt:       timePointer(time.Now().Add(90 * time.Minute)),
	}, {
		TargetUserID:    1,
		Scope:           domain.ScopeHotel,
		ActionType:      domain.TypeTradeLock,
		Reason:          "trade abuse",
		DurationMinutes: 60,
		Active:          true,
		CreatedAt:       time.Now().Add(-15 * time.Minute),
		ExpiresAt:       timePointer(time.Now().Add(45 * time.Minute)),
	}}}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	handled, err := rt.Handle(context.Background(), "conn1", packet.GetCFHStatusPacketID, nil)
	require.NoError(t, err)
	assert.True(t, handled)
	require.Contains(t, transport.packetIDs, packet.CFHSanctionStatusPacketID)
	body := transport.bodies[packet.CFHSanctionStatusPacketID][0]
	r := codec.NewReader(body)
	isNew, err := r.ReadBool()
	require.NoError(t, err)
	isActive, err := r.ReadBool()
	require.NoError(t, err)
	name, err := r.ReadString()
	require.NoError(t, err)
	lengthHours, err := r.ReadInt32()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	reason, err := r.ReadString()
	require.NoError(t, err)
	_, err = r.ReadString()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	_, err = r.ReadString()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	_, err = r.ReadInt32()
	require.NoError(t, err)
	hasCustomMute, err := r.ReadBool()
	require.NoError(t, err)
	tradeLockExpiry, err := r.ReadString()
	require.NoError(t, err)
	assert.True(t, isNew)
	assert.True(t, isActive)
	assert.Equal(t, "MUTE", name)
	assert.Equal(t, int32(2), lengthHours)
	assert.Equal(t, "spam", reason)
	assert.True(t, hasCustomMute)
	assert.NotEmpty(t, tradeLockExpiry)
}

// TestHandleGuideSessionCreateSendsGuideError verifies unsupported guide session requests return a deterministic error.
func TestHandleGuideSessionCreateSendsGuideError(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	w := codec.NewWriter()
	w.WriteInt32(1)
	require.NoError(t, w.WriteString("Need assistance"))
	handled, err := rt.Handle(context.Background(), "conn1", packet.GuideSessionCreatePacketID, w.Bytes())
	require.NoError(t, err)
	assert.True(t, handled)
	require.Contains(t, transport.packetIDs, packet.GuideSessionErrorPacketID)
	body := transport.bodies[packet.GuideSessionErrorPacketID][0]
	r := codec.NewReader(body)
	errorCode, err := r.ReadInt32()
	require.NoError(t, err)
	assert.Equal(t, int32(1), errorCode)
}

// TestHandleGetGuideReportingStatusSendsPendingTicket verifies room-report status requests expose a pending ticket when present.
func TestHandleGetGuideReportingStatusSendsPendingTicket(t *testing.T) {
	repo := &actionRepoStub{}
	svc, err := moderationapplication.NewService(repo)
	require.NoError(t, err)
	ticketSvc, err := moderationapplication.NewTicketService(reporterTicketRepoStub{})
	require.NoError(t, err)
	transport := &transportStub{}
	rt, err := moderationrealtime.NewRuntime(svc, sessionRegistryStub{}, transport, broadcasterStub{}, &closerStub{}, zap.NewNop())
	require.NoError(t, err)
	rt.SetTicketService(ticketSvc)
	rt.SetUserLookup(userLookupStub{})
	rt.SetRoomLookup(roomLookupStub{})
	handled, err := rt.Handle(context.Background(), "conn1", packet.GetGuideReportingStatusPacketID, nil)
	require.NoError(t, err)
	assert.True(t, handled)
	require.Contains(t, transport.packetIDs, packet.GuideReportingStatusPacketID)
	body := transport.bodies[packet.GuideReportingStatusPacketID][0]
	r := codec.NewReader(body)
	statusCode, err := r.ReadInt32()
	require.NoError(t, err)
	requestType, err := r.ReadInt32()
	require.NoError(t, err)
	secondsAgo, err := r.ReadInt32()
	require.NoError(t, err)
	isGuide, err := r.ReadBool()
	require.NoError(t, err)
	otherPartyName, err := r.ReadString()
	require.NoError(t, err)
	otherPartyFigure, err := r.ReadString()
	require.NoError(t, err)
	description, err := r.ReadString()
	require.NoError(t, err)
	roomName, err := r.ReadString()
	require.NoError(t, err)
	assert.Equal(t, int32(1), statusCode)
	assert.Equal(t, int32(1), requestType)
	assert.Positive(t, secondsAgo)
	assert.False(t, isGuide)
	assert.Equal(t, "beta", otherPartyName)
	assert.Equal(t, "hr-1", otherPartyFigure)
	assert.Equal(t, "Need help", description)
	assert.Equal(t, "Blue Room", roomName)
}

func timePointer(value time.Time) *time.Time {
	return &value
}
