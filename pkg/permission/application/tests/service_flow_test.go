package tests

import (
	"context"
	"strings"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkpermission "github.com/momlesstomato/pixel-sdk/events/permission"
	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	permissionstore "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/store"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestServiceResolveAccessAndPerks verifies multi-group access and perk resolution behavior.
func TestServiceResolveAccessAndPerks(t *testing.T) {
	service := openPermissionService(t)
	ctx := context.Background()
	defaultGroup, _ := service.CreateGroup(ctx, permissionapplication.CreateGroupInput{Name: "default", DisplayName: "Default", IsDefault: true, Priority: 1})
	adminGroup, _ := service.CreateGroup(ctx, permissionapplication.CreateGroupInput{Name: "admin", DisplayName: "Admin", Priority: 100})
	if _, err := service.AddPermissions(ctx, defaultGroup.Group.ID, []string{"perk.safe_chat"}); err != nil {
		t.Fatalf("expected default permissions success, got %v", err)
	}
	if _, err := service.AddPermissions(ctx, adminGroup.Group.ID, []string{"*", "role.ambassador"}); err != nil {
		t.Fatalf("expected admin permissions success, got %v", err)
	}
	if _, err := service.ReplaceUserGroups(ctx, 1, []int{defaultGroup.Group.ID, adminGroup.Group.ID}); err != nil {
		t.Fatalf("expected user assignment success, got %v", err)
	}
	access, err := service.ResolveAccess(ctx, 1)
	if err != nil || access.PrimaryGroup.ID != adminGroup.Group.ID {
		t.Fatalf("expected admin as primary group, got %+v err=%v", access, err)
	}
	if !access.PrimaryGroup.IsAmbassador {
		t.Fatalf("expected ambassador flag enabled through permission")
	}
	if granted, checkErr := service.HasPermission(ctx, 1, "room.enter.locked"); checkErr != nil || !granted {
		t.Fatalf("expected wildcard permission grant, granted=%v err=%v", granted, checkErr)
	}
	perks := service.ResolvePerks(access)
	if len(perks) == 0 {
		t.Fatalf("expected resolved perk list")
	}
}

// TestServiceReplaceUserGroupsCancellation verifies cancellable group-change event behavior.
func TestServiceReplaceUserGroupsCancellation(t *testing.T) {
	service := openPermissionService(t)
	ctx := context.Background()
	defaultGroup, _ := service.CreateGroup(ctx, permissionapplication.CreateGroupInput{Name: "default", DisplayName: "Default", IsDefault: true, Priority: 1})
	vipGroup, _ := service.CreateGroup(ctx, permissionapplication.CreateGroupInput{Name: "vip", DisplayName: "VIP", Priority: 10})
	service.SetEventFirer(func(event sdk.Event) {
		typed, ok := event.(*sdkpermission.UserGroupChanged)
		if ok && typed.NewGroupID == vipGroup.Group.ID {
			typed.Cancel()
		}
	})
	if _, err := service.ReplaceUserGroups(ctx, 1, []int{defaultGroup.Group.ID}); err != nil {
		t.Fatalf("expected baseline assignment success, got %v", err)
	}
	if _, err := service.ReplaceUserGroups(ctx, 1, []int{vipGroup.Group.ID}); err == nil {
		t.Fatalf("expected assignment cancellation error")
	}
	access, _ := service.ResolveAccess(ctx, 1)
	if access.PrimaryGroup.ID != defaultGroup.Group.ID {
		t.Fatalf("expected previous assignment to remain after cancellation")
	}
}

// openPermissionService creates a permission service backed by sqlite repository.
func openPermissionService(t *testing.T) *permissionapplication.Service {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&permissionmodel.Group{}, &permissionmodel.Grant{}, &permissionmodel.Assignment{}, &usermodel.Record{}); err != nil {
		t.Fatalf("expected sqlite migration success, got %v", err)
	}
	if err := database.Create(&usermodel.Record{Username: "user-a"}).Error; err != nil {
		t.Fatalf("expected user row insert success, got %v", err)
	}
	repository, _ := permissionstore.NewRepository(database)
	service, _ := permissionapplication.NewService(repository, nil, permissionapplication.Config{AmbassadorPermission: "role.ambassador"})
	return service
}
