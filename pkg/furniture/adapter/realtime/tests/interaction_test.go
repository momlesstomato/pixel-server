package tests

import (
	"context"
	"testing"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/furniture/adapter/realtime"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
)

// interactionRepo stores one mutable furniture item for realtime interaction tests.
type interactionRepo struct {
	repoStub
	item    furnituredomain.Item
	def     furnituredomain.Definition
	updates []string
}

// FindItemByID returns the mutable item.
func (r *interactionRepo) FindItemByID(_ context.Context, _ int) (furnituredomain.Item, error) {
	return r.item, nil
}

// FindDefinitionByID returns the mutable definition.
func (r *interactionRepo) FindDefinitionByID(_ context.Context, _ int) (furnituredomain.Definition, error) {
	return r.def, nil
}

// UpdateItemData stores the latest item state.
func (r *interactionRepo) UpdateItemData(_ context.Context, _ int, extraData string) error {
	r.item.ExtraData = extraData
	r.updates = append(r.updates, extraData)
	return nil
}

// UpdateItemInteractionData stores hidden item data.
func (r *interactionRepo) UpdateItemInteractionData(_ context.Context, _ int, interactionData string) error {
	r.item.InteractionData = interactionData
	return nil
}

// PlaceWallItem stores the latest wall placement.
func (r *interactionRepo) PlaceWallItem(_ context.Context, _ int, roomID int, wallPosition string) error {
	r.item.RoomID = roomID
	r.item.WallPosition = wallPosition
	return nil
}

// UpdateItemDefinition stores the latest transformed definition payload.
func (r *interactionRepo) UpdateItemDefinition(_ context.Context, _ int, definitionID int, extraData string, interactionData string) error {
	r.item.DefinitionID = definitionID
	r.item.ExtraData = extraData
	r.item.InteractionData = interactionData
	return nil
}

// TestHandleToggleMultistateRequiresProximity verifies floor interactions do not mutate distant items.
func TestHandleToggleMultistateRequiresProximity(t *testing.T) {
	repo := &interactionRepo{item: furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, X: 5, Y: 5, DefinitionID: 3, ExtraData: "0"}, def: furnituredomain.Definition{ID: 3, SpriteID: 100, InteractionModesCount: 2}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomEntityTileResolver(func(_ string) (int, int, int, bool) { return 5, 1, 1, true })
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.ToggleMultistatePacketID, encodeInt32(10)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.updates) != 0 {
		t.Fatalf("expected no item mutation, got %v", repo.updates)
	}
	if len(broadcast) != 0 {
		t.Fatalf("expected no broadcasts, got %v", broadcast)
	}
}

// TestHandleActivateDiceBroadcastsRollingAndFinalValue verifies dice activation emits rolling and completed room updates.
func TestHandleActivateDiceBroadcastsRollingAndFinalValue(t *testing.T) {
	repo := &interactionRepo{item: furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, X: 4, Y: 4, DefinitionID: 3, ExtraData: "0"}, def: furnituredomain.Definition{ID: 3, SpriteID: 100, InteractionType: furnituredomain.InteractionDice}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomEntityTileResolver(func(_ string) (int, int, int, bool) { return 5, 4, 3, true })
	rt.SetDiceRollDelay(time.Millisecond)
	rt.SetDiceRandomizer(func(int) int { return 4 })
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.ActivateDicePacketID, encodeInt32(10)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deadline := time.Now().Add(100 * time.Millisecond)
	for len(broadcast) < 4 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if repo.item.ExtraData != "5" {
		t.Fatalf("expected final dice value 5, got %s", repo.item.ExtraData)
	}
	if len(broadcast) < 4 {
		t.Fatalf("expected rolling and final broadcasts, got %v", broadcast)
	}
	if broadcast[0] != furnipacket.FloorItemUpdatePacketID || broadcast[1] != furnipacket.DiceValuePacketID {
		t.Fatalf("expected rolling broadcasts first, got %v", broadcast)
	}
	if broadcast[2] != furnipacket.FloorItemUpdatePacketID || broadcast[3] != furnipacket.DiceValuePacketID {
		t.Fatalf("expected final broadcasts last, got %v", broadcast)
	}
}

// TestHandleSetStackHeightBroadcastsUpdate verifies stack-helper changes require rights and acknowledge the new height.
func TestHandleSetStackHeightBroadcastsUpdate(t *testing.T) {
	repo := &interactionRepo{item: furnituredomain.Item{ID: 10, UserID: 1, RoomID: 5, X: 4, Y: 4, DefinitionID: 3}, def: furnituredomain.Definition{ID: 3, SpriteID: 100, InteractionType: furnituredomain.InteractionStackHelper}}
	svc, _ := furnitureapplication.NewService(repo)
	tp := &transportStub{}
	rt, _ := realtime.NewRuntime(svc, sessionStub{}, tp, nil)
	broadcast := make([]uint16, 0)
	rt.SetRoomFinder(func(_ string) (int, bool) { return 5, true })
	rt.SetRoomBroadcaster(func(_ int, pktID uint16, _ []byte) { broadcast = append(broadcast, pktID) })
	rt.SetRoomAccessChecker(func(_ context.Context, roomID, userID int) bool { return roomID == 5 && userID == 1 })
	if _, err := rt.Handle(context.Background(), "conn1", furnipacket.SetStackHeightPacketID, encodeInt32x2(10, 125)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.item.ExtraData != "1.25" {
		t.Fatalf("expected stored height 1.25, got %s", repo.item.ExtraData)
	}
	if len(broadcast) != 1 || broadcast[0] != furnipacket.FloorItemUpdatePacketID {
		t.Fatalf("expected floor update broadcast, got %v", broadcast)
	}
	if len(tp.sent) != 1 || tp.sent[0] != furnipacket.StackHeightUpdatePacketID {
		t.Fatalf("expected stack height ack %d, got %v", furnipacket.StackHeightUpdatePacketID, tp.sent)
	}
}

// encodeInt32 encodes one big-endian int32 payload.
func encodeInt32(value int32) []byte {
	return encodeInt32x2(value, 0)[:4]
}

// encodeInt32x2 encodes two big-endian int32 values into one payload.
func encodeInt32x2(a, b int32) []byte {
	buf := make([]byte, 8)
	writeInt32(buf[0:], a)
	writeInt32(buf[4:], b)
	return buf
}
