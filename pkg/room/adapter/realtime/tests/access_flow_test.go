package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/room/adapter/realtime"
	roomapp "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// transportRecord stores one emitted packet for assertion.
type transportRecord struct {
	connID   string
	packetID uint16
}

// transportCapture stores sent packets per connection.
type transportCapture struct{ sent []transportRecord }

// Send records one outgoing packet.
func (capture *transportCapture) Send(connID string, packetID uint16, _ []byte) error {
	capture.sent = append(capture.sent, transportRecord{connID: connID, packetID: packetID})
	return nil
}

type broadcastCapture struct{ sent map[string][]uint16 }

func (capture *broadcastCapture) Publish(_ context.Context, channel string, payload []byte) error {
	frames, err := codec.DecodeFrames(payload)
	if err != nil {
		return err
	}
	if capture.sent == nil {
		capture.sent = make(map[string][]uint16)
	}
	for _, frame := range frames {
		capture.sent[channel] = append(capture.sent[channel], frame.PacketID)
	}
	return nil
}

func (capture *broadcastCapture) Subscribe(context.Context, string) (<-chan []byte, coreconnection.Disposable, error) {
	return nil, nil, nil
}

// multiSessionStub stores deterministic session lookups.
type multiSessionStub struct {
	sessions map[string]coreconnection.Session
}

// Register is a no-op stub.
func (stub multiSessionStub) Register(coreconnection.Session) error { return nil }

// FindByConnID resolves one session by connection identifier.
func (stub multiSessionStub) FindByConnID(connID string) (coreconnection.Session, bool) {
	session, ok := stub.sessions[connID]
	return session, ok
}

// FindByUserID resolves one session by user identifier.
func (stub multiSessionStub) FindByUserID(userID int) (coreconnection.Session, bool) {
	for _, session := range stub.sessions {
		if session.UserID == userID {
			return session, true
		}
	}
	return coreconnection.Session{}, false
}

// Touch is a no-op stub.
func (stub multiSessionStub) Touch(string) error { return nil }

// Remove is a no-op stub.
func (stub multiSessionStub) Remove(string) {}

