package tests

import (
	"context"
	"testing"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/room/adapter/realtime"
	roomapp "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// newRuntimeWith builds a runtime with optional room data, rights, permissions, and sessions.
func newRuntimeWith(
	t *testing.T,
	rooms map[int]domain.Room,
	rights map[[2]int]bool,
	perms map[string]bool,
	sessions map[string]coreconnection.Session,
) (*realtime.Runtime, *transportStub) {
	t.Helper()
	models := &modelRepoStub{models: map[string]domain.RoomModel{
		"model_a": {Slug: "model_a", DoorX: 1, DoorY: 1, DoorDir: 2, Heightmap: "xxx\rx0x\rxxx"},
	}}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := roomapp.NewService(models, &banRepoStub{}, &rightsRepoStub{rights: rights}, mgr, zap.NewNop())
	require.NoError(t, err)
	if rooms != nil {
		svc.SetRoomRepository(&roomRepoStub{rooms: rooms})
	}
	tp := &transportStub{}
	var registry coreconnection.SessionRegistry = sessionStub{}
	if sessions != nil {
		registry = &multiSessionStub{sessions: sessions}
	}
	entitySvc, err := roomapp.NewEntityService(mgr, zap.NewNop())
	require.NoError(t, err)
	chatSvc, err := roomapp.NewChatService(zap.NewNop())
	require.NoError(t, err)
	rt, err := realtime.NewRuntime(svc, entitySvc, chatSvc, registry, tp, broadcasterStub{}, zap.NewNop())
	require.NoError(t, err)
	if perms != nil {
		rt.SetPermissionChecker(&permissionCheckerStub{allowed: perms})
	}
	t.Cleanup(func() { mgr.StopAll() })
	return rt, tp
}

// hasPacketID reports whether the transport received a packet with the given ID.
func hasPacketID(sent []uint16, id uint16) bool {
	for _, s := range sent {
		if s == id {
			return true
		}
	}
	return false
}

// passwordRoom builds a password-protected room fixture with a bcrypt hash.
func passwordRoom(t *testing.T) (map[int]domain.Room, string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	require.NoError(t, err)
	return map[int]domain.Room{
		42: {ID: 42, OwnerID: 99, State: domain.AccessPassword, Password: string(hash), ModelSlug: "model_a"},
	}, "secret"
}

// encodeOpenFlat builds the body bytes for an OpenFlatConnectionPacket.
func encodeOpenFlat(t *testing.T, roomID int32, password string) []byte {
	t.Helper()
	body, err := packet.OpenFlatConnectionPacket{RoomID: roomID, Password: password}.Encode()
	require.NoError(t, err)
	return body
}

// TestHandle_OpenFlat_WrongPassword_ShowsAlert verifies a GenericAlertPacket is sent for wrong password.
func TestHandle_OpenFlat_WrongPassword_ShowsAlert(t *testing.T) {
	rooms, _ := passwordRoom(t)
	rt, tp := newRuntimeWith(t, rooms, nil, nil, nil)
	body := encodeOpenFlat(t, 42, "wrong")
	handled, err := rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.True(t, hasPacketID(tp.sent, notificationpacket.GenericAlertPacketID))
	assert.True(t, hasPacketID(tp.sent, packet.CantConnectComposerID))
}

// TestHandle_OpenFlat_WrongPassword_CorrectEntryAfterReset verifies attempts reset on correct password.
func TestHandle_OpenFlat_WrongPassword_CorrectEntryAfterReset(t *testing.T) {
	rooms, _ := passwordRoom(t)
	rt, tp := newRuntimeWith(t, rooms, nil, nil, nil)
	body := encodeOpenFlat(t, 42, "wrong")
	rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	tp.sent = nil
	correct := encodeOpenFlat(t, 42, "secret")
	rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, correct)
	assert.False(t, hasPacketID(tp.sent, packet.CantConnectComposerID))
	assert.True(t, hasPacketID(tp.sent, packet.OpenConnectionComposerID))
}

