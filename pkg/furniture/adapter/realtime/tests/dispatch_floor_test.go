package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/core/codec"
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
	userBroadcast := &broadcasterStub{}
	rt.SetBroadcaster(userBroadcast)
	body := make([]byte, 8)
	body[3], body[7] = 1, 10
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PickupPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(*broadcast) != 1 || (*broadcast)[0] != furnipacket.FloorItemRemovePacketID {
		t.Fatalf("expected floor remove broadcast %d, got %v", furnipacket.FloorItemRemovePacketID, *broadcast)
	}
	if len(tp.sent) != 0 {
		t.Fatalf("expected no direct transport packets, got %v", tp.sent)
	}
	if len(userBroadcast.sent["broadcast:user:1"]) != 1 || userBroadcast.sent["broadcast:user:1"][0] != furnipacket.InventoryAddPacketID {
		t.Fatalf("expected owner inventory add packet %d, got %v", furnipacket.InventoryAddPacketID, userBroadcast.sent)
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

// TestHandleFloorUpdateAllowsRoomAuthorisedUser verifies room rights can move placed furniture even when the item owner differs.
func TestHandleFloorUpdateAllowsRoomAuthorisedUser(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 99, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100, StackHeight: 1.25}
	repo := foundRepoStub{item: item, def: def}
	rt, _, broadcast := buildRuntimeWithRoom(repo, 5)
	rt.SetRoomAccessChecker(func(_ context.Context, roomID, userID int) bool {
		return roomID == 5 && userID == 1
	})
	body := encodeInt32x4(10, 2, 3, 4)
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.FloorUpdatePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(*broadcast) != 1 || (*broadcast)[0] != furnipacket.FloorItemUpdatePacketID {
		t.Fatalf("expected floor update broadcast %d, got %v", furnipacket.FloorItemUpdatePacketID, *broadcast)
	}
}

// TestHandleFloorUpdateAllowsItemOwnerWithoutRoomRights verifies item owners can still move their own furniture in-room.
func TestHandleFloorUpdateAllowsItemOwnerWithoutRoomRights(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100}
	repo := foundRepoStub{item: item, def: def}
	rt, _, broadcast := buildRuntimeWithRoom(repo, 5)
	rt.SetRoomAccessChecker(func(_ context.Context, _, _ int) bool { return false })
	body := encodeInt32x4(10, 2, 3, 2)
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.FloorUpdatePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(*broadcast) != 1 || (*broadcast)[0] != furnipacket.FloorItemUpdatePacketID {
		t.Fatalf("expected floor update broadcast %d, got %v", furnipacket.FloorItemUpdatePacketID, *broadcast)
	}
}

// TestHandlePickupAllowsItemOwnerWithoutRoomRights verifies item owners can pick up their own furniture without room rights.
func TestHandlePickupAllowsItemOwnerWithoutRoomRights(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100, AllowRecycle: true, AllowTrade: true}
	repo := foundRepoStub{item: item, def: def}
	rt, tp, broadcast := buildRuntimeWithRoom(repo, 5)
	userBroadcast := &broadcasterStub{}
	rt.SetBroadcaster(userBroadcast)
	rt.SetRoomAccessChecker(func(_ context.Context, _, _ int) bool { return false })
	body := make([]byte, 8)
	body[3], body[7] = 1, 10
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PickupPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(*broadcast) != 1 || (*broadcast)[0] != furnipacket.FloorItemRemovePacketID {
		t.Fatalf("expected floor remove broadcast %d, got %v", furnipacket.FloorItemRemovePacketID, *broadcast)
	}
	if len(tp.sent) != 0 {
		t.Fatalf("expected no direct transport packets, got %v", tp.sent)
	}
	if len(userBroadcast.sent["broadcast:user:1"]) != 1 || userBroadcast.sent["broadcast:user:1"][0] != furnipacket.InventoryAddPacketID {
		t.Fatalf("expected owner inventory add packet %d, got %v", furnipacket.InventoryAddPacketID, userBroadcast.sent)
	}
}

