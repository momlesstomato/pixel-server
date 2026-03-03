package plugin

import (
	"pixel-server/pkg/plugin/event"
	"pixel-server/pkg/plugin/intercept"
	"pixel-server/pkg/plugin/roomsvc"

	"go.uber.org/zap"
)

// ServiceScope describes the hosting runtime where a plugin is enabled.
type ServiceScope struct {
	// Name is the service name (for example: "game", "auth", "catalog").
	Name string

	// NodeID is the instance identifier of the current process/pod.
	NodeID string

	// Version is the host service version string.
	Version string
}

// API is the server-facing interface injected into every plugin at OnEnable.
type API interface {
	// Scope returns service metadata describing where the plugin is running.
	Scope() ServiceScope

	// Events returns the in-process synchronous event bus.
	Events() event.Bus

	// Packets returns the packet interceptor registry for C2S/S2C hooks.
	Packets() intercept.Interceptor

	// Rooms returns a read-oriented room facade.
	Rooms() roomsvc.Service

	// Logger returns a structured logger namespaced to the plugin.
	Logger() *zap.Logger

	// Config returns raw bytes from the optional <plugin-name>.yml file.
	Config() []byte
}
