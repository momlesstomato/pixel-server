package ws

import (
	"context"
	"testing"
	"time"

	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/protocol"
)

// TestHandleBinaryPublishesPacket validates decode and publish behavior.
func TestHandleBinaryPublishesPacket(t *testing.T) {
	bus := local.New()
	gateway, err := NewGateway(bus, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan transport.Message, 1)
	_, err = bus.Subscribe(ctx, transport.PacketC2STopic("handshake-security", "s1"), func(_ context.Context, message transport.Message) error {
		out <- message
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	writer := codec.NewWriter(64)
	packet := protocol.HandshakeReleaseVersionPacket{
		ReleaseVersion: "NITRO-1-6-6",
		ClientType:     "HTML5",
		Platform:       2,
		DeviceCategory: 1,
	}
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	frame := codec.EncodeFrame(protocol.HeaderHandshakeReleaseVersionPacket, writer.Bytes())
	if err := gateway.handleBinary(ctx, "s1", frame); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case message := <-out:
		if message.Topic != "packet.c2s.handshake-security.s1" {
			t.Fatalf("unexpected topic: %s", message.Topic)
		}
		if string(message.Payload) != string(frame[4:]) {
			t.Fatalf("unexpected payload")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected published packet")
	}
}

// TestStartSessionOutputForward validates session output fan-out writes.
func TestStartSessionOutputForward(t *testing.T) {
	bus := local.New()
	gateway, err := NewGateway(bus, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := gateway.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	connection := &stubConnection{}
	if err := gateway.Sessions().Register("s1", connection); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := bus.Publish(ctx, transport.SessionOutputTopic("s1"), []byte("out")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := string(connection.last); got != "out" {
		t.Fatalf("unexpected session output: %s", got)
	}
}

// stubConnection stores the last payload written by session manager.
type stubConnection struct {
	// last stores the latest payload write.
	last []byte
}

// WriteBinary stores one binary payload.
func (s *stubConnection) WriteBinary(payload []byte) error {
	s.last = append([]byte(nil), payload...)
	return nil
}

// Close implements session.Connection close semantics.
func (s *stubConnection) Close() error {
	return nil
}
