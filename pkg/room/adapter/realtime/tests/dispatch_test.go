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
	rt, err := realtime.NewRuntime(svc, entitySvc, chatSvc, sessionStub{}, tp, zap.NewNop())
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
