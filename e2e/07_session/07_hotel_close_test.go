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
	corestatus "github.com/momlesstomato/pixel-server/core/status"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	statushotel "github.com/momlesstomato/pixel-server/pkg/status/application/hotelstatus"
	statusredisstore "github.com/momlesstomato/pixel-server/pkg/status/infrastructure/redisstore"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	redislib "github.com/redis/go-redis/v9"
)

// Test07HotelCloseBroadcastDisconnect verifies hotel-close broadcast disconnect behavior.
func Test07HotelCloseBroadcastDisconnect(t *testing.T) {
	database := openDatabase(t)
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	redisClient := redislib.NewClient(&redislib.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	broadcaster := broadcast.NewLocalBroadcaster()
	bus, _ := handshakerealtime.NewCloseSignalBus(broadcaster, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistry(redisClient)
	userRepository, _ := userstore.NewRepository(database)
	users, _ := userapplication.NewService(userRepository)
	created, err := users.Create(context.Background(), "hotel-close-user")
	if err != nil {
		t.Fatalf("expected user create success, got %v", err)
	}
	handler, _ := handshakerealtime.NewHandler(ticketValidator{values: map[string]int{"ticket-1": created.ID}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32))), bus, nil, 2*time.Second)
	handler.ConfigureBroadcaster(broadcaster)
	statusStore, _ := statusredisstore.NewStore(redisClient, "hotel:status:test")
	statusService, _ := statushotel.NewService(statusStore, broadcaster, corestatus.Config{OpenHour: 0, OpenMinute: 0, CloseHour: 23, CloseMinute: 59, BroadcastChannel: "broadcast:all", CountdownTickSeconds: 60, DefaultMaintenanceDurationMinutes: 15})
	handler.ConfigurePostAuth(statusService, users, users, "pixel-server")
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, connection)
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	_ = testkit.ReadFrame(t, connection)
	_ = testkit.ReadFrame(t, connection)
	if _, err := statusService.ScheduleClose(context.Background(), 0, 1, true); err != nil {
		t.Fatalf("expected schedule close success, got %v", err)
	}
	if _, err := statusService.Tick(context.Background()); err != nil {
		t.Fatalf("expected status tick success, got %v", err)
	}
	disconnectFrame := testkit.ReadFrameByID(t, connection, packetauth.DisconnectReasonPacketID)
	reason := packetauth.DisconnectReasonPacket{}
	if err := reason.Decode(disconnectFrame.Body); err != nil || reason.Reason != packetauth.DisconnectReasonHotelClosed {
		t.Fatalf("unexpected disconnect payload %#v err=%v", reason, err)
	}
	connection.SetReadDeadline(time.Now().Add(time.Second))
	_, _, readErr := connection.ReadMessage()
	var closeErr *gws.CloseError
	if !errors.As(readErr, &closeErr) || closeErr.Code != gws.ClosePolicyViolation {
		t.Fatalf("expected hotel closed policy violation close, got %v", readErr)
	}
}
