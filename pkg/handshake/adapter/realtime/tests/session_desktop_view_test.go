package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/application/navigation"
	packeterror "github.com/momlesstomato/pixel-server/pkg/session/packet/error"
	packetsnav "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	redislib "github.com/redis/go-redis/v9"
)

// TestDesktopViewWrongStateSendsConnectionError verifies desktop view state validation behavior.
func TestDesktopViewWrongStateSendsConnectionError(t *testing.T) {
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
	handler, _ := handshakerealtime.NewHandler(validatorMap{values: map[string]int{}}, registry, nil, bus, nil, 2*time.Second)
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsnav.DesktopViewRequestPacket{})
	frame := testkit.ReadFrameByID(t, connection, packeterror.ConnectionErrorPacketID)
	packet := packeterror.ConnectionErrorPacket{}
	if err := packet.Decode(frame.Body); err != nil || packet.MessageID != int32(packetsnav.DesktopViewRequestPacketID) || packet.ErrorCode != 3 {
		t.Fatalf("unexpected connection error payload %#v err=%v", packet, err)
	}
}

// TestDesktopViewInRoomSendsDesktopViewResponse verifies desktop view response behavior.
func TestDesktopViewInRoomSendsDesktopViewResponse(t *testing.T) {
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
	handler, _ := handshakerealtime.NewHandler(validatorMap{values: map[string]int{"ticket-1": 7}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32))), bus, nil, 2*time.Second)
	handler.ConfigureDesktopView(roomCheckerStub{inRoom: true})
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, connection)
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	_ = testkit.ReadFrame(t, connection)
	_ = testkit.ReadFrame(t, connection)
	testkit.SendPacket(t, connection, packetsnav.DesktopViewRequestPacket{})
	frame := testkit.ReadFrameByID(t, connection, packetsnav.DesktopViewResponsePacketID)
	if frame.PacketID != packetsnav.DesktopViewResponsePacketID {
		t.Fatalf("expected desktop view response packet")
	}
}

// roomCheckerStub defines deterministic room-presence behavior.
type roomCheckerStub struct {
	// inRoom stores deterministic room presence marker.
	inRoom bool
}

// IsInRoom returns deterministic room-presence output.
func (stub roomCheckerStub) IsInRoom(context.Context, int) (bool, error) {
	return stub.inRoom, nil
}

var _ sessionnavigation.RoomChecker = roomCheckerStub{}
