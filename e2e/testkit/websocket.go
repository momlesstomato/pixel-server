package testkit

import (
	"net"
	"testing"
	"time"

	"github.com/gofiber/contrib/websocket"
	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/codec"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// Packet defines encoded packet write behavior for websocket tests.
type Packet interface {
	// PacketID returns packet identifier.
	PacketID() uint16
	// Encode serializes packet body payload.
	Encode() ([]byte, error)
}

// StartWebSocket starts one websocket route and returns connected client plus cleanup.
func StartWebSocket(t *testing.T, handler func(*websocket.Conn)) (*gws.Conn, func()) {
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
	cleanup := func() {
		_ = connection.Close()
		_ = module.Dispose()
		_ = <-serverErrors
	}
	return connection, cleanup
}

// SendPacket writes one packet as websocket binary frame.
func SendPacket(t *testing.T, connection *gws.Conn, packet Packet) {
	t.Helper()
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected packet encode success, got %v", err)
	}
	if err := connection.WriteMessage(gws.BinaryMessage, codec.EncodeFrame(packet.PacketID(), body)); err != nil {
		t.Fatalf("expected websocket write success, got %v", err)
	}
}

// ReadFrame reads one protocol frame from websocket stream.
func ReadFrame(t *testing.T, connection *gws.Conn) codec.Frame {
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

// ReadFrameByID reads websocket frames until target packet id is found.
func ReadFrameByID(t *testing.T, connection *gws.Conn, packetID uint16) codec.Frame {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		frame := ReadFrame(t, connection)
		if frame.PacketID == packetID {
			return frame
		}
	}
	t.Fatalf("expected packet id %d", packetID)
	return codec.Frame{}
}
