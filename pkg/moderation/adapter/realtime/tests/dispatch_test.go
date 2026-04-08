package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/core/connection"
	moderationapplication "github.com/momlesstomato/pixel-server/pkg/moderation/application"
	moderationrealtime "github.com/momlesstomato/pixel-server/pkg/moderation/adapter/realtime"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
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

func (repo *actionRepoStub) List(_ context.Context, _ domain.ListFilter) ([]domain.Action, error) {
	return nil, nil
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
	return false, nil
}

func (repo *actionRepoStub) HasActiveIPBan(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (repo *actionRepoStub) HasActiveTradeLock(_ context.Context, _ int) (bool, error) {
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
	default:
		return connection.Session{}, false
	}
}

func (sessionRegistryStub) Touch(string) error { return nil }

func (sessionRegistryStub) Remove(string) {}

func (sessionRegistryStub) ListAll() ([]connection.Session, error) {
	return []connection.Session{{ConnID: "conn1", UserID: 1}, {ConnID: "conn2", UserID: 2}}, nil
}

type transportStub struct {
	packetIDs []uint16
}

func (transport *transportStub) Send(_ string, packetID uint16, _ []byte) error {
	transport.packetIDs = append(transport.packetIDs, packetID)
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
	case domain.PermKick, domain.PermMute, domain.PermWarn, domain.PermAmbassador:
		return true, nil
	default:
		return false, nil
	}
}

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