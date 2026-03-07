package loader

import (
	"pixelsv/pkg/plugin"

	"go.uber.org/zap"
)

// fakePlugin is a test double for plugin lifecycle tests.
type fakePlugin struct {
	// metadata contains static plugin identity.
	metadata plugin.Metadata
	// enableErr is returned by OnEnable when set.
	enableErr error
	// disableErr is returned by OnDisable when set.
	disableErr error
	// enabled counts OnEnable invocations.
	enabled int
	// disabled counts OnDisable invocations.
	disabled int
}

// Metadata returns configured metadata.
func (f *fakePlugin) Metadata() plugin.Metadata {
	return f.metadata
}

// OnEnable increments enable counter and returns configured error.
func (f *fakePlugin) OnEnable(api plugin.API) error {
	f.enabled++
	return f.enableErr
}

// OnDisable increments disable counter and returns configured error.
func (f *fakePlugin) OnDisable() error {
	f.disabled++
	return f.disableErr
}

// fakeAPI is a no-op API implementation for lifecycle tests.
type fakeAPI struct{}

// Scope returns an empty runtime scope.
func (f fakeAPI) Scope() plugin.Scope {
	return plugin.Scope{}
}

// Events returns nil because lifecycle tests do not use EventBus.
func (f fakeAPI) Events() plugin.EventBus {
	return nil
}

// Packets returns nil because lifecycle tests do not use interceptors.
func (f fakeAPI) Packets() plugin.PacketInterceptor {
	return nil
}

// Rooms returns nil because lifecycle tests do not use room service.
func (f fakeAPI) Rooms() plugin.RoomService {
	return nil
}

// HTTP returns nil because lifecycle tests do not use HTTP registration.
func (f fakeAPI) HTTP() plugin.RouteRegistrar {
	return nil
}

// Storage returns nil because lifecycle tests do not use plugin storage.
func (f fakeAPI) Storage() plugin.PluginStore {
	return nil
}

// Logger returns a no-op logger.
func (f fakeAPI) Logger() *zap.Logger {
	return zap.NewNop()
}

// Config returns empty plugin configuration.
func (f fakeAPI) Config() []byte {
	return nil
}

// newFakePlugin builds a fake plugin with dependency metadata.
func newFakePlugin(name string, dependsOn ...string) *fakePlugin {
	return &fakePlugin{metadata: plugin.Metadata{Name: name, DependsOn: dependsOn}}
}
