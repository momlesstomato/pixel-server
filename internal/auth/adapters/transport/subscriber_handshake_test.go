package transport

import (
	"context"
	"testing"
	"time"

	"pixelsv/internal/auth/adapters/memory"
	"pixelsv/internal/auth/app"
	authmessaging "pixelsv/internal/auth/messaging"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/protocol"
)

// TestSubscriberHandshakePacketFlow validates diffie and machine-id packet responses.
func TestSubscriberHandshakePacketFlow(t *testing.T) {
	service := app.NewService(memory.NewTicketStore(), nil)
	bus := local.New()
	subscriber := NewSubscriber(bus, service, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := subscriber.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := make(chan []byte, 3)
	_, _ = bus.Subscribe(ctx, sessionmessaging.OutputTopic("s1"), func(_ context.Context, message coretransport.Message) error {
		output <- message.Payload
		return nil
	})
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, &protocol.HandshakeInitDiffiePacket{})); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertSingleFrameHeader(t, output, 1347)
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, &protocol.HandshakeCompleteDiffiePacket{EncryptedPublicKey: "7"})); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertSingleFrameHeader(t, output, 3885)
	if err := bus.Publish(ctx, authmessaging.PacketIngressTopic("s1"), encodeBody(t, &protocol.SecurityMachineIdPacket{MachineId: "~invalid", Fingerprint: "fp", Capabilities: "cap"})); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	payload := readOutput(t, output)
	frames, err := codec.SplitFrames(payload)
	if err != nil || len(frames) != 1 || frames[0].Header != 1488 {
		t.Fatalf("unexpected machine-id output")
	}
	reader := codec.NewReader(frames[0].Payload)
	machineID, err := reader.ReadString()
	if err != nil || len(machineID) != 64 {
		t.Fatalf("unexpected machine id payload")
	}
}

// assertSingleFrameHeader validates one output payload frame header.
func assertSingleFrameHeader(t *testing.T, output <-chan []byte, expected uint16) {
	t.Helper()
	payload := readOutput(t, output)
	frames, err := codec.SplitFrames(payload)
	if err != nil || len(frames) != 1 || frames[0].Header != expected {
		t.Fatalf("unexpected frame header: %d", expected)
	}
}

// readOutput reads one output payload with timeout.
func readOutput(t *testing.T, output <-chan []byte) []byte {
	t.Helper()
	select {
	case payload := <-output:
		return payload
	case <-time.After(time.Second):
		t.Fatalf("expected output payload")
	}
	return nil
}
