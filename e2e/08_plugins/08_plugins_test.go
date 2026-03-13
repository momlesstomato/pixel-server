package plugins

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	"go.uber.org/zap"
)

// stubPlugin implements sdk.Plugin for end-to-end lifecycle testing.
type stubPlugin struct {
	// authEvents collects observed auth completed events.
	authEvents []*sdk.AuthCompleted
	// packetEvents collects observed packet received events.
	packetEvents []*sdk.PacketReceived
	// server stores the plugin api reference.
	server sdk.Server
}

// Manifest returns test plugin identity.
func (p *stubPlugin) Manifest() sdk.Manifest {
	return sdk.Manifest{Name: "e2e-stub", Author: "test", Version: "0.0.1"}
}

// Enable registers event subscriptions.
func (p *stubPlugin) Enable(server sdk.Server) error {
	p.server = server
	server.Events().Subscribe(func(e *sdk.AuthCompleted) {
		p.authEvents = append(p.authEvents, e)
	})
	server.Events().Subscribe(func(e *sdk.PacketReceived) {
		if e.PacketID == 9999 {
			p.packetEvents = append(p.packetEvents, e)
			e.Cancel()
		}
	})
	return nil
}

// Disable cleans up test plugin state.
func (p *stubPlugin) Disable() error {
	return nil
}

// Test08PluginEnableSubscribeFireVerify exercises the full plugin lifecycle.
func Test08PluginEnableSubscribeFireVerify(t *testing.T) {
	logger := zap.NewNop()
	dispatcher := coreplugin.NewDispatcher(logger)
	deps := coreplugin.ServerDependencies{}
	plug := &stubPlugin{}
	srv := coreplugin.NewServerImplForTest(plug.Manifest().Name, dispatcher, deps, logger)
	if err := plug.Enable(srv); err != nil {
		t.Fatalf("plugin enable failed: %v", err)
	}
	dispatcher.Fire(&sdk.ConnectionOpened{ConnID: "conn-a"})
	dispatcher.Fire(&sdk.AuthCompleted{ConnID: "conn-a", UserID: 42})
	dispatcher.Fire(&sdk.AuthCompleted{ConnID: "conn-b", UserID: 99})
	if len(plug.authEvents) != 2 {
		t.Fatalf("expected 2 auth events, got %d", len(plug.authEvents))
	}
	if plug.authEvents[0].UserID != 42 || plug.authEvents[1].UserID != 99 {
		t.Fatalf("unexpected user IDs: %d, %d", plug.authEvents[0].UserID, plug.authEvents[1].UserID)
	}
	blocked := &sdk.PacketReceived{ConnID: "conn-a", PacketID: 9999, Body: []byte{0x01}}
	dispatcher.Fire(blocked)
	if !blocked.Cancelled() {
		t.Fatalf("expected packet 9999 to be cancelled by plugin")
	}
	if len(plug.packetEvents) != 1 {
		t.Fatalf("expected 1 blocked packet event, got %d", len(plug.packetEvents))
	}
	allowed := &sdk.PacketReceived{ConnID: "conn-a", PacketID: 1234, Body: []byte{0x02}}
	dispatcher.Fire(allowed)
	if allowed.Cancelled() {
		t.Fatalf("expected packet 1234 to NOT be cancelled")
	}
	dispatcher.Fire(&sdk.ConnectionClosed{ConnID: "conn-a"})
	dispatcher.RemoveByOwner(plug.Manifest().Name)
	dispatcher.Fire(&sdk.AuthCompleted{ConnID: "conn-c", UserID: 50})
	if len(plug.authEvents) != 2 {
		t.Fatalf("expected no new events after owner removal, got %d", len(plug.authEvents))
	}
	_ = context.Background()
}
