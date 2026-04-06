package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
	"go.uber.org/zap"
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
	if !runtime.canManageItem(ctx, roomID, userID, int(itemID)) {
		return nil
	}
	original, err := runtime.service.FindItemByID(ctx, int(itemID))
	if err != nil {
		runtime.logger.Warn("find pickup item failed", zap.Int("item_id", int(itemID)), zap.Error(err))
		return nil
	}
	if original.RoomID == 0 {
		return nil
	}
	item, err := runtime.service.PickupItem(ctx, int(itemID), userID)
	if err != nil {
		runtime.logger.Warn("pickup floor item failed", zap.Int("item_id", int(itemID)), zap.Int("room_id", roomID), zap.Int("user_id", userID), zap.Error(err))
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		runtime.logger.Warn("find pickup item definition failed", zap.Int("definition_id", item.DefinitionID), zap.Error(err))
		return nil
	}
	removePkt := furnipacket.FloorItemRemovePacket{ItemID: item.ID, UserID: userID}
	encoded, err := removePkt.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.FloorItemRemovePacketID, encoded)
	oldEntries := runtime.seatEntriesFor(roomID, item.ID)
	if len(oldEntries) == 0 && (def.CanSit || def.CanLay) {
		oldEntries = seatEntriesFromFootprint(item.ID, original.X, original.Y, original.Dir, def.StackHeight, def.Width, def.Length, def.CanSit, def.CanLay)
	}
	if runtime.entityEvictor != nil {
		for _, tile := range uniqueSeatTiles(oldEntries) {
			runtime.entityEvictor(roomID, tile[0], tile[1])
		}
	}
	runtime.removeSeatEntries(roomID, item.ID)
	return runtime.sendUserPacket(ctx, connID, item.UserID, furnipacket.InventoryAddPacket{
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
	if !runtime.canManageItem(ctx, roomID, userID, int(itemID)) {
		return nil
	}
	original, err := runtime.service.FindItemByID(ctx, int(itemID))
	if err != nil {
		runtime.logger.Warn("find move item failed", zap.Int("item_id", int(itemID)), zap.Error(err))
		return nil
	}
	if (original.X != int(x) || original.Y != int(y)) && runtime.isTileOccupied(roomID, int(x), int(y)) {
		return nil
	}
	item, err := runtime.service.PlaceFloorItem(ctx, int(itemID), userID, roomID, int(x), int(y), int(dir))
	if err != nil {
		runtime.logger.Warn("move floor item failed", zap.Int("item_id", int(itemID)), zap.Int("room_id", roomID), zap.Int("user_id", userID), zap.Int("x", int(x)), zap.Int("y", int(y)), zap.Int("dir", int(dir)), zap.Error(err))
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		runtime.logger.Warn("find moved item definition failed", zap.Int("definition_id", item.DefinitionID), zap.Error(err))
		return nil
	}
	pkt := furnipacket.FloorItemUpdatePacket{
		ItemID: item.ID, SpriteID: def.SpriteID,
		X: item.X, Y: item.Y, Z: item.Z, Dir: item.Dir,
		StackHeight: def.StackHeight,
		ExtraData:   item.ExtraData, UserID: item.UserID,
	}
	encoded, err := pkt.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.FloorItemUpdatePacketID, encoded)
	if def.CanSit || def.CanLay {
		oldEntries := runtime.seatEntriesFor(roomID, item.ID)
		newEntries := seatEntriesFromFootprint(item.ID, item.X, item.Y, item.Dir, def.StackHeight, def.Width, def.Length, def.CanSit, def.CanLay)
		runtime.replaceSeatEntries(roomID, item.ID, newEntries)
		if len(oldEntries) > 0 && !sameSeatTiles(oldEntries, newEntries) {
			if runtime.entityEvictor != nil {
				for _, tile := range uniqueSeatTiles(oldEntries) {
					runtime.entityEvictor(roomID, tile[0], tile[1])
				}
			}
		} else if runtime.entityRotator != nil {
			for _, tile := range uniqueSeatTiles(newEntries) {
				runtime.entityRotator(roomID, tile[0], tile[1], item.Dir)
			}
		}
	} else {
		runtime.removeSeatEntries(roomID, item.ID)
	}
	return nil
}

// handleToggleMultistate processes a furniture state toggle (c2s 99).
func (runtime *Runtime) handleToggleMultistate(_ context.Context, _ string, _ []byte) error {
	return nil
}
