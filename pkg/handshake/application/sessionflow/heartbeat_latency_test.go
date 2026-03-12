package sessionflow

import (
	"context"
	"errors"
	"testing"
	"time"

	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
)

// flowTransportStub defines packet send and close capture behavior.
type flowTransportStub struct {
	// sent stores packet identifiers sent through transport.
	sent []uint16
	// closed stores closed connection identifiers.
	closed []string
}

// Send captures sent packet identifiers.
func (stub *flowTransportStub) Send(_ string, packetID uint16, _ []byte) error {
	stub.sent = append(stub.sent, packetID)
	return nil
}

// Close captures closed connection identifiers.
func (stub *flowTransportStub) Close(connID string, _ int, _ string) error {
	stub.closed = append(stub.closed, connID)
	return nil
}

// TestHeartbeatUseCaseSendsPingAndResetsOnPong verifies heartbeat ping/pong flow.
func TestHeartbeatUseCaseSendsPingAndResetsOnPong(t *testing.T) {
	transport := &flowTransportStub{}
	useCase, err := NewHeartbeatUseCase(transport, 10*time.Millisecond, 80*time.Millisecond)
	if err != nil {
		t.Fatalf("expected constructor success, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	pong := make(chan struct{}, 1)
	done := make(chan error, 1)
	go func() { done <- useCase.Run(ctx, "conn-1", pong) }()
	time.Sleep(20 * time.Millisecond)
	pong <- struct{}{}
	time.Sleep(20 * time.Millisecond)
	cancel()
	if runErr := <-done; runErr != nil {
		t.Fatalf("expected heartbeat stop success, got %v", runErr)
	}
	if len(transport.sent) == 0 {
		t.Fatalf("expected at least one ping packet")
	}
	if len(transport.closed) != 0 {
		t.Fatalf("expected no close calls, got %v", transport.closed)
	}
}

// TestHeartbeatUseCaseClosesOnPongTimeout verifies pong timeout behavior.
func TestHeartbeatUseCaseClosesOnPongTimeout(t *testing.T) {
	transport := &flowTransportStub{}
	useCase, err := NewHeartbeatUseCase(transport, 50*time.Millisecond, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("expected constructor success, got %v", err)
	}
	runErr := useCase.Run(context.Background(), "conn-1", make(chan struct{}))
	if !errors.Is(runErr, ErrPongTimeoutElapsed) {
		t.Fatalf("expected pong timeout error, got %v", runErr)
	}
	if len(transport.closed) != 1 || transport.closed[0] != "conn-1" {
		t.Fatalf("expected timeout close call, got %v", transport.closed)
	}
	if len(transport.sent) == 0 || transport.sent[0] != packetauth.DisconnectReasonPacketID {
		t.Fatalf("expected disconnect_reason packet before pong-timeout close, got %v", transport.sent)
	}
}

// TestLatencyUseCaseRespondsWithLatencyPacket verifies latency response behavior.
func TestLatencyUseCaseRespondsWithLatencyPacket(t *testing.T) {
	transport := &flowTransportStub{}
	useCase, err := NewLatencyUseCase(transport)
	if err != nil {
		t.Fatalf("expected constructor success, got %v", err)
	}
	if err := useCase.Respond("conn-1", 91); err != nil {
		t.Fatalf("expected response success, got %v", err)
	}
	if len(transport.sent) != 1 || transport.sent[0] != 10 {
		t.Fatalf("expected latency response packet id 10, got %v", transport.sent)
	}
}
