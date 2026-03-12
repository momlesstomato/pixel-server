package sessionflow

import (
	"context"
	"fmt"
	"time"

	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packetsession "github.com/momlesstomato/pixel-server/pkg/handshake/packet/session"
)

const defaultHeartbeatInterval = 30 * time.Second
const defaultHeartbeatTimeout = 60 * time.Second

// HeartbeatUseCase defines heartbeat ping/pong workflow behavior.
type HeartbeatUseCase struct {
	// transport sends ping packets and closes timed-out sessions.
	transport Transport
	// interval stores ping cadence interval.
	interval time.Duration
	// timeout stores max allowed pong silence window.
	timeout time.Duration
}

// NewHeartbeatUseCase creates heartbeat workflow behavior.
func NewHeartbeatUseCase(transport Transport, interval time.Duration, timeout time.Duration) (*HeartbeatUseCase, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	pingEvery := interval
	if pingEvery <= 0 {
		pingEvery = defaultHeartbeatInterval
	}
	pongTimeout := timeout
	if pongTimeout <= 0 {
		pongTimeout = defaultHeartbeatTimeout
	}
	return &HeartbeatUseCase{transport: transport, interval: pingEvery, timeout: pongTimeout}, nil
}

// Run sends ping packets and closes session when pong timeout is exceeded.
func (useCase *HeartbeatUseCase) Run(ctx context.Context, connID string, pong <-chan struct{}) error {
	packet := packetsession.ClientPingPacket{}
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	ticker := time.NewTicker(useCase.interval)
	timer := time.NewTimer(useCase.timeout)
	defer ticker.Stop()
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-pong:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(useCase.timeout)
		case <-ticker.C:
			if sendErr := useCase.transport.Send(connID, packet.PacketID(), body); sendErr != nil {
				return sendErr
			}
		case <-timer.C:
			reasonPacket := packetauth.DisconnectReasonPacket{Reason: packetauth.DisconnectReasonPongTimeout}
			reasonBody, encodeErr := reasonPacket.Encode()
			if encodeErr == nil {
				_ = useCase.transport.Send(connID, reasonPacket.PacketID(), reasonBody)
			}
			if closeErr := useCase.transport.Close(connID, PongTimeoutCloseCode, "pong timeout"); closeErr != nil {
				return closeErr
			}
			return ErrPongTimeoutElapsed
		}
	}
}