// TestHandlePlaceAllowsItemOwnerWithoutRoomRights verifies item owners can place inventory furniture without room rights.
func TestHandlePlaceAllowsItemOwnerWithoutRoomRights(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 0, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100}
	repo := foundRepoStub{item: item, def: def}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	userBroadcast := &broadcasterStub{}
	rt.SetBroadcaster(userBroadcast)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomAccessChecker(func(_ context.Context, _, _ int) bool { return false })
	w := codec.NewWriter()
	if err := w.WriteString("10 2 3 2"); err != nil {
		t.Fatalf("encode place payload: %v", err)
	}
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PlacePacketID, w.Bytes())
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(broadcast) != 1 || broadcast[0] != furnipacket.FloorItemAddPacketID {
		t.Fatalf("expected floor add broadcast %d, got %v", furnipacket.FloorItemAddPacketID, broadcast)
	}
	if len(tp.sent) != 0 {
		t.Fatalf("expected no direct transport packets, got %v", tp.sent)
	}
	if len(userBroadcast.sent["broadcast:user:1"]) != 1 || userBroadcast.sent["broadcast:user:1"][0] != furnipacket.InventoryRemovePacketID {
		t.Fatalf("expected owner inventory remove packet %d, got %v", furnipacket.InventoryRemovePacketID, userBroadcast.sent)
	}
}

// TestHandlePlaceRejectsOccupiedTile verifies 1258 ignores placement onto a player-occupied tile.
func TestHandlePlaceRejectsOccupiedTile(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 0, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100}
	repo := foundRepoStub{item: item, def: def}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	userBroadcast := &broadcasterStub{}
	rt.SetBroadcaster(userBroadcast)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomOccupancyChecker(func(roomID, x, y int) bool {
		return roomID == 5 && x == 2 && y == 3
	})
	w := codec.NewWriter()
	if err := w.WriteString("10 2 3 2"); err != nil {
		t.Fatalf("encode place payload: %v", err)
	}
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PlacePacketID, w.Bytes())
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(broadcast) != 0 {
		t.Fatalf("expected no floor add broadcast, got %v", broadcast)
	}
	if len(tp.sent) != 0 {
		t.Fatalf("expected no direct transport packets, got %v", tp.sent)
	}
	if len(userBroadcast.sent) != 0 {
		t.Fatalf("expected no inventory packets, got %v", userBroadcast.sent)
	}
}

// TestHandlePickupSendsInventoryAddToOwner verifies moderator pickup routes the inventory add to the furniture owner.
func TestHandlePickupSendsInventoryAddToOwner(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 2, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100, AllowRecycle: true, AllowTrade: true}
	repo := foundRepoStub{item: item, def: def}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	userBroadcast := &broadcasterStub{}
	rt, _ := realtime.NewRuntime(svc, ownerSessionStub{}, tp, nil)
	rt.SetBroadcaster(userBroadcast)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, _ uint16, _ []byte) {})
	rt.SetRoomAccessChecker(func(_ context.Context, _, _ int) bool { return true })
	body := make([]byte, 8)
	body[3], body[7] = 1, 10
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PickupPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(tp.sent) != 0 {
		t.Fatalf("expected no direct transport packets, got %v", tp.sent)
	}
	if len(userBroadcast.sent["broadcast:user:2"]) != 1 || userBroadcast.sent["broadcast:user:2"][0] != furnipacket.InventoryAddPacketID {
		t.Fatalf("expected owner inventory add packet %d, got %v", furnipacket.InventoryAddPacketID, userBroadcast.sent)
	}
}

// TestHandlePlaceIgnoresAlreadyPlacedItem verifies 1258 cannot re-place an item that is already in a room.
func TestHandlePlaceIgnoresAlreadyPlacedItem(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100}
	repo := foundRepoStub{item: item, def: def}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	userBroadcast := &broadcasterStub{}
	rt.SetBroadcaster(userBroadcast)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	w := codec.NewWriter()
	if err := w.WriteString("10 2 3 2"); err != nil {
		t.Fatalf("encode place payload: %v", err)
	}
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PlacePacketID, w.Bytes())
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(broadcast) != 0 {
		t.Fatalf("expected no floor add broadcast, got %v", broadcast)
	}
	if len(tp.sent) != 0 {
		t.Fatalf("expected no direct transport packets, got %v", tp.sent)
	}
	if len(userBroadcast.sent) != 0 {
		t.Fatalf("expected no inventory packets, got %v", userBroadcast.sent)
	}
}

// TestHandlePickupEvictsUsingOriginalTile verifies seated users are cleared from the furniture's old tile on pickup.
func TestHandlePickupEvictsUsingOriginalTile(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, X: 4, Y: 7, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100, StackHeight: 1.1, CanSit: true, AllowRecycle: true, AllowTrade: true}
	repo := foundRepoStub{item: item, def: def}
	rt, _, _ := buildRuntimeWithRoom(repo, 5)
	userBroadcast := &broadcasterStub{}
	rt.SetBroadcaster(userBroadcast)
	evictedX, evictedY := -1, -1
	rt.SetRoomEntityEvictor(func(_ int, x, y int) {
		evictedX, evictedY = x, y
	})
	body := make([]byte, 8)
	body[3], body[7] = 1, 10
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PickupPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if evictedX != 4 || evictedY != 7 {
		t.Fatalf("expected evictor to receive original tile 4,7 got %d,%d", evictedX, evictedY)
	}
}

