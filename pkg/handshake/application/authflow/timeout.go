package authflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

const defaultAuthTimeout = 30 * time.Second

// ErrAuthTimeoutElapsed defines auth-timeout expiration behavior.
var ErrAuthTimeoutElapsed = errors.New("authentication timeout elapsed")

// TimeoutUseCase defines auth timeout enforcement behavior.
type TimeoutUseCase struct {
	// transport closes connections when timeout expires.
	transport Transport
	// duration stores timeout length.
	duration time.Duration
	// after provides timer channel factory.
	after func(time.Duration) <-chan time.Time
}

// NewTimeoutUseCase creates auth-timeout behavior.
func NewTimeoutUseCase(transport Transport, duration time.Duration) (*TimeoutUseCase, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	normalized := duration
	if normalized <= 0 {
		normalized = defaultAuthTimeout
	}
	return &TimeoutUseCase{transport: transport, duration: normalized, after: time.After}, nil
}

// Wait closes connection when authentication signal is not received in time.
func (useCase *TimeoutUseCase) Wait(ctx context.Context, connID string, authenticated <-chan struct{}) error {
	select {
	case <-authenticated:
		return nil
	case <-ctx.Done():
		return nil
	case <-useCase.after(useCase.duration):
		packet := packetauth.DisconnectReasonPacket{Reason: packetauth.DisconnectReasonIdleNotAuthenticated}
		body, encodeErr := packet.Encode()
		if encodeErr == nil {
			_ = useCase.transport.Send(connID, packet.PacketID(), body)
		}
		if err := useCase.transport.Close(connID, AuthTimeoutCloseCode, "authentication timeout"); err != nil {
			return err
		}
		return ErrAuthTimeoutElapsed
	}
}
