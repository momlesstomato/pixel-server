package plugin

import (
	"context"
	"fmt"
	"sync"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"go.uber.org/zap"
)

// ServerDependencies holds infrastructure dependencies for plugin API implementations.
type ServerDependencies struct {
	// Registry stores session lookup and lifecycle operations.
	Registry coreconnection.SessionRegistry
	// Broadcaster stores cross-instance messaging behavior.
	Broadcaster broadcast.Broadcaster
	// BroadcastChannel stores the broadcast-all channel name.
	BroadcastChannel string
}

// serverImpl implements sdk.Server for one plugin.
type serverImpl struct {
	logger   sdk.Logger
	events   *pluginEventBus
	sessions *pluginSessionAPI
	packets  *pluginPacketAPI
}

// newServerImpl creates a Server implementation for one plugin.
func newServerImpl(name string, dispatcher *Dispatcher, deps ServerDependencies, logger *zap.Logger) *serverImpl {
	return &serverImpl{
		logger:   &pluginLogger{sugar: logger.Sugar().Named("plugin." + name)},
		events:   &pluginEventBus{dispatcher: dispatcher, owner: name},
		sessions: &pluginSessionAPI{registry: deps.Registry},
		packets:  &pluginPacketAPI{broadcaster: deps.Broadcaster, channel: deps.BroadcastChannel, handlers: &sync.Map{}},
	}
}

// Logger returns a logger scoped to the calling plugin.
func (s *serverImpl) Logger() sdk.Logger { return s.logger }

// Events returns the event subscription API.
func (s *serverImpl) Events() sdk.EventBus { return s.events }

// Sessions returns the session query and control API.
func (s *serverImpl) Sessions() sdk.SessionAPI { return s.sessions }

// Packets returns the packet send and handler registration API.
func (s *serverImpl) Packets() sdk.PacketAPI { return s.packets }

// pluginLogger wraps zap.SugaredLogger behind sdk.Logger.
type pluginLogger struct {
	sugar *zap.SugaredLogger
}

// Printf writes an informational log entry.
func (l *pluginLogger) Printf(format string, args ...any) { l.sugar.Infof(format, args...) }

// Errorf writes an error log entry.
func (l *pluginLogger) Errorf(format string, args ...any) { l.sugar.Errorf(format, args...) }

// pluginEventBus wraps dispatcher for one plugin owner.
type pluginEventBus struct {
	dispatcher *Dispatcher
	owner      string
}

// Subscribe registers a handler for events of type T.
func (b *pluginEventBus) Subscribe(handler any, opts ...sdk.HandlerOption) func() {
	return b.dispatcher.Subscribe(b.owner, handler, opts...)
}

// pluginSessionAPI provides session query and control to plugins.
type pluginSessionAPI struct {
	registry coreconnection.SessionRegistry
}

// FindByUserID returns session info for an online user.
func (a *pluginSessionAPI) FindByUserID(userID int) (sdk.SessionInfo, bool) {
	s, found := a.registry.FindByUserID(userID)
	if !found {
		return sdk.SessionInfo{}, false
	}
	return mapSessionInfo(s), true
}

// FindByConnID returns session info for a connection.
func (a *pluginSessionAPI) FindByConnID(connID string) (sdk.SessionInfo, bool) {
	s, found := a.registry.FindByConnID(connID)
	if !found {
		return sdk.SessionInfo{}, false
	}
	return mapSessionInfo(s), true
}

// Kick disconnects a session with a reason code.
func (a *pluginSessionAPI) Kick(connID string, _ int32) error {
	a.registry.Remove(connID)
	return nil
}

// Count returns the number of sessions.
func (a *pluginSessionAPI) Count() int {
	sessions, err := a.registry.ListAll()
	if err != nil {
		return 0
	}
	return len(sessions)
}

// mapSessionInfo converts a core session to SDK session info.
func mapSessionInfo(s coreconnection.Session) sdk.SessionInfo {
	return sdk.SessionInfo{ConnID: s.ConnID, UserID: s.UserID, MachineID: s.MachineID, InstanceID: s.InstanceID}
}

// pluginPacketAPI provides packet injection and handler registration.
type pluginPacketAPI struct {
	broadcaster broadcast.Broadcaster
	channel     string
	handlers    *sync.Map
}

// Send writes an encoded packet to a specific connection.
func (a *pluginPacketAPI) Send(connID string, _ uint16, body []byte) error {
	return a.broadcaster.Publish(context.Background(), "broadcast:conn:"+connID, body)
}

// Broadcast sends a packet to all authenticated sessions.
func (a *pluginPacketAPI) Broadcast(_ uint16, body []byte) error {
	ch := a.channel
	if ch == "" {
		ch = "broadcast:all"
	}
	return a.broadcaster.Publish(context.Background(), ch, body)
}

// Handle registers a handler for a custom inbound packet ID.
func (a *pluginPacketAPI) Handle(packetID uint16, handler sdk.PacketHandler) error {
	if _, loaded := a.handlers.LoadOrStore(packetID, handler); loaded {
		return fmt.Errorf("handler already registered for packet ID %d", packetID)
	}
	return nil
}
