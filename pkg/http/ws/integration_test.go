package ws

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	wsmiddleware "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	authmessaging "pixelsv/internal/auth/messaging"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/protocol"
)

// TestGatewayWebSocketFlow validates websocket connection lifecycle and routing.
func TestGatewayWebSocketFlow(t *testing.T) {
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
	ingress := make(chan transport.Message, 1)
	disconnects := make(chan transport.Message, 1)
	_, _ = bus.Subscribe(ctx, authmessaging.PacketIngressTopic("1"), func(_ context.Context, message transport.Message) error {
		ingress <- message
		return nil
	})
	_, _ = bus.Subscribe(ctx, sessionmessaging.TopicDisconnected, func(_ context.Context, message transport.Message) error {
		disconnects <- message
		return nil
	})
	app := fiber.New()
	app.Use("/ws", gateway.UpgradeMiddleware)
	app.Get("/ws", wsmiddleware.New(gateway.HandleConnection))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- app.Listener(listener) }()
	defer func() {
		_ = app.Shutdown()
		select {
		case <-errCh:
		case <-time.After(time.Second):
		}
	}()
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+listener.Addr().String()+"/ws", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	writer := codec.NewWriter(64)
	packet := protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1}
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, codec.EncodeFrame(protocol.HeaderHandshakeReleaseVersionPacket, writer.Bytes())); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case <-ingress:
	case <-time.After(time.Second):
		t.Fatalf("expected ingress publish")
	}
	if count := gateway.Sessions().Count(); count != 1 {
		t.Fatalf("expected one active session, got %d", count)
	}
	if err := bus.Publish(ctx, sessionmessaging.OutputTopic("1"), []byte("out")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if messageType != websocket.BinaryMessage || string(payload) != "out" {
		t.Fatalf("unexpected websocket output")
	}
	_ = conn.Close()
	select {
	case message := <-disconnects:
		if string(message.Payload) != "1" {
			t.Fatalf("unexpected disconnect payload: %s", string(message.Payload))
		}
	case <-time.After(time.Second):
		t.Fatalf("expected disconnect event")
	}
	deadline := time.Now().Add(time.Second)
	for gateway.Sessions().Count() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("expected session removal")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
