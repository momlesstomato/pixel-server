package room

import (
	"context"
	"fmt"
	"testing"

	roomapplication "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	roomstore "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/store"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupE2EWithRooms creates a test service backed by SQLite with a populated rooms table.
func setupE2EWithRooms(t *testing.T) (*roomapplication.Service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s_access?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err = db.AutoMigrate(&roommodel.RoomModel{}, &roommodel.RoomBan{}, &roommodel.RoomRight{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	createSQL := "CREATE TABLE IF NOT EXISTS rooms (" +
		"id INTEGER PRIMARY KEY AUTOINCREMENT," +
		"owner_id INTEGER NOT NULL DEFAULT 0," +
		"owner_name TEXT NOT NULL DEFAULT ''," +
		"name TEXT NOT NULL DEFAULT ''," +
		"description TEXT NOT NULL DEFAULT ''," +
		"state TEXT NOT NULL DEFAULT 'open'," +
		"category_id INTEGER NOT NULL DEFAULT 0," +
		"max_users INTEGER NOT NULL DEFAULT 25," +
		"score INTEGER NOT NULL DEFAULT 0," +
		"tags TEXT NOT NULL DEFAULT ''," +
		"trade_mode INTEGER NOT NULL DEFAULT 0," +
		"model_slug TEXT NOT NULL DEFAULT 'model_a'," +
		"password_hash TEXT NOT NULL DEFAULT ''," +
		"wall_height INTEGER NOT NULL DEFAULT -1," +
		"floor_thickness INTEGER NOT NULL DEFAULT 0," +
		"wall_thickness INTEGER NOT NULL DEFAULT 0," +
		"allow_pets BOOLEAN NOT NULL DEFAULT 1," +
		"allow_trading BOOLEAN NOT NULL DEFAULT 0)"
	if err = db.Exec(createSQL).Error; err != nil {
		t.Fatalf("create rooms table: %v", err)
	}
	modelStore, err := roomstore.NewModelStore(db)
	if err != nil { t.Fatalf("model store: %v", err) }
	banStore, err := roomstore.NewBanStore(db)
	if err != nil { t.Fatalf("ban store: %v", err) }
	rightsStore, err := roomstore.NewRightsStore(db)
	if err != nil { t.Fatalf("rights store: %v", err) }
	roomRepo, err := roomstore.NewRoomStore(db)
	if err != nil { t.Fatalf("room store: %v", err) }
	mgr := engine.NewManager(context.Background(), zap.NewNop(), noopBroadcast)
	svc, err := roomapplication.NewService(modelStore, banStore, rightsStore, mgr, zap.NewNop())
	if err != nil { t.Fatalf("new service: %v", err) }
	svc.SetRoomRepository(roomRepo)
	return svc, db
}

// Test13RoomAccessOpen verifies open rooms admit any user.
func Test13RoomAccessOpen(t *testing.T) {
	svc, _ := setupE2EWithRooms(t)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessOpen}
	if err := svc.CheckAccess(context.Background(), room, "", 99); err != nil {
		t.Fatalf("expected no error for open room, got: %v", err)
	}
}

// Test13RoomAccessOwnerBypass verifies room owners bypass all access restrictions.
func Test13RoomAccessOwnerBypass(t *testing.T) {
	svc, _ := setupE2EWithRooms(t)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked}
	if err := svc.CheckAccess(context.Background(), room, "", 10); err != nil {
		t.Fatalf("owner should bypass locked room, got: %v", err)
	}
}

// Test13RoomAccessLocked_NonOwnerDenied verifies locked rooms deny non-owners.
func Test13RoomAccessLocked_NonOwnerDenied(t *testing.T) {
	svc, _ := setupE2EWithRooms(t)
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessLocked}
	err := svc.CheckAccess(context.Background(), room, "", 99)
	if err != domain.ErrAccessDenied {
		t.Fatalf("expected ErrAccessDenied, got: %v", err)
	}
}

// Test13RoomAccessPassword_Valid verifies correct password grants entry.
func Test13RoomAccessPassword_Valid(t *testing.T) {
	svc, _ := setupE2EWithRooms(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil { t.Fatalf("hash: %v", err) }
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessPassword, Password: string(hash)}
	if err := svc.CheckAccess(context.Background(), room, "secret", 99); err != nil {
		t.Fatalf("valid password should pass, got: %v", err)
	}
}

// Test13RoomAccessPassword_Invalid verifies wrong password is rejected.
func Test13RoomAccessPassword_Invalid(t *testing.T) {
	svc, _ := setupE2EWithRooms(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil { t.Fatalf("hash: %v", err) }
	room := domain.Room{ID: 1, OwnerID: 10, State: domain.AccessPassword, Password: string(hash)}
	if err := svc.CheckAccess(context.Background(), room, "wrong", 99); err != domain.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got: %v", err)
	}
}

// Test13RoomSettings_OwnerCanSave verifies owner can persist room settings.
func Test13RoomSettings_OwnerCanSave(t *testing.T) {
	svc, db := setupE2EWithRooms(t)
	if err := db.Exec(`INSERT INTO rooms (id,owner_id,name,state) VALUES (10,42,'OldName','open')`).Error; err != nil {
		t.Fatalf("seed room: %v", err)
	}
	updated := domain.Room{Name: "NewName", State: domain.AccessOpen, MaxUsers: 25}
	if err := svc.SaveSettings(context.Background(), 10, 42, updated); err != nil {
		t.Fatalf("save settings: %v", err)
	}
}

// Test13RoomSettings_NonOwnerDenied verifies non-owner cannot update room settings.
func Test13RoomSettings_NonOwnerDenied(t *testing.T) {
	svc, db := setupE2EWithRooms(t)
	if err := db.Exec(`INSERT INTO rooms (id,owner_id,name,state) VALUES (11,42,'Room','open')`).Error; err != nil {
		t.Fatalf("seed room: %v", err)
	}
	updated := domain.Room{Name: "Hacked", State: domain.AccessOpen}
	err := svc.SaveSettings(context.Background(), 11, 99, updated)
	if err != domain.ErrAccessDenied {
		t.Fatalf("expected ErrAccessDenied for non-owner, got: %v", err)
	}
}
