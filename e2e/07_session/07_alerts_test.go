package session

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	gws "github.com/gorilla/websocket"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	packetalert "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
	redislib "github.com/redis/go-redis/v9"
)

// Test07TargetedAlertAndBanDisconnect verifies user-targeted alert and ban disconnect broadcast behavior.
func Test07TargetedAlertAndBanDisconnect(t *testing.T) {
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
	handler, _ := handshakerealtime.NewHandler(ticketValidator{values: map[string]int{"ticket-1": 7}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32))), bus, nil, 2*time.Second)
	handler.ConfigureBroadcaster(broadcaster)
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, connection)
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	_ = testkit.ReadFrame(t, connection)
	_ = testkit.ReadFrame(t, connection)
	service, _ := sessionnotification.NewService(broadcaster)
	if err := service.SendGenericAlert(context.Background(), 7, "hello from broadcast"); err != nil {
		t.Fatalf("expected generic alert publish success, got %v", err)
	}
	alertFrame := testkit.ReadFrameByID(t, connection, packetalert.GenericAlertPacketID)
	alert := packetalert.GenericAlertPacket{}
	if err := alert.Decode(alertFrame.Body); err != nil || alert.Message != "hello from broadcast" {
		t.Fatalf("unexpected alert payload %#v err=%v", alert, err)
	}
	if err := service.SendJustBannedDisconnect(context.Background(), 7); err != nil {
		t.Fatalf("expected just-banned disconnect publish success, got %v", err)
	}
	disconnectFrame := testkit.ReadFrameByID(t, connection, packetauth.DisconnectReasonPacketID)
	reason := packetauth.DisconnectReasonPacket{}
	if err := reason.Decode(disconnectFrame.Body); err != nil || reason.Reason != packetauth.DisconnectReasonJustBanned {
		t.Fatalf("unexpected disconnect payload %#v err=%v", reason, err)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, _, readErr := connection.ReadMessage()
	var closeErr *gws.CloseError
	if !errors.As(readErr, &closeErr) || closeErr.Code != gws.ClosePolicyViolation {
		t.Fatalf("expected ban close policy violation, got %v", readErr)
	}
}
