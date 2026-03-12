package tests

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gofiber/contrib/websocket"
	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	packetsession "github.com/momlesstomato/pixel-server/pkg/handshake/packet/session"
	packettelemetry "github.com/momlesstomato/pixel-server/pkg/handshake/packet/telemetry"
	redislib "github.com/redis/go-redis/v9"
)

// validatorMap defines deterministic ticket validation behavior.
type validatorMap struct {
	// values maps ticket values to user identifiers.
	values map[string]int
}

// Validate resolves ticket values using map lookup.
func (validator validatorMap) Validate(_ context.Context, ticket string) (int, error) {
	userID, found := validator.values[ticket]
	if !found {
		return 0, errors.New("invalid ticket")
	}
	return userID, nil
}

// packetWire defines packet serialization behavior for websocket frame writes.
type packetWire interface {
	// PacketID returns protocol packet identifier.
	PacketID() uint16
	// Encode serializes packet body payload.
	Encode() ([]byte, error)
}

// TestMilestone6FullHandshakeFlow verifies auth, heartbeat, and latency behavior.
func TestMilestone6FullHandshakeFlow(t *testing.T) {
	handler, cleanup := createHandler(t, map[string]int{"ticket-1": 7}, 300*time.Millisecond, 15*time.Millisecond, 70*time.Millisecond)
	defer cleanup()
	connection, closeConnection := startWebSocket(t, handler.Handle)
	defer closeConnection()
	sendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: "~bad", Fingerprint: "x", Capabilities: "y"})
	if frame := readFrame(t, connection); frame.PacketID != packetsecurity.ServerMachineIDPacketID {
		t.Fatalf("expected machine id response, got %d", frame.PacketID)
	}
	sendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	if frame := readFrame(t, connection); frame.PacketID != 2491 {
		t.Fatalf("expected authentication ok packet, got %d", frame.PacketID)
	}
	if frame := readFrame(t, connection); frame.PacketID != 3523 {
		t.Fatalf("expected identity accounts packet, got %d", frame.PacketID)
	}
	pingFrame := readFrameByID(t, connection, packetsession.ClientPingPacketID)
	pingPacket := packetsession.ClientPingPacket{}
	if err := pingPacket.Decode(pingFrame.Body); err != nil {
		t.Fatalf("expected ping decode success, got %v", err)
	}
	sendPacket(t, connection, packetsession.ClientPongPacket{})
	sendPacket(t, connection, packettelemetry.ClientLatencyTestPacket{RequestID: 91})
	latencyFrame := readFrameByID(t, connection, packettelemetry.ClientLatencyResponsePacketID)
	responsePacket := packettelemetry.ClientLatencyResponsePacket{}
	if err := responsePacket.Decode(latencyFrame.Body); err != nil {
		t.Fatalf("expected latency response decode success, got %v", err)
	}
	if responsePacket.RequestID != 91 {
		t.Fatalf("expected latency response id 91, got %d", responsePacket.RequestID)
	}
}

// TestMilestone6DuplicateLoginKick verifies duplicate login close behavior.
func TestMilestone6DuplicateLoginKick(t *testing.T) {
	handler, cleanup := createHandler(t, map[string]int{"ticket-a": 7, "ticket-b": 7}, 300*time.Millisecond, time.Second, 2*time.Second)
	defer cleanup()
	first, closeFirst := startWebSocket(t, handler.Handle)
	defer closeFirst()
	sendPacket(t, first, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = readFrame(t, first)
	sendPacket(t, first, packetsecurity.SSOTicketPacket{Ticket: "ticket-a"})
	_ = readFrame(t, first)
	_ = readFrame(t, first)
	second, closeSecond := startWebSocket(t, handler.Handle)
	defer closeSecond()
	sendPacket(t, second, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("b", 64), Fingerprint: "x", Capabilities: "y"})
	_ = readFrame(t, second)
	sendPacket(t, second, packetsecurity.SSOTicketPacket{Ticket: "ticket-b"})
	_ = readFrame(t, second)
	_ = readFrame(t, second)
	first.SetReadDeadline(time.Now().Add(time.Second))
	_, _, err := first.ReadMessage()
	if err == nil {
		t.Fatalf("expected duplicate login close on first connection")
	}
	var closeErr *gws.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != authflow.DuplicateLoginCloseCode {
		t.Fatalf("expected duplicate login close code %d, got %v", authflow.DuplicateLoginCloseCode, err)
	}
}

