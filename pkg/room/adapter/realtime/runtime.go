package realtime

import (
	"fmt"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	roomapplication "github.com/momlesstomato/pixel-server/pkg/room/application"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"go.uber.org/zap"
)

// Transport defines packet write behavior required by room realtime runtime.
type Transport interface {
	// Send writes one encoded packet to one connection identifier.
	Send(string, uint16, []byte) error
}

// Runtime defines room realm websocket packet behavior.
type Runtime struct {
	// service stores room application behavior.
	service *roomapplication.Service
	// entitySvc stores room entity mutation behavior.
	entitySvc *roomapplication.EntityService
	// chatSvc stores room chat behavior.
	chatSvc *roomapplication.ChatService
	// sessions stores authenticated connection lookup.
	sessions coreconnection.SessionRegistry
	// transport stores packet write behavior.
	transport Transport
	// logger stores runtime logging behavior.
	logger *zap.Logger
	// connRooms tracks which room each connection is in.
	connRooms map[string]int
}

// NewRuntime creates one room realtime runtime instance.
func NewRuntime(service *roomapplication.Service, entitySvc *roomapplication.EntityService, chatSvc *roomapplication.ChatService, sessions coreconnection.SessionRegistry, transport Transport, logger *zap.Logger) (*Runtime, error) {
	if service == nil {
		return nil, fmt.Errorf("room service is required")
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
	return &Runtime{
		service: service, entitySvc: entitySvc, chatSvc: chatSvc,
		sessions: sessions, transport: transport,
		logger: logger, connRooms: make(map[string]int),
	}, nil
}

// userID resolves authenticated user identifier for one connection.
func (rt *Runtime) userID(connID string) (int, bool) {
	session, found := rt.sessions.FindByConnID(connID)
	if !found || session.UserID <= 0 {
		return 0, false
	}
	return session.UserID, true
}

// sendPacket encodes and transmits one outgoing packet.
func (rt *Runtime) sendPacket(connID string, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return rt.transport.Send(connID, pkt.PacketID(), body)
}

// findEntityByConnID returns the room instance and entity for a connection.
func (rt *Runtime) findEntityByConnID(connID string, userID int) (*engine.Instance, *domain.RoomEntity) {
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil, nil
	}
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return nil, nil
	}
	entities := inst.Entities()
	for i := range entities {
		if entities[i].UserID == userID {
			return inst, &entities[i]
		}
	}
	return nil, nil
}

// Dispose releases per-connection resources.
func (rt *Runtime) Dispose(connID string) {
	delete(rt.connRooms, connID)
}
