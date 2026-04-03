package tests

import (
	"context"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func noopBroadcaster(_ int, _ []domain.RoomEntity, _ []byte) {}

func newModelRepo() *modelRepoStub {
	return &modelRepoStub{models: map[string]domain.RoomModel{
		"model_a": {
			Slug: "model_a", DoorX: 3, DoorY: 5, DoorDir: 2,
			Heightmap: "xxxx\rxxxx\rx00x\rxxxx",
		},
	}}
}

func newService(t *testing.T) *application.Service {
	t.Helper()
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	svc, err := application.NewService(newModelRepo(), &banRepoStub{}, &rightsRepoStub{}, mgr, zap.NewNop())
	require.NoError(t, err)
	return svc
}

// TestNewService_NilModelRepo verifies nil model repository is rejected.
func TestNewService_NilModelRepo(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	_, err := application.NewService(nil, &banRepoStub{}, &rightsRepoStub{}, mgr, zap.NewNop())
	assert.Error(t, err)
}

// TestNewService_NilBanRepo verifies nil ban repository is rejected.
func TestNewService_NilBanRepo(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	_, err := application.NewService(newModelRepo(), nil, &rightsRepoStub{}, mgr, zap.NewNop())
	assert.Error(t, err)
}

// TestNewService_NilRightsRepo verifies nil rights repository is rejected.
func TestNewService_NilRightsRepo(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	_, err := application.NewService(newModelRepo(), &banRepoStub{}, nil, mgr, zap.NewNop())
	assert.Error(t, err)
}

// TestNewService_NilManager verifies nil engine manager is rejected.
func TestNewService_NilManager(t *testing.T) {
	_, err := application.NewService(newModelRepo(), &banRepoStub{}, &rightsRepoStub{}, nil, zap.NewNop())
	assert.Error(t, err)
}

// TestLoadRoom_Success verifies room instance loads from model.
func TestLoadRoom_Success(t *testing.T) {
	svc := newService(t)
	defer svc.Manager().StopAll()
	room := domain.Room{ID: 1, ModelSlug: "model_a"}
	inst, err := svc.LoadRoom(context.Background(), room)
	require.NoError(t, err)
	assert.Equal(t, 1, inst.RoomID)
}

// TestLoadRoom_Cached verifies second load returns existing instance.
func TestLoadRoom_Cached(t *testing.T) {
	svc := newService(t)
	defer svc.Manager().StopAll()
	room := domain.Room{ID: 1, ModelSlug: "model_a"}
	inst1, _ := svc.LoadRoom(context.Background(), room)
	inst2, _ := svc.LoadRoom(context.Background(), room)
	assert.Equal(t, inst1, inst2)
}

// TestLoadRoom_InvalidModel verifies model not found error.
func TestLoadRoom_InvalidModel(t *testing.T) {
	svc := newService(t)
	room := domain.Room{ID: 1, ModelSlug: "nonexistent"}
	_, err := svc.LoadRoom(context.Background(), room)
	assert.ErrorIs(t, err, domain.ErrRoomModelNotFound)
}

// TestEnterRoom_Success verifies entity enters room instance.
func TestEnterRoom_Success(t *testing.T) {
	svc := newService(t)
	defer svc.Manager().StopAll()
	room := domain.Room{ID: 1, ModelSlug: "model_a"}
	inst, _ := svc.LoadRoom(context.Background(), room)
	time.Sleep(50 * time.Millisecond)
	entity := domain.NewPlayerEntity(0, 42, "c1", "user", "", "", "M", domain.Tile{})
	err := svc.EnterRoom(context.Background(), inst, &entity, 1, 42)
	require.NoError(t, err)
	assert.True(t, entity.VirtualID > 0)
}

// TestCheckBan_Active verifies active ban detection.
func TestCheckBan_Active(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	bans := &banRepoStub{banned: map[[2]int]bool{{1, 42}: true}}
	svc, _ := application.NewService(newModelRepo(), bans, &rightsRepoStub{}, mgr, zap.NewNop())
	assert.True(t, svc.CheckBan(context.Background(), 1, 42))
	assert.False(t, svc.CheckBan(context.Background(), 1, 99))
}

// TestHasRights_Granted verifies rights detection.
func TestHasRights_Granted(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	rights := &rightsRepoStub{rights: map[[2]int]bool{{1, 42}: true}}
	svc, _ := application.NewService(newModelRepo(), &banRepoStub{}, rights, mgr, zap.NewNop())
	assert.True(t, svc.HasRights(context.Background(), 1, 42))
	assert.False(t, svc.HasRights(context.Background(), 1, 99))
}
