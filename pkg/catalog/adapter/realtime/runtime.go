package realtime

import (
	"context"
	"fmt"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	catalogapp "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by catalog realtime runtime.
type Transport interface {
	// Send writes one encoded packet payload to one connection identifier.
	Send(string, uint16, []byte) error
}

// ClubOffersSender pushes subscription club offers to one connection.
type ClubOffersSender func(ctx context.Context, connID string, userID int) error

// InventoryItemSender pushes one purchased inventory item delta to one connection.
type InventoryItemSender func(ctx context.Context, connID string, userID int, itemID int) error

// Runtime defines catalog realm websocket packet behavior.
type Runtime struct {
	// service stores catalog application behavior.
	service *catalogapp.Service
	// sessions stores authenticated connection lookup behavior.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// logger stores runtime logging behavior.
	logger *zap.Logger
	// clubOffersSender pushes subscription offers when a club_buy page is served.
	clubOffersSender ClubOffersSender
	// inventoryItemSender pushes one purchased inventory delta to the buyer session.
	inventoryItemSender InventoryItemSender
}

// NewRuntime creates one catalog realtime runtime instance.
func NewRuntime(service *catalogapp.Service, sessions coreconnection.SessionRegistry, transport Transport, logger *zap.Logger) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("catalog service is required")
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

// SetClubOffersSender configures the optional callback invoked for club_buy catalog pages.
func (runtime *Runtime) SetClubOffersSender(fn ClubOffersSender) {
	runtime.clubOffersSender = fn
}

// SetInventoryItemSender configures the optional callback invoked after a successful purchase.
func (runtime *Runtime) SetInventoryItemSender(fn InventoryItemSender) {
	runtime.inventoryItemSender = fn
}

// Dispose releases per-connection resources.
func (runtime *Runtime) Dispose(_ string) {}

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