// TestHandleFloorUpdateEncodesDefinitionStackHeight verifies update packets expose the item's real stack height.
func TestHandleFloorUpdateEncodesDefinitionStackHeight(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100, StackHeight: 1.25}
	repo := foundRepoStub{item: item, def: def}
	rt, _, _ := buildRuntimeWithRoom(repo, 5)
	var encodedBody []byte
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, body []byte) {
		if pktID == furnipacket.FloorItemUpdatePacketID {
			encodedBody = body
		}
	})
	body := encodeInt32x4(10, 2, 3, 2)
	_, err := rt.Handle(context.Background(), "conn1", furnipacket.FloorUpdatePacketID, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := codec.NewReader(encodedBody)
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read item id: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read sprite id: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read x: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read y: %v", err)
	}
	if _, err = r.ReadInt32(); err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if _, err = r.ReadString(); err != nil {
		t.Fatalf("read z: %v", err)
	}
	stackHeight, err := r.ReadString()
	if err != nil {
		t.Fatalf("read stack height: %v", err)
	}
	if stackHeight != "1.25" {
		t.Fatalf("expected stack height 1.25, got %q", stackHeight)
	}
}

// TestHandleFloorUpdateRejectsOccupiedDestination verifies 248 cannot move furniture onto a player-occupied tile.
func TestHandleFloorUpdateRejectsOccupiedDestination(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, X: 1, Y: 1, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100}
	repo := foundRepoStub{item: item, def: def}
	rt, _, broadcast := buildRuntimeWithRoom(repo, 5)
	rt.SetRoomOccupancyChecker(func(roomID, x, y int) bool {
		return roomID == 5 && x == 2 && y == 3
	})
	body := encodeInt32x4(10, 2, 3, 2)
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.FloorUpdatePacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	if len(*broadcast) != 0 {
		t.Fatalf("expected no floor update broadcast, got %v", *broadcast)
	}
}

// TestHandleFloorUpdateAllowsRotateInPlaceOnOccupiedTile verifies rotation still works when a seated avatar occupies the furniture tile.
func TestHandleFloorUpdateAllowsRotateInPlaceOnOccupiedTile(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, X: 2, Y: 3, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100}
	repo := foundRepoStub{item: item, def: def}
	rt, _, broadcast := buildRuntimeWithRoom(repo, 5)
	rt.SetRoomOccupancyChecker(func(roomID, x, y int) bool {
		return roomID == 5 && x == 2 && y == 3
	})
	body := encodeInt32x4(10, 2, 3, 4)
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
	def := furnituredomain.Definition{ID: 3, SpriteID: 55, StackHeight: 0.75}
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

