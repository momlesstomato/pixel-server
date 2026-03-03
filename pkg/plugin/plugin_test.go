package plugin_test

import (
	"testing"

	"pixel-server/pkg/plugin"
	"pixel-server/pkg/plugin/event"
	"pixel-server/pkg/plugin/intercept"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// EventBus tests

func TestEventBus_PublishReceived(t *testing.T) {
	bus := event.NewBus()
	var got *event.Event
	bus.Subscribe(event.PlayerJoined, func(e *event.Event) { got = e })

	bus.Publish(&event.Event{Name: event.PlayerJoined, Payload: map[string]any{"playerID": int64(1)}})
	require.NotNil(t, got)
	payload, ok := got.Payload.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, int64(1), payload["playerID"])
}

func TestEventBus_CancelUnsubscribes(t *testing.T) {
	bus := event.NewBus()
	called := 0
	cancel := bus.Subscribe(event.PlayerChat, func(e *event.Event) { called++ })
	bus.Publish(&event.Event{Name: event.PlayerChat})
	assert.Equal(t, 1, called)

	cancel()
	bus.Publish(&event.Event{Name: event.PlayerChat})
	assert.Equal(t, 1, called, "should not be called after cancel")
}

func TestEventBus_EventCancellation_StopsChain(t *testing.T) {
	bus := event.NewBus()
	second := false
	bus.Subscribe(event.PlayerChat, func(e *event.Event) { e.Cancel() })
	bus.Subscribe(event.PlayerChat, func(e *event.Event) { second = true })

	evt := &event.Event{Name: event.PlayerChat}
	bus.Publish(evt)

	assert.True(t, evt.IsCancelled())
	assert.False(t, second, "second handler must not run after cancellation")
}

func TestEventBus_MultipleEvents_Independent(t *testing.T) {
	bus := event.NewBus()
	calls := map[string]int{}
	bus.Subscribe(event.Name("a"), func(e *event.Event) { calls["a"]++ })
	bus.Subscribe(event.Name("b"), func(e *event.Event) { calls["b"]++ })

	bus.Publish(&event.Event{Name: event.Name("a")})
	bus.Publish(&event.Event{Name: event.Name("a")})
	bus.Publish(&event.Event{Name: event.Name("b")})

	assert.Equal(t, 2, calls["a"])
	assert.Equal(t, 1, calls["b"])
}

// PacketInterceptor tests

func TestPacketInterceptor_BeforeHook(t *testing.T) {
	pi := intercept.NewInterceptor()
	called := false
	pi.Before(0x01, func(ctx *intercept.PacketContext) { called = true })

	ctx := &intercept.PacketContext{HeaderID: 0x01}
	pi.RunBefore(ctx)
	assert.True(t, called)
}

func TestPacketInterceptor_Cancel_StopsChain(t *testing.T) {
	pi := intercept.NewInterceptor()
	second := false
	pi.Before(0x02, func(ctx *intercept.PacketContext) { ctx.Cancel = true })
	pi.Before(0x02, func(ctx *intercept.PacketContext) { second = true })

	ctx := &intercept.PacketContext{HeaderID: 0x02}
	pi.RunBefore(ctx)
	assert.True(t, ctx.Cancel)
	assert.False(t, second)
}

func TestPacketInterceptor_AfterHook_Separate(t *testing.T) {
	pi := intercept.NewInterceptor()
	before, after := false, false
	pi.Before(0x03, func(ctx *intercept.PacketContext) { before = true })
	pi.After(0x03, func(ctx *intercept.PacketContext) { after = true })

	ctx := &intercept.PacketContext{HeaderID: 0x03}
	pi.RunBefore(ctx)
	pi.RunAfter(ctx)
	assert.True(t, before)
	assert.True(t, after)
}

func TestPacketInterceptor_CancelFunc(t *testing.T) {
	pi := intercept.NewInterceptor()
	calls := 0
	cancel := pi.Before(0x04, func(ctx *intercept.PacketContext) { calls++ })
	pi.RunBefore(&intercept.PacketContext{HeaderID: 0x04})
	assert.Equal(t, 1, calls)

	cancel()
	pi.RunBefore(&intercept.PacketContext{HeaderID: 0x04})
	assert.Equal(t, 1, calls, "must not be called after cancel")
}

func TestPacketInterceptor_UnknownHeader_Noop(t *testing.T) {
	pi := intercept.NewInterceptor()
	ctx := &intercept.PacketContext{HeaderID: 0xFF}
	pi.RunBefore(ctx)
	pi.RunAfter(ctx)
}

// Registry tests

type mockPlugin struct {
	meta     plugin.Meta
	enabled  bool
	disabled bool
}

func (m *mockPlugin) Meta() plugin.Meta           { return m.meta }
func (m *mockPlugin) OnEnable(_ plugin.API) error { m.enabled = true; return nil }
func (m *mockPlugin) OnDisable() error            { m.disabled = true; return nil }

func newTestRegistry(t *testing.T) (*plugin.Registry, *mockPlugin) {
	t.Helper()
	provider := &plugin.SimpleAPIProvider{
		Events:      event.NewBus(),
		Interceptor: intercept.NewInterceptor(),
		Log:         zap.NewNop(),
	}
	mp := &mockPlugin{meta: plugin.Meta{Name: "TestPlugin", Version: "0.1.0"}}
	reg := plugin.NewRegistry(provider, t.TempDir(), provider.Log)
	reg.InjectPlugin(mp)
	return reg, mp
}

func TestRegistry_EnableAll(t *testing.T) {
	reg, mp := newTestRegistry(t)
	require.NoError(t, reg.EnableAll())
	assert.True(t, mp.enabled)
}

func TestRegistry_DisableAll(t *testing.T) {
	reg, mp := newTestRegistry(t)
	require.NoError(t, reg.EnableAll())
	reg.DisableAll()
	assert.True(t, mp.disabled)
}

func TestRegistry_List(t *testing.T) {
	reg, _ := newTestRegistry(t)
	list := reg.List()
	require.Len(t, list, 1)
	assert.Equal(t, "TestPlugin", list[0].Name)
}