// ListAll returns all known sessions.
func (stub multiSessionStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

// permissionCheckerStub stores granted scopes by user identifier.
type permissionCheckerStub struct{ grants map[int]map[string]bool }

// HasPermission reports whether one scope is granted.
func (stub permissionCheckerStub) HasPermission(_ context.Context, userID int, scope string) (bool, error) {
	return stub.grants[userID][scope], nil
}

// roomRepoLocalStub stores room data for runtime tests.
type roomRepoLocalStub struct{ rooms map[int]domain.Room }

// FindByID resolves one room.
func (stub *roomRepoLocalStub) FindByID(_ context.Context, roomID int) (domain.Room, error) {
	room, ok := stub.rooms[roomID]
	if !ok {
		return domain.Room{}, domain.ErrRoomNotFound
	}
	return room, nil
}

// SaveSettings is a no-op stub.
func (stub *roomRepoLocalStub) SaveSettings(_ context.Context, _ domain.Room) error { return nil }

// SoftDelete is a no-op stub.
func (stub *roomRepoLocalStub) SoftDelete(_ context.Context, _ int) error { return nil }

// newAccessRuntime creates a room runtime with configurable rooms and sessions.
func newAccessRuntime(t *testing.T, rooms map[int]domain.Room, sessions map[string]coreconnection.Session, usernames map[int]string, grants map[int]map[string]bool, roomRights map[[2]int]bool) (*realtime.Runtime, *transportCapture) {
	rt, transport, _ := newAccessRuntimeWithBroadcast(t, rooms, sessions, usernames, grants, roomRights)
	return rt, transport
}

func newAccessRuntimeWithBroadcast(t *testing.T, rooms map[int]domain.Room, sessions map[string]coreconnection.Session, usernames map[int]string, grants map[int]map[string]bool, roomRights map[[2]int]bool) (*realtime.Runtime, *transportCapture, *broadcastCapture) {
	t.Helper()
	models := &modelRepoStub{models: map[string]domain.RoomModel{
		"model_a": {Slug: "model_a", DoorX: 1, DoorY: 1, DoorDir: 2, Heightmap: "xxx\rx0x\rxxx"},
	}}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := roomapp.NewService(models, &banRepoStub{}, &rightsRepoStub{rights: roomRights}, mgr, zap.NewNop())
	require.NoError(t, err)
	svc.SetRoomRepository(&roomRepoLocalStub{rooms: rooms})
	entitySvc, err := roomapp.NewEntityService(mgr, zap.NewNop())
	require.NoError(t, err)
	chatSvc, err := roomapp.NewChatService(zap.NewNop())
	require.NoError(t, err)
	transport := &transportCapture{}
	broadcaster := &broadcastCapture{sent: make(map[string][]uint16)}
	rt, err := realtime.NewRuntime(svc, entitySvc, chatSvc, multiSessionStub{sessions: sessions}, transport, broadcaster, zap.NewNop())
	require.NoError(t, err)
	rt.SetUsernameResolver(func(_ context.Context, userID int) (string, error) {
		if username, ok := usernames[userID]; ok {
			return username, nil
		}
		return "", nil
	})
	rt.SetPermissionChecker(permissionCheckerStub{grants: grants})
	t.Cleanup(func() { mgr.StopAll() })
	return rt, transport, broadcaster
}

func newSharedAccessRuntimesWithBroadcast(t *testing.T, rooms map[int]domain.Room, sessions map[string]coreconnection.Session, usernames map[int]string, grants map[int]map[string]bool, roomRights map[[2]int]bool) (*realtime.Runtime, *realtime.Runtime, *transportCapture, *transportCapture, *broadcastCapture) {
	t.Helper()
	models := &modelRepoStub{models: map[string]domain.RoomModel{
		"model_a": {Slug: "model_a", DoorX: 1, DoorY: 1, DoorDir: 2, Heightmap: "xxx\rx0x\rxxx"},
	}}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := roomapp.NewService(models, &banRepoStub{}, &rightsRepoStub{rights: roomRights}, mgr, zap.NewNop())
	require.NoError(t, err)
	svc.SetRoomRepository(&roomRepoLocalStub{rooms: rooms})
	entitySvc, err := roomapp.NewEntityService(mgr, zap.NewNop())
	require.NoError(t, err)
	chatSvc, err := roomapp.NewChatService(zap.NewNop())
	require.NoError(t, err)
	broadcaster := &broadcastCapture{sent: make(map[string][]uint16)}
	ownerTransport := &transportCapture{}
	visitorTransport := &transportCapture{}
	ownerRT, err := realtime.NewRuntime(svc, entitySvc, chatSvc, multiSessionStub{sessions: sessions}, ownerTransport, broadcaster, zap.NewNop())
	require.NoError(t, err)
	visitorRT, err := realtime.NewRuntime(svc, entitySvc, chatSvc, multiSessionStub{sessions: sessions}, visitorTransport, broadcaster, zap.NewNop())
	require.NoError(t, err)
	for _, runtime := range []*realtime.Runtime{ownerRT, visitorRT} {
		runtime.SetUsernameResolver(func(_ context.Context, userID int) (string, error) {
			if username, ok := usernames[userID]; ok {
				return username, nil
			}
			return "", nil
		})
		runtime.SetPermissionChecker(permissionCheckerStub{grants: grants})
	}
	t.Cleanup(func() { mgr.StopAll() })
	return ownerRT, visitorRT, ownerTransport, visitorTransport, broadcaster
}

// TestHandleOpenFlat_PasswordFailureCooldown verifies wrong-password feedback and cooldown packets.
func TestHandleOpenFlat_PasswordFailureCooldown(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	require.NoError(t, err)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessPassword, Password: string(hash), ModelSlug: "model_a"}
	rt, transport := newAccessRuntime(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{1: "visitor"}, nil, nil)
	body, err := packet.OpenFlatConnectionPacket{RoomID: 1, Password: "wrong"}.Encode()
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		handled, handleErr := rt.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, body)
		require.NoError(t, handleErr)
		require.True(t, handled)
	}
	packetIDs := make([]uint16, 0, len(transport.sent))
	for _, sent := range transport.sent {
		packetIDs = append(packetIDs, sent.packetID)
	}
	assert.Contains(t, packetIDs, packet.CantConnectComposerID)
	assert.Contains(t, packetIDs, notificationpacket.GenericErrorPacketID)
	assert.Contains(t, packetIDs, packet.FloodControlComposerID)
	assert.Contains(t, packetIDs, sessionnavigation.DesktopViewResponsePacketID)
	assert.Contains(t, packetIDs, notificationpacket.GenericAlertPacketID)
}

// TestHandleOpenFlat_LockedPermissionOverride verifies permission-based override bypasses doorbell access.
func TestHandleOpenFlat_LockedPermissionOverride(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked, ModelSlug: "model_a"}
	rt, transport := newAccessRuntime(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{1: "visitor"}, map[int]map[string]bool{
		1: {"pixels.room.access": true},
	}, nil)
	body, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	handled, handleErr := rt.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, body)
	require.NoError(t, handleErr)
	require.True(t, handled)
	packetIDs := make([]uint16, 0, len(transport.sent))
	for _, sent := range transport.sent {
		if sent.connID == "visitor" {
			packetIDs = append(packetIDs, sent.packetID)
		}
	}
	assert.Contains(t, packetIDs, packet.OpenConnectionComposerID)
	assert.NotContains(t, packetIDs, packet.CantConnectComposerID)
}

