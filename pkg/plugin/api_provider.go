package plugin

import (
	"os"
	"path/filepath"

	"pixel-server/pkg/plugin/event"
	"pixel-server/pkg/plugin/intercept"
	"pixel-server/pkg/plugin/roomsvc"

	"go.uber.org/zap"
)

type wrappedAPI struct {
	inner    API
	onCancel func(cancel event.CancelFunc)
}

func wrapAPI(base API, onCancel func(cancel event.CancelFunc)) API {
	return &wrappedAPI{inner: base, onCancel: onCancel}
}

func (a *wrappedAPI) Scope() ServiceScope { return a.inner.Scope() }

func (a *wrappedAPI) Events() event.Bus {
	return &trackedEventBus{inner: a.inner.Events(), onCancel: a.onCancel}
}

func (a *wrappedAPI) Packets() intercept.Interceptor {
	return &trackedPacketInterceptor{inner: a.inner.Packets(), onCancel: a.onCancel}
}

func (a *wrappedAPI) Rooms() roomsvc.Service { return a.inner.Rooms() }

func (a *wrappedAPI) Logger() *zap.Logger { return a.inner.Logger() }

func (a *wrappedAPI) Config() []byte { return a.inner.Config() }

type trackedEventBus struct {
	inner    event.Bus
	onCancel func(cancel event.CancelFunc)
}

func (b *trackedEventBus) Subscribe(name event.Name, handler event.Handler) event.CancelFunc {
	cancel := b.inner.Subscribe(name, handler)
	b.onCancel(cancel)
	return cancel
}

func (b *trackedEventBus) Publish(e *event.Event) {
	b.inner.Publish(e)
}

type trackedPacketInterceptor struct {
	inner    intercept.Interceptor
	onCancel func(cancel event.CancelFunc)
}

func (i *trackedPacketInterceptor) Before(headerID uint16, fn intercept.HookFunc) event.CancelFunc {
	cancel := i.inner.Before(headerID, fn)
	i.onCancel(cancel)
	return cancel
}

func (i *trackedPacketInterceptor) After(headerID uint16, fn intercept.HookFunc) event.CancelFunc {
	cancel := i.inner.After(headerID, fn)
	i.onCancel(cancel)
	return cancel
}

func (i *trackedPacketInterceptor) RunBefore(ctx *intercept.PacketContext) {
	i.inner.RunBefore(ctx)
}

func (i *trackedPacketInterceptor) RunAfter(ctx *intercept.PacketContext) {
	i.inner.RunAfter(ctx)
}

type simpleAPI struct {
	scope       ServiceScope
	events      event.Bus
	interceptor intercept.Interceptor
	rooms       roomsvc.Service
	log         *zap.Logger
	configBytes []byte
}

func (a *simpleAPI) Scope() ServiceScope            { return a.scope }
func (a *simpleAPI) Events() event.Bus              { return a.events }
func (a *simpleAPI) Packets() intercept.Interceptor { return a.interceptor }
func (a *simpleAPI) Rooms() roomsvc.Service         { return a.rooms }
func (a *simpleAPI) Logger() *zap.Logger            { return a.log }
func (a *simpleAPI) Config() []byte                 { return a.configBytes }

// SimpleAPIProvider is a baseline provider used by tests and bootstrap phases.
type SimpleAPIProvider struct {
	Scope       ServiceScope
	Events      event.Bus
	Interceptor intercept.Interceptor
	Rooms       roomsvc.Service
	Log         *zap.Logger
}

// PluginAPI creates a plugin-scoped API with optional config loading.
func (p *SimpleAPIProvider) PluginAPI(name string, configDir string) API {
	logger := p.Log.With(zap.String("plugin", name))
	cfgPath := filepath.Join(configDir, name+".yml")
	configBytes, _ := os.ReadFile(cfgPath)
	rooms := p.Rooms
	if rooms == nil {
		rooms = roomsvc.NopService{}
	}

	return &simpleAPI{
		scope:       p.Scope,
		events:      p.Events,
		interceptor: p.Interceptor,
		rooms:       rooms,
		log:         logger,
		configBytes: configBytes,
	}
}
