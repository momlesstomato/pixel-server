package realtime

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	messengerapplication "github.com/momlesstomato/pixel-server/pkg/messenger/application"
	packetrequest "github.com/momlesstomato/pixel-server/pkg/messenger/packet/request"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by messenger realtime runtime.
type Transport interface {
	// Send writes one encoded packet payload to one connection identifier.
	Send(string, uint16, []byte) error
}

// Runtime defines messenger realm websocket packet behavior.
type Runtime struct {
	// service stores messenger application behavior.
	service *messengerapplication.Service
	// sessions stores authenticated connection lookup behavior.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// logger stores runtime logging behavior.
	logger *zap.Logger
}

// Options defines runtime configuration for messenger realtime behavior.
type Options struct {
	// Logger stores optional logger override.
	Logger *zap.Logger
}

// NewRuntime creates one messenger realtime runtime instance.
func NewRuntime(service *messengerapplication.Service, sessions coreconnection.SessionRegistry, broadcaster broadcast.Broadcaster, transport Transport, options Options) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("messenger service is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if broadcaster == nil {
		return nil, fmt.Errorf("broadcaster is required")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	logger := options.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Runtime{service: service, sessions: sessions, transport: transport, logger: logger}, nil
}

// Dispose cleans up flood state for one connection.
func (runtime *Runtime) Dispose(connID string) {
	runtime.service.ClearFloodState(connID)
	session, ok := runtime.sessions.FindByConnID(connID)
	if !ok {
		return
	}
	go runtime.notifyFriendsStatus(session.UserID, false)
}

// sendPacket encodes and sends one packet to one connection.
func (runtime *Runtime) sendPacket(connID string, packet interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return runtime.transport.Send(connID, packet.PacketID(), body)
}

// userID resolves authenticated user identifier for one connection.
func (runtime *Runtime) userID(connID string) (int, bool) {
	session, found := runtime.sessions.FindByConnID(connID)
	if !found || session.UserID <= 0 {
		return 0, false
	}
	return session.UserID, true
}

// Handle dispatches one authenticated messenger packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	return runtime.dispatch(ctx, connID, userID, packetID, body)
}

// publishFriendUpdate encodes one friend update and delivers it to one user.
// Delivery goes through the shared user channel consumed by the handshake runtime.
func (runtime *Runtime) publishFriendUpdate(ctx context.Context, toUserID int, entries []packetrequest.FriendUpdateEntry) {
	composer := packetrequest.MessengerFriendUpdateComposer{Entries: entries}
	body, err := composer.Encode()
	if err != nil {
		return
	}
	frame := codec.EncodeFrame(composer.PacketID(), body)
	_ = runtime.service.NotifyFriendUpdate(ctx, toUserID, frame)
}

// notifyFriendsStatus broadcasts an online/offline status change to all online friends.
func (runtime *Runtime) notifyFriendsStatus(userID int, online bool) {
	ctx := context.Background()
	var username, figure, motto string
	if profiles, err := runtime.service.GetUserProfiles(ctx, []int{userID}); err == nil && len(profiles) > 0 {
		username, figure, motto = profiles[0].Username, profiles[0].Figure, profiles[0].Motto
	}
	friends, err := runtime.service.ListFriends(ctx, userID)
	if err != nil {
		return
	}
	entry := packetrequest.FriendUpdateEntry{
		Action: 0, FriendID: int32(userID), Username: username,
		Online: online, Figure: figure, Motto: motto,
	}
	for _, f := range friends {
		if _, ok := runtime.sessions.FindByUserID(f.UserTwoID); ok {
			runtime.publishFriendUpdate(ctx, f.UserTwoID, []packetrequest.FriendUpdateEntry{entry})
		}
	}
}
