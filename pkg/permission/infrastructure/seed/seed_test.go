package seed

import (
	"testing"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDefaultGroupAndPermissionSeeds verifies essential seed behavior.
func TestDefaultGroupAndPermissionSeeds(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&permissionmodel.Group{}, &permissionmodel.Grant{}); err != nil {
		t.Fatalf("expected schema migration success, got %v", err)
	}
	seeder := gormigrate.New(database, nil, []*gormigrate.Migration{Step01DefaultGroups(), Step02DefaultPermissions()})
	if err := seeder.Migrate(); err != nil {
		t.Fatalf("expected seed migrate success, got %v", err)
	}
	var groupCount int64
	if err := database.Model(&permissionmodel.Group{}).Count(&groupCount).Error; err != nil {
		t.Fatalf("expected group count query success, got %v", err)
	}
	if groupCount != 4 {
		t.Fatalf("expected four seeded groups, got %d", groupCount)
	}
	var grantCount int64
	if err := database.Model(&permissionmodel.Grant{}).Count(&grantCount).Error; err != nil {
		t.Fatalf("expected grant count query success, got %v", err)
	}
	if grantCount == 0 {
		t.Fatalf("expected seeded grants")
	}
}
