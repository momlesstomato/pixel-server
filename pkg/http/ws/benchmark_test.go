package ws

import (
	"context"
	"testing"

	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/protocol"
)

// BenchmarkHandleBinary measures websocket decode-and-publish overhead.
func BenchmarkHandleBinary(b *testing.B) {
	bus := local.New()
	gateway, err := NewGateway(bus, nil)
	if err != nil {
		b.Fatalf("expected no error, got %v", err)
	}
	ctx := context.Background()
	_, err = bus.Subscribe(ctx, transport.PacketC2STopic("handshake-security", "s1"), func(context.Context, transport.Message) error { return nil })
	if err != nil {
		b.Fatalf("expected no error, got %v", err)
	}
	writer := codec.NewWriter(64)
	packet := protocol.HandshakeReleaseVersionPacket{
		ReleaseVersion: "NITRO-1-6-6",
		ClientType:     "HTML5",
		Platform:       2,
		DeviceCategory: 1,
	}
	if err := packet.Encode(writer); err != nil {
		b.Fatalf("expected no error, got %v", err)
	}
	frame := codec.EncodeFrame(protocol.HeaderHandshakeReleaseVersionPacket, writer.Bytes())
	b.ReportAllocs()
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		if err := gateway.handleBinary(ctx, "s1", frame); err != nil {
			b.Fatalf("expected no error, got %v", err)
		}
	}
}
