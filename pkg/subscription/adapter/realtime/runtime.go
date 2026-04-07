package realtime

import (
	"context"
	"fmt"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	subapp "github.com/momlesstomato/pixel-server/pkg/subscription/application"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by subscription realtime runtime.
type Transport interface {
	// Send writes one encoded packet payload to one connection identifier.
	Send(string, uint16, []byte) error
}

// InventoryItemSender pushes one delivered furniture item delta to one connection.
type InventoryItemSender func(ctx context.Context, connID string, userID int, itemID int) error

// Runtime defines subscription realm websocket packet behavior.
type Runtime struct {
	// service stores subscription application behavior.
	service *subapp.Service
	// sessions stores authenticated connection lookup behavior.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// logger stores runtime logging behavior.
	logger *zap.Logger
	// inventoryItemSender pushes one newly delivered HC gift item to the client.
	inventoryItemSender InventoryItemSender
}

// NewRuntime creates one subscription realtime runtime instance.
func NewRuntime(service *subapp.Service, sessions coreconnection.SessionRegistry, transport Transport, logger *zap.Logger) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("subscription service is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Runtime{service: service, sessions: sessions, transport: transport, logger: logger}, nil
}

// userID resolves authenticated user identifier for one connection.
func (runtime *Runtime) userID(connID string) (int, bool) {
	session, found := runtime.sessions.FindByConnID(connID)
	if !found {
		return 0, false
	}
	return session.UserID, true
}

// Dispose releases per-connection resources.
func (runtime *Runtime) Dispose(_ string) {}

// SetInventoryItemSender configures the optional callback invoked after a club gift claim.
func (runtime *Runtime) SetInventoryItemSender(fn InventoryItemSender) {
	runtime.inventoryItemSender = fn
}

// SendClubOffers pushes available club offers to one connection.
// Called by external runtimes (e.g. catalog) when a club_buy page is served.
func (runtime *Runtime) SendClubOffers(ctx context.Context, connID string) error {
	return runtime.handleGetClubOffers(ctx, connID, 0)
}

// sendPacket encodes and transmits one outgoing packet.
func (runtime *Runtime) sendPacket(connID string, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return runtime.transport.Send(connID, pkt.PacketID(), body)
}