// TestSendRoomFloorItemsCachesLaySlots verifies multi-tile lay furniture resolves each lane to its own lay slot.
func TestSendRoomFloorItemsCachesLaySlots(t *testing.T) {
	item := furnituredomain.Item{ID: 7, UserID: 1, RoomID: 5, X: 3, Y: 4, Dir: 0, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 55, Width: 2, Length: 3, StackHeight: 1.4, CanLay: true}
	repo := foundRepoStub{item: item, def: def, repoStub: repoStub{items: []furnituredomain.Item{item}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	if err := rt.SendRoomFloorItems(context.Background(), "conn1", 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedTargets := map[[2]int][2]int{
		{3, 4}: {3, 4},
		{3, 5}: {3, 4},
		{3, 6}: {3, 4},
		{4, 4}: {4, 4},
		{4, 5}: {4, 4},
		{4, 6}: {4, 4},
	}
	for tile, expectedTarget := range expectedTargets {
		_, _, canSit, canLay := rt.TileSeatCheckerFor(5, tile[0], tile[1])
		targetX, targetY, ok := rt.ResolveSeatTargetFor(5, tile[0], tile[1])
		if !ok || targetX != expectedTarget[0] || targetY != expectedTarget[1] {
			t.Fatalf("expected tile %v to resolve to lay slot %v, got ok=%v target=[%d %d]", tile, expectedTarget, ok, targetX, targetY)
		}
		if canSit {
			t.Fatalf("expected lay footprint tile %v to remain non-sittable", tile)
		}
		if tile == expectedTarget {
			if !canLay {
				t.Fatalf("expected lay slot anchor %v to remain layable", tile)
			}
			continue
		}
		if canLay {
			t.Fatalf("expected non-anchor bed tile %v to redirect rather than lay in place", tile)
		}
	}
	_, _, _, canLay := rt.TileSeatCheckerFor(5, 5, 6)
	if canLay {
		t.Fatal("expected tile outside the bed footprint to remain non-layable")
	}
}

// TestResolveSeatTargetForLayFootprintFallsBackToFreeSlot verifies occupied bed lanes redirect to another free slot on the same item.
func TestResolveSeatTargetForLayFootprintFallsBackToFreeSlot(t *testing.T) {
	item := furnituredomain.Item{ID: 7, UserID: 1, RoomID: 5, X: 3, Y: 4, Dir: 0, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 55, Width: 2, Length: 3, StackHeight: 1.4, CanLay: true}
	repo := foundRepoStub{item: item, def: def, repoStub: repoStub{items: []furnituredomain.Item{item}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	rt.SetRoomOccupancyChecker(func(roomID, x, y int) bool {
		return roomID == 5 && x == 3 && y == 4
	})
	if err := rt.SendRoomFloorItems(context.Background(), "conn1", 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	targetX, targetY, ok := rt.ResolveSeatTargetFor(5, 3, 5)
	if !ok || targetX != 4 || targetY != 4 {
		t.Fatalf("expected occupied left slot to fall back to free right slot [4 4], got ok=%v target=[%d %d]", ok, targetX, targetY)
	}
}

// TestSendRoomFloorItemsCachesRotatedLaySlots verifies rotation maps each bed lane to the correct rotated slot.
func TestSendRoomFloorItemsCachesRotatedLaySlots(t *testing.T) {
	item := furnituredomain.Item{ID: 7, UserID: 1, RoomID: 5, X: 3, Y: 4, Dir: 2, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 55, Width: 2, Length: 3, StackHeight: 1.4, CanLay: true}
	repo := foundRepoStub{item: item, def: def, repoStub: repoStub{items: []furnituredomain.Item{item}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	if err := rt.SendRoomFloorItems(context.Background(), "conn1", 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedTargets := map[[2]int][2]int{
		{3, 4}: {5, 4},
		{4, 4}: {5, 4},
		{5, 4}: {5, 4},
		{3, 5}: {5, 5},
		{4, 5}: {5, 5},
		{5, 5}: {5, 5},
	}
	for tile, expectedTarget := range expectedTargets {
		_, _, _, canLay := rt.TileSeatCheckerFor(5, tile[0], tile[1])
		targetX, targetY, ok := rt.ResolveSeatTargetFor(5, tile[0], tile[1])
		if !ok || targetX != expectedTarget[0] || targetY != expectedTarget[1] {
			t.Fatalf("expected rotated tile %v to resolve to lay slot %v, got ok=%v target=[%d %d]", tile, expectedTarget, ok, targetX, targetY)
		}
		if tile == expectedTarget {
			if !canLay {
				t.Fatalf("expected rotated lay slot anchor %v to remain layable", tile)
			}
			continue
		}
		if canLay {
			t.Fatalf("expected rotated non-anchor bed tile %v to redirect rather than lay in place", tile)
		}
	}
}

// TestHandlePickupEvictsEntireLayFootprint verifies bed pickup clears seated avatars from every covered tile.
func TestHandlePickupEvictsEntireLayFootprint(t *testing.T) {
	item := furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, X: 4, Y: 7, Dir: 0, DefinitionID: 3}
	def := furnituredomain.Definition{ID: 3, SpriteID: 100, Width: 2, Length: 3, StackHeight: 1.4, CanLay: true, AllowRecycle: true, AllowTrade: true}
	repo := foundRepoStub{item: item, def: def, repoStub: repoStub{items: []furnituredomain.Item{item}}}
	rt, _, _ := buildRuntimeWithRoom(repo, 5)
	userBroadcast := &broadcasterStub{}
	rt.SetBroadcaster(userBroadcast)
	if err := rt.SendRoomFloorItems(context.Background(), "conn1", 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	evicted := make(map[[2]int]struct{})
	rt.SetRoomEntityEvictor(func(_ int, x, y int) {
		evicted[[2]int{x, y}] = struct{}{}
	})
	body := make([]byte, 8)
	body[3], body[7] = 1, 10
	handled, err := rt.Handle(context.Background(), "conn1", furnipacket.PickupPacketID, body)
	if err != nil || !handled {
		t.Fatalf("expected handled without error, got handled=%v err=%v", handled, err)
	}
	expected := [][2]int{{4, 7}, {5, 7}, {4, 8}, {5, 8}, {4, 9}, {5, 9}}
	if len(evicted) != len(expected) {
		t.Fatalf("expected %d evicted tiles, got %v", len(expected), evicted)
	}
	for _, tile := range expected {
		if _, ok := evicted[tile]; !ok {
			t.Fatalf("expected pickup eviction for tile %v, got %v", tile, evicted)
		}
	}
}