// TestHandleLetUserIn_OwnerApproval verifies doorbell approval completes room entry for the visitor.
func TestHandleLetUserIn_OwnerApproval(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked, ModelSlug: "model_a"}
	rt, transport := newAccessRuntime(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"owner":   {ConnID: "owner", UserID: 10},
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{10: "owner", 1: "visitor"}, nil, nil)
	ownerBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "owner", packet.OpenFlatConnectionPacketID, ownerBody)
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "owner", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	visitorBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, visitorBody)
	require.NoError(t, err)
	approvalBody, err := packet.FlatAccessibleComposer{Username: "visitor", Accessible: true}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "owner", packet.LetUserInPacketID, approvalBody)
	require.NoError(t, err)
	visitorPacketIDs := make([]uint16, 0, len(transport.sent))
	for _, sent := range transport.sent {
		if sent.connID == "visitor" {
			visitorPacketIDs = append(visitorPacketIDs, sent.packetID)
		}
	}
	assert.Contains(t, visitorPacketIDs, packet.FlatAccessibleComposerID)
	assert.Contains(t, visitorPacketIDs, packet.OpenConnectionComposerID)
}

// TestHandleOpenFlat_NoControllersPresentReturnsDesktopView verifies locked-room access falls back to hotel view when no one can answer the doorbell.
func TestHandleOpenFlat_NoControllersPresentReturnsDesktopView(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked, ModelSlug: "model_a"}
	rt, transport := newAccessRuntime(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{1: "visitor"}, nil, nil)
	body, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	handled, handleErr := rt.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, body)
	require.NoError(t, handleErr)
	require.True(t, handled)
	packetIDs := make([]uint16, 0, len(transport.sent))
	for _, sent := range transport.sent {
		if sent.connID == "visitor" {
			packetIDs = append(packetIDs, sent.packetID)
		}
	}
	assert.Contains(t, packetIDs, sessionnavigation.DesktopViewResponsePacketID)
	assert.Contains(t, packetIDs, notificationpacket.GenericAlertPacketID)
	assert.NotContains(t, packetIDs, packet.DoorbellComposerID)
}

// TestHandleOpenFlat_PermissionControllerReceivesDoorbell verifies present users with dotted room-access permission can answer the doorbell.
func TestHandleOpenFlat_PermissionControllerReceivesDoorbell(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked, ModelSlug: "model_a"}
	rt, transport := newAccessRuntime(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"staff":   {ConnID: "staff", UserID: 50},
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{50: "staff", 1: "visitor"}, map[int]map[string]bool{
		50: {"pixels.room.access": true},
	}, nil)
	staffBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "staff", packet.OpenFlatConnectionPacketID, staffBody)
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "staff", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	visitorBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, visitorBody)
	require.NoError(t, err)
	doorbellNotified := false
	for _, sent := range transport.sent {
		if sent.connID == "staff" && sent.packetID == packet.DoorbellComposerID {
			doorbellNotified = true
			break
		}
	}
	assert.True(t, doorbellNotified)
}

// TestHandleKickUserBroadcastsRemoval verifies room kicks remove the target entity immediately for other room users.
func TestHandleKickUserBroadcastsRemoval(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessOpen, ModelSlug: "model_a"}
	rt, _, broadcaster := newAccessRuntimeWithBroadcast(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"owner":   {ConnID: "owner", UserID: 10},
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{10: "owner", 1: "visitor"}, nil, nil)
	ownerBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "owner", packet.OpenFlatConnectionPacketID, ownerBody)
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "owner", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	visitorBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, visitorBody)
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "visitor", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	kickBody := codec.NewWriter()
	kickBody.WriteInt32(1)
	_, err = rt.Handle(context.Background(), "owner", packet.KickUserPacketID, kickBody.Bytes())
	require.NoError(t, err)
	assert.Contains(t, broadcaster.sent[sessionnotification.UserChannel(1)], notificationpacket.GenericErrorPacketID)
	assert.Contains(t, broadcaster.sent[sessionnotification.UserChannel(1)], sessionnavigation.DesktopViewResponsePacketID)
	assert.Contains(t, broadcaster.sent[sessionnotification.UserChannel(10)], packet.UserRemoveComposerID)
}

