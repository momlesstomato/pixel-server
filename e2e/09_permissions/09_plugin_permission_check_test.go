package permissions

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	"go.uber.org/zap"
)

// Test09PluginPermissionCheck verifies custom plugin permission checks through sdk permissions API.
func Test09PluginPermissionCheck(t *testing.T) {
	dispatcher := coreplugin.NewDispatcher(zap.NewNop())
	server := coreplugin.NewServerImplForTest("permission-plugin", dispatcher, coreplugin.ServerDependencies{Permissions: denyProvider{}}, zap.NewNop())
	plugin := &permissionPlugin{}
	if err := plugin.Enable(server); err != nil {
		t.Fatalf("expected plugin enable success, got %v", err)
	}
	event := &sdk.PacketReceived{ConnID: "c1", PacketID: 123, Body: []byte{1}}
	dispatcher.Fire(event)
	if !event.Cancelled() {
		t.Fatalf("expected packet cancellation when permission is denied")
	}
}

// permissionPlugin defines plugin behavior for permission-check testing.
type permissionPlugin struct{}

// Manifest returns plugin metadata.
func (permissionPlugin) Manifest() sdk.Manifest {
	return sdk.Manifest{Name: "permission-plugin", Author: "test", Version: "0.0.1"}
}

// Enable subscribes packet handlers guarded by permission checks.
func (permissionPlugin) Enable(server sdk.Server) error {
	server.Events().Subscribe(func(event *sdk.PacketReceived) {
		if event.PacketID == 123 && !server.Permissions().HasPermission(1, "custom.feature") {
			event.Cancel()
		}
	})
	return nil
}

// Disable completes plugin shutdown lifecycle.
func (permissionPlugin) Disable() error { return nil }

// denyProvider defines deterministic denied permission behavior.
type denyProvider struct{}

// HasPermission returns denied checks.
func (denyProvider) HasPermission(context.Context, int, string) (bool, error) {
	return false, nil
}

// EffectiveGroup returns deterministic missing group behavior.
func (denyProvider) EffectiveGroup(context.Context, int) (sdk.GroupInfo, bool, error) {
	return sdk.GroupInfo{}, false, nil
}