// TestMilestone6ExpiredSSORejected verifies unauthorized close behavior.
func TestMilestone6ExpiredSSORejected(t *testing.T) {
	handler, cleanup := createHandler(t, map[string]int{}, 300*time.Millisecond, 20*time.Millisecond, 80*time.Millisecond)
	defer cleanup()
	connection, closeConnection := startWebSocket(t, handler.Handle)
	defer closeConnection()
	sendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("c", 64), Fingerprint: "x", Capabilities: "y"})
	_ = readFrame(t, connection)
	sendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "expired-ticket"})
	if frame := readFrame(t, connection); frame.PacketID != 4000 {
		t.Fatalf("expected disconnect_reason packet, got %d", frame.PacketID)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, _, err := connection.ReadMessage()
	if err == nil {
		t.Fatalf("expected unauthorized close")
	}
	var closeErr *gws.CloseError
	if errors.As(err, &closeErr) {
		if closeErr.Code != authflow.UnauthorizedCloseCode {
			t.Fatalf("expected unauthorized close code %d, got %v", authflow.UnauthorizedCloseCode, err)
		}
		return
	}
	if !strings.Contains(err.Error(), "bad close code 1006") {
		t.Fatalf("expected unauthorized bad close code behavior for 1006, got %v", err)
	}
}

// createHandler builds a fully wired realtime handler for end-to-end tests.
func createHandler(t *testing.T, tickets map[string]int, authTimeout time.Duration, heartbeatInterval time.Duration, heartbeatTimeout time.Duration) (*realtime.Handler, func()) {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	bus, err := realtime.NewRedisCloseSignalBus(client, "handshake:test")
	if err != nil {
		t.Fatalf("expected close signal bus creation success, got %v", err)
	}
	registry, err := coreconnection.NewRedisSessionRegistry(client)
	if err != nil {
		t.Fatalf("expected session registry creation success, got %v", err)
	}
	policy := packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32)))
	handler, err := realtime.NewHandlerWithHeartbeat(validatorMap{values: tickets}, registry, policy, bus, nil, authTimeout, heartbeatInterval, heartbeatTimeout)
	if err != nil {
		t.Fatalf("expected handler creation success, got %v", err)
	}
	return handler, func() { _ = client.Close(); server.Close() }
}

// startWebSocket starts websocket server and returns connected client plus cleanup.
func startWebSocket(t *testing.T, handler func(*websocket.Conn)) (*gws.Conn, func()) {
	t.Helper()
	module := corehttp.New(corehttp.Options{})
	if err := module.RegisterWebSocket("/ws", handler); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected listener creation success, got %v", err)
	}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- module.App().Listener(listener) }()
	connection, _, err := gws.DefaultDialer.Dial("ws://"+listener.Addr().String()+"/ws", nil)
	if err != nil {
		t.Fatalf("expected websocket dial success, got %v", err)
	}
	return connection, func() { _ = connection.Close(); _ = module.Dispose(); _ = <-serverErrors }
}

// sendPacket writes one packet as websocket binary frame.
func sendPacket(t *testing.T, connection *gws.Conn, packet packetWire) {
	t.Helper()
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected packet encode success, got %v", err)
	}
	if err := connection.WriteMessage(gws.BinaryMessage, codec.EncodeFrame(packet.PacketID(), body)); err != nil {
		t.Fatalf("expected websocket write success, got %v", err)
	}
}

// readFrame reads one protocol frame from websocket stream.
func readFrame(t *testing.T, connection *gws.Conn) codec.Frame {
	t.Helper()
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("expected websocket read success, got %v", err)
	}
	frame, _, err := codec.DecodeFrame(payload)
	if err != nil {
		t.Fatalf("expected frame decode success, got %v", err)
	}
	return frame
}

// readFrameByID reads websocket frames until target packet id is found.
func readFrameByID(t *testing.T, connection *gws.Conn, packetID uint16) codec.Frame {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		frame := readFrame(t, connection)
		if frame.PacketID == packetID {
			return frame
		}
	}
	t.Fatalf("expected packet id %d", packetID)
	return codec.Frame{}
}
