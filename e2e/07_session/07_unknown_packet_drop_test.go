package session

import (
	"context"
	"net"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	packeterror "github.com/momlesstomato/pixel-server/pkg/session/packet/error"
	redislib "github.com/redis/go-redis/v9"
)

// noopUserRuntime is a user runtime that silently ignores all packets.
type noopUserRuntime struct{}

// Handle returns unhandled for every incoming packet.
func (noopUserRuntime) Handle(_ context.Context, _ string, _ uint16, _ []byte) (bool, error) {
	return false, nil
}

// Dispose is a no-op cleanup for the noop runtime.
func (noopUserRuntime) Dispose(_ string) {}

// Test07UnhandledAuthenticatedPacketIsDropped verifies that after authentication
// succeeds, packets with no registered handler are silently dropped rather than
// triggering a connection.error (1004) response or a protocol flood disconnect.
func Test07UnhandledAuthenticatedPacketIsDropped(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	client := redislib.NewClient(&redislib.Options{Addr: redisServer.Addr()})
	defer client.Close()
	broadcaster := broadcast.NewLocalBroadcaster()
	bus, _ := handshakerealtime.NewCloseSignalBus(broadcaster, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(client)
	handler, _ := handshakerealtime.NewHandler(
		ticketValidator{values: map[string]int{"valid-ticket": 7}},
		registry, nil, bus, nil, 3*time.Second,
	)
	handler.ConfigureUserRuntime(func(_ *handshakerealtime.Transport) (handshakerealtime.UserRuntime, error) {
		return noopUserRuntime{}, nil
	})
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "valid-ticket"})
	_ = testkit.ReadFrame(t, connection)
	_ = testkit.ReadFrame(t, connection)
	for i := 0; i < 11; i++ {
		if err := connection.WriteMessage(gws.BinaryMessage, codec.EncodeFrame(uint16(6000+i), []byte{})); err != nil {
			t.Fatalf("expected write success on packet %d, got %v", i, err)
		}
	}
	connection.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	for {
		_, payload, readErr := connection.ReadMessage()
		if readErr != nil {
			if netErr, ok := readErr.(net.Error); ok && netErr.Timeout() {
				return
			}
			t.Fatalf("expected deadline timeout (no 1004), got %v", readErr)
		}
		frame, _, decodeErr := codec.DecodeFrame(payload)
		if decodeErr != nil {
			continue
		}
		if frame.PacketID == packeterror.ConnectionErrorPacketID {
			t.Fatal("received unexpected connection.error 1004 for authenticated unknown packet")
		}
	}
}
