package session

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	corebroadcast "github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	corestatus "github.com/momlesstomato/pixel-server/core/status"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	statushotel "github.com/momlesstomato/pixel-server/pkg/status/application/hotelstatus"
	statusredisstore "github.com/momlesstomato/pixel-server/pkg/status/infrastructure/redisstore"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test07PostAuthBurstAndLoginStamp verifies post-auth burst and first-login-of-day behavior.
func Test07PostAuthBurstAndLoginStamp(t *testing.T) {
	database := openDatabase(t)
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	redisClient := redislib.NewClient(&redislib.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	broadcaster := corebroadcast.NewLocalBroadcaster()
	statusStore, _ := statusredisstore.NewStore(redisClient, "hotel:status:test")
	statusService, _ := statushotel.NewService(statusStore, broadcaster, corestatus.Config{OpenHour: 0, OpenMinute: 0, CloseHour: 23, CloseMinute: 59, BroadcastChannel: "broadcast:all", CountdownTickSeconds: 60, DefaultMaintenanceDurationMinutes: 15})
	userRepository, _ := userstore.NewRepository(database)
	users, _ := userapplication.NewService(userRepository)
	created, err := users.Create(context.Background(), "postauth-user")
	if err != nil {
		t.Fatalf("expected user create success, got %v", err)
	}
	bus, _ := handshakerealtime.NewRedisCloseSignalBus(redisClient, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistryWithOptions(redisClient, coreconnection.RedisSessionRegistryOptions{InstanceID: "pixel-server"})
	handler, _ := handshakerealtime.NewHandler(ticketValidator{values: map[string]int{"ticket-1": created.ID, "ticket-2": created.ID}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32))), bus, nil, 2*time.Second)
	handler.ConfigurePostAuth(statusService, users, users, "pixel-server")
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, connection)
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	first := []uint16{
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
	}
	if !equalIDs(first, []uint16{2491, 3523, 2033, 2725, 411, 2586, 3738, 513, 2875, 126, 793, 3928}) {
		t.Fatalf("unexpected first login packet sequence %v", first)
	}
	cleanup()
	connection, cleanup = testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrame(t, connection)
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-2"})
	second := []uint16{
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID, testkit.ReadFrame(t, connection).PacketID,
		testkit.ReadFrame(t, connection).PacketID,
	}
	if !equalIDs(second, []uint16{2491, 3523, 2033, 2725, 411, 2586, 3738, 513, 2875, 126, 3928}) {
		t.Fatalf("unexpected second login packet sequence %v", second)
	}
	var events []usermodel.LoginEvent
	if err := database.WithContext(context.Background()).Order("id asc").Find(&events).Error; err != nil {
		t.Fatalf("expected login event query success, got %v", err)
	}
	if len(events) != 2 || events[0].Holder != "pixel-server" || events[1].Holder != "pixel-server" {
		t.Fatalf("unexpected login event records %+v", events)
	}
}

// ticketValidator defines deterministic ticket validation behavior.
type ticketValidator struct{ values map[string]int }

// Validate resolves configured validator behavior.
func (validator ticketValidator) Validate(_ context.Context, ticket string) (int, error) {
	userID, found := validator.values[ticket]
	if !found {
		return 0, errors.New("invalid ticket")
	}
	return userID, nil
}

// openDatabase creates sqlite database with user and login event schemas.
func openDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite open success, got %v", err)
	}
	if err := database.AutoMigrate(&usermodel.Record{}, &usermodel.LoginEvent{}, &usermodel.Settings{}, &usermodel.Respect{}, &usermodel.WardrobeSlot{}, &usermodel.Ignore{}); err != nil {
		t.Fatalf("expected sqlite migration success, got %v", err)
	}
	return database
}

// equalIDs compares packet identifier slices.
func equalIDs(left []uint16, right []uint16) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
