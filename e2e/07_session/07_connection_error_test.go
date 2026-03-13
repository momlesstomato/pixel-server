package session

import (
	"errors"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packeterror "github.com/momlesstomato/pixel-server/pkg/session/packet/error"
	redislib "github.com/redis/go-redis/v9"
)

// Test07ConnectionErrorUnknownPacket verifies unknown packet protocol error behavior.
func Test07ConnectionErrorUnknownPacket(t *testing.T) {
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
	handler, _ := handshakerealtime.NewHandler(ticketValidator{values: map[string]int{}}, registry, nil, bus, nil, 2*time.Second)
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	if err := connection.WriteMessage(gws.BinaryMessage, codec.EncodeFrame(6553, []byte{})); err != nil {
		t.Fatalf("expected websocket write success, got %v", err)
	}
	frame := testkit.ReadFrameByID(t, connection, packeterror.ConnectionErrorPacketID)
	packet := packeterror.ConnectionErrorPacket{}
	if err := packet.Decode(frame.Body); err != nil {
		t.Fatalf("expected connection error decode success, got %v", err)
	}
	if packet.MessageID != 6553 || packet.ErrorCode != 1 {
		t.Fatalf("unexpected connection error payload %#v", packet)
	}
}

// Test07ConnectionErrorRateLimit verifies protocol error flood close behavior.
func Test07ConnectionErrorRateLimit(t *testing.T) {
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
	handler, _ := handshakerealtime.NewHandler(ticketValidator{values: map[string]int{}}, registry, nil, bus, nil, 3*time.Second)
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	for index := 0; index < 11; index++ {
		if err := connection.WriteMessage(gws.BinaryMessage, codec.EncodeFrame(uint16(6000+index), []byte{})); err != nil {
			t.Fatalf("expected websocket write success, got %v", err)
		}
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	for {
		_, _, readErr := connection.ReadMessage()
		if readErr == nil {
			continue
		}
		var closeErr *gws.CloseError
		if !errors.As(readErr, &closeErr) || closeErr.Code != gws.ClosePolicyViolation {
			if strings.Contains(readErr.Error(), "bad close code 1006") {
				return
			}
			t.Fatalf("expected protocol flood close behavior, got %v", readErr)
		}
		return
	}
}
