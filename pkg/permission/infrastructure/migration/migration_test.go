package migration

import (
	"testing"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestStep03UserPermissionGroupsBackfillsLegacyAssignments verifies migration backfill behavior.
func TestStep03UserPermissionGroupsBackfillsLegacyAssignments(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}, &permissionmodel.Group{}); err != nil {
		t.Fatalf("expected baseline migration success, got %v", err)
	}
	if err := database.Create(&permissionmodel.Group{ID: 1, Name: "default", DisplayName: "Default", IsDefault: true}).Error; err != nil {
		t.Fatalf("expected default group insert success, got %v", err)
	}
	user := usermodel.Record{Username: "user-a", GroupID: 1}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("expected legacy user insert success, got %v", err)
	}
	migrator := gormigrate.New(database, nil, []*gormigrate.Migration{Step03UserPermissionGroups()})
	if err := migrator.Migrate(); err != nil {
		t.Fatalf("expected migration success, got %v", err)
	}
	var assignments int64
	if err := database.Model(&permissionmodel.Assignment{}).Where("user_id = ? AND group_id = ?", user.ID, 1).Count(&assignments).Error; err != nil {
		t.Fatalf("expected assignment count query success, got %v", err)
	}
	if assignments != 1 {
		t.Fatalf("expected one backfilled assignment, got %d", assignments)
	}
}
