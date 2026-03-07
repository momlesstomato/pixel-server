package e2e_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"pixelsv/internal/auth"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	httpserver "pixelsv/pkg/http"
	"pixelsv/pkg/plugin/eventbus"
	"pixelsv/pkg/protocol"
)

// Test06AuthPhase3Flow validates auth realm registration and sso ticket handshake flow.
func Test06AuthPhase3Flow(t *testing.T) {
	bus := local.New()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	address := ln.Addr().String()
	_ = ln.Close()
	server, err := httpserver.New(httpserver.Config{Address: address, DisableStartupMessage: true, ReadTimeoutSeconds: 10, OpenAPIPath: "/openapi.json", SwaggerPath: "/swagger", APIKey: "secret"}, nil, bus)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authRuntime, err := auth.Register(ctx, server.App(), bus, eventbus.New(), nil, "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	ticket, _, err := authRuntime.Service.CreateTicket(17, 60)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	authenticated := make(chan []byte, 1)
	_, _ = bus.Subscribe(ctx, sessionmessaging.TopicAuthenticated, func(_ context.Context, message coretransport.Message) error {
		authenticated <- message.Payload
		return nil
	})
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe(ctx) }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("expected health endpoint up")
		}
		response, err := http.Get("http://" + address + "/health")
		if err == nil {
			_ = response.Body.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	connection, _, err := websocket.DefaultDialer.Dial("ws://"+address+"/ws", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer connection.Close()
	packet := &protocol.SecuritySsoTicketPacket{Ticket: ticket}
	release := &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1}
	if err := connection.WriteMessage(websocket.BinaryMessage, encodeFrame(t, release)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := connection.WriteMessage(websocket.BinaryMessage, encodeFrame(t, packet)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	select {
	case payload := <-authenticated:
		reader := codec.NewReader(payload)
		sessionID, _ := reader.ReadString()
		userID, _ := reader.ReadInt32()
		if sessionID != "1" || userID != 17 {
			t.Fatalf("unexpected auth payload")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected authenticated event")
	}
	messageType, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if messageType != websocket.BinaryMessage {
		t.Fatalf("unexpected websocket message type: %d", messageType)
	}
	frames, err := codec.SplitFrames(payload)
	if err != nil || len(frames) != 2 || frames[0].Header != 2491 || frames[1].Header != 3523 {
		t.Fatalf("unexpected auth output frame")
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

func encodeFrame(t *testing.T, packet protocol.Packet) []byte {
	t.Helper()
	writer := codec.NewWriter(64)
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	return codec.EncodeFrame(packet.HeaderID(), writer.Bytes())
}