// TestHandleRoomMuteUserBlocksChat verifies Nitro room mute user requests block subsequent chat from the target.
func TestHandleRoomMuteUserBlocksChat(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessOpen, ModelSlug: "model_a"}
	rt, transport := newAccessRuntime(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"owner":   {ConnID: "owner", UserID: 10},
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{10: "owner", 1: "visitor"}, nil, nil)
	ownerBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "owner", packet.OpenFlatConnectionPacketID, ownerBody)
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "owner", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	visitorBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, visitorBody)
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "visitor", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	muteBody := codec.NewWriter()
	muteBody.WriteInt32(1)
	muteBody.WriteInt32(1)
	muteBody.WriteInt32(2)
	_, err = rt.Handle(context.Background(), "owner", packet.RoomMuteUserPacketID, muteBody.Bytes())
	require.NoError(t, err)
	chatBody := codec.NewWriter()
	require.NoError(t, chatBody.WriteString("hello"))
	chatBody.WriteInt32(0)
	_, err = rt.Handle(context.Background(), "visitor", packet.ChatPacketID, chatBody.Bytes())
	require.NoError(t, err)
	visitorPacketIDs := make([]uint16, 0, len(transport.sent))
	for _, sent := range transport.sent {
		if sent.connID == "visitor" {
			visitorPacketIDs = append(visitorPacketIDs, sent.packetID)
		}
	}
	assert.Contains(t, visitorPacketIDs, notificationpacket.GenericAlertPacketID)
	assert.NotContains(t, visitorPacketIDs, packet.ChatComposerID)
}

// TestHandleKickUserBroadcastsRemovalAcrossRuntimes verifies owner kicks work when actor and target use different runtime instances.
func TestHandleKickUserBroadcastsRemovalAcrossRuntimes(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessOpen, ModelSlug: "model_a"}
	ownerRT, visitorRT, _, _, broadcaster := newSharedAccessRuntimesWithBroadcast(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"owner":   {ConnID: "owner", UserID: 10},
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{10: "owner", 1: "visitor"}, nil, nil)
	ownerBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = ownerRT.Handle(context.Background(), "owner", packet.OpenFlatConnectionPacketID, ownerBody)
	require.NoError(t, err)
	_, err = ownerRT.Handle(context.Background(), "owner", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	visitorBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = visitorRT.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, visitorBody)
	require.NoError(t, err)
	_, err = visitorRT.Handle(context.Background(), "visitor", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	kickBody := codec.NewWriter()
	kickBody.WriteInt32(1)
	_, err = ownerRT.Handle(context.Background(), "owner", packet.KickUserPacketID, kickBody.Bytes())
	require.NoError(t, err)
	assert.Contains(t, broadcaster.sent[sessionnotification.UserChannel(1)], notificationpacket.GenericErrorPacketID)
	assert.Contains(t, broadcaster.sent[sessionnotification.UserChannel(1)], sessionnavigation.DesktopViewResponsePacketID)
	assert.Contains(t, broadcaster.sent[sessionnotification.UserChannel(10)], packet.UserRemoveComposerID)
}

// TestHandleRoomMuteUserBlocksChatAcrossRuntimes verifies room-scoped mutes apply when actor and target use different runtime instances.
func TestHandleRoomMuteUserBlocksChatAcrossRuntimes(t *testing.T) {
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessOpen, ModelSlug: "model_a"}
	ownerRT, visitorRT, _, visitorTransport, _ := newSharedAccessRuntimesWithBroadcast(t, map[int]domain.Room{1: room}, map[string]coreconnection.Session{
		"owner":   {ConnID: "owner", UserID: 10},
		"visitor": {ConnID: "visitor", UserID: 1},
	}, map[int]string{10: "owner", 1: "visitor"}, nil, nil)
	ownerBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = ownerRT.Handle(context.Background(), "owner", packet.OpenFlatConnectionPacketID, ownerBody)
	require.NoError(t, err)
	_, err = ownerRT.Handle(context.Background(), "owner", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	visitorBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = visitorRT.Handle(context.Background(), "visitor", packet.OpenFlatConnectionPacketID, visitorBody)
	require.NoError(t, err)
	_, err = visitorRT.Handle(context.Background(), "visitor", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	muteBody := codec.NewWriter()
	muteBody.WriteInt32(1)
	muteBody.WriteInt32(1)
	muteBody.WriteInt32(2)
	_, err = ownerRT.Handle(context.Background(), "owner", packet.RoomMuteUserPacketID, muteBody.Bytes())
	require.NoError(t, err)
	chatBody := codec.NewWriter()
	require.NoError(t, chatBody.WriteString("hello"))
	chatBody.WriteInt32(0)
	_, err = visitorRT.Handle(context.Background(), "visitor", packet.ChatPacketID, chatBody.Bytes())
	require.NoError(t, err)
	visitorPacketIDs := make([]uint16, 0, len(visitorTransport.sent))
	for _, sent := range visitorTransport.sent {
		if sent.connID == "visitor" {
			visitorPacketIDs = append(visitorPacketIDs, sent.packetID)
		}
	}
	assert.Contains(t, visitorPacketIDs, notificationpacket.GenericAlertPacketID)
	assert.NotContains(t, visitorPacketIDs, packet.ChatComposerID)
}
