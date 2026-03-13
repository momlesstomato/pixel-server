package navigation

import (
	"context"
	"fmt"

	packetsnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
)

// Transport defines packet transport behavior.
type Transport interface {
	// Send writes one encoded packet payload to one connection.
	Send(string, uint16, []byte) error
}

// RoomChecker defines room-presence lookup behavior.
type RoomChecker interface {
	// IsInRoom reports whether one user currently occupies one room.
	IsInRoom(context.Context, int) (bool, error)
}

// DesktopViewUseCase defines room-exit desktop view workflow behavior.
type DesktopViewUseCase struct {
	// transport sends desktop view response packets.
	transport Transport
	// checker resolves whether user is currently in room context.
	checker RoomChecker
}

// NewDesktopViewUseCase creates one desktop view use case.
func NewDesktopViewUseCase(transport Transport, checker RoomChecker) (*DesktopViewUseCase, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	return &DesktopViewUseCase{transport: transport, checker: checker}, nil
}

// Run sends desktop view response when user is currently inside one room.
func (useCase *DesktopViewUseCase) Run(ctx context.Context, connID string, userID int) error {
	if userID <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	inRoom := false
	var err error
	if useCase.checker != nil {
		inRoom, err = useCase.checker.IsInRoom(ctx, userID)
		if err != nil {
			return err
		}
	}
	if !inRoom {
		return nil
	}
	packet := packetsnavigation.DesktopViewResponsePacket{}
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return useCase.transport.Send(connID, packet.PacketID(), body)
}
