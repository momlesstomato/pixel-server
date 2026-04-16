package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkroom "github.com/momlesstomato/pixel-sdk/events/room"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/room/adapter/realtime"
	roomapp "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// transportStub captures sent packets for assertion.
type transportStub struct{ sent []uint16 }

// Send appends the packet ID to the sent list.
func (t *transportStub) Send(_ string, id uint16, _ []byte) error {
	t.sent = append(t.sent, id)
	return nil
}

// sessionStub provides deterministic session lookup.
type sessionStub struct{}

// Register is a no-op stub.
func (sessionStub) Register(coreconnection.Session) error { return nil }

// FindByConnID returns session for conn1 only.
func (sessionStub) FindByConnID(id string) (coreconnection.Session, bool) {
	if id == "conn1" {
		return coreconnection.Session{ConnID: "conn1", UserID: 1}, true
	}
	return coreconnection.Session{}, false
}

// FindByUserID is a no-op stub.
func (sessionStub) FindByUserID(int) (coreconnection.Session, bool) {
	return coreconnection.Session{}, false
}

// Touch is a no-op stub.
func (sessionStub) Touch(string) error { return nil }

// Remove is a no-op stub.
func (sessionStub) Remove(string) {}

// ListAll is a no-op stub.
func (sessionStub) ListAll() ([]coreconnection.Session, error) { return nil, nil }

func noopBroadcaster(_ int, _ []domain.RoomEntity, _ []byte) {}

// broadcasterStub discards all publish calls.
type broadcasterStub struct{}

// Publish is a no-op stub.
func (broadcasterStub) Publish(_ context.Context, _ string, _ []byte) error { return nil }

// Subscribe is a no-op stub.
func (broadcasterStub) Subscribe(_ context.Context, _ string) (<-chan []byte, coreconnection.Disposable, error) {
	return nil, nil, nil
}

func newRuntime(t *testing.T) (*realtime.Runtime, *transportStub) {
	t.Helper()
	models := &modelRepoStub{models: map[string]domain.RoomModel{
		"model_a": {Slug: "model_a", DoorX: 1, DoorY: 1, DoorDir: 2,
			Heightmap: "xxx\rx0x\rxxx"},
	}}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := roomapp.NewService(models, &banRepoStub{}, &rightsRepoStub{}, mgr, zap.NewNop())
	require.NoError(t, err)
	tp := &transportStub{}
	entitySvc, err := roomapp.NewEntityService(mgr, zap.NewNop())
	require.NoError(t, err)
	chatSvc, err := roomapp.NewChatService(zap.NewNop())
	require.NoError(t, err)
	rt, err := realtime.NewRuntime(svc, entitySvc, chatSvc, sessionStub{}, tp, broadcasterStub{}, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { mgr.StopAll() })
	return rt, tp
}

// TestHandle_UnknownPacket verifies unknown packets are not handled.
func TestHandle_UnknownPacket(t *testing.T) {
	rt, _ := newRuntime(t)
	handled, err := rt.Handle(context.Background(), "conn1", 9999, nil)
	assert.NoError(t, err)
	assert.False(t, handled)
}

// TestHandle_UnauthenticatedConn verifies unauth returns false.
func TestHandle_UnauthenticatedConn(t *testing.T) {
	rt, _ := newRuntime(t)
	handled, err := rt.Handle(context.Background(), "unknown", packet.OpenFlatConnectionPacketID, nil)
	assert.NoError(t, err)
	assert.False(t, handled)
}