// TestHandle_OpenFlat_CooldownAppliedAfterThreeAttempts verifies cooldown triggers on the third attempt.
func TestHandle_OpenFlat_CooldownAppliedAfterThreeAttempts(t *testing.T) {
	rooms, _ := passwordRoom(t)
	rt, tp := newRuntimeWith(t, rooms, nil, nil, nil)
	body := encodeOpenFlat(t, 42, "wrong")
	for i := 0; i < 2; i++ {
		rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
		tp.sent = nil
	}
	rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	assert.True(t, hasPacketID(tp.sent, notificationpacket.GenericAlertPacketID))
	assert.True(t, hasPacketID(tp.sent, packet.CantConnectComposerID))
	tp.sent = nil
	rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	assert.True(t, hasPacketID(tp.sent, notificationpacket.GenericAlertPacketID))
	assert.True(t, hasPacketID(tp.sent, packet.CantConnectComposerID))
}

// TestHandle_OpenFlat_PermissionBypass_SkipsPassword verifies room.enter.bypass bypasses password check.
func TestHandle_OpenFlat_PermissionBypass_SkipsPassword(t *testing.T) {
	rooms, _ := passwordRoom(t)
	perms := map[string]bool{"room.enter.bypass": true}
	sess := map[string]coreconnection.Session{"conn1": {ConnID: "conn1", UserID: 1}}
	rt, tp := newRuntimeWith(t, rooms, nil, perms, sess)
	body := encodeOpenFlat(t, 42, "wrong")
	handled, err := rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.False(t, hasPacketID(tp.sent, packet.CantConnectComposerID))
	assert.True(t, hasPacketID(tp.sent, packet.OpenConnectionComposerID))
}

// TestDispose_CleansUpPasswordTracking verifies Dispose removes password state for the connection.
func TestDispose_CleansUpPasswordTracking(t *testing.T) {
	rooms, _ := passwordRoom(t)
	rt, tp := newRuntimeWith(t, rooms, nil, nil, nil)
	body := encodeOpenFlat(t, 42, "wrong")
	for i := 0; i < 3; i++ {
		rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	}
	rt.Dispose("conn1")
	tp.sent = nil
	rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	assert.True(t, hasPacketID(tp.sent, notificationpacket.GenericAlertPacketID))
	assert.False(t, hasPacketID(tp.sent, packet.CantConnectComposerID) && len(tp.sent) == 1)
}

// TestHandle_OpenFlat_LockedRoom_RightsHolderEntersDirect verifies rights holders bypass locked rooms.
func TestHandle_OpenFlat_LockedRoom_RightsHolderEntersDirect(t *testing.T) {
	rooms := map[int]domain.Room{
		10: {ID: 10, OwnerID: 99, State: domain.AccessLocked, ModelSlug: "model_a"},
	}
	rights := map[[2]int]bool{{10, 1}: true}
	sess := map[string]coreconnection.Session{"conn1": {ConnID: "conn1", UserID: 1}}
	rt, tp := newRuntimeWith(t, rooms, rights, nil, sess)
	body := encodeOpenFlat(t, 10, "")
	handled, err := rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.False(t, hasPacketID(tp.sent, packet.CantConnectComposerID))
	assert.True(t, hasPacketID(tp.sent, packet.OpenConnectionComposerID))
}

// TestHandle_OpenFlat_LockedRoom_NonRightsHolderRingsDoorbell verifies non-rights-holders trigger doorbell.
func TestHandle_OpenFlat_LockedRoom_NonRightsHolderRingsDoorbell(t *testing.T) {
	rooms := map[int]domain.Room{
		10: {ID: 10, OwnerID: 99, State: domain.AccessLocked, ModelSlug: "model_a"},
	}
	sess := map[string]coreconnection.Session{"conn1": {ConnID: "conn1", UserID: 1}}
	rt, tp := newRuntimeWith(t, rooms, nil, nil, sess)
	body := encodeOpenFlat(t, 10, "")
	handled, err := rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, body)
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.True(t, hasPacketID(tp.sent, packet.CantConnectComposerID))
}
