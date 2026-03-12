package authentication

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	gws "github.com/gorilla/websocket"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	redislib "github.com/redis/go-redis/v9"
)

// validatorMap defines deterministic ticket validation behavior.
type validatorMap struct{ values map[string]int }

// Validate resolves ticket values using map lookup.
func (validator validatorMap) Validate(_ context.Context, ticket string) (int, error) {
	userID, found := validator.values[ticket]
	if !found {
		return 0, errors.New("invalid ticket")
	}
	return userID, nil
}

// Test05ExpiredSSOClosesConnection verifies unauthorized close behavior for invalid tickets.
func Test05ExpiredSSOClosesConnection(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	client := redislib.NewClient(&redislib.Options{Addr: redisServer.Addr()})
	defer client.Close()
	bus, _ := handshakerealtime.NewRedisCloseSignalBus(client, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(client)
	handler, _ := handshakerealtime.NewHandler(validatorMap{values: map[string]int{}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("e", 32))), bus, nil, 2*time.Second)
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, connection)
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "expired"})
	frame := testkit.ReadFrame(t, connection)
	if frame.PacketID != 4000 {
		t.Fatalf("expected disconnect reason packet, got %d", frame.PacketID)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, _, err = connection.ReadMessage()
	if err == nil {
		t.Fatalf("expected unauthorized close")
	}
	var closeErr *gws.CloseError
	if errors.As(err, &closeErr) && closeErr.Code != authflow.UnauthorizedCloseCode {
		t.Fatalf("expected unauthorized close code %d, got %v", authflow.UnauthorizedCloseCode, err)
	}
}
