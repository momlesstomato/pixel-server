package plugin

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// API defines the plugin-safe runtime surface exposed by pixelsv.
type API interface {
	// Scope returns host runtime identity information.
	Scope() Scope
	// Events returns the in-process event bus for domain event hooks.
	Events() EventBus
	// Packets returns packet before/after interception hooks.
	Packets() PacketInterceptor
	// Rooms returns ECS-safe room operations or nil when unavailable for role.
	Rooms() RoomService
	// HTTP returns a route registrar under plugin-scoped API paths.
	HTTP() RouteRegistrar
	// Storage returns plugin-scoped key-value storage.
	Storage() PluginStore
	// Logger returns a logger scoped to plugin identity.
	Logger() *zap.Logger
	// Config returns plugin raw config bytes.
	Config() []byte
}

// EventBus defines plugin event publish and subscribe operations.
type EventBus interface {
	// On registers a handler for one event name.
	On(event string, handler EventHandler) Registration
	// Emit notifies listeners with the provided event.
	Emit(event *Event) error
}

// PacketInterceptor exposes before/after packet processing hooks.
type PacketInterceptor interface {
	// Before registers a pre-handler for one packet header identifier.
	Before(headerID uint16, hook PacketHook) Registration
	// After registers a post-handler for one packet header identifier.
	After(headerID uint16, hook PacketHook) Registration
	// BeforeAll registers a pre-handler for every packet.
	BeforeAll(hook PacketHook) Registration
	// AfterAll registers a post-handler for every packet.
	AfterAll(hook PacketHook) Registration
	// RunBefore executes pre-hooks and returns false when cancelled.
	RunBefore(ctx PacketContext) bool
	// RunAfter executes post-hooks after packet handling.
	RunAfter(ctx PacketContext)
}

// PacketHook handles one packet interception callback.
type PacketHook func(ctx PacketContext) bool

// PacketContext carries packet details to interception hooks.
type PacketContext struct {
	// SessionID is the sender session identifier.
	SessionID string
	// HeaderID is the binary packet header identifier.
	HeaderID uint16
	// Payload is the packet payload bytes.
	Payload []byte
	// Realm is the logical target realm for the packet.
	Realm string
}

// RouteRegistrar exposes plugin HTTP route registration.
type RouteRegistrar interface {
	// Group returns a router scoped under the plugin API prefix.
	Group() fiber.Router
}

// PluginStore defines plugin-scoped key-value operations.
type PluginStore interface {
	// Get returns the stored bytes for a key.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores bytes under one key with optional expiration.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes key data.
	Delete(ctx context.Context, key string) error
}
