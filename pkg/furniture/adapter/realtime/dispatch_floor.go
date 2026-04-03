package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
)

// handlePickup processes a furniture pickup request (c2s 3456).
// It validates ownership, clears room placement, and returns the item to inventory.
func (runtime *Runtime) handlePickup(ctx context.Context, connID string, body []byte) error {
	userID, _ := runtime.userID(connID)
	r := codec.NewReader(body)
	_, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	itemID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	if runtime.roomFinder == nil || runtime.roomBroadcaster == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return nil
	}
	item, err := runtime.service.PickupItem(ctx, int(itemID), userID)
	if err != nil {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return nil
	}
	removePkt := furnipacket.FloorItemRemovePacket{ItemID: item.ID, UserID: userID}
	encoded, err := removePkt.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.FloorItemRemovePacketID, encoded)
	if runtime.entityEvictor != nil {
		runtime.entityEvictor(roomID, item.X, item.Y)
	}
	runtime.removeSeatEntry(roomID, item.ID)
	return runtime.sendPacket(connID, furnipacket.InventoryAddPacket{
		ItemID: item.ID, SpriteID: def.SpriteID, ExtraData: item.ExtraData,
		AllowRecycle: def.AllowRecycle, AllowTrade: def.AllowTrade,
		AllowInventoryStack:  def.AllowInventoryStack,
		AllowMarketplaceSell: def.AllowMarketplaceSell,
	})
}

// handleFloorUpdate processes a furniture move/rotate request (c2s 248).
// It validates ownership, updates placement, and broadcasts updated position.
func (runtime *Runtime) handleFloorUpdate(ctx context.Context, connID string, body []byte) error {
	userID, _ := runtime.userID(connID)
	r := codec.NewReader(body)
	itemID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	x, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	y, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	dir, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	if runtime.roomFinder == nil || runtime.roomBroadcaster == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return nil
	}
	item, err := runtime.service.PlaceFloorItem(ctx, int(itemID), userID, roomID, int(x), int(y), int(dir))
	if err != nil {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return nil
	}
	pkt := furnipacket.FloorItemUpdatePacket{
		ItemID: item.ID, SpriteID: def.SpriteID,
		X: item.X, Y: item.Y, Z: item.Z, Dir: item.Dir,
		ExtraData: item.ExtraData, UserID: item.UserID,
	}
	encoded, err := pkt.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.FloorItemUpdatePacketID, encoded)
	if def.CanSit {
		old, hadOld := runtime.seatEntryFor(roomID, item.ID)
		runtime.addSeatEntry(roomID, item.ID, item.X, item.Y, item.Dir, def.StackHeight, true)
		if hadOld && (old.x != item.X || old.y != item.Y) {
			if runtime.entityEvictor != nil {
				runtime.entityEvictor(roomID, old.x, old.y)
			}
		} else if runtime.entityRotator != nil {
			runtime.entityRotator(roomID, item.X, item.Y, item.Dir)
		}
	} else {
		runtime.removeSeatEntry(roomID, item.ID)
	}
	return nil
}

// handleToggleMultistate processes a furniture state toggle (c2s 99).
func (runtime *Runtime) handleToggleMultistate(_ context.Context, _ string, _ []byte) error {
	return nil
}
