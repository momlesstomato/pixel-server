package sessionflow

import (
	"fmt"

	packettelemetry "github.com/momlesstomato/pixel-server/pkg/handshake/packet/telemetry"
)

// LatencyUseCase defines latency request/response workflow behavior.
type LatencyUseCase struct {
	// transport sends latency response packets to client.
	transport Transport
}

// NewLatencyUseCase creates latency workflow behavior.
func NewLatencyUseCase(transport Transport) (*LatencyUseCase, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	return &LatencyUseCase{transport: transport}, nil
}

// Respond echoes one latency request identifier back to connection.
func (useCase *LatencyUseCase) Respond(connID string, requestID int32) error {
	if connID == "" {
		return fmt.Errorf("connection id is required")
	}
	packet := packettelemetry.ClientLatencyResponsePacket{RequestID: requestID}
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return useCase.transport.Send(connID, packet.PacketID(), body)
}
