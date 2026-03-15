package tests

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkpermission "github.com/momlesstomato/pixel-sdk/events/permission"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	"go.uber.org/zap"
)

// TestPermissionAPIHasPermission verifies plugin permission API behavior.
func TestPermissionAPIHasPermission(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	server := coreplugin.NewServerImplForTest("test", dispatcher, coreplugin.ServerDependencies{Permissions: providerStub{}, EmitPermissionChecked: true}, zap.NewNop())
	var event *sdkpermission.PermissionChecked
	server.Events().Subscribe(func(value *sdkpermission.PermissionChecked) { event = value })
	granted := server.Permissions().HasPermission(7, "room.enter")
	if !granted {
		t.Fatalf("expected granted permission")
	}
	if event == nil || !event.Granted || event.UserID != 7 || event.Permission != "room.enter" {
		t.Fatalf("expected permission checked event payload, got %+v", event)
	}
}

// TestPermissionAPIGetGroup verifies plugin effective-group lookup behavior.
func TestPermissionAPIGetGroup(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	server := coreplugin.NewServerImplForTest("test", dispatcher, coreplugin.ServerDependencies{Permissions: providerStub{}}, zap.NewNop())
	group, found := server.Permissions().GetGroup(5)
	if !found || group.ID != 2 || group.Name != "vip" {
		t.Fatalf("expected effective group payload, got %+v found=%v", group, found)
	}
}

// providerStub defines deterministic plugin permission provider behavior.
type providerStub struct{}

// HasPermission returns deterministic permission checks.
func (providerStub) HasPermission(context.Context, int, string) (bool, error) {
	return true, nil
}

// EffectiveGroup returns deterministic group payload.
func (providerStub) EffectiveGroup(context.Context, int) (sdk.GroupInfo, bool, error) {
	return sdk.GroupInfo{ID: 2, Name: "vip", ClubLevel: 2}, true, nil
}
