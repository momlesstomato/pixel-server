package seeds

import (
	"testing"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionseed "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/seed"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestStep03TestUsersCreatesBootstrapUsers verifies test user seed behavior.
func TestStep03TestUsersCreatesBootstrapUsers(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_loc=auto"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&permissionmodel.Group{}, &permissionmodel.Grant{}, &usermodel.Record{}); err != nil {
		t.Fatalf("expected schema migration success, got %v", err)
	}
	steps := []*gormigrate.Migration{
		permissionseed.Step01DefaultGroups(),
		permissionseed.Step02DefaultPermissions(),
		Step03TestUsers(),
	}
	seeder := gormigrate.New(database, nil, steps)
	if err := seeder.Migrate(); err != nil {
		t.Fatalf("expected seed migrate success, got %v", err)
	}
	var users []usermodel.Record
	if err := database.Find(&users).Error; err != nil {
		t.Fatalf("expected user query success, got %v", err)
	}
	expectedNames := map[string]string{"alice": "Alice", "bob": "Bob", "charlie": "Charlie", "dave": "Dave"}
	found := map[string]bool{}
	for _, user := range users {
		found[user.Username] = true
		if realName, ok := expectedNames[user.Username]; ok {
			if user.RealName != realName {
				t.Errorf("expected real name %q for %q, got %q", realName, user.Username, user.RealName)
			}
			if user.GroupID == 0 {
				t.Errorf("expected non-zero group id for user %q", user.Username)
			}
		}
	}
	for name := range expectedNames {
		if !found[name] {
			t.Errorf("expected test user %q to be seeded", name)
		}
	}
}

// TestStep03TestUsersRollbackRemovesUsers verifies test user rollback behavior.
func TestStep03TestUsersRollbackRemovesUsers(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_loc=auto"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&permissionmodel.Group{}, &permissionmodel.Grant{}, &usermodel.Record{}); err != nil {
		t.Fatalf("expected schema migration success, got %v", err)
	}
	steps := []*gormigrate.Migration{
		permissionseed.Step01DefaultGroups(),
		permissionseed.Step02DefaultPermissions(),
		Step03TestUsers(),
	}
	seeder := gormigrate.New(database, nil, steps)
	if err := seeder.Migrate(); err != nil {
		t.Fatalf("expected seed migrate success, got %v", err)
	}
	if err := Step03TestUsers().Rollback(database); err != nil {
		t.Fatalf("expected rollback success, got %v", err)
	}
	names := []string{"alice", "bob", "charlie", "dave"}
	for _, name := range names {
		var user usermodel.Record
		if err := database.Where("username = ?", name).First(&user).Error; err == nil {
			t.Errorf("expected user %q to be invisible after rollback", name)
		}
	}
}
