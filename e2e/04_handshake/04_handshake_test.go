package handshake

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
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

// Test04HandshakeAuthenticatesConnection verifies machine-id and authentication packet flow.
func Test04HandshakeAuthenticatesConnection(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	client := redislib.NewClient(&redislib.Options{Addr: redisServer.Addr()})
	defer client.Close()
	bus, _ := handshakerealtime.NewRedisCloseSignalBus(client, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(client)
	handler, _ := handshakerealtime.NewHandler(validatorMap{values: map[string]int{"ticket-1": 7}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32))), bus, nil, 2*time.Second)
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: "~bad", Fingerprint: "x", Capabilities: "y"})
	machineFrame := testkit.ReadFrameByID(t, connection, packetsecurity.ServerMachineIDPacketID)
	if machineFrame.PacketID != packetsecurity.ServerMachineIDPacketID {
		t.Fatalf("expected machine id response packet")
	}
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	first := testkit.ReadFrame(t, connection)
	second := testkit.ReadFrame(t, connection)
	if first.PacketID != 2491 || second.PacketID != 3523 {
		t.Fatalf("expected authentication packets 2491/3523, got %d/%d", first.PacketID, second.PacketID)
	}
	if _, found := registry.FindByUserID(7); !found {
		t.Fatalf("expected authenticated session stored")
	}
}
