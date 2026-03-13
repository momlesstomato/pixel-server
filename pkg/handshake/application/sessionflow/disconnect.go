package sessionflow

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	sdk "github.com/momlesstomato/pixel-sdk"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
)

// DisconnectUseCase defines connection-disconnect workflow behavior.
type DisconnectUseCase struct {
	// sessions stores and removes session lifecycle state.
	sessions SessionRegistry
	// transport closes target websocket connections.
	transport Transport
	// fire dispatches optional plugin lifecycle events.
	fire func(sdk.Event)
}

// NewDisconnectUseCase creates disconnect workflow behavior.
func NewDisconnectUseCase(sessions SessionRegistry, transport Transport) (*DisconnectUseCase, error) {
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	return &DisconnectUseCase{sessions: sessions, transport: transport}, nil
}

// SetEventFirer sets optional plugin event dispatch behavior.
func (useCase *DisconnectUseCase) SetEventFirer(fire func(sdk.Event)) {
	useCase.fire = fire
}

// Disconnect marks one session as disconnecting, removes it, and closes connection.
func (useCase *DisconnectUseCase) Disconnect(connID string) error {
	if connID == "" {
		return fmt.Errorf("connection id is required")
	}
	session, found := useCase.sessions.FindByConnID(connID)
	if found {
		if useCase.fire != nil {
			event := &sdk.SessionDisconnecting{ConnID: connID, UserID: session.UserID}
			useCase.fire(event)
			if event.Cancelled() {
				return nil
			}
		}
		session.State = coreconnection.StateDisconnecting
		if err := useCase.sessions.Register(session); err != nil {
			return err
		}
	}
	useCase.sessions.Remove(connID)
	return useCase.transport.Close(connID, websocket.CloseNormalClosure, "client disconnect")
}

// Cleanup removes session records for abrupt disconnect events.
func (useCase *DisconnectUseCase) Cleanup(connID string) {
	if connID == "" {
		return
	}
	useCase.sessions.Remove(connID)
}
