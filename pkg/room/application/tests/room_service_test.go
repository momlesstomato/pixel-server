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
	"golang.org/x/crypto/bcrypt"
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

// TestCheckAccess_Open verifies open rooms allow any user entry.
func TestCheckAccess_Open(t *testing.T) {
	svc := newService(t)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessOpen}
	assert.NoError(t, svc.CheckAccess(context.Background(), room, "", 99))
}

// TestCheckAccess_OwnerBypass verifies the room owner bypasses all access restrictions.
func TestCheckAccess_OwnerBypass(t *testing.T) {
	svc := newService(t)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked}
	assert.NoError(t, svc.CheckAccess(context.Background(), room, "", 10))
}

// TestCheckAccess_Locked_DeniedForNonOwner verifies locked rooms deny non-owners.
func TestCheckAccess_Locked_DeniedForNonOwner(t *testing.T) {
	svc := newService(t)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked}
	assert.ErrorIs(t, svc.CheckAccess(context.Background(), room, "", 99), domain.ErrAccessDenied)
}

// TestCheckAccess_Locked_RightsHolderBypasses verifies rights holders enter locked rooms without doorbell.
func TestCheckAccess_Locked_RightsHolderBypasses(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	rights := &rightsRepoStub{rights: map[[2]int]bool{{1, 42}: true}}
	svc, err := application.NewService(newModelRepo(), &banRepoStub{}, rights, mgr, zap.NewNop())
	require.NoError(t, err)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked}
	assert.NoError(t, svc.CheckAccess(context.Background(), room, "", 42))
}

// TestCheckAccess_Locked_NonRightsHolderDenied verifies users without rights still hit the doorbell.
func TestCheckAccess_Locked_NonRightsHolderDenied(t *testing.T) {
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcaster)
	rights := &rightsRepoStub{rights: map[[2]int]bool{{1, 42}: true}}
	svc, err := application.NewService(newModelRepo(), &banRepoStub{}, rights, mgr, zap.NewNop())
	require.NoError(t, err)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked}
	assert.ErrorIs(t, svc.CheckAccess(context.Background(), room, "", 99), domain.ErrAccessDenied)
}

// TestCheckAccess_Password_Valid verifies valid password admits non-owner entry.
func TestCheckAccess_Password_Valid(t *testing.T) {
	svc := newService(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	require.NoError(t, err)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessPassword, Password: string(hash)}
	assert.NoError(t, svc.CheckAccess(context.Background(), room, "secret", 99))
}

// TestCheckAccess_Password_Invalid verifies wrong password returns ErrInvalidPassword.
func TestCheckAccess_Password_Invalid(t *testing.T) {
	svc := newService(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	require.NoError(t, err)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessPassword, Password: string(hash)}
	assert.ErrorIs(t, svc.CheckAccess(context.Background(), room, "wrong", 99), domain.ErrInvalidPassword)
}

// TestFindRoom_NoRepository verifies ErrRoomNotFound when no repository is set.
func TestFindRoom_NoRepository(t *testing.T) {
	svc := newService(t)
	_, err := svc.FindRoom(context.Background(), 1)
	assert.ErrorIs(t, err, domain.ErrRoomNotFound)
}

// TestFindRoom_Found verifies room data is returned by the repository.
func TestFindRoom_Found(t *testing.T) {
	svc := newService(t)
	repo := &roomRepoStub{rooms: map[int]domain.Room{1: {ID: 1, Name: "Test"}}}
	svc.SetRoomRepository(repo)
	room, err := svc.FindRoom(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Test", room.Name)
}

// TestSaveSettings_OwnerCanSave verifies the room owner can update settings.
func TestSaveSettings_OwnerCanSave(t *testing.T) {
	svc := newService(t)
	repo := &roomRepoStub{rooms: map[int]domain.Room{1: {ID: 1, OwnerID: 10}}}
	svc.SetRoomRepository(repo)
	assert.NoError(t, svc.SaveSettings(context.Background(), 1, 10, domain.Room{Name: "New"}))
}

// TestSaveSettings_NonOwnerDenied verifies non-owner cannot update room settings.
func TestSaveSettings_NonOwnerDenied(t *testing.T) {
	svc := newService(t)
	repo := &roomRepoStub{rooms: map[int]domain.Room{1: {ID: 1, OwnerID: 10}}}
	svc.SetRoomRepository(repo)
	assert.ErrorIs(t, svc.SaveSettings(context.Background(), 1, 99, domain.Room{Name: "New"}), domain.ErrAccessDenied)
}
