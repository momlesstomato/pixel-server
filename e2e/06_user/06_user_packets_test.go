package user

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
	userrealtime "github.com/momlesstomato/pixel-server/pkg/user/adapter/realtime"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	packetignore "github.com/momlesstomato/pixel-server/pkg/user/packet/ignore"
	packetname "github.com/momlesstomato/pixel-server/pkg/user/packet/name"
	packetprofileview "github.com/momlesstomato/pixel-server/pkg/user/packet/profileview"
	packetwardrobe "github.com/momlesstomato/pixel-server/pkg/user/packet/wardrobe"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test06UserRealtimePacketFlow verifies milestone 3 and 4 websocket packet handlers.
func Test06UserRealtimePacketFlow(t *testing.T) {
	database := openUserPacketDatabase(t)
	repository, _ := userstore.NewRepository(database)
	users, _ := userapplication.NewService(repository)
	actor, _ := users.Create(context.Background(), "actor")
	target, _ := users.Create(context.Background(), "target")
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
	bus, _ := handshakerealtime.NewRedisCloseSignalBus(redisClient, "handshake:test")
	registry, _ := coreconnection.NewRedisSessionRegistryWithOptions(redisClient, coreconnection.RedisSessionRegistryOptions{InstanceID: "pixel-server"})
	handler, _ := handshakerealtime.NewHandler(packetValidator{values: map[string]int{"ticket-1": actor.ID}}, registry, packetsecurity.NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32))), bus, nil, 2*time.Second)
	handler.ConfigurePostAuth(statusService, users, users, "pixel-server")
	handler.ConfigureUserRuntime(func(transport *handshakerealtime.Transport) (handshakerealtime.UserRuntime, error) {
		return userrealtime.NewRuntime(users, registry, transport, userrealtime.Options{})
	})
	connection, cleanup := testkit.StartWebSocket(t, handler.Handle)
	defer cleanup()
	testkit.SendPacket(t, connection, packetsecurity.ClientMachineIDPacket{MachineID: strings.Repeat("a", 64), Fingerprint: "x", Capabilities: "y"})
	_ = testkit.ReadFrameByID(t, connection, packetsecurity.ServerMachineIDPacketID)
	testkit.SendPacket(t, connection, packetsecurity.SSOTicketPacket{Ticket: "ticket-1"})
	_ = testkit.ReadFrameByID(t, connection, 3928)
	testkit.SendPacket(t, connection, packetprofileview.UserGetProfilePacket{UserID: int32(target.ID), OpenProfileWindow: true})
	_ = testkit.ReadFrameByID(t, connection, packetprofileview.UserProfilePacketID)
	testkit.SendPacket(t, connection, packetwardrobe.UserGetWardrobePacket{PageID: 1})
	_ = testkit.ReadFrameByID(t, connection, packetwardrobe.UserWardrobePagePacketID)
	testkit.SendPacket(t, connection, packetignore.UserIgnorePacket{Username: "target"})
	_ = testkit.ReadFrameByID(t, connection, packetignore.UserIgnoreResultPacketID)
	_ = testkit.ReadFrameByID(t, connection, packetignore.UserIgnoredUsersPacketID)
	testkit.SendPacket(t, connection, packetname.UserNameInputPacket{Name: "new_name"})
	_ = testkit.ReadFrameByID(t, connection, packetname.UserCheckNameResultPacketID)
}

// packetValidator defines deterministic ticket validation behavior.
type packetValidator struct{ values map[string]int }

// Validate resolves configured validator behavior.
func (validator packetValidator) Validate(_ context.Context, ticket string) (int, error) {
	userID, found := validator.values[ticket]
	if !found {
		return 0, errors.New("invalid ticket")
	}
	return userID, nil
}

// openUserPacketDatabase creates sqlite database with user packet schemas.
func openUserPacketDatabase(t *testing.T) *gorm.DB {
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
