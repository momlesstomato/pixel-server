package realtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	packetignore "github.com/momlesstomato/pixel-server/pkg/user/packet/ignore"
	packetname "github.com/momlesstomato/pixel-server/pkg/user/packet/name"
	packetprofile "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by user realtime runtime.
type Transport interface {
	// Send writes one encoded packet payload to one connection identifier.
	Send(string, uint16, []byte) error
}

// PacketIDs defines runtime-configurable packet mappings for variable packet IDs.
type PacketIDs struct {
	// SettingsRoomInvites stores user.settings_room_invites packet identifier.
	SettingsRoomInvites uint16
	// SettingsOldChat stores user.settings_old_chat packet identifier.
	SettingsOldChat uint16
	// Unignore stores user.unignore packet identifier.
	Unignore uint16
	// IgnoreByID stores user.ignore_id packet identifier.
	IgnoreByID uint16
	// ApproveName stores user.approve_name packet identifier.
	ApproveName uint16
}

// Options defines runtime configuration for user realtime behavior.
type Options struct {
	// Debounce stores settings persistence coalesce window.
	Debounce time.Duration
	// PacketIDs stores variable protocol packet identifiers.
	PacketIDs PacketIDs
	// Logger stores optional logger override.
	Logger *zap.Logger
}

// Runtime defines user realm websocket packet behavior.
type Runtime struct {
	// service stores user application behavior.
	service *userapplication.Service
	// sessions stores authenticated connection lookup behavior.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// debounce stores settings write coalesce window.
	debounce time.Duration
	// packetIDs stores variable protocol packet identifiers.
	packetIDs PacketIDs
	// logger stores runtime logging behavior.
	logger *zap.Logger
	// mutex guards pending settings state.
	mutex sync.Mutex
	// pending stores staged settings writes by connection identifier.
	pending map[string]*pendingSettings
}

// pendingSettings defines one staged settings write payload.
type pendingSettings struct {
	// userID stores owning user identifier.
	userID int
	// patch stores merged pending settings patch.
	patch settingsPatch
	// timer stores scheduled flush timer.
	timer *time.Timer
}

// NewRuntime creates one user realtime runtime instance.
func NewRuntime(service *userapplication.Service, sessions coreconnection.SessionRegistry, transport Transport, options Options) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("user service is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	debounce := options.Debounce
	if debounce <= 0 {
		debounce = 2 * time.Second
	}
	logger := options.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	packetIDs := options.PacketIDs
	if packetIDs.SettingsRoomInvites == 0 {
		packetIDs.SettingsRoomInvites = packetprofile.UserSettingsRoomInvitesPacketID
	}
	if packetIDs.SettingsOldChat == 0 {
		packetIDs.SettingsOldChat = packetprofile.UserSettingsOldChatPacketID
	}
	if packetIDs.Unignore == 0 {
		packetIDs.Unignore = packetignore.UserUnignorePacketIDDefault
	}
	if packetIDs.IgnoreByID == 0 {
		packetIDs.IgnoreByID = packetignore.UserIgnoreByIDPacketIDDefault
	}
	if packetIDs.ApproveName == 0 {
		packetIDs.ApproveName = packetname.UserApproveNamePacketIDDefault
	}
	return &Runtime{
		service: service, sessions: sessions, transport: transport, debounce: debounce,
		packetIDs: packetIDs, logger: logger, pending: map[string]*pendingSettings{},
	}, nil
}

// Dispose flushes staged settings writes for one disposed connection.
func (runtime *Runtime) Dispose(connID string) {
	runtime.flushPending(context.Background(), connID)
}
