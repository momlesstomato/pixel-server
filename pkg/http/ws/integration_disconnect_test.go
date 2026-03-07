package ws

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	wsmiddleware "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
)

// TestGatewayWritesDisconnectReasonOnMalformedFrame validates runtime disconnect signaling.
func TestGatewayWritesDisconnectReasonOnMalformedFrame(t *testing.T) {
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
	disconnects := make(chan transport.Message, 1)
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
	defer conn.Close()
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte{0, 0, 0, 10, 0, 1}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	messageType, payload, err := conn.ReadMessage()
	if err != nil || messageType != websocket.BinaryMessage {
		t.Fatalf("expected disconnect frame, got %v %d", err, messageType)
	}
	frames, err := codec.SplitFrames(payload)
	if err != nil || len(frames) != 1 || frames[0].Header != 4000 {
		t.Fatalf("unexpected disconnect payload")
	}
	reader := codec.NewReader(frames[0].Payload)
	reason, err := reader.ReadInt32()
	if err != nil || reason != sessionmessaging.DisconnectReasonGeneric {
		t.Fatalf("unexpected disconnect reason: %d %v", reason, err)
	}
	select {
	case message := <-disconnects:
		if string(message.Payload) != "1" {
			t.Fatalf("unexpected disconnect payload: %s", string(message.Payload))
		}
	case <-time.After(time.Second):
		t.Fatalf("expected disconnect event")
	}
}
