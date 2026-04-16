package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/furniture/adapter/realtime"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
)

// advancedRepo stores deterministic mutable furniture data for advanced interaction tests.
type advancedRepo struct {
	// items stores mutable item rows keyed by identifier.
	items map[int]furnituredomain.Item
	// defs stores immutable definition rows keyed by identifier.
	defs map[int]furnituredomain.Definition
	// itemDataUpdates stores visible state transitions keyed by placed item identifier.
	itemDataUpdates map[int][]string
}

// FindDefinitionByID returns one definition by identifier.
func (r *advancedRepo) FindDefinitionByID(_ context.Context, id int) (furnituredomain.Definition, error) {
	return r.defs[id], nil
}

// FindDefinitionByName returns one definition by name.
func (r *advancedRepo) FindDefinitionByName(_ context.Context, name string) (furnituredomain.Definition, error) {
	for _, def := range r.defs {
		if def.ItemName == name {
			return def, nil
		}
	}
	return furnituredomain.Definition{}, nil
}

// ListDefinitions returns all definitions.
func (r *advancedRepo) ListDefinitions(_ context.Context) ([]furnituredomain.Definition, error) {
	result := make([]furnituredomain.Definition, 0, len(r.defs))
	for _, def := range r.defs {
		result = append(result, def)
	}
	return result, nil
}

// CreateDefinition returns the input definition unchanged.
func (r *advancedRepo) CreateDefinition(_ context.Context, def furnituredomain.Definition) (furnituredomain.Definition, error) {
	return def, nil
}

// UpdateDefinition returns the current definition.
func (r *advancedRepo) UpdateDefinition(_ context.Context, id int, _ furnituredomain.DefinitionPatch) (furnituredomain.Definition, error) {
	return r.defs[id], nil
}

// DeleteDefinition returns nil.
func (r *advancedRepo) DeleteDefinition(_ context.Context, _ int) error { return nil }

// FindItemByID returns one item by identifier.
func (r *advancedRepo) FindItemByID(_ context.Context, id int) (furnituredomain.Item, error) {
	return r.items[id], nil
}

