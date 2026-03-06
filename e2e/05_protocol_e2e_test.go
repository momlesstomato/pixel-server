package e2e_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	httpserver "pixelsv/pkg/http"
	"pixelsv/pkg/protocol"
)

// Test05WebSocketProtocolFlow validates websocket protocol ingress and egress.
func Test05WebSocketProtocolFlow(t *testing.T) {
	bus := local.New()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	address := ln.Addr().String()
	_ = ln.Close()
	cfg := httpserver.Config{Address: address, DisableStartupMessage: true, ReadTimeoutSeconds: 10, OpenAPIPath: "/openapi.json", SwaggerPath: "/swagger", APIKey: "secret"}
	server, err := httpserver.New(cfg, nil, bus)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe(ctx) }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("expected health endpoint up")
		}
		resp, err := http.Get("http://" + address + "/health")
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	ingress := make(chan transport.Message, 1)
	_, err = bus.Subscribe(ctx, transport.PacketC2STopic("handshake-security", "1"), func(_ context.Context, message transport.Message) error {
		ingress <- message
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+address+"/ws", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer conn.Close()
	writer := codec.NewWriter(64)
	packet := protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1}
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	inbound := codec.EncodeFrame(protocol.HeaderHandshakeReleaseVersionPacket, writer.Bytes())
	if err := conn.WriteMessage(websocket.BinaryMessage, inbound); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case message := <-ingress:
		if message.Topic != "packet.c2s.handshake-security.1" {
			t.Fatalf("unexpected topic: %s", message.Topic)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected ingress publish")
	}
	outbound := codec.EncodeFrame(1347, []byte("ok"))
	if err := bus.Publish(ctx, transport.SessionOutputTopic("1"), outbound); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if messageType != websocket.BinaryMessage {
		t.Fatalf("expected binary message type, got %d", messageType)
	}
	if string(payload) != string(outbound) {
		t.Fatalf("unexpected outbound payload")
	}
	cancel()
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("unexpected shutdown error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected server shutdown")
	}
}
