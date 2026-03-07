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

// Test07HandshakeSecurityFullFlow validates the full handshake-security packet slice.
func Test07HandshakeSecurityFullFlow(t *testing.T) {
	bus := local.New()
	address := openLocalAddress(t)
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
	ticket, _, err := authRuntime.Service.CreateTicket(21, 60)
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
	waitHealth(t, address)
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+address+"/ws", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer conn.Close()
	sendPacket(t, conn, &protocol.HandshakeInitDiffiePacket{})
	assertOneHeader(t, conn, 1347)
	sendPacket(t, conn, &protocol.HandshakeCompleteDiffiePacket{EncryptedPublicKey: "7"})
	assertOneHeader(t, conn, 3885)
	sendPacket(t, conn, &protocol.SecurityMachineIdPacket{MachineId: "~invalid", Fingerprint: "fp", Capabilities: "cap"})
	assertOneHeader(t, conn, 1488)
	sendPacket(t, conn, &protocol.HandshakeReleaseVersionPacket{ReleaseVersion: "NITRO-1-6-6", ClientType: "HTML5", Platform: 2, DeviceCategory: 1})
	sendPacket(t, conn, &protocol.HandshakeClientVariablesPacket{ClientId: 7, ClientUrl: "https://localhost", ExternalVariablesUrl: "https://localhost/ext"})
	sendPacket(t, conn, &protocol.HandshakeClientLatencyMeasurePacket{})
	sendPacket(t, conn, &protocol.HandshakeClientPolicyPacket{})
	sendPacket(t, conn, &protocol.SecuritySsoTicketPacket{Ticket: ticket})
	select {
	case payload := <-authenticated:
		reader := codec.NewReader(payload)
		sessionID, _ := reader.ReadString()
		userID, _ := reader.ReadInt32()
		if sessionID != "1" || userID != 21 {
			t.Fatalf("unexpected auth payload")
		}
	case <-time.After(time.Second):
		t.Fatalf("expected authenticated event")
	}
	messageType, payload, err := conn.ReadMessage()
	if err != nil || messageType != websocket.BinaryMessage {
		t.Fatalf("expected binary output, got %v %d", err, messageType)
	}
	frames, err := codec.SplitFrames(payload)
	if err != nil || len(frames) != 2 || frames[0].Header != 2491 || frames[1].Header != 3523 {
		t.Fatalf("unexpected auth output frames")
	}
	cancel()
	select {
	case shutdownErr := <-errCh:
		if shutdownErr != nil && !errors.Is(shutdownErr, http.ErrServerClosed) {
			t.Fatalf("unexpected shutdown error: %v", shutdownErr)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected server shutdown")
	}
}

// openLocalAddress reserves an ephemeral local tcp address.
func openLocalAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	return address
}

// waitHealth blocks until HTTP health endpoint is reachable.
func waitHealth(t *testing.T, address string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("expected health endpoint up")
		}
		response, err := http.Get("http://" + address + "/health")
		if err == nil {
			_ = response.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// sendPacket writes one encoded packet as websocket binary message.
func sendPacket(t *testing.T, conn *websocket.Conn, packet protocol.Packet) {
	t.Helper()
	writer := codec.NewWriter(128)
	if err := packet.Encode(writer); err != nil {
		t.Fatalf("encode packet failed: %v", err)
	}
	frame := codec.EncodeFrame(packet.HeaderID(), writer.Bytes())
	if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		t.Fatalf("write packet failed: %v", err)
	}
}

// assertOneHeader reads one output message and validates single frame header.
func assertOneHeader(t *testing.T, conn *websocket.Conn, expected uint16) {
	t.Helper()
	messageType, payload, err := conn.ReadMessage()
	if err != nil || messageType != websocket.BinaryMessage {
		t.Fatalf("expected binary output, got %v %d", err, messageType)
	}
	frames, err := codec.SplitFrames(payload)
	if err != nil || len(frames) != 1 || frames[0].Header != expected {
		t.Fatalf("unexpected output header: %d", expected)
	}
}
