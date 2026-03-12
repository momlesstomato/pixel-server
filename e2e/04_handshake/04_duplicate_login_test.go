package handshake

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

// duplicateValidatorMap defines deterministic ticket validation behavior.
type duplicateValidatorMap struct {
	// values maps ticket values to user identifiers.
	values map[string]int
}

// Validate resolves ticket values using map lookup.
func (validator duplicateValidatorMap) Validate(_ context.Context, ticket string) (int, error) {
	userID, found := validator.values[ticket]
	if !found {
		return 0, errors.New("invalid ticket")
	}
	return userID, nil
}

// Test04DuplicateLoginClosesPreviousConnection verifies duplicate-login kick behavior.
func Test04DuplicateLoginClosesPreviousConnection(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	client := redislib.NewClient(&redislib.Options{Addr: redisServer.Addr()})
	defer client.Close()
	bus, _ := handshakerealtime.NewRedisCloseSignalBus(client, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(client)
	handler, _ := handshakerealtime.NewHandler(duplicateValidatorMap{values: map[string]int{"ticket-a": 7, "ticket-b": 7}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("b", 32))), bus, nil, 2*time.Second)
	first, closeFirst := testkit.StartWebSocket(t, handler.Handle)
	defer closeFirst()
	testkit.SendPacket(t, first, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, first)
	testkit.SendPacket(t, first, packetsecurity.SSOTicketPacket{Ticket: "ticket-a"})
	_ = testkit.ReadFrame(t, first)
	_ = testkit.ReadFrame(t, first)
	second, closeSecond := testkit.StartWebSocket(t, handler.Handle)
	defer closeSecond()
	testkit.SendPacket(t, second, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("c", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, second)
	testkit.SendPacket(t, second, packetsecurity.SSOTicketPacket{Ticket: "ticket-b"})
	_ = testkit.ReadFrame(t, second)
	_ = testkit.ReadFrame(t, second)
	first.SetReadDeadline(time.Now().Add(time.Second))
	_, _, err = first.ReadMessage()
	if err == nil {
		t.Fatalf("expected duplicate login close on first connection")
	}
	var closeErr *gws.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != authflow.DuplicateLoginCloseCode {
		t.Fatalf("expected duplicate close code %d, got %v", authflow.DuplicateLoginCloseCode, err)
	}
}
