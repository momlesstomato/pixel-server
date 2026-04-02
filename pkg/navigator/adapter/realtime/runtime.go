package realtime

import (
	"fmt"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	navapp "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by navigator realtime runtime.
type Transport interface {
	// Send writes one encoded packet payload to one connection identifier.
	Send(string, uint16, []byte) error
}

// Runtime defines navigator realm websocket packet behavior.
type Runtime struct {
	// service stores navigator application behavior.
	service *navapp.Service
	// sessions stores authenticated connection lookup behavior.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// logger stores runtime logging behavior.
	logger *zap.Logger
}

// NewRuntime creates one navigator realtime runtime instance.
func NewRuntime(service *navapp.Service, sessions coreconnection.SessionRegistry, transport Transport, logger *zap.Logger) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("navigator service is required")
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
