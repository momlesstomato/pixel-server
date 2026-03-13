package hotelstatus

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/codec"
	corestatus "github.com/momlesstomato/pixel-server/core/status"
	statusredisstore "github.com/momlesstomato/pixel-server/pkg/status/infrastructure/redisstore"
	redislib "github.com/redis/go-redis/v9"
)

// TestServiceScheduleClosePublishesPackets verifies close scheduling and broadcast behavior.
func TestServiceScheduleClosePublishesPackets(t *testing.T) {
	service, messages, cleanup := createService(t)
	defer cleanup()
	now := time.Date(2026, time.March, 12, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	status, err := service.ScheduleClose(context.Background(), 5, 10, true)
	if err != nil || status.State != "closing" {
		t.Fatalf("expected closing schedule success, got %#v err=%v", status, err)
	}
	first := readPacketID(t, messages)
	second := readPacketID(t, messages)
	third := readPacketID(t, messages)
	if first != 1050 || second != 2771 || third != 1350 {
		t.Fatalf("unexpected packet sequence %d %d %d", first, second, third)
	}
}

// TestServiceTickTransitionsClosingAndClosedStates verifies countdown transitions.
func TestServiceTickTransitionsClosingAndClosedStates(t *testing.T) {
	service, messages, cleanup := createService(t)
	defer cleanup()
	now := time.Date(2026, time.March, 12, 11, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	_, err := service.ScheduleClose(context.Background(), 0, 1, true)
	if err != nil {
		t.Fatalf("expected schedule close success, got %v", err)
	}
	for index := 0; index < 3; index++ {
		_ = readPacketID(t, messages)
	}
	closed, err := service.Tick(context.Background())
	if err != nil || closed.State != "closed" {
		t.Fatalf("expected closed transition, got %#v err=%v", closed, err)
	}
	if packetID := readPacketID(t, messages); packetID != 3728 {
		t.Fatalf("expected closed_and_opens packet, got %d", packetID)
	}
	if packetID := readPacketID(t, messages); packetID != 4000 {
		t.Fatalf("expected disconnect_reason packet, got %d", packetID)
	}
	now = now.Add(2 * time.Minute)
	opened, err := service.Tick(context.Background())
	if err != nil || opened.State != "open" {
		t.Fatalf("expected reopen transition, got %#v err=%v", opened, err)
	}
}

// TestServiceTickClosedWithoutThrowUsersSkipsDisconnect verifies optional kick behavior.
func TestServiceTickClosedWithoutThrowUsersSkipsDisconnect(t *testing.T) {
	service, messages, cleanup := createService(t)
	defer cleanup()
	now := time.Date(2026, time.March, 12, 11, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	_, err := service.ScheduleClose(context.Background(), 0, 1, false)
	if err != nil {
		t.Fatalf("expected schedule close success, got %v", err)
	}
	for index := 0; index < 3; index++ {
		_ = readPacketID(t, messages)
	}
	_, err = service.Tick(context.Background())
	if err != nil {
		t.Fatalf("expected closed transition success, got %v", err)
	}
	if packetID := readPacketID(t, messages); packetID != 3728 {
		t.Fatalf("expected closed_and_opens packet, got %d", packetID)
	}
	select {
	case payload := <-messages:
		t.Fatalf("expected no additional packets, got %v", payload)
	case <-time.After(20 * time.Millisecond):
	}
}

// createService builds one service with Redis-backed store and local broadcaster.
func createService(t *testing.T) (*Service, <-chan []byte, func()) {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	store, _ := statusredisstore.NewStore(client, "hotel:status:test")
	broadcaster := broadcast.NewLocalBroadcaster()
	config := corestatus.Config{OpenHour: 0, OpenMinute: 0, CloseHour: 23, CloseMinute: 59, BroadcastChannel: "broadcast:all", CountdownTickSeconds: 60, DefaultMaintenanceDurationMinutes: 15}
	service, _ := NewService(store, broadcaster, config)
	messages, disposable, _ := broadcaster.Subscribe(context.Background(), "broadcast:all")
	cleanup := func() {
		_ = disposable.Dispose()
		_ = client.Close()
		server.Close()
	}
	return service, messages, cleanup
}

// readPacketID decodes one protocol frame and returns packet identifier.
func readPacketID(t *testing.T, messages <-chan []byte) uint16 {
	t.Helper()
	select {
	case payload := <-messages:
		frame, _, err := codec.DecodeFrame(payload)
		if err != nil {
			t.Fatalf("expected frame decode success, got %v", err)
		}
		return frame.PacketID
	case <-time.After(time.Second):
		t.Fatalf("expected packet payload delivery")
		return 0
	}
}
