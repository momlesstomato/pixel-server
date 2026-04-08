package seed

import (
	"testing"

	navmodel "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDemoAdminRoomOwnerBackfillMovesSeededAdminRooms verifies shipped admin demo rooms are reassigned to demo_admin and can be rolled back.
func TestDemoAdminRoomOwnerBackfillMovesSeededAdminRooms(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}, &navmodel.Room{}); err != nil {
		t.Fatalf("expected schema migration success, got %v", err)
	}
	testAdmin := usermodel.Record{Username: "test_admin"}
	demoAdmin := usermodel.Record{Username: "demo_admin"}
	testDefault := usermodel.Record{Username: "test_default"}
	if err := database.Create(&testAdmin).Error; err != nil {
		t.Fatalf("expected test_admin create success, got %v", err)
	}
	if err := database.Create(&demoAdmin).Error; err != nil {
		t.Fatalf("expected demo_admin create success, got %v", err)
	}
	if err := database.Create(&testDefault).Error; err != nil {
		t.Fatalf("expected test_default create success, got %v", err)
	}
	rooms := []navmodel.Room{
		{OwnerID: testAdmin.ID, OwnerName: testAdmin.Username, Name: "Inappropriate to hotel staff", Description: "Official welcome room", Tags: "welcome,official"},
		{OwnerID: testAdmin.ID, OwnerName: testAdmin.Username, Name: "Staff Office", Description: "Staff-only room", Tags: "staff"},
		{OwnerID: testAdmin.ID, OwnerName: testAdmin.Username, Name: "Custom Room", Description: "Custom description", Tags: "custom"},
		{OwnerID: testDefault.ID, OwnerName: testDefault.Username, Name: "Chill Room", Description: "Relax and chat", Tags: "chill,social"},
	}
	if err := database.Create(&rooms).Error; err != nil {
		t.Fatalf("expected room seed success, got %v", err)
	}
	step := Step03DemoAdminRoomOwnerBackfill()
	if err := step.Migrate(database); err != nil {
		t.Fatalf("expected owner backfill success, got %v", err)
	}
	assertRoomOwner(t, database, "Official welcome room", "welcome,official", demoAdmin.ID, demoAdmin.Username)
	assertRoomOwner(t, database, "Staff-only room", "staff", demoAdmin.ID, demoAdmin.Username)
	assertRoomOwner(t, database, "Custom description", "custom", testAdmin.ID, testAdmin.Username)
	assertRoomOwner(t, database, "Relax and chat", "chill,social", testDefault.ID, testDefault.Username)
	if err := step.Rollback(database); err != nil {
		t.Fatalf("expected owner backfill rollback success, got %v", err)
	}
	assertRoomOwner(t, database, "Official welcome room", "welcome,official", testAdmin.ID, testAdmin.Username)
	assertRoomOwner(t, database, "Staff-only room", "staff", testAdmin.ID, testAdmin.Username)
	assertRoomOwner(t, database, "Custom description", "custom", testAdmin.ID, testAdmin.Username)
	assertRoomOwner(t, database, "Relax and chat", "chill,social", testDefault.ID, testDefault.Username)
}

// assertRoomOwner verifies one room owner assignment by stable seeded markers.
func assertRoomOwner(t *testing.T, database *gorm.DB, description string, tags string, ownerID uint, ownerName string) {
	t.Helper()
	var room navmodel.Room
	query := database.Where("description = ? AND tags = ?", description, tags).Limit(1).Find(&room)
	if query.Error != nil {
		t.Fatalf("expected room query success, got %v", query.Error)
	}
	if query.RowsAffected == 0 {
		t.Fatalf("expected room with description %q and tags %q", description, tags)
	}
	if room.OwnerID != ownerID || room.OwnerName != ownerName {
		t.Fatalf("expected owner %d/%q, got %d/%q", ownerID, ownerName, room.OwnerID, room.OwnerName)
	}
}