// TestHandle_OpenFlat_CantConnect verifies bad body returns true.
func TestHandle_OpenFlat_BadBody(t *testing.T) {
	rt, tp := newRuntime(t)
	handled, err := rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, []byte{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Empty(t, tp.sent)
}

// TestHandle_MoveAvatar_NoRoom verifies move without room entry.
func TestHandle_MoveAvatar_NoRoom(t *testing.T) {
	rt, _ := newRuntime(t)
	body := make([]byte, 8)
	body[3], body[7] = 2, 2
	handled, err := rt.Handle(context.Background(), "conn1", packet.MoveAvatarPacketID, body)
	assert.NoError(t, err)
	assert.True(t, handled)
}

// TestQueueTeleporterEntry_SpawnsWithoutImmediateExitWalk verifies pending teleporter entry enters the target room on the spawn tile without immediately walking out.
func TestQueueTeleporterEntry_SpawnsWithoutImmediateExitWalk(t *testing.T) {
	models := &modelRepoStub{models: map[string]domain.RoomModel{
		"model_a": {Slug: "model_a", DoorX: 1, DoorY: 1, DoorDir: 2, Heightmap: "xxxxx\rx000x\rxxxxx"},
	}}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := roomapp.NewService(models, &banRepoStub{}, &rightsRepoStub{}, mgr, zap.NewNop())
	require.NoError(t, err)
	svc.SetRoomRepository(&roomRepoLocalStub{rooms: map[int]domain.Room{1: {ID: 1, OwnerID: 1, State: domain.AccessOpen, ModelSlug: "model_a"}}})
	entitySvc, err := roomapp.NewEntityService(mgr, zap.NewNop())
	require.NoError(t, err)
	chatSvc, err := roomapp.NewChatService(zap.NewNop())
	require.NoError(t, err)
	tp := &transportStub{}
	rt, err := realtime.NewRuntime(svc, entitySvc, chatSvc, sessionStub{}, tp, broadcasterStub{}, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { mgr.StopAll() })
	baselinePackets := len(tp.sent)
	require.NoError(t, rt.QueueTeleporterEntry(context.Background(), "conn1", 1, 2, 1, 0, 2, 3, 1))
	inst, ok := mgr.Get(1)
	require.True(t, ok)
	require.Empty(t, inst.Entities())
	queuedPackets := tp.sent[baselinePackets:]
	assert.NotContains(t, queuedPackets, packet.UsersComposerID)
	assert.NotContains(t, queuedPackets, packet.UserUpdateComposerID)
	_, err = rt.Handle(context.Background(), "conn1", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	entities := inst.Entities()
	require.Len(t, entities, 1)
	assert.Equal(t, 2, entities[0].Position.X)
	assert.Equal(t, 1, entities[0].Position.Y)
	assert.Nil(t, entities[0].GoalPosition)
	assert.Empty(t, entities[0].Path)
	require.NotEmpty(t, tp.sent)
	assert.Equal(t, packet.OpenConnectionComposerID, tp.sent[0])
}

// TestQueueTeleporterEntry_UsesRoomLifecycle verifies cross-room teleports leave through room events and defer target entry until room entry data is requested.
func TestQueueTeleporterEntry_UsesRoomLifecycle(t *testing.T) {
	models := &modelRepoStub{models: map[string]domain.RoomModel{
		"model_a": {Slug: "model_a", DoorX: 1, DoorY: 1, DoorDir: 2, Heightmap: "xxxxx\rx000x\rxxxxx"},
	}}
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := roomapp.NewService(models, &banRepoStub{}, &rightsRepoStub{}, mgr, zap.NewNop())
	require.NoError(t, err)
	events := make([]string, 0, 4)
	svc.SetEventFirer(func(event sdk.Event) {
		switch event.(type) {
		case *sdkroom.RoomLeaving:
			events = append(events, "leaving")
		case *sdkroom.RoomLeft:
			events = append(events, "left")
		case *sdkroom.RoomEntering:
			events = append(events, "entering")
		case *sdkroom.RoomEntered:
			events = append(events, "entered")
		}
	})
	svc.SetRoomRepository(&roomRepoLocalStub{rooms: map[int]domain.Room{
		1: {ID: 1, OwnerID: 1, State: domain.AccessOpen, ModelSlug: "model_a"},
		2: {ID: 2, OwnerID: 1, State: domain.AccessOpen, ModelSlug: "model_a"},
	}})
	entitySvc, err := roomapp.NewEntityService(mgr, zap.NewNop())
	require.NoError(t, err)
	chatSvc, err := roomapp.NewChatService(zap.NewNop())
	require.NoError(t, err)
	tp := &transportStub{}
	rt, err := realtime.NewRuntime(svc, entitySvc, chatSvc, sessionStub{}, tp, broadcasterStub{}, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { mgr.StopAll() })
	openBody, err := packet.OpenFlatConnectionPacket{RoomID: 1}.Encode()
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "conn1", packet.OpenFlatConnectionPacketID, openBody)
	require.NoError(t, err)
	_, err = rt.Handle(context.Background(), "conn1", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	events = events[:0]
	inst, ok := mgr.Get(1)
	require.True(t, ok)
	require.Len(t, inst.Entities(), 1)
	baselinePackets := len(tp.sent)
	require.NoError(t, rt.QueueTeleporterEntry(context.Background(), "conn1", 2, 2, 1, 0, 2, 3, 1))
	assert.Empty(t, inst.Entities())
	targetInst, ok := mgr.Get(2)
	require.True(t, ok)
	assert.Empty(t, targetInst.Entities())
	queuedPackets := tp.sent[baselinePackets:]
	assert.NotContains(t, queuedPackets, packet.UsersComposerID)
	assert.NotContains(t, queuedPackets, packet.UserUpdateComposerID)
	assert.Equal(t, []string{"leaving", "left"}, events)
	_, err = rt.Handle(context.Background(), "conn1", packet.GetRoomEntryDataPacketID, nil)
	require.NoError(t, err)
	targetEntities := targetInst.Entities()
	require.Len(t, targetEntities, 1)
	assert.Equal(t, 2, targetEntities[0].Position.X)
	assert.Equal(t, 1, targetEntities[0].Position.Y)
	assert.Equal(t, []string{"leaving", "left", "entering", "entered"}, events)
	require.NotEmpty(t, queuedPackets)
	assert.Contains(t, queuedPackets, packet.OpenConnectionComposerID)
}

// TestDispose verifies connection cleanup for a connection not in any room.
func TestDispose(t *testing.T) {
	rt, _ := newRuntime(t)
	rt.Dispose("conn1")
}

// TestDisposeIsIdempotent verifies repeated Dispose calls do not panic.
func TestDisposeIsIdempotent(t *testing.T) {
	rt, _ := newRuntime(t)
	rt.Dispose("conn1")
	rt.Dispose("conn1")
}

// TestHandle_KickUser_NotInRoom verifies kick when requester is not in a room.
func TestHandle_KickUser_NotInRoom(t *testing.T) {
	rt, _ := newRuntime(t)
	w := codec.NewWriter()
	w.WriteInt32(2)
	handled, err := rt.Handle(context.Background(), "conn1", packet.KickUserPacketID, w.Bytes())
	assert.NoError(t, err)
	assert.True(t, handled)
}

// TestHandle_BanUser_NotInRoom verifies ban when requester is not in a room.
func TestHandle_BanUser_NotInRoom(t *testing.T) {
	rt, _ := newRuntime(t)
	w := codec.NewWriter()
	w.WriteInt32(2)
	w.WriteInt32(1)
	_ = w.WriteString("RWUAM_BAN_USER_HOUR")
	handled, err := rt.Handle(context.Background(), "conn1", packet.BanUserPacketID, w.Bytes())
	assert.NoError(t, err)
	assert.True(t, handled)
}
