package realtime

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
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
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

// TestHandlerAuthenticatesConnection verifies machine-id and authentication packet flow.
func TestHandlerAuthenticatesConnection(t *testing.T) {
	server := startMiniRedis(t)
	defer server.Close()
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	defer client.Close()
	bus, _ := NewRedisCloseSignalBus(client, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(client)
	policy := packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32)))
	handler, _ := NewHandler(validatorMap{values: map[string]int{"ticket-1": 7}}, registry, policy, bus, nil, time.Second)
	connection, closeFn := startWebSocket(t, handler.Handle)
	defer closeFn()
	sendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: "~bad", Fingerprint: "x", Capabilities: "y"})
	if packetID := readPacketID(t, connection); packetID != packetsecurity.ServerMachineIDPacketID {
		t.Fatalf("expected machine_id response, got %d", packetID)
	}
	sendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	first := readPacketID(t, connection)
	second := readPacketID(t, connection)
	if first != 2491 || second != 3523 {
		t.Fatalf("expected authentication packets 2491/3523, got %d/%d", first, second)
	}
	if _, found := registry.FindByUserID(7); !found {
		t.Fatalf("expected authenticated session stored")
	}
}

// TestHandlerClosesOnAuthTimeout verifies auth timeout close behavior.
func TestHandlerClosesOnAuthTimeout(t *testing.T) {
	server := startMiniRedis(t)
	defer server.Close()
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	defer client.Close()
	bus, _ := NewRedisCloseSignalBus(client, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(client)
	handler, _ := NewHandler(validatorMap{values: map[string]int{}}, registry, packetsecurity.NewMachineIDPolicy(nil), bus, nil, 20*time.Millisecond)
	connection, closeFn := startWebSocket(t, handler.Handle)
	defer closeFn()
	packetID := readPacketID(t, connection)
	if packetID != 4000 {
		t.Fatalf("expected disconnect_reason packet, got %d", packetID)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, _, err := connection.ReadMessage()
	var closeErr *gws.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != authflow.AuthTimeoutCloseCode {
		t.Fatalf("expected auth timeout close code, got %v", err)
	}
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

// sendPacket writes one handshake packet as websocket binary frame.
func sendPacket(t *testing.T, connection *gws.Conn, packet interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) {
	t.Helper()
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected packet encode success, got %v", err)
	}
	if err := connection.WriteMessage(gws.BinaryMessage, codec.EncodeFrame(packet.PacketID(), body)); err != nil {
		t.Fatalf("expected websocket write success, got %v", err)
	}
}

// readPacketID reads one websocket binary message and decodes its first packet identifier.
func readPacketID(t *testing.T, connection *gws.Conn) uint16 {
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
	return frame.PacketID
}

// startMiniRedis creates one isolated Redis test server.
func startMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup, got %v", err)
	}
	return server
}