// ListItemsByUserID returns all items owned by one user.
func (r *advancedRepo) ListItemsByUserID(_ context.Context, userID int) ([]furnituredomain.Item, error) {
	result := make([]furnituredomain.Item, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

// CreateItem stores one item and returns it.
func (r *advancedRepo) CreateItem(_ context.Context, item furnituredomain.Item) (furnituredomain.Item, error) {
	r.items[item.ID] = item
	return item, nil
}

// DeleteItem removes one item.
func (r *advancedRepo) DeleteItem(_ context.Context, id int) error {
	delete(r.items, id)
	return nil
}

// TransferItem updates the item owner.
func (r *advancedRepo) TransferItem(_ context.Context, itemID int, newUserID int) error {
	item := r.items[itemID]
	item.UserID = newUserID
	r.items[itemID] = item
	return nil
}

// UpdateItemData updates one visible item payload.
func (r *advancedRepo) UpdateItemData(_ context.Context, itemID int, extraData string) error {
	item := r.items[itemID]
	item.ExtraData = extraData
	r.items[itemID] = item
	if r.itemDataUpdates == nil {
		r.itemDataUpdates = make(map[int][]string)
	}
	r.itemDataUpdates[itemID] = append(r.itemDataUpdates[itemID], extraData)
	return nil
}

// UpdateItemInteractionData updates one hidden interaction payload.
func (r *advancedRepo) UpdateItemInteractionData(_ context.Context, itemID int, interactionData string) error {
	item := r.items[itemID]
	item.InteractionData = interactionData
	r.items[itemID] = item
	return nil
}

// PlaceItem updates one floor placement row.
func (r *advancedRepo) PlaceItem(_ context.Context, itemID int, roomID int, x int, y int, z float64, dir int) error {
	item := r.items[itemID]
	item.RoomID = roomID
	item.X = x
	item.Y = y
	item.Z = z
	item.Dir = dir
	item.WallPosition = ""
	r.items[itemID] = item
	return nil
}

// PlaceWallItem updates one wall placement row.
func (r *advancedRepo) PlaceWallItem(_ context.Context, itemID int, roomID int, wallPosition string) error {
	item := r.items[itemID]
	item.RoomID = roomID
	item.WallPosition = wallPosition
	r.items[itemID] = item
	return nil
}

// UpdateItemDefinition updates one transformed item payload.
func (r *advancedRepo) UpdateItemDefinition(_ context.Context, itemID int, definitionID int, extraData string, interactionData string) error {
	item := r.items[itemID]
	item.DefinitionID = definitionID
	item.ExtraData = extraData
	item.InteractionData = interactionData
	r.items[itemID] = item
	return nil
}

// ListItemsByRoomID returns all placed room items.
func (r *advancedRepo) ListItemsByRoomID(_ context.Context, roomID int) ([]furnituredomain.Item, error) {
	result := make([]furnituredomain.Item, 0)
	for _, item := range r.items {
		if item.RoomID == roomID {
			result = append(result, item)
		}
	}
	return result, nil
}

// CountItemsByUserID returns one user inventory count.
func (r *advancedRepo) CountItemsByUserID(_ context.Context, userID int) (int, error) {
	items, _ := r.ListItemsByUserID(context.Background(), userID)
	return len(items), nil
}

// TestHandleSetItemDataBroadcastsWallStateAndData verifies sticky-note saves update both wall state and hidden note text.
func TestHandleSetItemDataBroadcastsWallStateAndData(t *testing.T) {
	repo := &advancedRepo{items: map[int]furnituredomain.Item{10: {ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3, ExtraData: "FFFF33", WallPosition: ":w=1,1 l=1,1 l", InteractionData: "hello"}}, defs: map[int]furnituredomain.Definition{3: {ID: 3, ItemType: furnituredomain.ItemTypeWall, InteractionType: furnituredomain.InteractionPostIt, SpriteID: 100}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomAccessChecker(func(_ context.Context, roomID, userID int) bool { return roomID == 5 && userID == 1 })
	w := codec.NewWriter()
	w.WriteInt32(10)
	_ = w.WriteString("9CFF9C")
	_ = w.WriteString("updated note")
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.SetItemDataPacketID, w.Bytes()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.items[10].ExtraData != "9CFF9C" || repo.items[10].InteractionData != "updated note" {
		t.Fatalf("expected wall data and note text update, got %+v", repo.items[10])
	}
	if len(broadcast) != 2 || broadcast[0] != furnipacket.WallItemUpdatePacketID || broadcast[1] != furnipacket.ItemDataUpdatePacketID {
		t.Fatalf("expected wall update and item data broadcasts, got %v", broadcast)
	}
}

// TestHandleDimmerSettingsAndToggle verifies moodlight requests send presets and toggles update the wall item state.
func TestHandleDimmerSettingsAndToggle(t *testing.T) {
	raw, _ := furnituredomain.InteractionData{Dimmer: &furnituredomain.DimmerData{Enabled: false, SelectedPresetID: 1, Presets: []furnituredomain.DimmerPresetData{{PresetID: 1, Type: 1, Color: "#000000", Brightness: 255}}}}.Encode()
	repo := &advancedRepo{items: map[int]furnituredomain.Item{10: {ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3, ExtraData: "0,1,1,#000000,255", WallPosition: ":w=1,1 l=1,1 l", InteractionData: raw}}, defs: map[int]furnituredomain.Definition{3: {ID: 3, ItemType: furnituredomain.ItemTypeWall, InteractionType: furnituredomain.InteractionDimmer, SpriteID: 200}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomAccessChecker(func(_ context.Context, roomID, userID int) bool { return roomID == 5 && userID == 1 })
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.DimmerSettingsPacketID, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tp.sent) != 1 || tp.sent[0] != furnipacket.DimmerPresetsPacketID {
		t.Fatalf("expected dimmer presets packet, got %v", tp.sent)
	}
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.DimmerTogglePacketID, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(repo.items[10].ExtraData, "1,") {
		t.Fatalf("expected enabled dimmer extra data, got %s", repo.items[10].ExtraData)
	}
	if len(broadcast) != 1 || broadcast[0] != furnipacket.WallItemUpdatePacketID {
		t.Fatalf("expected wall item update broadcast, got %v", broadcast)
	}
	if len(tp.sent) != 2 || tp.sent[1] != furnipacket.DimmerPresetsPacketID {
		t.Fatalf("expected refreshed dimmer presets packet, got %v", tp.sent)
	}
}

// TestHandleToggleMultistateTeleporterForwardsCrossRoom verifies teleporter use walks to the booth, forwards rooms, and exits after arrival.
func TestHandleToggleMultistateTeleporterForwardsCrossRoom(t *testing.T) {
	leftRaw, _ := furnituredomain.InteractionData{Teleporter: &furnituredomain.TeleporterData{RoomID: 5, ItemID: 11}}.Encode()
	rightRaw, _ := furnituredomain.InteractionData{Teleporter: &furnituredomain.TeleporterData{RoomID: 8, ItemID: 10}}.Encode()
	repo := &advancedRepo{items: map[int]furnituredomain.Item{10: {ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3, X: 4, Y: 4, Z: 0, Dir: 2, ExtraData: "0", InteractionData: leftRaw}, 11: {ID: 11, UserID: 1, RoomID: 8, DefinitionID: 3, X: 1, Y: 1, Z: 0, Dir: 2, ExtraData: "0", InteractionData: rightRaw}}, defs: map[int]furnituredomain.Definition{3: {ID: 3, ItemType: furnituredomain.ItemTypeFloor, InteractionType: furnituredomain.InteractionTeleport, SpriteID: 100}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	approached := make(chan []int, 1)
	exited := make(chan []int, 1)
	forwarded := make(chan []int, 1)
	currentRoomID, currentX, currentY := 5, 1, 1
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, _ uint16, _ []byte) {})
	rt.SetRoomEntityTileResolver(func(_ string) (int, int, int, bool) { return currentRoomID, currentX, currentY, true })
	rt.SetRoomEntitySnapshotter(func(roomID int) []realtime.RoomEntitySnapshot {
		if roomID != 5 || currentRoomID != 5 {
			return nil
		}
		return []realtime.RoomEntitySnapshot{{ConnID: "conn1", VirtualID: 7, UserID: 1, X: currentX, Y: currentY, Z: 0, Dir: 2}}
	})
	rt.SetRoomEntityWalker(func(_ context.Context, _ string, x, y int) error {
		currentX = x
		currentY = y
		if currentRoomID == 5 {
			approached <- []int{x, y}
			return nil
		}
		exited <- []int{x, y}
		return nil
	})
	rt.SetRoomEntityWarper(func(_ context.Context, roomID, virtualID, x, y int, _ float64, _ int, _ bool, animate bool) error {
		if roomID != 5 || virtualID != 7 {
			return nil
		}
		currentX = x
		currentY = y
		if !animate {
			t.Fatalf("expected source teleporter entry warp to animate")
		}
		return nil
	})
	rt.SetTeleporterForwarder(func(_ context.Context, _ string, roomID, spawnX, spawnY int, _ float64, _ int, exitX, exitY int) error {
		currentRoomID = roomID
		currentX = spawnX
		currentY = spawnY
		forwarded <- []int{roomID, spawnX, spawnY, exitX, exitY}
		return nil
	})
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.ToggleMultistatePacketID, encodeInt32(10)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case values := <-approached:
		if values[0] != 5 || values[1] != 4 {
			t.Fatalf("unexpected teleporter approach tile %v", values)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected teleporter approach walk")
	}
	select {
	case values := <-forwarded:
		if values[0] != 8 || values[1] != 1 || values[2] != 1 || values[3] != 2 || values[4] != 1 {
			t.Fatalf("unexpected teleporter forward payload %v", values)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("expected teleporter forward callback")
	}
	select {
	case values := <-exited:
		if values[0] != 2 || values[1] != 1 {
			t.Fatalf("unexpected teleporter exit walk %v", values)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected teleporter exit walk")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(repo.itemDataUpdates[10]) >= 2 && len(repo.itemDataUpdates[11]) >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := repo.itemDataUpdates[10]; len(got) < 2 || got[0] != "1" || got[1] != "2" {
		t.Fatalf("expected source teleporter to open then transfer, got %v", got)
	}
	if got := repo.itemDataUpdates[11]; len(got) < 2 || got[0] != "2" || got[1] != "1" {
		t.Fatalf("expected destination teleporter to receive then release, got %v", got)
	}
	leftData, err := furnituredomain.ParseInteractionData(repo.items[10].InteractionData)
	if err != nil || leftData.Teleporter == nil || leftData.Teleporter.RoomID != 8 {
		t.Fatalf("expected refreshed current teleporter metadata, got %q", repo.items[10].InteractionData)
	}
	rightData, err := furnituredomain.ParseInteractionData(repo.items[11].InteractionData)
	if err != nil || rightData.Teleporter == nil || rightData.Teleporter.RoomID != 5 {
		t.Fatalf("expected refreshed partner teleporter metadata, got %q", repo.items[11].InteractionData)
	}
}

// TestProcessRoomTickRollerBroadcastsMovement verifies rollers move one floor item and one avatar with a rolling packet.
func TestProcessRoomTickRollerBroadcastsMovement(t *testing.T) {
	repo := &advancedRepo{items: map[int]furnituredomain.Item{1: {ID: 1, UserID: 1, RoomID: 5, DefinitionID: 3, X: 1, Y: 1, Z: 0, Dir: 2}, 2: {ID: 2, UserID: 1, RoomID: 5, DefinitionID: 4, X: 1, Y: 1, Z: 0.5, Dir: 2}}, defs: map[int]furnituredomain.Definition{3: {ID: 3, ItemType: furnituredomain.ItemTypeFloor, InteractionType: furnituredomain.InteractionRoller, SpriteID: 10, StackHeight: 0.5}, 4: {ID: 4, ItemType: furnituredomain.ItemTypeFloor, SpriteID: 20, StackHeight: 1}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	warped := make(chan []int, 1)
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomEntitySnapshotter(func(roomID int) []realtime.RoomEntitySnapshot {
		if roomID != 5 {
			return nil
		}
		return []realtime.RoomEntitySnapshot{{ConnID: "conn1", VirtualID: 8, UserID: 1, X: 1, Y: 1, Z: 0.5, Dir: 2}}
	})
	rt.SetRoomEntityWarper(func(_ context.Context, roomID, virtualID, x, y int, _ float64, _ int, _ bool, _ bool) error {
		warped <- []int{roomID, virtualID, x, y}
		return nil
	})
	rt.ProcessRoomTick(5)
	if repo.items[2].X != 2 || repo.items[2].Y != 1 {
		t.Fatalf("expected rolled item to move to 2,1, got %+v", repo.items[2])
	}
	if len(broadcast) != 1 || broadcast[0] != furnipacket.RoomRollingPacketID {
		t.Fatalf("expected room rolling packet, got %v", broadcast)
	}
	select {
	case values := <-warped:
		if values[0] != 5 || values[1] != 8 || values[2] != 2 || values[3] != 1 {
			t.Fatalf("unexpected roller warp payload %v", values)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected roller avatar warp callback")
	}
}

// TestProcessRoomTickRollerSkipsInactiveMultistate verifies multistate rollers only tick while switched on.
func TestProcessRoomTickRollerSkipsInactiveMultistate(t *testing.T) {
	repo := &advancedRepo{items: map[int]furnituredomain.Item{1: {ID: 1, UserID: 1, RoomID: 5, DefinitionID: 3, X: 1, Y: 1, Z: 0, Dir: 2, ExtraData: "0"}, 2: {ID: 2, UserID: 1, RoomID: 5, DefinitionID: 4, X: 1, Y: 1, Z: 0.5, Dir: 2}}, defs: map[int]furnituredomain.Definition{3: {ID: 3, ItemType: furnituredomain.ItemTypeFloor, InteractionType: furnituredomain.InteractionRoller, InteractionModesCount: 2, SpriteID: 10, StackHeight: 0.5}, 4: {ID: 4, ItemType: furnituredomain.ItemTypeFloor, SpriteID: 20, StackHeight: 1}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomEntitySnapshotter(func(roomID int) []realtime.RoomEntitySnapshot {
		if roomID != 5 {
			return nil
		}
		return nil
	})
	rt.ProcessRoomTick(5)
	if repo.items[2].X != 1 || repo.items[2].Y != 1 {
		t.Fatalf("expected inactive roller to leave item in place, got %+v", repo.items[2])
	}
	if len(broadcast) != 0 {
		t.Fatalf("expected no room rolling packet, got %v", broadcast)
	}
}

// TestProcessRoomTickRollerSkipsWalkingAvatar verifies rollers do not warp avatars already walking across the tile.
func TestProcessRoomTickRollerSkipsWalkingAvatar(t *testing.T) {
	repo := &advancedRepo{items: map[int]furnituredomain.Item{1: {ID: 1, UserID: 1, RoomID: 5, DefinitionID: 3, X: 1, Y: 1, Z: 0, Dir: 2}}, defs: map[int]furnituredomain.Definition{3: {ID: 3, ItemType: furnituredomain.ItemTypeFloor, InteractionType: furnituredomain.InteractionRoller, SpriteID: 10, StackHeight: 0.5}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	warped := false
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomEntitySnapshotter(func(roomID int) []realtime.RoomEntitySnapshot {
		if roomID != 5 {
			return nil
		}
		return []realtime.RoomEntitySnapshot{{ConnID: "conn1", VirtualID: 8, UserID: 1, X: 1, Y: 1, Z: 0.5, Dir: 2, IsWalking: true}}
	})
	rt.SetRoomEntityWarper(func(_ context.Context, roomID, virtualID, x, y int, _ float64, _ int, _ bool, _ bool) error {
		warped = true
		t.Fatalf("unexpected roller warp for walking avatar room=%d virtual=%d x=%d y=%d", roomID, virtualID, x, y)
		return nil
	})
	rt.ProcessRoomTick(5)
	if warped {
		t.Fatal("expected walking avatar to be ignored by roller")
	}
	if len(broadcast) != 0 {
		t.Fatalf("expected no room rolling packet, got %v", broadcast)
	}
}

// TestHandleOpenPresentTransformsGift verifies present opening transforms the item and sends the opened payload.
func TestHandleOpenPresentTransformsGift(t *testing.T) {
	raw, _ := furnituredomain.InteractionData{Gift: &furnituredomain.GiftData{DefinitionID: 4, ProductCode: "revealed_chair"}}.Encode()
	repo := &advancedRepo{items: map[int]furnituredomain.Item{10: {ID: 10, UserID: 1, RoomID: 5, DefinitionID: 3, X: 4, Y: 4, Z: 0, Dir: 2, InteractionData: raw}}, defs: map[int]furnituredomain.Definition{3: {ID: 3, ItemType: furnituredomain.ItemTypeFloor, InteractionType: furnituredomain.InteractionGift, SpriteID: 100}, 4: {ID: 4, ItemType: furnituredomain.ItemTypeFloor, SpriteID: 200, ItemName: "revealed_chair"}}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomEntityTileResolver(func(_ string) (int, int, int, bool) { return 5, 4, 4, true })
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.OpenPresentPacketID, encodeInt32(10)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.items[10].DefinitionID != 4 {
		t.Fatalf("expected transformed definition, got %+v", repo.items[10])
	}
	if len(broadcast) != 1 || broadcast[0] != furnipacket.FloorItemUpdatePacketID {
		t.Fatalf("expected floor update broadcast, got %v", broadcast)
	}
	if len(tp.sent) != 1 || tp.sent[0] != furnipacket.GiftOpenedPacketID {
		t.Fatalf("expected gift opened packet, got %v", tp.sent)
	}
}
