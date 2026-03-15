package tests

import (
	"context"
	"errors"
	"testing"

	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	permissionstore "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/store"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRepositoryGroupPermissionAndAssignmentFlow verifies core repository behavior.
func TestRepositoryGroupPermissionAndAssignmentFlow(t *testing.T) {
	database := openPermissionDatabase(t)
	repository, _ := permissionstore.NewRepository(database)
	ctx := context.Background()
	group, err := repository.CreateGroup(ctx, permissiondomain.Group{Name: "default", DisplayName: "Default", IsDefault: true, Priority: 1})
	if err != nil {
		t.Fatalf("expected group create success, got %v", err)
	}
	if err := repository.AddGroupPermissions(ctx, group.ID, []string{"perk.safe_chat", "perk.safe_chat", "room.enter"}); err != nil {
		t.Fatalf("expected permission add success, got %v", err)
	}
	permissions, err := repository.ListGroupPermissions(ctx, group.ID)
	if err != nil || len(permissions) != 2 {
		t.Fatalf("expected two deduplicated permissions, got %v err=%v", permissions, err)
	}
	if err := database.Create(&usermodel.Record{Username: "user-a"}).Error; err != nil {
		t.Fatalf("expected user create success, got %v", err)
	}
	if err := repository.ReplaceUserGroups(ctx, 1, []int{group.ID}); err != nil {
		t.Fatalf("expected assignment replace success, got %v", err)
	}
	count, err := repository.CountGroupUsers(ctx, group.ID)
	if err != nil || count != 1 {
		t.Fatalf("expected one assigned user, got count=%d err=%v", count, err)
	}
	if err := repository.DeleteGroup(ctx, group.ID); !errors.Is(err, permissiondomain.ErrCannotDeleteDefaultGroup) {
		t.Fatalf("expected default-group delete rejection, got %v", err)
	}
}

// TestRepositorySwitchDefaultGroup verifies default group switch behavior.
func TestRepositorySwitchDefaultGroup(t *testing.T) {
	database := openPermissionDatabase(t)
	repository, _ := permissionstore.NewRepository(database)
	ctx := context.Background()
	first, _ := repository.CreateGroup(ctx, permissiondomain.Group{Name: "default", DisplayName: "Default", IsDefault: true})
	second, _ := repository.CreateGroup(ctx, permissiondomain.Group{Name: "vip", DisplayName: "VIP", IsDefault: false})
	if err := repository.SwitchDefaultGroup(ctx, second.ID); err != nil {
		t.Fatalf("expected default switch success, got %v", err)
	}
	defaultGroup, err := repository.FindDefaultGroup(ctx)
	if err != nil || defaultGroup.ID != second.ID {
		t.Fatalf("expected second group as default, got %+v err=%v", defaultGroup, err)
	}
	updatedFirst, _ := repository.FindGroupByID(ctx, first.ID)
	if updatedFirst.IsDefault {
		t.Fatalf("expected first group default flag to be false")
	}
}

// openPermissionDatabase creates sqlite database with permission and user schemas.
func openPermissionDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&permissionmodel.Group{}, &permissionmodel.Grant{}, &permissionmodel.Assignment{}, &usermodel.Record{}); err != nil {
		t.Fatalf("expected sqlite migration success, got %v", err)
	}
	return database
}
