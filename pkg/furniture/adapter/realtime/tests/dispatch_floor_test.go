package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/furniture/adapter/realtime"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
)

// foundRepoStub returns a fixed item and definition for deterministic test behavior.
type foundRepoStub struct {
	repoStub
	// item stores the fixed item returned by FindItemByID.
	item furnituredomain.Item
	// def stores the fixed definition returned by FindDefinitionByID.
	def furnituredomain.Definition
	// defErr stores an optional error for FindDefinitionByID.
	defErr error
}

// FindItemByID returns the stub item.
func (r foundRepoStub) FindItemByID(_ context.Context, _ int) (furnituredomain.Item, error) {
	return r.item, nil
}

// FindDefinitionByID returns the stub definition or configured error.
func (r foundRepoStub) FindDefinitionByID(_ context.Context, _ int) (furnituredomain.Definition, error) {
	return r.def, r.defErr
}

// ListItemsByRoomID returns the underlying repoStub item list for the room.
func (r foundRepoStub) ListItemsByRoomID(_ context.Context, _ int) ([]furnituredomain.Item, error) {
	return r.repoStub.items, nil
}

// compile-time assertion that foundRepoStub satisfies domain.Repository.
var _ furnituredomain.Repository = foundRepoStub{}

// buildRuntimeWithRoom creates a runtime wired with room finder and broadcaster stubs.
func buildRuntimeWithRoom(repo furnituredomain.Repository, roomID int) (*realtime.Runtime, *transportStub, *[]uint16) {
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return roomID, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	return rt, tp, &broadcast
}

// encodeInt32x4 encodes four big-endian int32 values into a 16-byte slice.
func encodeInt32x4(a, b, c, d int32) []byte {
	buf := make([]byte, 16)
	writeInt32(buf[0:], a)
	writeInt32(buf[4:], b)
	writeInt32(buf[8:], c)
	writeInt32(buf[12:], d)
	return buf
}

// writeInt32 writes one big-endian int32 into a four-byte slice.
func writeInt32(b []byte, v int32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

// TestHandlePickupBroadcastsRemoveAndAddsToInventory verifies pickup broadcasts 2703 and sends 104.
func TestHandlePickupBroadcastsRemoveAndAddsToInventory(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100, AllowRecycle: true, AllowTrade: true}
	repo := foundRepoStub{item: item, def: def}
	rt, tp, broadcast := buildRuntimeWithRoom(repo, 5)
	body := make([]byte, 8)
	body[3], body[7] = 1, 10
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PickupPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(*broadcast) != 1 || (*broadcast)[0] != furnipacket.FloorItemRemovePacketID {
		t.Fatalf("expected floor remove broadcast %d, got %v", furnipacket.FloorItemRemovePacketID, *broadcast)
	}
	if len(tp.sent) != 1 || tp.sent[0] != furnipacket.InventoryAddPacketID {
		t.Fatalf("expected inventory add packet %d, got %v", furnipacket.InventoryAddPacketID, tp.sent)
	}
}

// TestHandlePickupNoRoomFinderIsNoop verifies pickup without room finder does nothing.
func TestHandlePickupNoRoomFinderIsNoop(t *testing.T) {
	rt, tp := buildRuntime(repoStub{})
	body := make([]byte, 8)
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PickupPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(tp.sent) != 0 {
		t.Fatalf("expected no packets without room finder, got %v", tp.sent)
	}
}

// TestHandleFloorUpdateBroadcastsNewPosition verifies 248 triggers a broadcast with updated coords.
func TestHandleFloorUpdateBroadcastsNewPosition(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100}
	repo := foundRepoStub{item: item, def: def}
	rt, _, broadcast := buildRuntimeWithRoom(repo, 5)
	body := encodeInt32x4(10, 2, 3, 2)
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.FloorUpdatePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(*broadcast) != 1 || (*broadcast)[0] != furnipacket.FloorItemUpdatePacketID {
		t.Fatalf("expected floor update broadcast %d, got %v", furnipacket.FloorItemUpdatePacketID, *broadcast)
	}
}

// TestSendRoomFloorItemsSends1778 verifies SendRoomFloorItems encodes and sends packet 1778.
func TestSendRoomFloorItemsSends1778(t *testing.T) {
	item := furnituredomain.Item{ID: 7, UserID: 1, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 55}
	repo := foundRepoStub{item: item, def: def, repoStub: repoStub{items: []furnituredomain.Item{item}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	if err := rt.SendRoomFloorItems(context.Background(), "conn1", 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tp.sent) != 1 || tp.sent[0] != furnipacket.FurnitureFloorPacketID {
		t.Fatalf("expected floor packet %d, got %v", furnipacket.FurnitureFloorPacketID, tp.sent)
	}
}
